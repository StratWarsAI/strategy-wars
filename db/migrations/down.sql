-- Migration Down Script

-- Drop Users Table Indexes
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_wallet_address;

-- Drop Strategies Table Indexes
DROP INDEX IF EXISTS idx_strategies_user_public;
DROP INDEX IF EXISTS idx_strategies_name;

-- Drop Duels Table Indexes
DROP INDEX IF EXISTS idx_duels_status_time;

-- Drop Votes Table Indexes
DROP INDEX IF EXISTS idx_votes_duel_strategy;

-- Drop User Scores Table Indexes
DROP INDEX IF EXISTS idx_user_scores_points_wins;

-- Drop Comments Table Indexes
DROP INDEX IF EXISTS idx_comments_strategy_user;

-- Drop Notifications Table Indexes
DROP INDEX IF EXISTS idx_notifications_user_type;

-- Drop Strategy Metrics Table Indexes
DROP INDEX IF EXISTS idx_strategy_metrics_duel;

-- Drop Tokens Table Indexes
DROP INDEX IF EXISTS idx_tokens_creator_timestamp;
DROP INDEX IF EXISTS idx_tokens_market_cap;

-- Drop Trades Table Indexes
DROP INDEX IF EXISTS idx_trades_token_timestamp;
DROP INDEX IF EXISTS idx_trades_user_timestamp;

-- Drop Simulated Trades Table Indexes
DROP INDEX IF EXISTS idx_simulated_trades_strategy_token;
DROP INDEX IF EXISTS idx_simulated_trades_timestamps;

-- Drop tables (in reverse order of creation to handle dependencies)
DROP TABLE IF EXISTS simulated_trades;
DROP TABLE IF EXISTS strategy_metrics;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS user_scores;
DROP TABLE IF EXISTS duels;
DROP TABLE IF EXISTS strategies;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS users;