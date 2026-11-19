package postgres

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateUp_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test_table (id INT);"),
		},
		"000001_create_test.down.sql": &fstest.MapFile{
			Data: []byte("DROP TABLE test_table;"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Current Database
	mock.ExpectQuery("SELECT CURRENT_DATABASE\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("testdb"))

	// 2. Initial driver lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Check schema_migrations in information_schema.tables
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema\\.tables").
		WithArgs("public", "schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 4. Initial driver unlock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Migrate lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 6. Query current version (empty)
	mock.ExpectQuery("SELECT version, dirty FROM \"public\"\\.\"schema_migrations\" LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}))

	// 7. Set dirty true for version 1
	mock.ExpectBegin()
	mock.ExpectExec("TRUNCATE \"public\"\\.\"schema_migrations\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"public\"\\.\"schema_migrations\"").
		WithArgs(int64(1), true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 8. Execute migration content directly
	mock.ExpectExec("CREATE TABLE test_table \\(id INT\\);").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 9. Set dirty false for version 1 (TRUNCATE + INSERT (1, false))
	mock.ExpectBegin()
	mock.ExpectExec("TRUNCATE \"public\"\\.\"schema_migrations\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"public\"\\.\"schema_migrations\"").
		WithArgs(int64(1), false).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 10. Release advisory lock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateUp(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMigrateUp_AlreadyUpToDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test_table (id INT);"),
		},
		"000001_create_test.down.sql": &fstest.MapFile{
			Data: []byte("DROP TABLE test_table;"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Current Database
	mock.ExpectQuery("SELECT CURRENT_DATABASE\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("testdb"))

	// 2. Initial lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Check schema_migrations
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema\\.tables").
		WithArgs("public", "schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 4. Initial unlock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Migrate lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 6. Version already at 1 (not dirty)
	mock.ExpectQuery("SELECT version, dirty FROM \"public\"\\.\"schema_migrations\" LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(1, false))

	// 7. Release advisory lock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateUp(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMigrateDown_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test_table (id INT);"),
		},
		"000001_create_test.down.sql": &fstest.MapFile{
			Data: []byte("DROP TABLE test_table;"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Current Database
	mock.ExpectQuery("SELECT CURRENT_DATABASE\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("testdb"))

	// 2. Initial lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Check table
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema\\.tables").
		WithArgs("public", "schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 4. Initial unlock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Migrate lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 6. Current version is 1
	mock.ExpectQuery("SELECT version, dirty FROM \"public\"\\.\"schema_migrations\" LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(1, false))

	// 7. Revert to -1 with dirty=true
	mock.ExpectBegin()
	mock.ExpectExec("TRUNCATE \"public\"\\.\"schema_migrations\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"public\"\\.\"schema_migrations\"").
		WithArgs(int64(-1), true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 8. Execute down migration SQL
	mock.ExpectExec("DROP TABLE test_table;").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 9. Revert version to nil (TRUNCATE)
	mock.ExpectBegin()
	mock.ExpectExec("TRUNCATE \"public\"\\.\"schema_migrations\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 10. Release advisory lock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateDown(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
