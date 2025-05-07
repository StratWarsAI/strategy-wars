export interface WebSocketMessage {
    type: string;
    [key: string]: any;
  }
  
  export interface PerformanceUpdateEvent extends WebSocketMessage {
    type: 'performance_update';
    timestamp: number;
    strategy_id: number;
    winRate: number;
    roi: number;
    netPnl: number;
    totalTrades: number;
  }
  
  export interface TradeExecutedEvent extends WebSocketMessage {
    type: 'trade_executed';
    strategy_id: number;
    tokenId: number;
    tokenSymbol: string;
    tokenName: string;
    tokenMint: string;
    imageUrl?: string;
    twitterUrl?: string;
    websiteUrl?: string;
    action: 'buy' | 'sell';
    price: number;
    amount: number;
    timestamp: number;
    currentBalance: number;
    entryMarketCap?: number;
    usdMarketCap?: number;
    signalData?: Record<string, any>;
  }
  
  export interface TradeClosedEvent extends WebSocketMessage {
    type: 'trade_closed';
    strategy_id: number;
    tokenId: number;
    tokenSymbol: string;
    tokenName: string;
    tokenMint: string;
    imageUrl?: string;
    twitterUrl?: string;
    websiteUrl?: string;
    action: string;
    entryPrice: number;
    exitPrice: number;
    exitReason: string;
    profitLoss: number;
    profitLossPct: number;
    timestamp: number;
    entryMarketCap?: number;
    exitMarketCap?: number;
    usdMarketCap?: number;
  }
  
  export interface SimulationStatusEvent extends WebSocketMessage {
    type: 'simulation_status';
    strategy_id: number;
    status: 'starting' | 'running' | 'completed' | 'failed';
    timestamp: number;
    totalTrades: number;
    activeTrades: number;
    profitableTrades: number;
    losingTrades: number;
    winRate: number;
    roi: number;
    currentBalance: number;
    initialBalance: number;
  }
  
  export interface SimulationStartedEvent extends WebSocketMessage {
    type: 'simulation_started';
    strategy_id: number;
    timestamp: number;
  }
  
  export interface SimulationCompletedEvent extends WebSocketMessage {
    type: 'simulation_completed';
    strategy_id: number;
    totalIterations: number;
    executionTimeSec: number;
    timestamp: number;
  }

export interface SimulationBalanceDepletedEvent extends WebSocketMessage {
    type: 'simulation_balance_depleted';
    strategy_id: number;
    timestamp: number;
    remainingBalance: number;
    positionSize: number;
  }

export interface AIAnalysisEvent extends WebSocketMessage {
    type: 'ai_analysis';
    strategy_id: number;
    strategy_name: string;
    timestamp: number;
    analysis: string;
    rating: string;
    roi: number;
    win_rate: number;
    total_trades: number;
    max_drawdown: number;
    net_pnl: number;
    avg_trade_profit: number;
  }