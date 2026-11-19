package pluggy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
)

// ItemStatus represents the state of a connected bank item in Pluggy.
type ItemStatus string

const (
	ItemStatusUpdated          ItemStatus = "UPDATED"
	ItemStatusUpdating         ItemStatus = "UPDATING"
	ItemStatusWaitingUserInput ItemStatus = "WAITING_USER_INPUT"
	ItemStatusLoginError       ItemStatus = "LOGIN_ERROR"
	ItemStatusOutdated         ItemStatus = "OUTDATED"
)

// ConnectTokenOptions defines configuration options for launching the Connect Widget.
type ConnectTokenOptions struct {
	WebhookURL       string `json:"webhookUrl,omitempty"`
	ClientUserID     string `json:"clientUserId,omitempty"`
	OAuthRedirectURI string `json:"oauthRedirectUri,omitempty"`
	AvoidDuplicates  bool   `json:"avoidDuplicates,omitempty"`
}

// ConnectTokenRequest payload sent to POST /connect_token.
type ConnectTokenRequest struct {
	ItemID  string               `json:"itemId,omitempty"`
	Options *ConnectTokenOptions `json:"options,omitempty"`
}

// ConnectTokenResponse returned from POST /connect_token.
type ConnectTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

// PluggyItem represents an account aggregation item in Pluggy.
type PluggyItem struct {
	ID              string     `json:"id"`
	ConnectorID     int        `json:"connectorId"`
	Status          ItemStatus `json:"status"`
	ExecutionStatus string     `json:"executionStatus"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	LastUpdatedAt   *time.Time `json:"lastUpdatedAt,omitempty"`
	Error           *ItemError `json:"error,omitempty"`
	ClientUserID    string     `json:"clientUserId,omitempty"`
}

// ItemError contains error details when an item synchronization fails.
type ItemError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PluggyClient implements domain.OpenFinanceProvider for the Pluggy aggregation API v2.
type PluggyClient struct {
	cfg        *bootstrap.Config
	httpClient *http.Client
	apiKey     string
	apiKeyExp  time.Time
}

// NewPluggyClient initializes the HTTP client for Pluggy.
func NewPluggyClient(cfg *bootstrap.Config) (*PluggyClient, error) {
	return &PluggyClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *PluggyClient) GetName() string {
	return "pluggy"
}

// Authenticate exchanges clientId and clientSecret for an apiKey.
func (c *PluggyClient) Authenticate(ctx context.Context) error {
	if c.apiKey != "" && time.Now().Before(c.apiKeyExp) {
		return nil
	}

	authURL := fmt.Sprintf("%s/auth", c.cfg.PluggyBaseURL)
	payload := map[string]string{
		"clientId":     c.cfg.PluggyClientID,
		"clientSecret": c.cfg.PluggyClientSecret,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pluggy auth failed: status %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var authResp struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	c.apiKey = authResp.APIKey
	c.apiKeyExp = time.Now().Add(2 * time.Hour) // Pluggy tokens typically valid for 2 hours

	return nil
}

// CreateConnectToken generates an accessToken for the client-side Pluggy Connect Widget.
func (c *PluggyClient) CreateConnectToken(ctx context.Context, reqPayload *ConnectTokenRequest) (*ConnectTokenResponse, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	urlEndpoint := fmt.Sprintf("%s/connect_token", c.cfg.PluggyBaseURL)
	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create connect token failed: status %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var tokenResp ConnectTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// GetItem retrieves item details and current execution status from Pluggy.
func (c *PluggyClient) GetItem(ctx context.Context, itemID string) (*PluggyItem, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/items/%s", c.cfg.PluggyBaseURL, url.PathEscape(itemID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pluggy get item failed: status %d", resp.StatusCode)
	}

	var item PluggyItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	return &item, nil
}

// DeleteItem revokes and deletes an Item connection in Pluggy.
func (c *PluggyClient) DeleteItem(ctx context.Context, itemID string) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/items/%s", c.cfg.PluggyBaseURL, url.PathEscape(itemID))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("pluggy delete item failed: status %d", resp.StatusCode)
	}

	return nil
}

// ListAccounts retrieves bank accounts linked to a connection (Item ID).
func (c *PluggyClient) ListAccounts(ctx context.Context, connectionID string) ([]domain.BankAccount, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/accounts?itemId=%s", c.cfg.PluggyBaseURL, url.QueryEscape(connectionID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pluggy accounts fetch failed: status %d", resp.StatusCode)
	}

	var pluggyAccounts struct {
		Results []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Number   string `json:"number"`
			Agency   string `json:"agency"`
			Type     string `json:"type"`
			Currency string `json:"currencyCode"`
			BankData struct {
				Name string `json:"name"`
			} `json:"connector"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pluggyAccounts); err != nil {
		return nil, err
	}

	accounts := make([]domain.BankAccount, 0, len(pluggyAccounts.Results))
	for _, acc := range pluggyAccounts.Results {
		accounts = append(accounts, domain.BankAccount{
			ID:         acc.ID,
			BrandName:  acc.BankData.Name,
			BranchCode: acc.Agency,
			Number:     acc.Number,
			Type:       acc.Type,
			Currency:   acc.Currency,
		})
	}

	return accounts, nil
}

// RawPluggyTransaction is the wire representation of a transaction in Pluggy API v2.
type RawPluggyTransaction struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"` // "DEBIT" or "CREDIT"
	Date        string  `json:"date"`
	Currency    string  `json:"currencyCode"`
	PaymentData struct {
		PaymentMethod string `json:"paymentMethod"`
	} `json:"paymentData"`
}

// CursorPageResponse is the wire structure of GET /v2/transactions.
type CursorPageResponse struct {
	Results []RawPluggyTransaction `json:"results"`
	Next    *string                `json:"next"`
}

// FetchTransactions retrieves transactions using official Pluggy API v2 with date filters.
func (c *PluggyClient) FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*domain.TransactionPage, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	u, err := url.Parse(fmt.Sprintf("%s/v2/transactions", c.cfg.PluggyBaseURL))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("accountId", accountID)
	if !fromDate.IsZero() {
		q.Set("dateFrom", fromDate.Format("2006-01-02"))
	}
	if !toDate.IsZero() {
		q.Set("dateTo", toDate.Format("2006-01-02"))
	}
	u.RawQuery = q.Encode()

	return c.fetchTransactionsURL(ctx, accountID, u.String())
}

// FetchTransactionsFromLink fetches transactions from a relative or absolute URL (such as createdTransactionsLink).
func (c *PluggyClient) FetchTransactionsFromLink(ctx context.Context, link string) ([]domain.BankTransaction, *string, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, nil, err
	}

	var fetchURL string
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		fetchURL = link
	} else if strings.HasPrefix(link, "/") {
		fetchURL = fmt.Sprintf("%s%s", c.cfg.PluggyBaseURL, link)
	} else {
		fetchURL = fmt.Sprintf("%s/v2/transactions%s", c.cfg.PluggyBaseURL, link)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("fetch transactions failed: status %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var pageResp CursorPageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, nil, err
	}

	txs := make([]domain.BankTransaction, 0, len(pageResp.Results))
	for _, raw := range pageResp.Results {
		txs = append(txs, mapRawTransaction(raw, ""))
	}

	return txs, pageResp.Next, nil
}

// FetchAllTransactionsFromLink exhaustively pages a cursor link until all records are collected.
func (c *PluggyClient) FetchAllTransactionsFromLink(ctx context.Context, link string) ([]domain.BankTransaction, error) {
	var allTxs []domain.BankTransaction
	currentLink := link

	for currentLink != "" {
		txs, next, err := c.FetchTransactionsFromLink(ctx, currentLink)
		if err != nil {
			return nil, err
		}
		allTxs = append(allTxs, txs...)

		if next == nil || *next == "" {
			break
		}
		currentLink = *next
	}

	return allTxs, nil
}

func (c *PluggyClient) fetchTransactionsURL(ctx context.Context, accountID, reqURL string) (*domain.TransactionPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pluggy v2 transactions fetch failed: status %d", resp.StatusCode)
	}

	var pageResp CursorPageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, err
	}

	txs := make([]domain.BankTransaction, 0, len(pageResp.Results))
	for _, raw := range pageResp.Results {
		txs = append(txs, mapRawTransaction(raw, accountID))
	}

	return &domain.TransactionPage{
		Transactions: txs,
		CurrentPage:  1,
		TotalPages:   1,
		TotalRecords: len(txs),
	}, nil
}

func mapRawTransaction(raw RawPluggyTransaction, defaultAccountID string) domain.BankTransaction {
	amountDec := decimal.NewFromFloat(raw.Amount).Abs()

	creditDebit := domain.CreditDebitCredit
	if raw.Type == "DEBIT" || raw.Amount < 0 {
		creditDebit = domain.CreditDebitDebit
	}

	dt, parseErr := time.Parse(time.RFC3339, raw.Date)
	if parseErr != nil {
		dt, _ = time.Parse("2006-01-02", raw.Date)
	}

	return domain.BankTransaction{
		ID:              raw.ID,
		AccountID:       defaultAccountID,
		Amount:          amountDec,
		CreditDebitType: creditDebit,
		Description:     raw.Description,
		PaymentType:     raw.PaymentData.PaymentMethod,
		Currency:        raw.Currency,
		TransactionDate: dt,
		BookingDateTime: dt,
	}
}

// HealthCheck validates connectivity and active credentials with the Pluggy API.
func (c *PluggyClient) HealthCheck(ctx context.Context) error {
	return c.Authenticate(ctx)
}
