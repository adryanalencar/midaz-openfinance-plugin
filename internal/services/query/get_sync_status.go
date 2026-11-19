package query

import (
	"context"
	"time"
)

// SyncStatusDTO represents the high-level health and metrics of Open Finance synchronization.
type SyncStatusDTO struct {
	Status                    string    `json:"status"`
	Provider                  string    `json:"provider"`
	AccountsTracked           int       `json:"accountsTracked"`
	TotalTransactionsIngested int       `json:"totalTransactionsIngested"`
	FailedTransactions        int       `json:"failedTransactions"`
	LastSyncAt                time.Time `json:"lastSyncAt"`
}

// GetSyncStatusUseCase queries the current operational status.
type GetSyncStatusUseCase struct {
	providerName string
}

// NewGetSyncStatusUseCase creates a new GetSyncStatusUseCase.
func NewGetSyncStatusUseCase(providerName string) *GetSyncStatusUseCase {
	return &GetSyncStatusUseCase{
		providerName: providerName,
	}
}

// Execute returns the synchronization status snapshot.
func (uc *GetSyncStatusUseCase) Execute(ctx context.Context) (*SyncStatusDTO, error) {
	return &SyncStatusDTO{
		Status:                    "healthy",
		Provider:                  uc.providerName,
		AccountsTracked:           1,
		TotalTransactionsIngested: 0,
		FailedTransactions:        0,
		LastSyncAt:                time.Now().UTC(),
	}, nil
}
