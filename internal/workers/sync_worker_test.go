package workers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// mockStorageRepo implements domain.StorageRepository in-memory for testing
type mockStorageRepo struct {
	connections   []domain.BankConnection
	checkpoints   map[string]*domain.SyncCheckpoint
	processedMap  map[string]bool
	recordedTxs   []*domain.ProcessedTransaction
	mu            sync.Mutex
}

func newMockStorageRepo() *mockStorageRepo {
	return &mockStorageRepo{
		connections:  make([]domain.BankConnection, 0),
		checkpoints:  make(map[string]*domain.SyncCheckpoint),
		processedMap: make(map[string]bool),
		recordedTxs:  make([]*domain.ProcessedTransaction, 0),
	}
}

func (m *mockStorageRepo) SaveConnection(ctx context.Context, conn *domain.BankConnection) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections = append(m.connections, *conn)
	return nil
}

func (m *mockStorageRepo) FindConnectionByID(ctx context.Context, id string) (*domain.BankConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.connections {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

func (m *mockStorageRepo) FindConnectionByAccountID(ctx context.Context, accountID string) (*domain.BankConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.connections {
		if c.BankAccountID == accountID || c.ID == accountID {
			return &c, nil
		}
	}
	return nil, nil
}

func (m *mockStorageRepo) UpdateConnectionStatus(ctx context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, c := range m.connections {
		if c.ID == id || c.BankAccountID == id {
			m.connections[i].Status = status
			return nil
		}
	}
	return nil
}

func (m *mockStorageRepo) ListActiveConnections(ctx context.Context) ([]domain.BankConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connections, nil
}

func (m *mockStorageRepo) UpdateConnectionTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error {
	return nil
}

func (m *mockStorageRepo) GetCheckpoint(ctx context.Context, connectionID string) (*domain.SyncCheckpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.checkpoints[connectionID], nil
}

func (m *mockStorageRepo) SaveCheckpoint(ctx context.Context, cp *domain.SyncCheckpoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkpoints[cp.ConnectionID] = cp
	return nil
}

func (m *mockStorageRepo) IsTransactionProcessed(ctx context.Context, bankTxID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.processedMap[bankTxID], nil
}

func (m *mockStorageRepo) RecordProcessedTransaction(ctx context.Context, tx *domain.ProcessedTransaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedMap[tx.BankTransactionID] = true
	m.recordedTxs = append(m.recordedTxs, tx)
	return nil
}

// mockWorkerProvider implements domain.OpenFinanceProvider for worker testing
type mockWorkerProvider struct {
	txs []domain.BankTransaction
}

func (m *mockWorkerProvider) GetName() string { return "mock" }
func (m *mockWorkerProvider) ListAccounts(ctx context.Context, connectionID string) ([]domain.BankAccount, error) {
	return []domain.BankAccount{}, nil
}
func (m *mockWorkerProvider) FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*domain.TransactionPage, error) {
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
func (m *mockWorkerProvider) HealthCheck(ctx context.Context) error { return nil }

// TestSyncWorkerCycleSuccess tests the automated sync cycle
func TestSyncWorkerCycleSuccess(t *testing.T) {
	midazServerCalls := 0
	midazServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		midazServerCalls++
		w.WriteHeader(http.StatusCreated)
	}))
	defer midazServer.Close()

	cfg := &bootstrap.Config{
		MidazTransactionURL:       midazServer.URL,
		MidazClearingAccountAlias: "@external/BRL:openfinance",
		LookbackWindowHours:       24,
	}

	storage := newMockStorageRepo()
	_ = storage.SaveConnection(context.Background(), &domain.BankConnection{
		ID:                "conn-01",
		BankCompe:         "341",
		BankAccountID:     "acc-itau-01",
		MidazAccountAlias: "itau_operating_acc",
		Status:            "ACTIVE",
	})

	amount, _ := decimal.NewFromString("350.00")
	provider := &mockWorkerProvider{
		txs: []domain.BankTransaction{
			{
				ID:              "TX-WORKER-01",
				Amount:          amount,
				CreditDebitType: domain.CreditDebitCredit,
				Description:     "Pix Inflow Test",
				BookingDateTime: time.Now().UTC(),
			},
		},
	}

	midazClient := midaz.NewMidazClient(cfg)
	logger := &libLog.NoneLogger{}

	worker := NewSyncWorker(cfg, storage, provider, midazClient, logger)

	// Execute single sync cycle
	worker.runSyncCycle(context.Background())

	assert.Equal(t, 1, midazServerCalls)
	assert.Len(t, storage.recordedTxs, 1)
	assert.Equal(t, "TX-WORKER-01", storage.recordedTxs[0].BankTransactionID)

	cp, err := storage.GetCheckpoint(context.Background(), "conn-01")
	assert.NoError(t, err)
	assert.NotNil(t, cp)
	assert.Equal(t, "SUCCESS", cp.Status)
	assert.Equal(t, 1, cp.RecordsSynced)
}

// TestSyncWorkerDeduplication tests that already processed transactions are skipped
func TestSyncWorkerDeduplication(t *testing.T) {
	midazServerCalls := 0
	midazServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		midazServerCalls++
		w.WriteHeader(http.StatusCreated)
	}))
	defer midazServer.Close()

	cfg := &bootstrap.Config{
		MidazTransactionURL:       midazServer.URL,
		MidazClearingAccountAlias: "@external/BRL:openfinance",
	}

	storage := newMockStorageRepo()
	_ = storage.SaveConnection(context.Background(), &domain.BankConnection{
		ID:                "conn-01",
		BankCompe:         "341",
		BankAccountID:     "acc-itau-01",
		MidazAccountAlias: "itau_operating_acc",
		Status:            "ACTIVE",
	})

	// Pre-mark TX as already processed
	storage.processedMap["TX-ALREADY-DONE"] = true

	amount, _ := decimal.NewFromString("50.00")
	provider := &mockWorkerProvider{
		txs: []domain.BankTransaction{
			{
				ID:              "TX-ALREADY-DONE",
				Amount:          amount,
				CreditDebitType: domain.CreditDebitDebit,
			},
		},
	}

	midazClient := midaz.NewMidazClient(cfg)
	worker := NewSyncWorker(cfg, storage, provider, midazClient, &libLog.NoneLogger{})

	worker.runSyncCycle(context.Background())

	// Midaz should NOT have been called because transaction is already marked as processed
	assert.Equal(t, 0, midazServerCalls)
}
