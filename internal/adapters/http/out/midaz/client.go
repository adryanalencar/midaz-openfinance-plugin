package midaz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
)

// Client handles communication with the Midaz Onboarding and Transaction APIs.
type Client struct {
	cfg        *bootstrap.Config
	httpClient *http.Client
}

// NewMidazClient initializes the Midaz API client.
func NewMidazClient(cfg *bootstrap.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// EnsureClearingAccount verifies that the @external account exists in the target ledger, creating it if necessary.
func (c *Client) EnsureClearingAccount(ctx context.Context) error {
	// 1. Check if account already exists
	checkURL := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/accounts?alias=%s",
		c.cfg.MidazOnboardingURL, c.cfg.MidazOrganizationID, c.cfg.MidazLedgerID, c.cfg.MidazClearingAccountAlias)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil // Account already present
	}

	// 2. If not found, create external account
	createURL := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/accounts",
		c.cfg.MidazOnboardingURL, c.cfg.MidazOrganizationID, c.cfg.MidazLedgerID)

	payload := map[string]any{
		"name":      "Open Finance External Clearing",
		"assetCode": "BRL",
		"type":      "external",
		"alias":     c.cfg.MidazClearingAccountAlias,
		"metadata": map[string]string{
			"source": "midaz-plugin-openfinance",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reqCreate, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	c.setHeaders(reqCreate)

	respCreate, err := c.httpClient.Do(reqCreate)
	if err != nil {
		return err
	}
	defer respCreate.Body.Close()

	if respCreate.StatusCode != http.StatusCreated && respCreate.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create clearing account in Midaz: status %d", respCreate.StatusCode)
	}

	return nil
}

// ListExternalAccounts retrieves all external clearing accounts in the ledger.
func (c *Client) ListExternalAccounts(ctx context.Context) ([]domain.ExternalAccount, error) {
	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/accounts?limit=100",
		c.cfg.MidazOnboardingURL, c.cfg.MidazOrganizationID, c.cfg.MidazLedgerID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for external accounts: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts from Midaz: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("midaz returned status %d fetching accounts", resp.StatusCode)
	}

	var response struct {
		Items []struct {
			ID        string `json:"id"`
			Alias     string `json:"alias"`
			Name      string `json:"name"`
			AssetCode string `json:"assetCode"`
			Type      string `json:"type"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Midaz accounts response: %w", err)
	}

	result := make([]domain.ExternalAccount, 0)
	for _, item := range response.Items {
		if strings.EqualFold(item.Type, "external") {
			result = append(result, domain.ExternalAccount{
				ID:        item.ID,
				Alias:     item.Alias,
				Name:      item.Name,
				AssetCode: item.AssetCode,
				Type:      item.Type,
			})
		}
	}

	return result, nil
}

// PostDoubleEntryTransaction posts an inflow or outflow transaction to Midaz balancing against the clearing account.
func (c *Client) PostDoubleEntryTransaction(ctx context.Context, tx domain.BankTransaction, targetAccountAlias string, clearingAccountAlias ...string) error {
	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/transactions/json",
		c.cfg.MidazTransactionURL, c.cfg.MidazOrganizationID, c.cfg.MidazLedgerID)

	scale := int32(2)
	integerValue := tx.Amount.Mul(decimal.NewFromInt(100)).IntPart()

	clearingAlias := c.cfg.MidazClearingAccountAlias
	if len(clearingAccountAlias) > 0 && clearingAccountAlias[0] != "" {
		clearingAlias = clearingAccountAlias[0]
	}

	var sourceAlias, destAlias string

	if tx.CreditDebitType == domain.CreditDebitCredit {
		// INFLOW: From @external clearing account TO target bank account
		sourceAlias = clearingAlias
		destAlias = targetAccountAlias
	} else {
		// OUTFLOW: From target bank account TO @external clearing account
		sourceAlias = targetAccountAlias
		destAlias = clearingAlias
	}

	payload := map[string]any{
		"chartOfAccountsGroupName": "OPEN_FINANCE_SYNC",
		"description":              tx.Description,
		"code":                     fmt.Sprintf("OFP-%s", tx.ID),
		"metadata": map[string]any{
			"provider":            c.cfg.Provider,
			"open_finance_tx_id":  tx.ID,
			"payment_type":        tx.PaymentType,
			"credit_debit_type":   string(tx.CreditDebitType),
			"original_date":       tx.BookingDateTime.Format(time.RFC3339),
		},
		"send": map[string]any{
			"asset": "BRL",
			"value": integerValue,
			"scale": scale,
			"source": map[string]any{
				"from": []map[string]any{
					{
						"accountAlias": sourceAlias,
						"amount": map[string]any{
							"asset": "BRL",
							"value": tx.Amount.StringFixed(2),
						},
					},
				},
			},
			"distribute": map[string]any{
				"to": []map[string]any{
					{
						"accountAlias": destAlias,
						"amount": map[string]any{
							"asset": "BRL",
							"value": tx.Amount.StringFixed(2),
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		// Already ingested in Midaz (idempotent)
		return nil
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("midaz transaction rejected: HTTP %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.MidazAuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.MidazAuthToken)
	}
}
