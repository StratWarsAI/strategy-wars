-- Migration Down Script

-- Drop Strategy Generations Table Indexes
DROP INDEX IF EXISTS idx_strategy_generations_parent;
DROP INDEX IF EXISTS idx_strategy_generations_child;
DROP INDEX IF EXISTS idx_strategy_generations_number;

-- Drop Simulation Results Table Indexes
DROP INDEX IF EXISTS idx_simulation_results_strategy;
DROP INDEX IF EXISTS idx_simulation_results_simulation;
DROP INDEX IF EXISTS idx_simulation_results_roi;
DROP INDEX IF EXISTS idx_simulation_results_performance;
DROP INDEX IF EXISTS idx_simulation_results_rank;

-- Drop Strategy Metrics Table Indexes
DROP INDEX IF EXISTS idx_strategy_metrics_strategy;
DROP INDEX IF EXISTS idx_strategy_metrics_simulation;
DROP INDEX IF EXISTS idx_strategy_metrics_win_rate;

-- Drop Simulated Trades Table Indexes
DROP INDEX IF EXISTS idx_simulated_trades_strategy_token;
DROP INDEX IF EXISTS idx_simulated_trades_simulation;
DROP INDEX IF EXISTS idx_simulated_trades_status;
DROP INDEX IF EXISTS idx_simulated_trades_timestamps;
DROP INDEX IF EXISTS idx_simulated_trades_profit_loss;

-- Drop Simulation Events Table Indexes
DROP INDEX IF EXISTS idx_simulation_events_strategy;
DROP INDEX IF EXISTS idx_simulation_events_simulation;
DROP INDEX IF EXISTS idx_simulation_events_type;
DROP INDEX IF EXISTS idx_simulation_events_timestamp;

-- Drop Trades Table Indexes
DROP INDEX IF EXISTS idx_trades_token_timestamp;
DROP INDEX IF EXISTS idx_trades_signature;
DROP INDEX IF EXISTS idx_trades_is_buy;

-- Drop Tokens Table Indexes
DROP INDEX IF EXISTS idx_tokens_mint_address;
DROP INDEX IF EXISTS idx_tokens_creator_timestamp;
DROP INDEX IF EXISTS idx_tokens_market_cap;
DROP INDEX IF EXISTS idx_tokens_usd_market_cap;

-- Drop Simulation Runs Table Indexes
DROP INDEX IF EXISTS idx_simulation_runs_status;
DROP INDEX IF EXISTS idx_simulation_runs_time_range;
DROP INDEX IF EXISTS idx_simulation_runs_winner;

-- Drop Strategies Table Indexes
DROP INDEX IF EXISTS idx_strategies_name;
DROP INDEX IF EXISTS idx_strategies_ai_enhanced;
DROP INDEX IF EXISTS idx_strategies_complexity;
DROP INDEX IF EXISTS idx_strategies_risk;

-- Drop tables (in reverse order of creation to handle dependencies)
DROP TABLE IF EXISTS strategy_generations;
DROP TABLE IF EXISTS simulation_results;
DROP TABLE IF EXISTS strategy_metrics;
DROP TABLE IF EXISTS simulation_events;
DROP TABLE IF EXISTS simulated_trades;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS simulation_runs;
DROP TABLE IF EXISTS strategies;