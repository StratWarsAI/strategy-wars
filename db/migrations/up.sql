-- Migration Up Script


-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    wallet_address TEXT UNIQUE,
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create strategies table
CREATE TABLE IF NOT EXISTS strategies (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    config JSONB NOT NULL,
    user_id INTEGER,
    is_public BOOLEAN DEFAULT FALSE,
    vote_count INTEGER DEFAULT 0,
    win_count INTEGER DEFAULT 0,
    last_win_time TIMESTAMP WITH TIME ZONE,
    tags TEXT[],
    complexity_score INTEGER DEFAULT 5,
    risk_score INTEGER DEFAULT 5,
    ai_enhanced BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create duels table (for 10-minute battle periods)
CREATE TABLE IF NOT EXISTS duels (
    id SERIAL PRIMARY KEY,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    voting_end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    winner_strategy_id INTEGER REFERENCES strategies(id),
    status TEXT NOT NULL, -- 'voting', 'simulating', 'completed'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create votes table (for strategy voting)
CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    duel_id INTEGER REFERENCES duels(id) NOT NULL,
    strategy_id INTEGER REFERENCES strategies(id) NOT NULL,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(duel_id, user_id) -- Each user can only vote once per duel
);

-- Create user_scores table (for leaderboard)
CREATE TABLE IF NOT EXISTS user_scores (
    user_id INTEGER REFERENCES users(id) PRIMARY KEY,
    total_points INTEGER DEFAULT 0,
    win_count INTEGER DEFAULT 0,
    strategy_count INTEGER DEFAULT 0,
    vote_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create comments table (for strategy discussions)
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id) NOT NULL,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    parent_id INTEGER REFERENCES comments(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    type TEXT NOT NULL, -- 'strategy_win', 'vote', 'comment', etc.
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    related_id INTEGER, -- Could be a strategy_id, duel_id, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create strategy_metrics table (for AI analysis)
CREATE TABLE IF NOT EXISTS strategy_metrics (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id) NOT NULL,
    duel_id INTEGER REFERENCES duels(id),
    win_rate DECIMAL(5, 2),
    avg_profit DECIMAL(10, 2),
    avg_loss DECIMAL(10, 2),
    max_drawdown DECIMAL(10, 2),
    total_trades INTEGER,
    successful_trades INTEGER,
    risk_score INTEGER, -- 1-10 score calculated by AI
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
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
    duel_id INTEGER REFERENCES duels(id),
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

-- Users Table
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_wallet_address ON users(wallet_address);

-- Strategies Table (Combined Index)
CREATE INDEX IF NOT EXISTS idx_strategies_user_public ON strategies(user_id, is_public);
CREATE INDEX IF NOT EXISTS idx_strategies_name ON strategies(name);

-- Duels Table
CREATE INDEX IF NOT EXISTS idx_duels_status_time ON duels(status, start_time);

-- Votes Table
CREATE INDEX IF NOT EXISTS idx_votes_duel_strategy ON votes(duel_id, strategy_id);

-- User Scores Table
CREATE INDEX IF NOT EXISTS idx_user_scores_points_wins ON user_scores(total_points, win_count);

-- Comments Table
CREATE INDEX IF NOT EXISTS idx_comments_strategy_user ON comments(strategy_id, user_id);

-- Notifications Table
CREATE INDEX IF NOT EXISTS idx_notifications_user_type ON notifications(user_id, type);

-- Strategy Metrics Table
CREATE INDEX IF NOT EXISTS idx_strategy_metrics_duel ON strategy_metrics(duel_id, strategy_id);

-- Tokens Table
CREATE INDEX IF NOT EXISTS idx_tokens_creator_timestamp ON tokens(creator_address, created_timestamp);
CREATE INDEX IF NOT EXISTS idx_tokens_market_cap ON tokens(market_cap);

-- Trades Table
CREATE INDEX IF NOT EXISTS idx_trades_token_timestamp ON trades(token_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_trades_user_timestamp ON trades(user_address, timestamp);

-- Simulated Trades Table
CREATE INDEX IF NOT EXISTS idx_simulated_trades_strategy_token ON simulated_trades(strategy_id, token_id);
CREATE INDEX IF NOT EXISTS idx_simulated_trades_timestamps ON simulated_trades(entry_timestamp, exit_timestamp);