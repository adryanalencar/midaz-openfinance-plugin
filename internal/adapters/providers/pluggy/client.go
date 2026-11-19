package pluggy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
)

// PluggyClient implements domain.OpenFinanceProvider for the Pluggy aggregation API.
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
			ID         string `json:"id"`
			Name       string `json:"name"`
			Number     string `json:"number"`
			Agency     string `json:"agency"`
			Type       string `json:"type"`
			Currency   string `json:"currencyCode"`
			BankData   struct {
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

func (c *PluggyClient) FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*domain.TransactionPage, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	u, err := url.Parse(fmt.Sprintf("%s/transactions", c.cfg.PluggyBaseURL))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("accountId", accountID)
	if !fromDate.IsZero() {
		q.Set("from", fromDate.Format("2006-01-02"))
	}
	if !toDate.IsZero() {
		q.Set("to", toDate.Format("2006-01-02"))
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("pageSize", strconv.Itoa(pageSize))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
		return nil, fmt.Errorf("pluggy transactions fetch failed: status %d", resp.StatusCode)
	}

	var pluggyTx struct {
		Total        int `json:"total"`
		TotalPages   int `json:"totalPages"`
		Page         int `json:"page"`
		Results      []struct {
			ID           string  `json:"id"`
			Description  string  `json:"description"`
			Amount       float64 `json:"amount"`
			Type         string  `json:"type"` // "DEBIT" or "CREDIT"
			Date         string  `json:"date"`
			CurrencyCode string  `json:"currencyCode"`
			PaymentData  struct {
				PaymentMethod string `json:"paymentMethod"`
			} `json:"paymentData"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pluggyTx); err != nil {
		return nil, err
	}

	txs := make([]domain.BankTransaction, 0, len(pluggyTx.Results))
	for _, raw := range pluggyTx.Results {
		// Pluggy returns positive/negative numbers or type DEBIT/CREDIT
		amountDec := decimal.NewFromFloat(raw.Amount).Abs()

		creditDebit := domain.CreditDebitCredit
		if raw.Type == "DEBIT" || raw.Amount < 0 {
			creditDebit = domain.CreditDebitDebit
		}

		dt, parseErr := time.Parse(time.RFC3339, raw.Date)
		if parseErr != nil {
			dt, _ = time.Parse("2006-01-02", raw.Date)
		}

		txs = append(txs, domain.BankTransaction{
			ID:              raw.ID,
			AccountID:       accountID,
			Amount:          amountDec,
			CreditDebitType: creditDebit,
			Description:     raw.Description,
			PaymentType:     raw.PaymentData.PaymentMethod,
			Currency:        raw.CurrencyCode,
			TransactionDate: dt,
			BookingDateTime: dt,
		})
	}

	return &domain.TransactionPage{
		Transactions: txs,
		CurrentPage:  pluggyTx.Page,
		TotalPages:   pluggyTx.TotalPages,
		TotalRecords: pluggyTx.Total,
	}, nil
}

func (c *PluggyClient) HealthCheck(ctx context.Context) error {
	return nil
}
