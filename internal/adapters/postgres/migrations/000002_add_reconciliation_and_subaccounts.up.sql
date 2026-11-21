-- 000002_add_reconciliation_and_subaccounts.up.sql

ALTER TABLE bank_connections
ADD COLUMN IF NOT EXISTS mother_account_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS mother_account_alias VARCHAR(150),
ADD COLUMN IF NOT EXISTS reconciliation_enabled BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS reconciliation_receivables_account VARCHAR(150),
ADD COLUMN IF NOT EXISTS matcher_context_id VARCHAR(150);

ALTER TABLE processed_transactions
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'POSTED_DIRECT';

CREATE INDEX IF NOT EXISTS idx_bank_connections_mother ON bank_connections(mother_account_id);
CREATE INDEX IF NOT EXISTS idx_bank_connections_reconcile ON bank_connections(reconciliation_enabled);
