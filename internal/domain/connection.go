package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// BankConnection stores the authorization and link between an external bank and a Midaz account.
type BankConnection struct {
	ID                string     `json:"id"`
	Provider          string     `json:"provider"` // "openfinance" or "pluggy"
	BankCompe         string     `json:"bankCompe"`
	BankAccountID     string     `json:"bankAccountId"`
	MidazAccountAlias string     `json:"midazAccountAlias"`
	ConsentID         string     `json:"consentId"`
	AccessToken       string     `json:"accessToken"`
	RefreshToken      string     `json:"refreshToken"`
	TokenExpiresAt    time.Time  `json:"tokenExpiresAt"`
	Status            string     `json:"status"` // "ACTIVE", "EXPIRED", "REVOKED"
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// SyncCheckpoint records the incremental cursor for continuous bank sync.
type SyncCheckpoint struct {
	ID             string     `json:"id"`
	ConnectionID   string     `json:"connectionId"`
	LastBookingDate time.Time `json:"lastBookingDate"`
	PaginationKey  string     `json:"paginationKey"`
	Status         string     `json:"status"` // "IDLE", "RUNNING", "SUCCESS", "FAILED"
	RecordsSynced  int        `json:"recordsSynced"`
	LastSyncedAt   time.Time  `json:"lastSyncedAt"`
}

// ProcessedTransaction represents an ingested transaction ledger entry for deduplication.
type ProcessedTransaction struct {
	ID                   string          `json:"id"`
	ConnectionID         string          `json:"connectionId"`
	BankTransactionID    string          `json:"bankTransactionId"`
	MidazTransactionCode string          `json:"midazTransactionCode"`
	Amount               decimal.Decimal `json:"amount"`
	CreditDebit          string          `json:"creditDebit"`
	SyncedAt             time.Time       `json:"syncedAt"`
}

// StorageRepository defines the persistence operations for connections, checkpoints, and idempotency.
type StorageRepository interface {
	// Connection operations
	SaveConnection(ctx context.Context, conn *BankConnection) error
	FindConnectionByID(ctx context.Context, id string) (*BankConnection, error)
	ListActiveConnections(ctx context.Context) ([]BankConnection, error)
	UpdateConnectionTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error

	// Checkpoint operations
	GetCheckpoint(ctx context.Context, connectionID string) (*SyncCheckpoint, error)
	SaveCheckpoint(ctx context.Context, cp *SyncCheckpoint) error

	// Idempotency operations
	IsTransactionProcessed(ctx context.Context, bankTxID string) (bool, error)
	RecordProcessedTransaction(ctx context.Context, tx *ProcessedTransaction) error
}
