-- 000003_add_clearing_account_alias.up.sql

ALTER TABLE bank_connections
ADD COLUMN IF NOT EXISTS clearing_account_alias VARCHAR(255);
