package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_PostgreSQLDSN_DiscreteParams(t *testing.T) {
	cfg := &Config{
		DBHost:     "db.internal",
		DBPort:     "5432",
		DBUser:     "midaz_user",
		DBPassword: "secret_password",
		DBName:     "openfinance_db",
		DBSSLMode:  "disable",
	}

	assert.False(t, cfg.IsExternalDatabase())
	expectedDSN := "host=db.internal port=5432 user=midaz_user password=secret_password dbname=openfinance_db sslmode=disable"
	assert.Equal(t, expectedDSN, cfg.PostgreSQLDSN())
	assert.Equal(t, "host=db.internal port=5432 user=midaz_user dbname=openfinance_db sslmode=disable", cfg.RedactedPostgreSQLDSN())
}

func TestConfig_PostgreSQLDSN_DatabaseURLPriority(t *testing.T) {
	url := "postgres://midaz_user:secret_password@aws-rds.amazonaws.com:5432/custom_db?sslmode=require"
	cfg := &Config{
		DatabaseURL: url,
		DBHost:      "localhost", // Should be ignored because DatabaseURL has priority
	}

	assert.True(t, cfg.IsExternalDatabase())
	assert.Equal(t, url, cfg.PostgreSQLDSN())
	assert.Equal(t, "DATABASE_URL (configured)", cfg.RedactedPostgreSQLDSN())
}

func TestConfig_PostgreSQLDSN_DBURIFallback(t *testing.T) {
	uri := "postgresql://neon_user:neon_pass@neon.tech/neondb?sslmode=require"
	cfg := &Config{
		DBURI:  uri,
		DBHost: "localhost",
	}

	assert.True(t, cfg.IsExternalDatabase())
	assert.Equal(t, uri, cfg.PostgreSQLDSN())
	assert.Equal(t, "DB_URI (configured)", cfg.RedactedPostgreSQLDSN())
}
