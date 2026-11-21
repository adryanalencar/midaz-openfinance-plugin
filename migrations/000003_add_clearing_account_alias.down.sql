-- 000003_add_clearing_account_alias.down.sql

ALTER TABLE bank_connections
DROP COLUMN IF EXISTS clearing_account_alias;
