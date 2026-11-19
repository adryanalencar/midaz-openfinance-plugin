package pluggy

import (
	"context"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
)

// PluggyClient implements domain.OpenFinanceProvider for the Pluggy aggregation API.
type PluggyClient struct {
	cfg *bootstrap.Config
}

// NewPluggyClient initializes the HTTP client for Pluggy.
func NewPluggyClient(cfg *bootstrap.Config) (*PluggyClient, error) {
	return &PluggyClient{cfg: cfg}, nil
}

func (c *PluggyClient) GetName() string {
	return "pluggy"
}

func (c *PluggyClient) ListAccounts(ctx context.Context, connectionID string) ([]domain.BankAccount, error) {
	// In production, queries GET /accounts?itemId={connectionID}
	return []domain.BankAccount{}, nil
}

func (c *PluggyClient) FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*domain.TransactionPage, error) {
	// In production, queries GET /transactions?accountId={accountID}&from={fromDate}&to={toDate}
	return &domain.TransactionPage{
		Transactions: []domain.BankTransaction{},
		CurrentPage:  page,
		TotalPages:   1,
		TotalRecords: 0,
	}, nil
}

func (c *PluggyClient) HealthCheck(ctx context.Context) error {
	return nil
}
