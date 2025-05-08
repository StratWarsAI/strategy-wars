export interface TopPerformer {
  id: number;
  name: string;
  roi: number;
  trades: number;
}

export interface DashboardStats {
  total_balance: number;
  balance_change: number;
  balance_change_percent: number;
  total_profits: number;
  total_trades: number;
  winning_trades: number;
  losing_trades: number;
  win_rate: number;
  active_trade_count: number;
  avg_hold_time: string;
  top_performer: TopPerformer;
  market_status: string;
  volatility_index: number;
  last_updated: string;
}

export interface PerformanceDataPoint {
  date: string;
  balance: number;
}

export interface StrategyDistributionItem {
  id: number;
  name: string;
  trades: number;
  profit: number;
  color: string;
}

export interface RecentPerformanceStats {
  period: string;
  trades: number;
  profit: number;
  win_rate: number;
  best_trade: string;
}

export interface DashboardCharts {
  performance_data: PerformanceDataPoint[];
  strategy_data: StrategyDistributionItem[];
  recent_stats: RecentPerformanceStats[];
}

export interface CompleteDashboard {
  stats: DashboardStats;
  charts: DashboardCharts;
}