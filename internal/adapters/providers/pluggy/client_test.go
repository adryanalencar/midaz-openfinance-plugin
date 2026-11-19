package pluggy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestPluggyClientFullFlow tests authentication, account listing, and transaction mapping
func TestPluggyClientFullFlow(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey": "pluggy-test-api-key"}`))
		case "/accounts":
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "pluggy-test-api-key", r.Header.Get("X-API-KEY"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"id": "plg-acc-1",
						"name": "Conta Corrente",
						"number": "998877",
						"agency": "1234",
						"type": "CHECKING",
						"currencyCode": "BRL",
						"connector": {"name": "Nubank"}
					}
				]
			}`))
		case "/transactions":
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "pluggy-test-api-key", r.Header.Get("X-API-KEY"))
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total":      1,
				"totalPages": 1,
				"page":       1,
				"results": []map[string]any{
					{
						"id":           "plg-tx-1",
						"description":  "Depósito Recebido",
						"amount":       250.00,
						"type":         "CREDIT",
						"date":         "2026-11-19T10:00:00Z",
						"currencyCode": "BRL",
						"paymentData": map[string]string{
							"paymentMethod": "PIX",
						},
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		PluggyBaseURL:      mockServer.URL,
		PluggyClientID:     "client-id",
		PluggyClientSecret: "client-secret",
	}

	client, err := NewPluggyClient(cfg)
	assert.NoError(t, err)

	// 1. List Accounts
	accounts, err := client.ListAccounts(context.Background(), "conn-01")
	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "plg-acc-1", accounts[0].ID)
	assert.Equal(t, "Nubank", accounts[0].BrandName)

	// 2. Fetch Transactions
	page, err := client.FetchTransactions(context.Background(), "plg-acc-1", time.Time{}, time.Time{}, 1, 100)
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Len(t, page.Transactions, 1)

	tx := page.Transactions[0]
	assert.Equal(t, "plg-tx-1", tx.ID)
	assert.Equal(t, domain.CreditDebitCredit, tx.CreditDebitType)
	assert.Equal(t, "250.00", tx.Amount.StringFixed(2))
	assert.Equal(t, "PIX", tx.PaymentType)
}
