package statement

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestListAccountsSuccess tests retrieving and parsing accounts from the Open Finance endpoint
func TestListAccountsSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/open-banking/accounts/v2/accounts", r.URL.Path)

		resp := AccountListResponse{
			Data: []AccountData{
				{
					BrandName:   "Banco Itaú",
					CompanyCnpj: "60701190000104",
					Type:        "CONTA_DEPOSITO_A_VISTA",
					CompeCode:   "341",
					BranchCode:  "0001",
					Number:      "12345678",
					CheckDigit:  "9",
					AccountID:   "acc-itau-001",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewStatementClient(mockServer.Client(), mockServer.URL)
	accounts, err := client.ListAccounts(context.Background(), "mock-token")

	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "acc-itau-001", accounts[0].ID)
	assert.Equal(t, "341", accounts[0].CompeCode)
	assert.Equal(t, "12345678", accounts[0].Number)
}

// TestFetchTransactionsSuccess tests retrieving transactions and normalizing into domain.BankTransaction
func TestFetchTransactionsSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/open-banking/accounts/v2/accounts/acc-01/transactions", r.URL.Path)
		assert.Equal(t, "2026-11-01", r.URL.Query().Get("fromBookingDate"))
		assert.Equal(t, "2026-11-19", r.URL.Query().Get("toBookingDate"))

		var resp AccountTransactionsResponse
		resp.Data = []AccountTransactionData{
			{
				TransactionID:                  "TX-100",
				CompletedAuthorisedPaymentType: "PIX",
				CreditDebitType:                "CREDITO",
				TransactionName:                "PIX RECEBIDO EMPRESA ABC",
				Type:                           "PIX",
				TransactionDateTime:            "2026-11-19T10:30:00.000Z",
				PartieCnpjCpf:                  "12345678000199",
			},
		}
		resp.Data[0].TransactionAmount.Amount = "250.75"
		resp.Data[0].TransactionAmount.Currency = "BRL"
		resp.Meta.TotalPages = 1
		resp.Meta.TotalRecords = 1

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewStatementClient(mockServer.Client(), mockServer.URL)
	fromDate, _ := time.Parse("2006-01-02", "2026-11-01")
	toDate, _ := time.Parse("2006-01-02", "2026-11-19")

	page, err := client.FetchTransactions(context.Background(), "mock-token", "acc-01", fromDate, toDate, 1, 100, "")

	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Len(t, page.Transactions, 1)

	tx := page.Transactions[0]
	assert.Equal(t, "TX-100", tx.ID)
	assert.Equal(t, domain.CreditDebitCredit, tx.CreditDebitType)
	assert.Equal(t, "250.75", tx.Amount.StringFixed(2))
	assert.Equal(t, "PIX RECEBIDO EMPRESA ABC", tx.Description)
	assert.Equal(t, "12345678000199", tx.RawMetadata["partie_cnpj_cpf"])
}

// TestFetchCurrentTransactionsSuccess tests querying the intraday transactions endpoint
func TestFetchCurrentTransactionsSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/open-banking/accounts/v2/accounts/acc-01/transactions-current", r.URL.Path)

		var resp AccountTransactionsResponse
		resp.Data = []AccountTransactionData{
			{
				TransactionID:   "TX-INTRADAY-1",
				CreditDebitType: "DEBITO",
				TransactionName: "PAGTO PIX LOJA",
			},
		}
		resp.Data[0].TransactionAmount.Amount = "45.00"
		resp.Data[0].TransactionAmount.Currency = "BRL"
		resp.Data[0].TransactionDateTime = "2026-11-19T14:00:00.000Z"

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewStatementClient(mockServer.Client(), mockServer.URL)
	page, err := client.FetchCurrentTransactions(context.Background(), "mock-token", "acc-01", 1, 50)

	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Len(t, page.Transactions, 1)
	assert.Equal(t, "TX-INTRADAY-1", page.Transactions[0].ID)
	assert.Equal(t, domain.CreditDebitDebit, page.Transactions[0].CreditDebitType)
}

// TestDoWithRetryOnHTTP429 tests automatic backoff when encountering rate-limiting
func TestDoWithRetryOnHTTP429(t *testing.T) {
	attempts := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": []}`))
	}))
	defer mockServer.Close()

	client := NewStatementClient(mockServer.Client(), mockServer.URL)
	accounts, err := client.ListAccounts(context.Background(), "mock-token")

	assert.NoError(t, err)
	assert.NotNil(t, accounts)
	assert.Equal(t, 2, attempts, "Should have retried after receiving 429")
}
