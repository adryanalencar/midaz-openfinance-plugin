package in

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorageRepo struct {
	mock.Mock
}

func (m *MockStorageRepo) SaveConnection(ctx context.Context, conn *domain.BankConnection) error {
	return nil
}
func (m *MockStorageRepo) FindConnectionByID(ctx context.Context, id string) (*domain.BankConnection, error) {
	return nil, nil
}
func (m *MockStorageRepo) FindConnectionByAccountID(ctx context.Context, accountID string) (*domain.BankConnection, error) {
	return nil, nil
}
func (m *MockStorageRepo) ListActiveConnections(ctx context.Context) ([]domain.BankConnection, error) {
	return nil, nil
}
func (m *MockStorageRepo) UpdateConnectionTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error {
	return nil
}
func (m *MockStorageRepo) UpdateConnectionStatus(ctx context.Context, id, status string) error {
	return nil
}
func (m *MockStorageRepo) GetCheckpoint(ctx context.Context, connectionID string) (*domain.SyncCheckpoint, error) {
	return nil, nil
}
func (m *MockStorageRepo) SaveCheckpoint(ctx context.Context, cp *domain.SyncCheckpoint) error {
	return nil
}
func (m *MockStorageRepo) IsTransactionProcessed(ctx context.Context, bankTxID string) (bool, error) {
	args := m.Called(ctx, bankTxID)
	return args.Bool(0), args.Error(1)
}
func (m *MockStorageRepo) RecordProcessedTransaction(ctx context.Context, tx *domain.ProcessedTransaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func TestInternalHandler_MarkProcessed(t *testing.T) {
	repo := new(MockStorageRepo)
	ih := NewInternalHandler(repo)
	app := NewRouter(nil, nil, nil, nil, ih)

	repo.On("RecordProcessedTransaction", mock.Anything, mock.MatchedBy(func(tx *domain.ProcessedTransaction) bool {
		return tx.BankTransactionID == "OFP-SICREDI-9988" &&
			tx.MidazTransactionCode == "MIDAZ-MULTI-777" &&
			tx.Status == "SETTLED_VIA_RECEIVABLES"
	})).Return(nil)

	payload := MarkProcessedRequest{
		BankTransactionID:    "OFP-SICREDI-9988",
		MidazTransactionCode: "MIDAZ-MULTI-777",
		Status:               "SETTLED_VIA_RECEIVABLES",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/internal/transactions/mark-processed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	repo.AssertExpectations(t)
}
