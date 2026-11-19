-- ==============================================================================
-- Open Finance Plugin: Initial PostgreSQL Schema
-- ==============================================================================

CREATE TABLE IF NOT EXISTS bank_connections (
    id VARCHAR(36) PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,            -- 'openfinance' or 'pluggy'
    bank_compe VARCHAR(10) NOT NULL,          -- e.g. '341', '001', '260'
    bank_account_id VARCHAR(100) NOT NULL,    -- Account ID at banking institution
    midaz_account_alias VARCHAR(150) NOT NULL,-- Target account alias in Midaz
    consent_id VARCHAR(150),
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE', -- 'ACTIVE', 'EXPIRED', 'REVOKED'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_connections_status ON bank_connections(status);
CREATE INDEX IF NOT EXISTS idx_bank_connections_provider ON bank_connections(provider);

CREATE TABLE IF NOT EXISTS sync_checkpoints (
    id VARCHAR(36) PRIMARY KEY,
    connection_id VARCHAR(36) NOT NULL REFERENCES bank_connections(id) ON DELETE CASCADE,
    last_booking_date DATE,
    pagination_key VARCHAR(255),
    status VARCHAR(32) NOT NULL DEFAULT 'IDLE', -- 'IDLE', 'RUNNING', 'SUCCESS', 'FAILED'
    records_synced INTEGER NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_sync_checkpoints_conn UNIQUE (connection_id)
);

CREATE TABLE IF NOT EXISTS processed_transactions (
    id VARCHAR(36) PRIMARY KEY,
    connection_id VARCHAR(36) NOT NULL REFERENCES bank_connections(id) ON DELETE CASCADE,
    bank_transaction_id VARCHAR(100) NOT NULL,
    midaz_transaction_code VARCHAR(120) NOT NULL,
    amount NUMERIC(18, 4) NOT NULL,
    credit_debit VARCHAR(10) NOT NULL,
    synced_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_processed_tx_id UNIQUE (bank_transaction_id)
);

CREATE INDEX IF NOT EXISTS idx_processed_tx_midaz_code ON processed_transactions(midaz_transaction_code);
