package command

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWebhookStorage struct {
	connections   map[string]*domain.BankConnection
	processedTxs  map[string]bool
	recordedTxs   []*domain.ProcessedTransaction
	savedCheckpts map[string]*domain.SyncCheckpoint
	mu            sync.Mutex
}

func newMockWebhookStorage() *mockWebhookStorage {
	return &mockWebhookStorage{
		connections:   make(map[string]*domain.BankConnection),
		processedTxs:  make(map[string]bool),
		recordedTxs:   make([]*domain.ProcessedTransaction, 0),
		savedCheckpts: make(map[string]*domain.SyncCheckpoint),
	}
}

func (m *mockWebhookStorage) SaveConnection(ctx context.Context, conn *domain.BankConnection) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[conn.ID] = conn
	m.connections[conn.BankAccountID] = conn
	return nil
}

func (m *mockWebhookStorage) FindConnectionByID(ctx context.Context, id string) (*domain.BankConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connections[id], nil
}

func (m *mockWebhookStorage) FindConnectionByAccountID(ctx context.Context, accountID string) (*domain.BankConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connections[accountID], nil
}

func (m *mockWebhookStorage) ListActiveConnections(ctx context.Context) ([]domain.BankConnection, error) {
	return nil, nil
}

func (m *mockWebhookStorage) UpdateConnectionTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error {
	return nil
}

func (m *mockWebhookStorage) UpdateConnectionStatus(ctx context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.connections[id]; ok {
		c.Status = status
	}
	return nil
}

func (m *mockWebhookStorage) GetCheckpoint(ctx context.Context, connectionID string) (*domain.SyncCheckpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.savedCheckpts[connectionID], nil
}

func (m *mockWebhookStorage) SaveCheckpoint(ctx context.Context, cp *domain.SyncCheckpoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.savedCheckpts[cp.ConnectionID] = cp
	return nil
}

func (m *mockWebhookStorage) IsTransactionProcessed(ctx context.Context, bankTxID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.processedTxs[bankTxID], nil
}

func (m *mockWebhookStorage) RecordProcessedTransaction(ctx context.Context, tx *domain.ProcessedTransaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedTxs[tx.BankTransactionID] = true
	m.recordedTxs = append(m.recordedTxs, tx)
	return nil
}

func TestHandlePluggyWebhook_UnauthorizedClientID(t *testing.T) {
	cfg := &bootstrap.Config{
		PluggyClientID: "authorized-client-id",
	}

	uc := NewHandlePluggyWebhookUseCase(cfg, nil, nil, nil, &libLog.NoneLogger{})
	err := uc.Execute(context.Background(), &pluggy.WebhookEvent{
		Event:    "item/updated",
		ClientID: "malicious-client-id",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized webhook client_id")
}

func TestHandlePluggyWebhook_TransactionsCreated_Success(t *testing.T) {
	// 1. Mock Pluggy Server for auth and v2 transactions
	mockPluggy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey": "pluggy-key"}`))
		case "/v2/transactions":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(pluggy.CursorPageResponse{
				Results: []pluggy.RawPluggyTransaction{
					{
						ID:          "tx-webhook-01",
						Description: "Pix Inflow Webhook",
						Amount:      350.00,
						Type:        "CREDIT",
						Date:        "2026-11-19T10:00:00Z",
						Currency:    "BRL",
					},
				},
				Next: nil,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockPluggy.Close()

	// 2. Mock Midaz Server for posting double-entry transactions
	midazCalls := 0
	mockMidaz := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		midazCalls++
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockMidaz.Close()

	cfg := &bootstrap.Config{
		PluggyBaseURL:             mockPluggy.URL,
		PluggyClientID:            "client-id-123",
		MidazTransactionURL:       mockMidaz.URL,
		MidazClearingAccountAlias: "@external/BRL:openfinance",
	}

	pluggyClient, err := pluggy.NewPluggyClient(cfg)
	require.NoError(t, err)

	midazClient := midaz.NewMidazClient(cfg)
	storage := newMockWebhookStorage()

	// Seed pre-existing connection
	_ = storage.SaveConnection(context.Background(), &domain.BankConnection{
		ID:                "conn-01",
		BankAccountID:     "acc-pluggy-01",
		MidazAccountAlias: "midaz_itau_acc",
		Status:            "ACTIVE",
	})

	uc := NewHandlePluggyWebhookUseCase(cfg, pluggyClient, storage, midazClient, &libLog.NoneLogger{})

	// Fire webhook event
	err = uc.Execute(context.Background(), &pluggy.WebhookEvent{
		Event:                     "transactions/created",
		EventID:                   "evt-101",
		ClientID:                  "client-id-123",
		ItemID:                    "conn-01",
		AccountID:                 "acc-pluggy-01",
		CreatedTransactionsLinkV2: "/v2/transactions?accountId=acc-pluggy-01",
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, midazCalls)
	assert.Len(t, storage.recordedTxs, 1)
	assert.Equal(t, "tx-webhook-01", storage.recordedTxs[0].BankTransactionID)
	assert.Equal(t, "OFP-PLG-tx-webhook-01", storage.recordedTxs[0].MidazTransactionCode)

	cp, err := storage.GetCheckpoint(context.Background(), "conn-01")
	assert.NoError(t, err)
	assert.Equal(t, 1, cp.RecordsSynced)
	assert.Equal(t, "SUCCESS", cp.Status)
}

func TestHandlePluggyWebhook_ItemStatusEvents(t *testing.T) {
	storage := newMockWebhookStorage()
	_ = storage.SaveConnection(context.Background(), &domain.BankConnection{
		ID:                "item-01",
		BankAccountID:     "acc-01",
		MidazAccountAlias: "operating_acc",
		Status:            "ACTIVE",
	})

	cfg := &bootstrap.Config{
		PluggyClientID: "client-id-123",
	}

	uc := NewHandlePluggyWebhookUseCase(cfg, nil, storage, nil, &libLog.NoneLogger{})

	// 1. item/error
	err := uc.Execute(context.Background(), &pluggy.WebhookEvent{
		Event:    "item/error",
		ClientID: "client-id-123",
		ItemID:   "item-01",
	})
	assert.NoError(t, err)
	conn, _ := storage.FindConnectionByID(context.Background(), "item-01")
	assert.Equal(t, "LOGIN_ERROR", conn.Status)

	// 2. item/updated
	err = uc.Execute(context.Background(), &pluggy.WebhookEvent{
		Event:    "item/updated",
		ClientID: "client-id-123",
		ItemID:   "item-01",
	})
	assert.NoError(t, err)
	conn, _ = storage.FindConnectionByID(context.Background(), "item-01")
	assert.Equal(t, "ACTIVE", conn.Status)
}
