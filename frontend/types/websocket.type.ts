export interface WebSocketMessage {
    type: string;
    [key: string]: any;
  }
  
  export interface PerformanceUpdateEvent extends WebSocketMessage {
    type: 'performance_update';
    timestamp: number;
    strategy_id: number;
    win_rate: number;
    roi: number;
    net_pnl: number;
    total_trades: number;
  }
  
  export interface TradeExecutedEvent extends WebSocketMessage {
    type: 'trade_executed';
    token_id: number;
    token_symbol: string;
    token_name: string;
    token_mint: string;
    image_url?: string;
    action: 'buy' | 'sell';
    price: number;
    amount: number;
    timestamp: number;
    current_balance: number;
  }
  
  export interface TradeClosedEvent extends WebSocketMessage {
    type: 'trade_closed';
    token_id: number;
    token_symbol: string;
    token_name: string;
    entry_price: number;
    exit_price: number;
    exit_reason: string;
    profit_loss: number;
    profit_loss_pct: number;
    timestamp: number;
  }
  
  export interface SimulationStatusEvent extends WebSocketMessage {
    type: 'simulation_status';
    strategy_id: number;
    status: 'starting' | 'running' | 'completed' | 'failed';
    timestamp: number;
  }
  
  export interface SimulationStartedEvent extends WebSocketMessage {
    type: 'simulation_started';
    strategy_id: number;
    timestamp: number;
  }
  
  export interface SimulationCompletedEvent extends WebSocketMessage {
    type: 'simulation_completed';
    strategy_id: number;
    total_iterations: number;
    execution_time_sec: number;
    timestamp: number;
  }