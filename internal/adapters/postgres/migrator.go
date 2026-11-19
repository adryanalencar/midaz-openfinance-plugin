package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/postgres/migrations"
)

const (
	// AdvisoryLockID is a constant hash representing plugin-openfinance schema migrations.
	AdvisoryLockID = 7429148
)

// MigrationFile represents a parsed SQL migration file.
type MigrationFile struct {
	Version   int64
	Name      string
	Direction string // "up" or "down"
	Filename  string
	Content   string
}

// Migrator manages transactional database schema migrations.
type Migrator struct {
	db     *sql.DB
	fs     fs.FS
	logger libLog.Logger
}

// NewMigrator creates a new database migrator with embedded migrations.
func NewMigrator(db *sql.DB, logger libLog.Logger) *Migrator {
	return NewMigratorWithFS(db, migrations.MigrationFS, logger)
}

// NewMigratorWithFS creates a migrator with a custom filesystem (useful for testing).
func NewMigratorWithFS(db *sql.DB, fsys fs.FS, logger libLog.Logger) *Migrator {
	if logger == nil {
		logger = &libLog.NoneLogger{}
	}
	return &Migrator{
		db:     db,
		fs:     fsys,
		logger: logger,
	}
}

// EnsureSchemaMigrationsTable initializes the migration tracking table.
func (m *Migrator) EnsureSchemaMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// GetAppliedVersions returns a set of all versions that have already run.
func (m *Migrator) GetAppliedVersions(ctx context.Context) (map[int64]bool, error) {
	if err := m.EnsureSchemaMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	query := `SELECT version FROM schema_migrations;`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, nil
}

// parseMigrations reads and validates all migration files from the filesystem.
func (m *Migrator) parseMigrations(direction string) ([]MigrationFile, error) {
	entries, err := fs.ReadDir(m.fs, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	var files []MigrationFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if !strings.HasSuffix(filename, fmt.Sprintf(".%s.sql", direction)) {
			continue
		}

		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			continue
		}

		ver, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in file %s: %w", filename, err)
		}

		content, err := fs.ReadFile(m.fs, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		name := strings.TrimSuffix(filepath.Base(filename), fmt.Sprintf(".%s.sql", direction))

		files = append(files, MigrationFile{
			Version:   ver,
			Name:      name,
			Direction: direction,
			Filename:  filename,
			Content:   string(content),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if direction == "up" {
			return files[i].Version < files[j].Version
		}
		return files[i].Version > files[j].Version
	})

	return files, nil
}

// AcquireAdvisoryLock acquires an exclusive PostgreSQL advisory lock.
func (m *Migrator) AcquireAdvisoryLock(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, "SELECT pg_advisory_lock($1);", AdvisoryLockID)
	return err
}

// ReleaseAdvisoryLock releases the PostgreSQL advisory lock.
func (m *Migrator) ReleaseAdvisoryLock(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, "SELECT pg_advisory_unlock($1);", AdvisoryLockID)
	return err
}

// MigrateUp executes all pending migrations sequentially within individual transactions.
func (m *Migrator) MigrateUp(ctx context.Context) error {
	if err := m.AcquireAdvisoryLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire migration advisory lock: %w", err)
	}
	defer func() {
		_ = m.ReleaseAdvisoryLock(ctx)
	}()

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return err
	}

	migrations, err := m.parseMigrations("up")
	if err != nil {
		return err
	}

	appliedCount := 0
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		m.logger.Infof("Executing migration up %s...", migration.Filename)

		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start transaction for %s: %w", migration.Filename, err)
		}

		if _, err := tx.ExecContext(ctx, migration.Content); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", migration.Filename, err)
		}

		recordQuery := `INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, NOW());`
		if _, err := tx.ExecContext(ctx, recordQuery, migration.Version, migration.Name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s in schema_migrations: %w", migration.Filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Filename, err)
		}

		m.logger.Infof("Successfully applied migration %06d (%s)", migration.Version, migration.Name)
		appliedCount++
	}

	if appliedCount == 0 {
		m.logger.Infof("Database schema is up to date (no pending migrations)")
	} else {
		m.logger.Infof("Applied %d database migration(s) successfully", appliedCount)
	}

	return nil
}

// MigrateDown rolls back the most recent applied migration.
func (m *Migrator) MigrateDown(ctx context.Context) error {
	if err := m.AcquireAdvisoryLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire migration advisory lock: %w", err)
	}
	defer func() {
		_ = m.ReleaseAdvisoryLock(ctx)
	}()

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return err
	}

	migrations, err := m.parseMigrations("down")
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if !applied[migration.Version] {
			continue
		}

		m.logger.Infof("Rolling back migration %s...", migration.Filename)

		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start rollback transaction for %s: %w", migration.Filename, err)
		}

		if _, err := tx.ExecContext(ctx, migration.Content); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("rollback %s failed: %w", migration.Filename, err)
		}

		deleteQuery := `DELETE FROM schema_migrations WHERE version = $1;`
		if _, err := tx.ExecContext(ctx, deleteQuery, migration.Version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to delete migration record %s: %w", migration.Filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit rollback for %s: %w", migration.Filename, err)
		}

		m.logger.Infof("Successfully rolled back migration %06d (%s)", migration.Version, migration.Name)
		return nil
	}

	m.logger.Infof("No migrations to rollback")
	return nil
}
