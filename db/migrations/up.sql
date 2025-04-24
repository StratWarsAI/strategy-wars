-- Migration Up Script

-- Create tokens table
CREATE TABLE IF NOT EXISTS tokens (
    id SERIAL PRIMARY KEY,
    mint_address TEXT UNIQUE NOT NULL,
    creator_address TEXT NOT NULL,
    name TEXT,
    symbol TEXT,
    image_url TEXT,
    twitter_url TEXT,
    website_url TEXT,
    telegram_url TEXT,
    metadata_url TEXT,
    king_of_the_hill_timestamp BIGINT,
    completed BOOLEAN DEFAULT FALSE,
    created_timestamp BIGINT NOT NULL,
    market_cap DECIMAL(20, 8),
    usd_market_cap DECIMAL(20, 8),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trades table
CREATE TABLE IF NOT EXISTS trades (
    id SERIAL PRIMARY KEY,
    token_id INTEGER REFERENCES tokens(id),
    signature TEXT UNIQUE NOT NULL,
    sol_amount DECIMAL(20, 9) NOT NULL,
    token_amount DECIMAL(40, 0) NOT NULL,
    is_buy BOOLEAN NOT NULL,
    user_address TEXT NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create strategies table
CREATE TABLE IF NOT EXISTS strategies (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    config JSONB NOT NULL,
    user_id INTEGER,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create simulated_trades table
CREATE TABLE IF NOT EXISTS simulated_trades (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id),
    token_id INTEGER REFERENCES tokens(id),
    entry_price DECIMAL(20, 9) NOT NULL,
    exit_price DECIMAL(20, 9),
    entry_timestamp BIGINT NOT NULL,
    exit_timestamp BIGINT,
    position_size DECIMAL(20, 9) NOT NULL,
    profit_loss DECIMAL(20, 9),
    status TEXT NOT NULL,
    exit_reason TEXT,
    entry_usd_market_cap DECIMAL(20, 9) DEFAULT 0,
    exit_usd_market_cap DECIMAL(20, 9) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add indexes for tokens table
CREATE INDEX idx_tokens_creator_address ON tokens(creator_address);
CREATE INDEX idx_tokens_created_timestamp ON tokens(created_timestamp);
CREATE INDEX idx_tokens_market_cap ON tokens(market_cap);
CREATE INDEX idx_tokens_usd_market_cap ON tokens(usd_market_cap);
CREATE INDEX idx_tokens_completed ON tokens(completed);
CREATE INDEX idx_tokens_king_of_the_hill_timestamp ON tokens(king_of_the_hill_timestamp);

-- Add indexes for trades table
CREATE INDEX idx_trades_token_id ON trades(token_id);
CREATE INDEX idx_trades_user_address ON trades(user_address);
CREATE INDEX idx_trades_timestamp ON trades(timestamp);
CREATE INDEX idx_trades_is_buy ON trades(is_buy);
CREATE INDEX idx_trades_token_id_timestamp ON trades(token_id, timestamp);
CREATE INDEX idx_trades_user_address_timestamp ON trades(user_address, timestamp);

-- Add indexes for strategies table
CREATE INDEX idx_strategies_user_id ON strategies(user_id);
CREATE INDEX idx_strategies_is_public ON strategies(is_public);
CREATE INDEX idx_strategies_name ON strategies(name);

-- Add indexes for simulated_trades table
CREATE INDEX idx_simulated_trades_strategy_id ON simulated_trades(strategy_id);
CREATE INDEX idx_simulated_trades_token_id ON simulated_trades(token_id);
CREATE INDEX idx_simulated_trades_status ON simulated_trades(status);
CREATE INDEX idx_simulated_trades_entry_timestamp ON simulated_trades(entry_timestamp);
CREATE INDEX idx_simulated_trades_exit_timestamp ON simulated_trades(exit_timestamp);
CREATE INDEX idx_simulated_trades_strategy_id_token_id ON simulated_trades(strategy_id, token_id);
CREATE INDEX idx_simulated_trades_profit_loss ON simulated_trades(profit_loss);