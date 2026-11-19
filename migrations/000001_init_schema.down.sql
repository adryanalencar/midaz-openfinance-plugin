-- ==============================================================================
-- Open Finance Plugin: Drop Initial Schema (Rollback)
-- ==============================================================================

DROP TABLE IF EXISTS processed_transactions CASCADE;
DROP TABLE IF EXISTS sync_checkpoints CASCADE;
DROP TABLE IF EXISTS bank_connections CASCADE;
