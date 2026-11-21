package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Repository implements domain.StorageRepository backed by PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL storage repository.
func NewPostgresRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveConnection inserts or updates a bank connection.
func (r *Repository) SaveConnection(ctx context.Context, conn *domain.BankConnection) error {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = now
	}
	conn.UpdatedAt = now

	query := `
		INSERT INTO bank_connections (
			id, provider, bank_compe, bank_account_id, midaz_account_alias,
			mother_account_id, mother_account_alias, reconciliation_enabled,
			reconciliation_receivables_account, matcher_context_id,
			consent_id, access_token, refresh_token, token_expires_at, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expires_at = EXCLUDED.token_expires_at,
			status = EXCLUDED.status,
			reconciliation_enabled = EXCLUDED.reconciliation_enabled,
			reconciliation_receivables_account = EXCLUDED.reconciliation_receivables_account,
			matcher_context_id = EXCLUDED.matcher_context_id,
			updated_at = EXCLUDED.updated_at;
	`

	_, err := r.db.ExecContext(ctx, query,
		conn.ID, conn.Provider, conn.BankCompe, conn.BankAccountID, conn.MidazAccountAlias,
		conn.MotherAccountID, conn.MotherAccountAlias, conn.ReconciliationEnabled,
		conn.ReconciliationReceivablesAccount, conn.MatcherContextID,
		conn.ConsentID, conn.AccessToken, conn.RefreshToken, conn.TokenExpiresAt, conn.Status,
		conn.CreatedAt, conn.UpdatedAt,
	)
	return err
}

// FindConnectionByID finds a connection by its primary ID.
func (r *Repository) FindConnectionByID(ctx context.Context, id string) (*domain.BankConnection, error) {
	query := `
		SELECT id, provider, bank_compe, bank_account_id, midaz_account_alias,
		       mother_account_id, mother_account_alias, reconciliation_enabled,
		       reconciliation_receivables_account, matcher_context_id,
		       consent_id, access_token, refresh_token, token_expires_at, status, created_at, updated_at
		FROM bank_connections
		WHERE id = $1;
	`

	var conn domain.BankConnection
	var motherID, motherAlias, recAccount, matcherCtx sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conn.ID, &conn.Provider, &conn.BankCompe, &conn.BankAccountID, &conn.MidazAccountAlias,
		&motherID, &motherAlias, &conn.ReconciliationEnabled, &recAccount, &matcherCtx,
		&conn.ConsentID, &conn.AccessToken, &conn.RefreshToken, &conn.TokenExpiresAt, &conn.Status,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if motherID.Valid {
		conn.MotherAccountID = motherID.String
	}
	if motherAlias.Valid {
		conn.MotherAccountAlias = motherAlias.String
	}
	if recAccount.Valid {
		conn.ReconciliationReceivablesAccount = recAccount.String
	}
	if matcherCtx.Valid {
		conn.MatcherContextID = matcherCtx.String
	}

	return &conn, nil
}

// FindConnectionByAccountID finds a bank connection by its external account ID or primary ID.
func (r *Repository) FindConnectionByAccountID(ctx context.Context, accountID string) (*domain.BankConnection, error) {
	query := `
		SELECT id, provider, bank_compe, bank_account_id, midaz_account_alias,
		       mother_account_id, mother_account_alias, reconciliation_enabled,
		       reconciliation_receivables_account, matcher_context_id,
		       consent_id, access_token, refresh_token, token_expires_at, status, created_at, updated_at
		FROM bank_connections
		WHERE bank_account_id = $1 OR id = $1
		LIMIT 1;
	`

	var conn domain.BankConnection
	var motherID, motherAlias, recAccount, matcherCtx sql.NullString

	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&conn.ID, &conn.Provider, &conn.BankCompe, &conn.BankAccountID, &conn.MidazAccountAlias,
		&motherID, &motherAlias, &conn.ReconciliationEnabled, &recAccount, &matcherCtx,
		&conn.ConsentID, &conn.AccessToken, &conn.RefreshToken, &conn.TokenExpiresAt, &conn.Status,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if motherID.Valid {
		conn.MotherAccountID = motherID.String
	}
	if motherAlias.Valid {
		conn.MotherAccountAlias = motherAlias.String
	}
	if recAccount.Valid {
		conn.ReconciliationReceivablesAccount = recAccount.String
	}
	if matcherCtx.Valid {
		conn.MatcherContextID = matcherCtx.String
	}

	return &conn, nil
}

// UpdateConnectionStatus updates the operational status of a bank connection.
func (r *Repository) UpdateConnectionStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE bank_connections
		SET status = $1, updated_at = $2
		WHERE id = $3 OR bank_account_id = $3;
	`
	_, err := r.db.ExecContext(ctx, query, status, time.Now().UTC(), id)
	return err
}

// ListActiveConnections returns all active bank connections scheduled for synchronization.
func (r *Repository) ListActiveConnections(ctx context.Context) ([]domain.BankConnection, error) {
	query := `
		SELECT id, provider, bank_compe, bank_account_id, midaz_account_alias,
		       mother_account_id, mother_account_alias, reconciliation_enabled,
		       reconciliation_receivables_account, matcher_context_id,
		       consent_id, access_token, refresh_token, token_expires_at, status, created_at, updated_at
		FROM bank_connections
		WHERE status = 'ACTIVE';
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []domain.BankConnection
	for rows.Next() {
		var conn domain.BankConnection
		var motherID, motherAlias, recAccount, matcherCtx sql.NullString

		if err := rows.Scan(
			&conn.ID, &conn.Provider, &conn.BankCompe, &conn.BankAccountID, &conn.MidazAccountAlias,
			&motherID, &motherAlias, &conn.ReconciliationEnabled, &recAccount, &matcherCtx,
			&conn.ConsentID, &conn.AccessToken, &conn.RefreshToken, &conn.TokenExpiresAt, &conn.Status,
			&conn.CreatedAt, &conn.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if motherID.Valid {
			conn.MotherAccountID = motherID.String
		}
		if motherAlias.Valid {
			conn.MotherAccountAlias = motherAlias.String
		}
		if recAccount.Valid {
			conn.ReconciliationReceivablesAccount = recAccount.String
		}
		if matcherCtx.Valid {
			conn.MatcherContextID = matcherCtx.String
		}

		connections = append(connections, conn)
	}

	return connections, rows.Err()
}

// UpdateConnectionTokens refreshes authentication credentials for a bank connection.
func (r *Repository) UpdateConnectionTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error {
	query := `
		UPDATE bank_connections
		SET access_token = $1, refresh_token = $2, token_expires_at = $3, updated_at = $4
		WHERE id = $5;
	`
	_, err := r.db.ExecContext(ctx, query, accessToken, refreshToken, expiresAt, time.Now().UTC(), id)
	return err
}

// GetCheckpoint retrieves the last sync checkpoint for a given connection.
func (r *Repository) GetCheckpoint(ctx context.Context, connectionID string) (*domain.SyncCheckpoint, error) {
	query := `
		SELECT id, connection_id, last_booking_date, pagination_key, status, records_synced, last_synced_at
		FROM sync_checkpoints
		WHERE connection_id = $1;
	`

	var cp domain.SyncCheckpoint
	var lastBookingDate sql.NullTime
	var lastSyncedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, connectionID).Scan(
		&cp.ID, &cp.ConnectionID, &lastBookingDate, &cp.PaginationKey, &cp.Status, &cp.RecordsSynced, &lastSyncedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lastBookingDate.Valid {
		cp.LastBookingDate = lastBookingDate.Time
	}
	if lastSyncedAt.Valid {
		cp.LastSyncedAt = lastSyncedAt.Time
	}

	return &cp, nil
}

// SaveCheckpoint saves or updates a checkpoint cursor for continuous incremental sync.
func (r *Repository) SaveCheckpoint(ctx context.Context, cp *domain.SyncCheckpoint) error {
	if cp.ID == "" {
		cp.ID = uuid.New().String()
	}
	cp.LastSyncedAt = time.Now().UTC()

	query := `
		INSERT INTO sync_checkpoints (
			id, connection_id, last_booking_date, pagination_key, status, records_synced, last_synced_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (connection_id) DO UPDATE SET
			last_booking_date = EXCLUDED.last_booking_date,
			pagination_key = EXCLUDED.pagination_key,
			status = EXCLUDED.status,
			records_synced = EXCLUDED.records_synced,
			last_synced_at = EXCLUDED.last_synced_at,
			updated_at = EXCLUDED.updated_at;
	`

	_, err := r.db.ExecContext(ctx, query,
		cp.ID, cp.ConnectionID, cp.LastBookingDate, cp.PaginationKey, cp.Status, cp.RecordsSynced, cp.LastSyncedAt, cp.LastSyncedAt,
	)
	return err
}

// IsTransactionProcessed verifies if a transaction from the bank was already posted to Midaz.
func (r *Repository) IsTransactionProcessed(ctx context.Context, bankTxID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM processed_transactions WHERE bank_transaction_id = $1);`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, bankTxID).Scan(&exists)
	return exists, err
}

// RecordProcessedTransaction persists an ingested transaction ID for deduplication.
func (r *Repository) RecordProcessedTransaction(ctx context.Context, tx *domain.ProcessedTransaction) error {
	if tx.ID == "" {
		tx.ID = uuid.New().String()
	}
	if tx.SyncedAt.IsZero() {
		tx.SyncedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO processed_transactions (
			id, connection_id, bank_transaction_id, midaz_transaction_code, amount, credit_debit, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (bank_transaction_id) DO NOTHING;
	`

	amountFloat, _ := tx.Amount.Float64()

	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.ConnectionID, tx.BankTransactionID, tx.MidazTransactionCode, amountFloat, tx.CreditDebit, tx.SyncedAt,
	)
	return err
}

// ConnectPostgres connects to PostgreSQL and validates connectivity.
func ConnectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	return db, nil
}
