package in

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/command"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPluggyHandlerStorage struct {
	savedConns map[string]*domain.BankConnection
}

func (m *mockPluggyHandlerStorage) SaveConnection(ctx context.Context, conn *domain.BankConnection) error {
	m.savedConns[conn.ID] = conn
	return nil
}
func (m *mockPluggyHandlerStorage) FindConnectionByID(ctx context.Context, id string) (*domain.BankConnection, error) {
	return m.savedConns[id], nil
}
func (m *mockPluggyHandlerStorage) FindConnectionByAccountID(ctx context.Context, accountID string) (*domain.BankConnection, error) {
	return nil, nil
}
func (m *mockPluggyHandlerStorage) ListActiveConnections(ctx context.Context) ([]domain.BankConnection, error) {
	return nil, nil
}
func (m *mockPluggyHandlerStorage) UpdateConnectionTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error {
	return nil
}
func (m *mockPluggyHandlerStorage) UpdateConnectionStatus(ctx context.Context, id, status string) error {
	return nil
}
func (m *mockPluggyHandlerStorage) GetCheckpoint(ctx context.Context, connectionID string) (*domain.SyncCheckpoint, error) {
	return nil, nil
}
func (m *mockPluggyHandlerStorage) SaveCheckpoint(ctx context.Context, cp *domain.SyncCheckpoint) error {
	return nil
}
func (m *mockPluggyHandlerStorage) IsTransactionProcessed(ctx context.Context, bankTxID string) (bool, error) {
	return false, nil
}
func (m *mockPluggyHandlerStorage) RecordProcessedTransaction(ctx context.Context, tx *domain.ProcessedTransaction) error {
	return nil
}

func TestPluggyEndpoints(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey": "pluggy-key"}`))
		case "/connect_token":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"accessToken": "widget-token"}`))
		case "/items/item-123":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "item-123", "status": "UPDATED"}`))
		case "/accounts":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"results": [{"id": "acc-123", "name": "Conta Corrente", "connector": {"name": "Nubank"}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		PluginHost:         "http://localhost:8085",
		PluggyBaseURL:      mockServer.URL,
		PluggyClientID:     "client-1",
		PluggyClientSecret: "secret-1",
	}

	pluggyClient, err := pluggy.NewPluggyClient(cfg)
	require.NoError(t, err)

	storage := &mockPluggyHandlerStorage{savedConns: make(map[string]*domain.BankConnection)}
	createConnCmd := command.NewCreatePluggyConnectionUseCase(pluggyClient, storage)

	mh := NewManifestHandler(cfg)
	sh := NewSyncHandler(nil, nil)
	ph := NewPluggyHandler(pluggyClient, createConnCmd)
	app := NewRouter(mh, sh, nil, ph, nil)

	// 1. POST /v1/connect-token/pluggy
	reqToken := httptest.NewRequest(http.MethodPost, "/v1/connect-token/pluggy", nil)
	respToken, err := app.Test(reqToken, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respToken.StatusCode)

	// 2. POST /v1/connections/pluggy - Validation Error
	reqBadConn := httptest.NewRequest(http.MethodPost, "/v1/connections/pluggy", bytes.NewBuffer([]byte(`{}`)))
	reqBadConn.Header.Set("Content-Type", "application/json")
	respBadConn, err := app.Test(reqBadConn, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, respBadConn.StatusCode)

	// 3. POST /v1/connections/pluggy - Success
	validPayload := map[string]string{
		"itemId":            "item-123",
		"accountId":         "acc-123",
		"midazAccountAlias": "nubank_checking",
	}
	bodyBytes, _ := json.Marshal(validPayload)
	reqValidConn := httptest.NewRequest(http.MethodPost, "/v1/connections/pluggy", bytes.NewBuffer(bodyBytes))
	reqValidConn.Header.Set("Content-Type", "application/json")
	respValidConn, err := app.Test(reqValidConn, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, respValidConn.StatusCode)
}
