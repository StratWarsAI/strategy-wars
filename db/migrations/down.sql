-- Migration Down Script

-- Drop indexes for simulated_trades table
DROP INDEX IF EXISTS idx_simulated_trades_profit_loss;
DROP INDEX IF EXISTS idx_simulated_trades_strategy_id_token_id;
DROP INDEX IF EXISTS idx_simulated_trades_exit_timestamp;
DROP INDEX IF EXISTS idx_simulated_trades_entry_timestamp;
DROP INDEX IF EXISTS idx_simulated_trades_status;
DROP INDEX IF EXISTS idx_simulated_trades_token_id;
DROP INDEX IF EXISTS idx_simulated_trades_strategy_id;

-- Drop indexes for strategies table
DROP INDEX IF EXISTS idx_strategies_name;
DROP INDEX IF EXISTS idx_strategies_is_public;
DROP INDEX IF EXISTS idx_strategies_user_id;

-- Drop indexes for trades table
DROP INDEX IF EXISTS idx_trades_user_address_timestamp;
DROP INDEX IF EXISTS idx_trades_token_id_timestamp;
DROP INDEX IF EXISTS idx_trades_is_buy;
DROP INDEX IF EXISTS idx_trades_timestamp;
DROP INDEX IF EXISTS idx_trades_user_address;
DROP INDEX IF EXISTS idx_trades_token_id;

-- Drop indexes for tokens table
DROP INDEX IF EXISTS idx_tokens_king_of_the_hill_timestamp;
DROP INDEX IF EXISTS idx_tokens_completed;
DROP INDEX IF EXISTS idx_tokens_usd_market_cap;
DROP INDEX IF EXISTS idx_tokens_market_cap;
DROP INDEX IF EXISTS idx_tokens_created_timestamp;
DROP INDEX IF EXISTS idx_tokens_creator_address;

-- Drop tables (in reverse order of creation to handle dependencies)
DROP TABLE IF EXISTS simulated_trades;
DROP TABLE IF EXISTS strategies;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS tokens;