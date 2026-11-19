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
	"github.com/stretchr/testify/require"
)

func TestPluggyClient_FullLifecycle(t *testing.T) {
	var authCalls, connectTokenCalls, getItemCalls, deleteItemCalls, accountsCalls, v2TxCalls int

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			authCalls++
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey": "pluggy-test-key-123"}`))

		case "/connect_token":
			connectTokenCalls++
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "pluggy-test-key-123", r.Header.Get("X-API-KEY"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"accessToken": "widget-access-token-jwt"}`))

		case "/items/item-uuid-01":
			if r.Method == http.MethodGet {
				getItemCalls++
				assert.Equal(t, "pluggy-test-key-123", r.Header.Get("X-API-KEY"))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"id": "item-uuid-01",
					"connectorId": 201,
					"status": "UPDATED",
					"executionStatus": "SUCCESS",
					"createdAt": "2026-11-19T10:00:00Z",
					"updatedAt": "2026-11-19T11:00:00Z"
				}`))
			} else if r.Method == http.MethodDelete {
				deleteItemCalls++
				assert.Equal(t, "pluggy-test-key-123", r.Header.Get("X-API-KEY"))
				w.WriteHeader(http.StatusNoContent)
			}

		case "/accounts":
			accountsCalls++
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "pluggy-test-key-123", r.Header.Get("X-API-KEY"))
			assert.Equal(t, "item-uuid-01", r.URL.Query().Get("itemId"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"id": "acc-nubank-01",
						"name": "Conta Digital",
						"number": "1234567-8",
						"agency": "0001",
						"type": "CHECKING",
						"currencyCode": "BRL",
						"connector": {"name": "Nubank"}
					}
				]
			}`))

		case "/v2/transactions":
			v2TxCalls++
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "pluggy-test-key-123", r.Header.Get("X-API-KEY"))

			after := r.URL.Query().Get("after")
			if after == "" {
				// Page 1
				nextCursor := "?after=cursor-page-2"
				resp := CursorPageResponse{
					Results: []RawPluggyTransaction{
						{
							ID:          "tx-page1-01",
							Description: "PIX Recebido 1",
							Amount:      100.50,
							Type:        "CREDIT",
							Date:        "2026-11-19T10:00:00Z",
							Currency:    "BRL",
							PaymentData: struct {
								PaymentMethod string `json:"paymentMethod"`
							}{PaymentMethod: "PIX"},
						},
					},
					Next: &nextCursor,
				}
				_ = json.NewEncoder(w).Encode(resp)
			} else if after == "cursor-page-2" {
				// Page 2 (last page)
				resp := CursorPageResponse{
					Results: []RawPluggyTransaction{
						{
							ID:          "tx-page2-02",
							Description: "Pagamento Boleto",
							Amount:      -50.00,
							Type:        "DEBIT",
							Date:        "2026-11-19T10:30:00Z",
							Currency:    "BRL",
						},
					},
					Next: nil,
				}
				_ = json.NewEncoder(w).Encode(resp)
			}

		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		PluggyBaseURL:      mockServer.URL,
		PluggyClientID:     "my-client-id",
		PluggyClientSecret: "my-client-secret",
	}

	client, err := NewPluggyClient(cfg)
	require.NoError(t, err)

	// 1. HealthCheck
	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, authCalls)

	// 2. Create Connect Token
	tokenResp, err := client.CreateConnectToken(context.Background(), &ConnectTokenRequest{
		Options: &ConnectTokenOptions{
			WebhookURL:   "https://my-app.com/webhook",
			ClientUserID: "user-123",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "widget-access-token-jwt", tokenResp.AccessToken)
	assert.Equal(t, 1, connectTokenCalls)

	// 3. Get Item
	item, err := client.GetItem(context.Background(), "item-uuid-01")
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "item-uuid-01", item.ID)
	assert.Equal(t, ItemStatusUpdated, item.Status)
	assert.Equal(t, 1, getItemCalls)

	// 4. List Accounts
	accounts, err := client.ListAccounts(context.Background(), "item-uuid-01")
	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "acc-nubank-01", accounts[0].ID)
	assert.Equal(t, "Nubank", accounts[0].BrandName)

	// 5. FetchTransactions (v2)
	page, err := client.FetchTransactions(context.Background(), "acc-nubank-01", time.Time{}, time.Time{}, 1, 100)
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Len(t, page.Transactions, 1)
	assert.Equal(t, "tx-page1-01", page.Transactions[0].ID)
	assert.Equal(t, domain.CreditDebitCredit, page.Transactions[0].CreditDebitType)
	assert.Equal(t, "100.50", page.Transactions[0].Amount.StringFixed(2))

	// 6. FetchAllTransactionsFromLink (Paginating until next is nil)
	allTxs, err := client.FetchAllTransactionsFromLink(context.Background(), "/v2/transactions")
	assert.NoError(t, err)
	assert.Len(t, allTxs, 2)
	assert.Equal(t, "tx-page1-01", allTxs[0].ID)
	assert.Equal(t, "tx-page2-02", allTxs[1].ID)
	assert.Equal(t, domain.CreditDebitDebit, allTxs[1].CreditDebitType)
	assert.Equal(t, "50.00", allTxs[1].Amount.StringFixed(2))

	// 7. Delete Item
	err = client.DeleteItem(context.Background(), "item-uuid-01")
	assert.NoError(t, err)
	assert.Equal(t, 1, deleteItemCalls)
}
