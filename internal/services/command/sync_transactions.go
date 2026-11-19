package command

import (
	"context"
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
)

// SyncTransactionsInput defines the arguments for running a synchronization cycle.
type SyncTransactionsInput struct {
	BankAccountID          string
	TargetMidazAccountAlias string
	FromDate               time.Time
	ToDate                 time.Time
	PageSize               int
}

// SyncResult captures the statistics of a completed sync cycle.
type SyncResult struct {
	BankAccountID string    `json:"bankAccountId"`
	RecordsFetched int       `json:"recordsFetched"`
	RecordsPosted  int       `json:"recordsPosted"`
	RecordsFailed  int       `json:"recordsFailed"`
	SyncedAt       time.Time `json:"syncedAt"`
}

// SyncTransactionsUseCase coordinates the pull from provider and idempotent double-entry push to Midaz.
type SyncTransactionsUseCase struct {
	provider    domain.OpenFinanceProvider
	midazClient *midaz.Client
}

// NewSyncTransactionsUseCase creates a new SyncTransactionsUseCase.
func NewSyncTransactionsUseCase(provider domain.OpenFinanceProvider, midazClient *midaz.Client) *SyncTransactionsUseCase {
	return &SyncTransactionsUseCase{
		provider:    provider,
		midazClient: midazClient,
	}
}

// Execute pulls transactions from the provider and posts them to Midaz with double-entry guarantees.
func (uc *SyncTransactionsUseCase) Execute(ctx context.Context, input SyncTransactionsInput) (*SyncResult, error) {
	if input.PageSize <= 0 {
		input.PageSize = 100
	}

	result := &SyncResult{
		BankAccountID: input.BankAccountID,
		SyncedAt:      time.Now().UTC(),
	}

	currentPage := 1
	for {
		page, err := uc.provider.FetchTransactions(ctx, input.BankAccountID, input.FromDate, input.ToDate, currentPage, input.PageSize)
		if err != nil {
			return result, fmt.Errorf("provider error fetching page %d: %w", currentPage, err)
		}

		if page == nil || len(page.Transactions) == 0 {
			break
		}

		result.RecordsFetched += len(page.Transactions)

		for _, tx := range page.Transactions {
			postErr := uc.midazClient.PostDoubleEntryTransaction(ctx, tx, input.TargetMidazAccountAlias)
			if postErr != nil {
				result.RecordsFailed++
			} else {
				result.RecordsPosted++
			}
		}

		if page.TotalPages > 0 && currentPage >= page.TotalPages {
			break
		}
		currentPage++
	}

	return result, nil
}
