package midaz

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestEnsureClearingAccountAlreadyExists tests when @external account is already present in Midaz
func TestEnsureClearingAccountAlreadyExists(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Query().Get("alias"), "@external/BRL:openfinance")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items": [{"id": "acc-123"}]}`))
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		MidazOnboardingURL:        mockServer.URL,
		MidazOrganizationID:       "org-1",
		MidazLedgerID:             "led-1",
		MidazClearingAccountAlias: "@external/BRL:openfinance",
	}

	client := NewMidazClient(cfg)
	err := client.EnsureClearingAccount(context.Background())

	assert.NoError(t, err)
}

// TestEnsureClearingAccountCreateSuccess tests provisioning @external account when not found
func TestEnsureClearingAccountCreateSuccess(t *testing.T) {
	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method == http.MethodGet {
			// Not found
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost {
			var body map[string]any
			err := json.NewDecoder(r.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, "external", body["type"])
			assert.Equal(t, "@external/BRL:openfinance", body["alias"])
			assert.Equal(t, "BRL", body["assetCode"])

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "new-acc-uuid"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		MidazOnboardingURL:        mockServer.URL,
		MidazOrganizationID:       "org-1",
		MidazLedgerID:             "led-1",
		MidazClearingAccountAlias: "@external/BRL:openfinance",
	}

	client := NewMidazClient(cfg)
	err := client.EnsureClearingAccount(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

// TestPostDoubleEntryTransactionInflow tests that bank CREDITO debits @external and credits target account
func TestPostDoubleEntryTransactionInflow(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)

		assert.Equal(t, "OFP-TX-INFLOW-01", payload["code"])

		send := payload["send"].(map[string]any)
		assert.Equal(t, float64(50000), send["value"]) // R$ 500.00 -> 50000
		assert.Equal(t, float64(2), send["scale"])

		// Source must be @external
		source := send["source"].(map[string]any)
		fromList := source["from"].([]any)
		from := fromList[0].(map[string]any)
		assert.Equal(t, "@external/BRL:openfinance", from["accountAlias"])

		// Distribute must be target account
		distribute := send["distribute"].(map[string]any)
		toList := distribute["to"].([]any)
		to := toList[0].(map[string]any)
		assert.Equal(t, "itau_checking_acc", to["accountAlias"])

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		MidazTransactionURL:       mockServer.URL,
		MidazOrganizationID:       "org-1",
		MidazLedgerID:             "led-1",
		MidazClearingAccountAlias: "@external/BRL:openfinance",
		Provider:                  "openfinance",
	}

	amount, _ := decimal.NewFromString("500.00")
	tx := domain.BankTransaction{
		ID:              "TX-INFLOW-01",
		Amount:          amount,
		CreditDebitType: domain.CreditDebitCredit,
		Description:     "PIX RECEBIDO CLIENTE",
		PaymentType:     "PIX",
		Currency:        "BRL",
		BookingDateTime: time.Now(),
	}

	client := NewMidazClient(cfg)
	err := client.PostDoubleEntryTransaction(context.Background(), tx, "itau_checking_acc")

	assert.NoError(t, err)
}

// TestPostDoubleEntryTransactionOutflow tests that bank DEBITO debits target account and credits @external
func TestPostDoubleEntryTransactionOutflow(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)

		send := payload["send"].(map[string]any)
		assert.Equal(t, float64(12050), send["value"]) // R$ 120.50 -> 12050

		// Source must be target account
		source := send["source"].(map[string]any)
		fromList := source["from"].([]any)
		from := fromList[0].(map[string]any)
		assert.Equal(t, "itau_checking_acc", from["accountAlias"])

		// Distribute must be @external
		distribute := send["distribute"].(map[string]any)
		toList := distribute["to"].([]any)
		to := toList[0].(map[string]any)
		assert.Equal(t, "@external/BRL:openfinance", to["accountAlias"])

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		MidazTransactionURL:       mockServer.URL,
		MidazOrganizationID:       "org-1",
		MidazLedgerID:             "led-1",
		MidazClearingAccountAlias: "@external/BRL:openfinance",
		Provider:                  "openfinance",
	}

	amount, _ := decimal.NewFromString("120.50")
	tx := domain.BankTransaction{
		ID:              "TX-OUTFLOW-01",
		Amount:          amount,
		CreditDebitType: domain.CreditDebitDebit,
		Description:     "PIX ENVIADO FORNECEDOR",
		PaymentType:     "PIX",
		Currency:        "BRL",
		BookingDateTime: time.Now(),
	}

	client := NewMidazClient(cfg)
	err := client.PostDoubleEntryTransaction(context.Background(), tx, "itau_checking_acc")

	assert.NoError(t, err)
}

// TestPostDoubleEntryTransactionIdempotentConflict tests that HTTP 409 is treated gracefully as already synced
func TestPostDoubleEntryTransactionIdempotentConflict(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict) // Midaz duplicate code conflict
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		MidazTransactionURL:       mockServer.URL,
		MidazClearingAccountAlias: "@external/BRL:openfinance",
	}

	amount, _ := decimal.NewFromString("50.00")
	tx := domain.BankTransaction{
		ID:              "TX-DUPLICATE",
		Amount:          amount,
		CreditDebitType: domain.CreditDebitCredit,
	}

	client := NewMidazClient(cfg)
	err := client.PostDoubleEntryTransaction(context.Background(), tx, "acc_target")

	// Must NOT return error, because 409 means already ingested in Midaz
	assert.NoError(t, err)
}

func TestPostDoubleEntryTransactionCustomClearing(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)

		send := payload["send"].(map[string]any)
		source := send["source"].(map[string]any)
		fromList := source["from"].([]any)
		from := fromList[0].(map[string]any)
		assert.Equal(t, "@external/BRL:custom_clearing", from["accountAlias"])

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		MidazTransactionURL:       mockServer.URL,
		MidazOrganizationID:       "org-1",
		MidazLedgerID:             "led-1",
		MidazClearingAccountAlias: "@external/BRL:openfinance",
	}

	amount, _ := decimal.NewFromString("100.00")
	tx := domain.BankTransaction{
		ID:              "TX-CUSTOM-01",
		Amount:          amount,
		CreditDebitType: domain.CreditDebitCredit,
		BookingDateTime: time.Now(),
	}

	client := NewMidazClient(cfg)
	err := client.PostDoubleEntryTransaction(context.Background(), tx, "checking_acc", "@external/BRL:custom_clearing")
	assert.NoError(t, err)
}

func TestListExternalAccounts(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		resp := map[string]any{
			"items": []map[string]any{
				{"id": "acc-1", "alias": "@external/BRL:universal", "name": "Universal Clearing", "assetCode": "BRL", "type": "external"},
				{"id": "acc-2", "alias": "@checking/001", "name": "Checking Account", "assetCode": "BRL", "type": "deposit"},
				{"id": "acc-3", "alias": "@external/BRL:openfinance", "name": "Open Finance Clearing", "assetCode": "BRL", "type": "external"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		MidazOnboardingURL:  mockServer.URL,
		MidazOrganizationID: "org-1",
		MidazLedgerID:       "led-1",
	}

	client := NewMidazClient(cfg)
	accounts, err := client.ListExternalAccounts(context.Background())
	assert.NoError(t, err)
	assert.Len(t, accounts, 2)
	assert.Equal(t, "@external/BRL:universal", accounts[0].Alias)
	assert.Equal(t, "@external/BRL:openfinance", accounts[1].Alias)
}

