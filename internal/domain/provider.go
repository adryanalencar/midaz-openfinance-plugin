package domain

import (
	"context"
	"time"
)

// OpenFinanceProvider is the unified port interface implemented by all banking providers
// (e.g., Direct Open Finance Brasil or aggregators like Pluggy).
type OpenFinanceProvider interface {
	// GetName returns the unique name of the provider driver.
	GetName() string

	// ListAccounts retrieves all authorized bank accounts for a given connection/consent.
	ListAccounts(ctx context.Context, connectionID string) ([]BankAccount, error)

	// FetchTransactions queries paginated bank transactions within the requested date window.
	FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*TransactionPage, error)

	// HealthCheck tests connectivity and certificate validity with the upstream provider.
	HealthCheck(ctx context.Context) error
}
