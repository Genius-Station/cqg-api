
CREATE TABLE IF NOT EXISTS accounts (
    id               SERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL UNIQUE,
    username         VARCHAR(255) NOT NULL UNIQUE,
    password         VARCHAR(255) NOT NULL,
    access_token     TEXT,
    autotrade_active BOOLEAN DEFAULT FALSE,
    valid_credentials BOOLEAN DEFAULT FALSE,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);


CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER set_account_updated_at
BEFORE UPDATE ON accounts
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();


CREATE TABLE IF NOT EXISTS account_config (
    id                       SERIAL PRIMARY KEY,
    account_id               BIGINT NOT NULL,  
    daily_profit_cap         NUMERIC,
    daily_profit_cap_quantifier VARCHAR(50),
    monoside                 VARCHAR(50),
    loose_limit              NUMERIC,
    leverage                 NUMERIC,
    loose_limit_quantifier   VARCHAR(50),
    profitability            NUMERIC,
    max_lots                 INTEGER,
    created_at               TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TRIGGER set_account_config_updated_at
BEFORE UPDATE ON account_config
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

ALTER TABLE account_config
ADD CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE;


CREATE TABLE IF NOT EXISTS symbols (
    id SERIAL PRIMARY KEY,     
    symbol_name VARCHAR(255) NOT NULL UNIQUE,   
    name VARCHAR(255) NOT NULL,       
    market VARCHAR(255) NOT NULL,     
    category VARCHAR(255) NOT NULL 
);

INSERT INTO symbols (symbol_name, name, market, category)
VALUES
    ('6B', 'British Bound', 'CME', 'Currencies'),
    ('6C', 'Canadian Dollar', 'CME', 'Currencies'),
    ('6E', 'Euro FX', 'CME', 'Currencies'),
    ('DX-M', 'US Dollar Index', 'ICE', 'Currencies'),
    ('QM', 'E-Mini Crude Oil', 'NYMEX', 'Energies'),
    ('QG', 'E-Mini Natural Gas', 'NYMEX', 'Energies'),
    ('FGBL', 'Euro-Bund', 'EUREX', 'Financials'),
    ('YM', 'E-Mini Dow', 'CBOT', 'Indices'),
    ('NQ', 'E-Mini NASDAQ 100', 'CME', 'Indices'),
    ('RTY', 'E-Mini Russell 100', 'CME', 'Indices'),
    ('ES', 'E-Mini S&P 500', 'CME', 'Indices'),
    ('FDXM', 'Mini-DAX', 'EUREX', 'Indices'),
    ('MYM', 'Micro E-mini Dow', 'CBOT', 'Indices'),
    ('MNQ', 'Micro E-mini Nasdaq-100', 'CME', 'Indices'),
    ('M2K', 'Micro E-mini Russell 2000', 'CME', 'Indices'),
    ('MES', 'Micro E-mini S&P 500', 'CME', 'Indices'),
    ('NKD', 'Nikkei 225/USD', 'CME', 'Indices'),
    ('MBT', 'CME Micro Bitcoin', 'CME', 'Indices'),
    ('BTC', 'CME Bitcoin Futures', 'CME', 'Indices'),
    ('QO', 'E-mini Gold', 'COMEX', 'Metals'),
    ('QI', 'E-Mini Silver', 'COMEX', 'Metals');
    
CREATE TABLE IF NOT EXISTS account_symbols (
    id SERIAL PRIMARY KEY, 
    account_id BIGINT NOT NULL,        
    symbol_id INT NOT NULL,            
    is_active BOOLEAN DEFAULT TRUE,   
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);


CREATE TRIGGER set_account_symbols_updated_at
BEFORE UPDATE ON account_symbols
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();