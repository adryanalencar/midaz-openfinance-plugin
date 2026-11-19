package postgres

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateUp_AllPendingApplied(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test_table (id INT);"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Advisory Lock
	mock.ExpectExec("SELECT pg_advisory_lock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Ensure schema_migrations table
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Query applied versions (empty)
	mock.ExpectQuery("SELECT version FROM schema_migrations;").
		WillReturnRows(sqlmock.NewRows([]string{"version"}))

	// 4. Begin TX for 000001
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test_table \\(id INT\\);").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO schema_migrations").
		WithArgs(int64(1), "000001_create_test").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 5. Release Advisory Lock
	mock.ExpectExec("SELECT pg_advisory_unlock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateUp(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMigrateUp_AlreadyApplied(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test_table (id INT);"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Advisory Lock
	mock.ExpectExec("SELECT pg_advisory_lock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Ensure schema_migrations table
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Query applied versions (already has version 1)
	mock.ExpectQuery("SELECT version FROM schema_migrations;").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1))

	// No TX expected since version 1 is already applied

	// 4. Release Advisory Lock
	mock.ExpectExec("SELECT pg_advisory_unlock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateUp(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMigrateUp_ExecutionFailsAndRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_bad_sql.up.sql": &fstest.MapFile{
			Data: []byte("INVALID SQL SYNTAX;"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Advisory Lock
	mock.ExpectExec("SELECT pg_advisory_lock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Ensure table
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Query applied (empty)
	mock.ExpectQuery("SELECT version FROM schema_migrations;").
		WillReturnRows(sqlmock.NewRows([]string{"version"}))

	// 4. Begin TX, fail on invalid SQL, rollback
	mock.ExpectBegin()
	mock.ExpectExec("INVALID SQL SYNTAX;").
		WillReturnError(errors.New("syntax error at or near INVALID"))
	mock.ExpectRollback()

	// 5. Release Advisory Lock in defer
	mock.ExpectExec("SELECT pg_advisory_unlock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateUp(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error at or near INVALID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMigrateDown_RollbackLatest(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.down.sql": &fstest.MapFile{
			Data: []byte("DROP TABLE test_table;"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Advisory Lock
	mock.ExpectExec("SELECT pg_advisory_lock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Ensure schema_migrations
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Query applied versions (returns 1)
	mock.ExpectQuery("SELECT version FROM schema_migrations;").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1))

	// 4. Begin TX for rollback
	mock.ExpectBegin()
	mock.ExpectExec("DROP TABLE test_table;").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM schema_migrations WHERE version = \\$1;").
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 5. Release Lock
	mock.ExpectExec("SELECT pg_advisory_unlock\\(\\$1\\);").
		WithArgs(AdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateDown(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
