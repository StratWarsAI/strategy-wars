-- Migration Up Script


-- Create strategies table
CREATE TABLE IF NOT EXISTS strategies (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    config JSONB NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    vote_count INTEGER DEFAULT 0,
    win_count INTEGER DEFAULT 0,
    last_win_time TIMESTAMP WITH TIME ZONE,
    tags TEXT[],
    complexity_score INTEGER DEFAULT 5,
    risk_score INTEGER DEFAULT 5,
    ai_enhanced BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create simulation_runs table
CREATE TABLE IF NOT EXISTS simulation_runs (
    id SERIAL PRIMARY KEY,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    winner_strategy_id INTEGER REFERENCES strategies(id),
    status TEXT NOT NULL, -- 'preparing', 'running', 'completed', 'failed'
    simulation_parameters JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

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

-- Create simulated_trades table
CREATE TABLE IF NOT EXISTS simulated_trades (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id),
    token_id INTEGER REFERENCES tokens(id),
    simulation_run_id INTEGER REFERENCES simulation_runs(id),
    entry_price DECIMAL(20, 9) NOT NULL,
    exit_price DECIMAL(20, 9),
    entry_timestamp BIGINT NOT NULL,
    exit_timestamp BIGINT,
    position_size DECIMAL(20, 9) NOT NULL,
    profit_loss DECIMAL(20, 9),
    status TEXT NOT NULL, -- 'open', 'closed', 'canceled'
    exit_reason TEXT,
    entry_usd_market_cap DECIMAL(20, 9) DEFAULT 0,
    exit_usd_market_cap DECIMAL(20, 9) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create strategy_metrics table (keep as is, but update reference)
CREATE TABLE IF NOT EXISTS strategy_metrics (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id) NOT NULL,
    simulation_run_id INTEGER REFERENCES simulation_runs(id),
    win_rate DECIMAL(5, 2),
    avg_profit DECIMAL(10, 2),
    avg_loss DECIMAL(10, 2),
    max_drawdown DECIMAL(10, 2),
    total_trades INTEGER,
    successful_trades INTEGER,
    risk_score INTEGER,
    roi DECIMAL(10, 4),
    current_balance DECIMAL(20, 9),
    initial_balance DECIMAL(20, 9),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create simulation_results table
CREATE TABLE IF NOT EXISTS simulation_results (
    id SERIAL PRIMARY KEY,
    simulation_run_id INTEGER REFERENCES simulation_runs(id),
    strategy_id INTEGER REFERENCES strategies(id),
    roi DECIMAL(10, 4),
    trade_count INTEGER,
    win_rate DECIMAL(5, 4),
    max_drawdown DECIMAL(10, 4),
    performance_rating TEXT, -- 'excellent', 'good', 'average', 'poor', 'very_poor'
    analysis TEXT,
    rank INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create simulation_events table
CREATE TABLE IF NOT EXISTS simulation_events (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER NOT NULL REFERENCES strategies(id),
    simulation_run_id INTEGER NOT NULL REFERENCES simulation_runs(id),
    event_type TEXT NOT NULL,
    event_data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create strategy_generations table
CREATE TABLE IF NOT EXISTS strategy_generations (
    id SERIAL PRIMARY KEY,
    generation_number INTEGER,
    parent_strategy_id INTEGER REFERENCES strategies(id),
    child_strategy_id INTEGER REFERENCES strategies(id),
    improvement_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Strategies Table Indexes
CREATE INDEX IF NOT EXISTS idx_strategies_name ON strategies(name);
CREATE INDEX IF NOT EXISTS idx_strategies_ai_enhanced ON strategies(ai_enhanced);
CREATE INDEX IF NOT EXISTS idx_strategies_complexity ON strategies(complexity_score);
CREATE INDEX IF NOT EXISTS idx_strategies_risk ON strategies(risk_score);

-- Simulation Runs Table Indexes
CREATE INDEX IF NOT EXISTS idx_simulation_runs_status ON simulation_runs(status);
CREATE INDEX IF NOT EXISTS idx_simulation_runs_time_range ON simulation_runs(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_simulation_runs_winner ON simulation_runs(winner_strategy_id);

-- Tokens Table Indexes
CREATE INDEX IF NOT EXISTS idx_tokens_mint_address ON tokens(mint_address);
CREATE INDEX IF NOT EXISTS idx_tokens_creator_timestamp ON tokens(creator_address, created_timestamp);
CREATE INDEX IF NOT EXISTS idx_tokens_market_cap ON tokens(market_cap);
CREATE INDEX IF NOT EXISTS idx_tokens_usd_market_cap ON tokens(usd_market_cap);

-- Trades Table Indexes
CREATE INDEX IF NOT EXISTS idx_trades_token_timestamp ON trades(token_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_trades_signature ON trades(signature);
CREATE INDEX IF NOT EXISTS idx_trades_is_buy ON trades(is_buy);

-- Simulated Trades Table Indexes
CREATE INDEX IF NOT EXISTS idx_simulated_trades_strategy_token ON simulated_trades(strategy_id, token_id);
CREATE INDEX IF NOT EXISTS idx_simulated_trades_simulation ON simulated_trades(simulation_run_id);
CREATE INDEX IF NOT EXISTS idx_simulated_trades_status ON simulated_trades(status);
CREATE INDEX IF NOT EXISTS idx_simulated_trades_timestamps ON simulated_trades(entry_timestamp, exit_timestamp);
CREATE INDEX IF NOT EXISTS idx_simulated_trades_profit_loss ON simulated_trades(profit_loss);

-- Strategy Metrics Table Indexes
CREATE INDEX IF NOT EXISTS idx_strategy_metrics_strategy ON strategy_metrics(strategy_id);
CREATE INDEX IF NOT EXISTS idx_strategy_metrics_simulation ON strategy_metrics(simulation_run_id);
CREATE INDEX IF NOT EXISTS idx_strategy_metrics_win_rate ON strategy_metrics(win_rate);

-- Simulation Results Table Indexes
CREATE INDEX IF NOT EXISTS idx_simulation_results_strategy ON simulation_results(strategy_id);
CREATE INDEX IF NOT EXISTS idx_simulation_results_simulation ON simulation_results(simulation_run_id);
CREATE INDEX IF NOT EXISTS idx_simulation_results_roi ON simulation_results(roi);
CREATE INDEX IF NOT EXISTS idx_simulation_results_performance ON simulation_results(performance_rating);
CREATE INDEX IF NOT EXISTS idx_simulation_results_rank ON simulation_results(simulation_run_id, rank);

-- Simulation Events Table Indexes
CREATE INDEX IF NOT EXISTS idx_simulation_events_strategy ON simulation_events(strategy_id);
CREATE INDEX IF NOT EXISTS idx_simulation_events_simulation ON simulation_events(simulation_run_id);
CREATE INDEX IF NOT EXISTS idx_simulation_events_type ON simulation_events(event_type);
CREATE INDEX IF NOT EXISTS idx_simulation_events_timestamp ON simulation_events(timestamp);

-- Strategy Generations Table Indexes
CREATE INDEX IF NOT EXISTS idx_strategy_generations_parent ON strategy_generations(parent_strategy_id);
CREATE INDEX IF NOT EXISTS idx_strategy_generations_child ON strategy_generations(child_strategy_id);
CREATE INDEX IF NOT EXISTS idx_strategy_generations_number ON strategy_generations(generation_number);
