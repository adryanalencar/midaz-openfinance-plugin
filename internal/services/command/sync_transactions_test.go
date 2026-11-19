package command

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// mockProvider implements domain.OpenFinanceProvider for testing
type mockProvider struct {
	txs []domain.BankTransaction
}

func (m *mockProvider) GetName() string { return "mock" }
func (m *mockProvider) ListAccounts(ctx context.Context, connectionID string) ([]domain.BankAccount, error) {
	return []domain.BankAccount{}, nil
}
func (m *mockProvider) FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*domain.TransactionPage, error) {
	if page > 1 {
		return &domain.TransactionPage{Transactions: []domain.BankTransaction{}, CurrentPage: page, TotalPages: 1}, nil
	}
	return &domain.TransactionPage{
		Transactions: m.txs,
		CurrentPage:  1,
		TotalPages:   1,
		TotalRecords: len(m.txs),
	}, nil
}
func (m *mockProvider) HealthCheck(ctx context.Context) error { return nil }

// TestSyncTransactionsExecuteSuccess tests the end-to-end sync execution from provider to Midaz
func TestSyncTransactionsExecuteSuccess(t *testing.T) {
	postedCodes := make([]string, 0)
	midazServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
	}))
	defer midazServer.Close()

	cfg := &bootstrap.Config{
		MidazTransactionURL:       midazServer.URL,
		MidazOrganizationID:       "org-uuid",
		MidazLedgerID:             "ledger-uuid",
		MidazClearingAccountAlias: "@external/BRL:openfinance",
	}

	midazClient := midaz.NewMidazClient(cfg)

	tx1Amount, _ := decimal.NewFromString("100.00")
	tx2Amount, _ := decimal.NewFromString("50.00")

	provider := &mockProvider{
		txs: []domain.BankTransaction{
			{
				ID:              "TX-001",
				Amount:          tx1Amount,
				CreditDebitType: domain.CreditDebitCredit,
				Description:     "Pix Inflow",
			},
			{
				ID:              "TX-002",
				Amount:          tx2Amount,
				CreditDebitType: domain.CreditDebitDebit,
				Description:     "Pix Outflow",
			},
		},
	}

	useCase := NewSyncTransactionsUseCase(provider, midazClient)

	input := SyncTransactionsInput{
		BankAccountID:           "bank-acc-1",
		TargetMidazAccountAlias: "itau_checking_acc",
		PageSize:                10,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.RecordsFetched)
	assert.Equal(t, 2, result.RecordsPosted)
	assert.Equal(t, 0, result.RecordsFailed)
	assert.Empty(t, postedCodes)
}
