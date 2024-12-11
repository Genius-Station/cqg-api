DROP TRIGGER IF EXISTS set_account_symbols_updated_at ON account_symbols;
DROP TRIGGER IF EXISTS set_account_config_updated_at ON account_config;
DROP TRIGGER IF EXISTS set_account_updated_at ON accounts;

DROP TABLE IF EXISTS account_symbols;
DROP TABLE IF EXISTS account_config;
DROP TABLE IF EXISTS symbols;
DROP TABLE IF EXISTS accounts;

DROP FUNCTION IF EXISTS update_timestamp;
