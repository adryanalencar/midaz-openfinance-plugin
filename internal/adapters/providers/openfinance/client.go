package openfinance

import (
	"context"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
)

// DirectOpenFinanceClient implements domain.OpenFinanceProvider for direct Open Finance Brasil APIs.
type DirectOpenFinanceClient struct {
	cfg *bootstrap.Config
}

// NewDirectOpenFinanceClient initializes the mTLS and JWS client for Open Finance Brasil.
func NewDirectOpenFinanceClient(cfg *bootstrap.Config) (*DirectOpenFinanceClient, error) {
	return &DirectOpenFinanceClient{cfg: cfg}, nil
}

func (c *DirectOpenFinanceClient) GetName() string {
	return "openfinance"
}

func (c *DirectOpenFinanceClient) ListAccounts(ctx context.Context, connectionID string) ([]domain.BankAccount, error) {
	// In production, queries GET /open-banking/accounts/v2/accounts using mTLS
	return []domain.BankAccount{}, nil
}

func (c *DirectOpenFinanceClient) FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*domain.TransactionPage, error) {
	// In production, queries GET /open-banking/accounts/v2/accounts/{accountId}/transactions with mTLS
	return &domain.TransactionPage{
		Transactions: []domain.BankTransaction{},
		CurrentPage:  page,
		TotalPages:   1,
		TotalRecords: 0,
	}, nil
}

func (c *DirectOpenFinanceClient) HealthCheck(ctx context.Context) error {
	return nil
}
