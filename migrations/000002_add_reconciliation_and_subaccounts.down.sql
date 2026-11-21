-- 000002_add_reconciliation_and_subaccounts.down.sql

DROP INDEX IF EXISTS idx_bank_connections_reconcile;
DROP INDEX IF EXISTS idx_bank_connections_mother;

ALTER TABLE processed_transactions
DROP COLUMN IF EXISTS status;

ALTER TABLE bank_connections
DROP COLUMN IF EXISTS matcher_context_id,
DROP COLUMN IF EXISTS reconciliation_receivables_account,
DROP COLUMN IF EXISTS reconciliation_enabled,
DROP COLUMN IF EXISTS mother_account_alias,
DROP COLUMN IF EXISTS mother_account_id;
