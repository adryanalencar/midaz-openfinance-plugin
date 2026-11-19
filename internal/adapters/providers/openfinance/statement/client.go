package statement

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance/crypto"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
)

// Client handles accounts and transaction endpoints specified by accounts-2.4.2.yml.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewStatementClient creates an accounts and statement client.
func NewStatementClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// ListAccounts queries GET /open-banking/accounts/v2/accounts.
func (c *Client) ListAccounts(ctx context.Context, token string) ([]domain.BankAccount, error) {
	endpoint := fmt.Sprintf("%s/open-banking/accounts/v2/accounts", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	crypto.InjectFAPIHeaders(req, "", "")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list accounts: status %d", resp.StatusCode)
	}

	var accountList AccountListResponse
	if err := json.NewDecoder(resp.Body).Decode(&accountList); err != nil {
		return nil, err
	}

	domainAccounts := make([]domain.BankAccount, 0, len(accountList.Data))
	for _, acc := range accountList.Data {
		domainAccounts = append(domainAccounts, domain.BankAccount{
			ID:         acc.AccountID,
			BrandName:  acc.BrandName,
			CompeCode:  acc.CompeCode,
			BranchCode: acc.BranchCode,
			Number:     acc.Number,
			CheckDigit: acc.CheckDigit,
			Type:       acc.Type,
			Currency:   "BRL",
		})
	}

	return domainAccounts, nil
}

// FetchTransactions queries GET /open-banking/accounts/v2/accounts/{accountId}/transactions.
func (c *Client) FetchTransactions(ctx context.Context, token, accountID string, fromDate, toDate time.Time, page, pageSize int, paginationKey string) (*domain.TransactionPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/open-banking/accounts/v2/accounts/%s/transactions", c.baseURL, accountID))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if !fromDate.IsZero() {
		q.Set("fromBookingDate", fromDate.Format("2006-01-02"))
	}
	if !toDate.IsZero() {
		q.Set("toBookingDate", toDate.Format("2006-01-02"))
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("pageSize", strconv.Itoa(pageSize))
	}
	if paginationKey != "" {
		q.Set("pagination-key", paginationKey)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	crypto.InjectFAPIHeaders(req, "", "")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch transactions: status %d", resp.StatusCode)
	}

	var txResponse AccountTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&txResponse); err != nil {
		return nil, err
	}

	domainTxs := make([]domain.BankTransaction, 0, len(txResponse.Data))
	for _, raw := range txResponse.Data {
		amount, parseErr := decimal.NewFromString(raw.TransactionAmount.Amount)
		if parseErr != nil {
			amount = decimal.Zero
		}

		dt, parseErr := time.Parse(time.RFC3339Nano, raw.TransactionDateTime)
		if parseErr != nil {
			dt, _ = time.Parse("2006-01-02T15:04:05Z", raw.TransactionDateTime)
		}

		domainTxs = append(domainTxs, domain.BankTransaction{
			ID:              raw.TransactionID,
			AccountID:       accountID,
			Amount:          amount,
			CreditDebitType: domain.CreditDebitType(raw.CreditDebitType),
			Description:     raw.TransactionName,
			PaymentType:     raw.CompletedAuthorisedPaymentType,
			Currency:        raw.TransactionAmount.Currency,
			TransactionDate: dt,
			BookingDateTime: dt,
			RawMetadata: map[string]any{
				"partie_cnpj_cpf":   raw.PartieCnpjCpf,
				"partie_compe_code": raw.PartieCompeCode,
				"partie_branch":     raw.PartieBranchCode,
				"partie_number":     raw.PartieNumber,
				"type":              raw.Type,
			},
		})
	}

	return &domain.TransactionPage{
		Transactions: domainTxs,
		CurrentPage:  page,
		TotalPages:   txResponse.Meta.TotalPages,
		TotalRecords: txResponse.Meta.TotalRecords,
	}, nil
}

// FetchCurrentTransactions queries the intraday transactions endpoint (/transactions-current).
func (c *Client) FetchCurrentTransactions(ctx context.Context, token, accountID string, page, pageSize int) (*domain.TransactionPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/open-banking/accounts/v2/accounts/%s/transactions-current", c.baseURL, accountID))
	if err != nil {
		return nil, err
	}

	q := u.Query()
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

	crypto.InjectFAPIHeaders(req, "", "")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch current transactions: status %d", resp.StatusCode)
	}

	var txResponse AccountTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&txResponse); err != nil {
		return nil, err
	}

	domainTxs := make([]domain.BankTransaction, 0, len(txResponse.Data))
	for _, raw := range txResponse.Data {
		amount, _ := decimal.NewFromString(raw.TransactionAmount.Amount)
		dt, _ := time.Parse(time.RFC3339Nano, raw.TransactionDateTime)

		domainTxs = append(domainTxs, domain.BankTransaction{
			ID:              raw.TransactionID,
			AccountID:       accountID,
			Amount:          amount,
			CreditDebitType: domain.CreditDebitType(raw.CreditDebitType),
			Description:     raw.TransactionName,
			PaymentType:     raw.CompletedAuthorisedPaymentType,
			Currency:        raw.TransactionAmount.Currency,
			TransactionDate: dt,
			BookingDateTime: dt,
		})
	}

	return &domain.TransactionPage{
		Transactions: domainTxs,
		CurrentPage:  page,
		TotalPages:   txResponse.Meta.TotalPages,
		TotalRecords: txResponse.Meta.TotalRecords,
	}, nil
}

// doWithRetry executes HTTP requests with automatic handling of 429 Too Many Requests
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfterHeader := resp.Header.Get("Retry-After")
			waitDuration := time.Duration(1<<attempt) * time.Second

			if seconds, parseErr := strconv.Atoi(retryAfterHeader); parseErr == nil && seconds > 0 {
				waitDuration = time.Duration(seconds) * time.Second
			}

			resp.Body.Close()
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(waitDuration):
				continue
			}
		}

		return resp, nil
	}

	return c.httpClient.Do(req)
}
