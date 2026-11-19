package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSaveConnection tests inserting a new bank connection into postgres
func TestSaveConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	conn := &domain.BankConnection{
		ID:                "conn-uuid-001",
		Provider:          "openfinance",
		BankCompe:         "341",
		BankAccountID:     "itau-acc-123",
		MidazAccountAlias: "midaz_itau_acc",
		Status:            "ACTIVE",
	}

	mock.ExpectExec("INSERT INTO bank_connections").
		WithArgs(
			conn.ID, conn.Provider, conn.BankCompe, conn.BankAccountID, conn.MidazAccountAlias,
			conn.ConsentID, conn.AccessToken, conn.RefreshToken, conn.TokenExpiresAt, conn.Status,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveConnection(context.Background(), conn)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestListActiveConnections tests listing active connections for the sync worker
func TestListActiveConnections(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "provider", "bank_compe", "bank_account_id", "midaz_account_alias",
		"consent_id", "access_token", "refresh_token", "token_expires_at", "status", "created_at", "updated_at",
	}).AddRow(
		"conn-1", "openfinance", "341", "acc-100", "midaz_acc_1",
		"consent-1", "token-1", "refresh-1", time.Now(), "ACTIVE", time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT (.+) FROM bank_connections WHERE status = 'ACTIVE'").
		WillReturnRows(rows)

	conns, err := repo.ListActiveConnections(context.Background())
	assert.NoError(t, err)
	assert.Len(t, conns, 1)
	assert.Equal(t, "conn-1", conns[0].ID)
	assert.Equal(t, "341", conns[0].BankCompe)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestIsTransactionProcessed tests checking transaction deduplication
func TestIsTransactionProcessed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	// Case 1: Transaction exists
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("TX-EXISTS-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.IsTransactionProcessed(context.Background(), "TX-EXISTS-1")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Case 2: Transaction does not exist
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("TX-NEW-2").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err = repo.IsTransactionProcessed(context.Background(), "TX-NEW-2")
	assert.NoError(t, err)
	assert.False(t, exists)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRecordProcessedTransaction tests saving ingested transaction
func TestRecordProcessedTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	amt, _ := decimal.NewFromString("100.50")
	tx := &domain.ProcessedTransaction{
		ConnectionID:         "conn-1",
		BankTransactionID:    "TX-100",
		MidazTransactionCode: "OFP-TX-100",
		Amount:               amt,
		CreditDebit:          "CREDITO",
	}

	mock.ExpectExec("INSERT INTO processed_transactions").
		WithArgs(sqlmock.AnyArg(), tx.ConnectionID, tx.BankTransactionID, tx.MidazTransactionCode, 100.50, tx.CreditDebit, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.RecordProcessedTransaction(context.Background(), tx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
