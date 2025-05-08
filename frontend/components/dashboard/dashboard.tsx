'use client';

import { useState, useEffect } from 'react';
import { useRouter, usePathname, useSearchParams } from 'next/navigation';
import { 
  LineChart, BarChart, ArrowUp, ArrowDown, ArrowRight,
  Activity, Wallet, BarChart2, Percent, CheckCircle, XCircle, Link, Clock
} from 'lucide-react';
import { DashboardHeader } from '@/components/dashboard/dashboard-header';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { getCompleteDashboard } from '@/lib/api';
import { CompleteDashboard } from '@/types/dashboard.type';
import { useWebSocket } from '@/lib/context/web-socket-context';
import { useEffect as useReactEffect } from 'react';
import { AnimatedValue, AnimatedCurrency, AnimatedPercentage } from '@/components/ui/animated-value';
import { motion, AnimatePresence } from 'framer-motion';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend
} from 'recharts';
import { 
  PerformanceUpdateEvent, 
  TradeExecutedEvent, 
  TradeClosedEvent, 
  SimulationStatusEvent 
} from '@/types/websocket.type';

interface DashboardProps {
  initialData: CompleteDashboard;
  defaultTimeframe: string;
}

export default function Dashboard({ initialData, defaultTimeframe }: DashboardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { isConnected, lastMessage } = useWebSocket();
  
  const [selectedTimeframe, setSelectedTimeframe] = useState(defaultTimeframe);
  const [isLoading, setIsLoading] = useState(false);
  const [dashboardData, setDashboardData] = useState<CompleteDashboard>(initialData);
  const [hasLiveUpdates, setHasLiveUpdates] = useState(false);
  
  // Extract data from the dashboard response for easier use in the UI
  const statsData = {
    totalBalance: dashboardData.stats.total_balance,
    balanceChange: dashboardData.stats.balance_change,
    balanceChangePercent: dashboardData.stats.balance_change_percent,
    totalProfits: dashboardData.stats.total_profits,
    totalTrades: dashboardData.stats.total_trades,
    winningTrades: dashboardData.stats.winning_trades,
    losingTrades: dashboardData.stats.losing_trades,
    winRate: dashboardData.stats.win_rate,
    activeTradeCount: dashboardData.stats.active_trade_count,
    avgHoldTime: dashboardData.stats.avg_hold_time,
    topPerformer: {
      id: dashboardData.stats.top_performer.id,
      name: dashboardData.stats.top_performer.name,
      roi: dashboardData.stats.top_performer.roi,
      trades: dashboardData.stats.top_performer.trades
    },
    marketStatus: dashboardData.stats.market_status,
    volatilityIndex: dashboardData.stats.volatility_index,
    lastUpdated: new Date(dashboardData.stats.last_updated)
  };
  
  const performanceData = dashboardData.charts.performance_data;
  const strategyData = dashboardData.charts.strategy_data;
  const recentStats = dashboardData.charts.recent_stats;
  
  // Fetch dashboard data when timeframe changes
  const fetchDashboardData = async () => {
    setIsLoading(true);
    try {
      const data = await getCompleteDashboard(selectedTimeframe);
      setDashboardData(data);
    } catch (error) {
      console.error("Error fetching dashboard data:", error);
    } finally {
      setIsLoading(false);
    }
  };
  
  // Update URL when timeframe changes
  useEffect(() => {
    const newParams = new URLSearchParams(searchParams);
    newParams.set('timeframe', selectedTimeframe);
    router.push(`${pathname}?${newParams.toString()}`);
    
    // Fetch new data when timeframe changes
    fetchDashboardData();
  }, [selectedTimeframe]);
  
  // State to track per-strategy metrics
  const [strategyMetrics, setStrategyMetrics] = useState<{[strategyId: string]: any}>({});
  
  // WebSocket updates handling
  useEffect(() => {
    if (!lastMessage) return;
    
    setHasLiveUpdates(true);
    
    switch (lastMessage.type) {
      case 'trade_executed': {
        // Update active trade count when a new trade is executed
        const tradeEvent = lastMessage as TradeExecutedEvent;
        
        setDashboardData(prev => {
          if (!prev) return prev;
          
          // Update dashboard with new trade info
          return {
            ...prev,
            stats: {
              ...prev.stats,
              active_trade_count: prev.stats.active_trade_count + 1,
              total_trades: prev.stats.total_trades + 1,
              // Don't override total_balance from a single trade event
              last_updated: new Date().toISOString()
            }
          };
        });
        break;
      }
      
      case 'trade_closed': {
        // Update stats when a trade is closed
        const closeEvent = lastMessage as TradeClosedEvent;
        
        setDashboardData(prev => {
          if (!prev) return prev;
          
          // Calculate new win/loss stats
          const isWin = closeEvent.profitLoss > 0;
          const newWinningTrades = isWin ? prev.stats.winning_trades + 1 : prev.stats.winning_trades;
          const newLosingTrades = !isWin ? prev.stats.losing_trades + 1 : prev.stats.losing_trades;
          const newWinRate = ((newWinningTrades / (prev.stats.total_trades)) * 100);
          
          // Update dashboard with trade results
          return {
            ...prev,
            stats: {
              ...prev.stats,
              active_trade_count: Math.max(0, prev.stats.active_trade_count - 1),
              winning_trades: newWinningTrades,
              losing_trades: newLosingTrades,
              win_rate: newWinRate,
              total_profits: isWin ? prev.stats.total_profits + closeEvent.profitLoss : prev.stats.total_profits,
              last_updated: new Date().toISOString()
            }
          };
        });
        break;
      }
      
      case 'performance_update': {
        // Update overall performance metrics
        const perfEvent = lastMessage as PerformanceUpdateEvent;
        
        setDashboardData(prev => {
          if (!prev) return prev;
          
          return {
            ...prev,
            stats: {
              ...prev.stats,
              win_rate: perfEvent.winRate,
              balance_change: perfEvent.netPnl,
              total_trades: perfEvent.totalTrades,
              last_updated: new Date().toISOString()
            }
          };
        });
        break;
      }
      
      case 'simulation_status': {
        // Update status based on simulation status events
        const simEvent = lastMessage as SimulationStatusEvent;
        
        // First, update the per-strategy metrics
        setStrategyMetrics(prev => {
          // Store this strategy's metrics
          return {
            ...prev,
            [simEvent.strategy_id.toString()]: {
              activeTrades: simEvent.activeTrades,
              totalTrades: simEvent.totalTrades,
              profitableTrades: simEvent.profitableTrades,
              losingTrades: simEvent.losingTrades,
              winRate: simEvent.winRate,
              currentBalance: simEvent.currentBalance
            }
          };
        });
        
        // Then update the dashboard based on aggregated metrics
        setDashboardData(prev => {
          if (!prev) return prev;
          
          // Don't directly use the event data to update total_trades/active_trade_count
          // Instead, only refresh the dashboard when explicitly requested
          // This prevents a single strategy's simulation_status from overriding the aggregate counts
          return {
            ...prev,
            stats: {
              ...prev.stats,
              // Refresh metrics to indicate we received updates
              last_updated: new Date().toISOString()
            }
          };
        });
        
        // Refresh the dashboard data to get accurate aggregate counts
        fetchDashboardData();
        break;
      }
    }
    
    // Clear live update indicator after 2 seconds
    const timer = setTimeout(() => setHasLiveUpdates(false), 2000);
    return () => clearTimeout(timer);
  }, [lastMessage]);
  
  // Format currency values
  const formatCurrency = (value: number) => {
    return `${value.toFixed(2)} SOL`;
  };
  
  // Format percentage values
  const formatPercent = (value: number) => {
    return `${value > 0 ? '+' : ''}${value.toFixed(2)}%`;
  };
  
  // Format time
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };
  
  // Refresh data function
  const refreshData = () => {
    fetchDashboardData();
  };
  
  return (
    <div className="space-y-6">
      {/* Dashboard Header with timeframe controls */}
      <DashboardHeader
        selectedTimeframe={selectedTimeframe}
        onTimeframeChange={setSelectedTimeframe}
        lastUpdated={statsData.lastUpdated}
        isConnected={isConnected}
        isLoading={isLoading}
        hasLiveUpdates={hasLiveUpdates}
        onRefresh={refreshData}
      />

      {/* Main Stats Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4">
        <Card className="overflow-hidden">
          <CardHeader className="pb-2">
            <CardDescription>Total Balance</CardDescription>
            <CardTitle className="text-2xl">
              <AnimatedCurrency value={statsData.totalBalance} />
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className={`flex items-center gap-1 text-sm ${statsData.balanceChange >= 0 ? 'text-green-500' : 'text-red-500'}`}>
              <motion.div 
                initial={{ scale: 1 }}
                animate={{ scale: hasLiveUpdates ? [1, 1.2, 1] : 1 }}
                transition={{ duration: 0.3 }}
              >
                {statsData.balanceChange >= 0 ? (
                  <ArrowUp className="h-4 w-4" />
                ) : (
                  <ArrowDown className="h-4 w-4" />
                )}
              </motion.div>
              <AnimatedCurrency value={Math.abs(statsData.balanceChange)} />
              <span>(</span>
              <AnimatedPercentage value={statsData.balanceChangePercent} />
              <span>)</span>
            </div>
          </CardContent>
          <motion.div 
            className={`h-1.5 w-full ${statsData.balanceChange >= 0 ? 'bg-green-500' : 'bg-red-500'}`}
            animate={{ 
              opacity: hasLiveUpdates ? [0.6, 1, 0.6] : 0.6,
            }}
            transition={{ duration: 1, repeat: hasLiveUpdates ? 1 : 0 }}
          ></motion.div>
        </Card>
        
        <Card className="overflow-hidden">
          <CardHeader className="pb-2">
            <CardDescription>Win Rate</CardDescription>
            <CardTitle className="text-2xl">
              <AnimatedPercentage value={statsData.winRate} showSign={false} />
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center gap-1">
                <motion.div
                  animate={{ scale: hasLiveUpdates && statsData.winningTrades > 0 ? [1, 1.2, 1] : 1 }}
                  transition={{ duration: 0.3 }}
                >
                  <CheckCircle className="h-4 w-4 text-green-500" /> 
                </motion.div>
                <AnimatedValue 
                  value={statsData.winningTrades} 
                  precision={0} 
                  suffix=" wins" 
                />
              </div>
              <div className="flex items-center gap-1">
                <XCircle className="h-4 w-4 text-red-500" /> 
                <AnimatedValue 
                  value={statsData.losingTrades} 
                  precision={0} 
                  suffix=" losses" 
                />
              </div>
            </div>
            <motion.div
              className="mt-2 h-1.5 w-full bg-secondary rounded-full overflow-hidden"
              animate={{ opacity: hasLiveUpdates ? [0.7, 1, 0.7] : 0.7 }}
              transition={{ duration: 0.8 }}
            >
              <motion.div 
                className="h-full bg-primary rounded-full"
                style={{ width: `${statsData.winRate}%` }}
                animate={{ 
                  width: `${statsData.winRate}%`,
                  opacity: hasLiveUpdates ? [0.8, 1, 0.8] : 0.8 
                }}
                transition={{ duration: 0.8 }}
              />
            </motion.div>
          </CardContent>
        </Card>
        
        <Card className="overflow-hidden">
          <CardHeader className="pb-2">
            <CardDescription>Trading Activity</CardDescription>
            <CardTitle className="text-2xl flex items-center gap-2">
              <motion.div
                animate={hasLiveUpdates ? { scale: [1, 1.05, 1] } : { scale: 1 }}
                transition={{ duration: 0.5 }}
              >
                <AnimatedValue 
                  value={statsData.totalTrades} 
                  precision={0}
                  duration={0.8}
                />
              </motion.div>
              <Badge variant="outline" className="text-xs font-normal">Total</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center gap-1">
                <motion.div
                  animate={{ 
                    rotate: hasLiveUpdates ? [0, 15, -15, 0] : 0,
                    scale: hasLiveUpdates ? [1, 1.1, 1] : 1 
                  }}
                  transition={{ duration: 0.5 }}
                >
                  <Activity className="h-4 w-4 text-primary" />
                </motion.div>
                <motion.div
                  animate={
                    hasLiveUpdates ? 
                    { backgroundColor: ['rgba(0,0,0,0)', 'rgba(var(--primary-rgb), 0.1)', 'rgba(0,0,0,0)'] } : 
                    { backgroundColor: 'rgba(0,0,0,0)' }
                  }
                  transition={{ duration: 0.8 }}
                  className="px-1 rounded"
                >
                  <AnimatedValue 
                    value={statsData.activeTradeCount} 
                    precision={0} 
                    suffix=" active trades"
                    duration={0.8}
                  />
                </motion.div>
              </div>
              <div className="flex items-center gap-1 text-muted-foreground">
                <Clock className="h-4 w-4" />
                <span>Avg: {statsData.avgHoldTime}</span>
              </div>
            </div>
          </CardContent>
          <motion.div 
            className="bg-primary h-1.5 w-full"
            animate={{ 
              opacity: hasLiveUpdates ? [0.7, 1, 0.7] : 0.7,
            }}
            transition={{ duration: 0.8 }}
          ></motion.div>
        </Card>
        
        <Card className="overflow-hidden">
          <CardHeader className="pb-2">
            <CardDescription>Top Strategy</CardDescription>
            <CardTitle className="text-xl line-clamp-1">
              <AnimatePresence mode="wait">
                <motion.span
                  key={statsData.topPerformer.id}
                  initial={{ opacity: 0, y: 5 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -5 }}
                  transition={{ duration: 0.3 }}
                >
                  {statsData.topPerformer.id === 0 ? 
                    "No strategies" : 
                    statsData.topPerformer.name}
                </motion.span>
              </AnimatePresence>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center gap-1">
                <motion.div
                  animate={{ 
                    rotate: hasLiveUpdates && statsData.topPerformer.id !== 0 ? [0, 10, -10, 0] : 0,
                  }}
                  transition={{ duration: 0.5 }}
                >
                  <Percent className="h-4 w-4 text-primary" /> 
                </motion.div>
                {statsData.topPerformer.id === 0 ? (
                  <span>N/A</span>
                ) : (
                  <AnimatedPercentage value={statsData.topPerformer.roi} />
                )}
              </div>
              <div className="flex items-center gap-1 text-muted-foreground">
                <BarChart className="h-4 w-4" />
                <AnimatedValue 
                  value={statsData.topPerformer.trades} 
                  precision={0} 
                  suffix=" trades" 
                />
              </div>
            </div>
          </CardContent>
          <motion.div 
            className="bg-primary h-1.5 w-full"
            animate={{ 
              opacity: hasLiveUpdates ? [0.7, 1, 0.7] : 0.7,
            }}
            transition={{ duration: 0.8 }}
          ></motion.div>
        </Card>
      </div>
      
      {/* Stats Charts */}
      <motion.div 
        className="grid grid-cols-1 gap-4 md:grid-cols-3"
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ 
          duration: 0.6,
          delay: 0.2,
          ease: "easeOut"
        }}
      >
        <Card className="md:col-span-2 overflow-hidden">
          <CardHeader>
            <CardTitle>Balance History</CardTitle>
            <CardDescription>Portfolio performance over time</CardDescription>
          </CardHeader>
          <CardContent className="h-[250px] sm:h-[300px]">
            {performanceData.length > 1 ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.8, delay: 0.3 }}
                className="w-full h-full"
              >
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart
                    data={performanceData}
                    margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                    <XAxis dataKey="date" />
                    <YAxis />
                    <Tooltip 
                      formatter={(value) => [`${Number(value).toFixed(2)} SOL`, 'Balance']}
                      labelFormatter={(label) => `Date: ${label}`}
                    />
                    <Area 
                      type="monotone" 
                      dataKey="balance" 
                      stroke="var(--primary)" 
                      fill="var(--primary)" 
                      fillOpacity={0.3} 
                      animationDuration={1500}
                      animationBegin={300}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </motion.div>
            ) : (
              <motion.div 
                className="flex h-full w-full items-center justify-center rounded-md border border-dashed border-muted-foreground/50 p-8 text-center"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.8, delay: 0.3 }}
              >
                <div className="space-y-2">
                  <motion.div
                    initial={{ scale: 0.9, opacity: 0 }}
                    animate={{ scale: 1, opacity: 0.6 }}
                    transition={{ duration: 0.5, delay: 0.5 }}
                  >
                    <BarChart2 className="mx-auto h-10 w-10 text-muted-foreground/60" />
                  </motion.div>
                  <p className="text-sm text-muted-foreground">No performance data available</p>
                  <p className="text-xs text-muted-foreground/60">
                    Starting: 100 SOL → Current: {formatCurrency(statsData.totalBalance)}
                  </p>
                </div>
              </motion.div>
            )}
          </CardContent>
          <motion.div 
            className="bg-primary/20 h-1 w-full"
            initial={{ scaleX: 0 }}
            animate={{ scaleX: 1 }}
            transition={{ duration: 0.8, delay: 0.4 }}
          />
        </Card>
        
        <Card className="overflow-hidden">
          <CardHeader>
            <CardTitle>Strategy Distribution</CardTitle>
            <CardDescription>Profits by strategy</CardDescription>
          </CardHeader>
          <CardContent>
            {strategyData.length > 0 && strategyData[0].id !== 0 ? (
              <div className="space-y-4">
                {strategyData.map((strategy, index) => (
                  <motion.div 
                    key={strategy.id || index} 
                    className="space-y-1"
                    initial={{ opacity: 0, x: -5 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ 
                      duration: 0.4, 
                      delay: 0.2 + (index * 0.1),
                      ease: "easeOut"
                    }}
                  >
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium">{strategy.name}</span>
                      <AnimatedCurrency value={strategy.profit} />
                    </div>
                    <motion.div 
                      className="h-2 w-full rounded-full bg-secondary overflow-hidden"
                      animate={{ opacity: hasLiveUpdates ? [0.7, 1, 0.7] : 0.7 }}
                      transition={{ duration: 0.8 }}
                    >
                      <motion.div 
                        className="h-full rounded-full" 
                        style={{ backgroundColor: strategy.color }}
                        initial={{ width: 0 }}
                        animate={{ 
                          width: `${statsData.totalProfits > 0 ? (strategy.profit / statsData.totalProfits) * 100 : 0}%`,
                          opacity: hasLiveUpdates ? [0.8, 1, 0.8] : 0.8 
                        }}
                        transition={{ 
                          duration: 0.8,
                          width: { type: "spring", stiffness: 50, damping: 10, delay: 0.3 + (index * 0.1) }
                        }}
                      />
                    </motion.div>
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <AnimatedValue
                        value={strategy.trades}
                        precision={0}
                        suffix=" trades"
                      />
                      <AnimatedValue
                        value={statsData.totalProfits > 0 ? (strategy.profit / statsData.totalProfits) * 100 : 0}
                        precision={1}
                        suffix="% of profits"
                      />
                    </div>
                  </motion.div>
                ))}
              </div>
            ) : (
              <div className="flex h-32 w-full items-center justify-center rounded-md border border-dashed border-muted-foreground/50 p-8 text-center">
                <motion.div 
                  className="space-y-2"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.3 }}
                >
                  <p className="text-sm text-muted-foreground">No strategy data available</p>
                </motion.div>
              </div>
            )}
          </CardContent>
          <motion.div 
            className="bg-primary/20 h-1 w-full"
            initial={{ scaleX: 0 }}
            animate={{ scaleX: 1 }}
            transition={{ duration: 0.8, delay: 0.5 }}
          />
        </Card>
      </motion.div>
      
      {/* Market Stats */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7, delay: 0.3 }}
      >
        <Card>
          <CardHeader>
            <CardTitle>Market Conditions</CardTitle>
            <CardDescription>Current crypto market insights</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
            <div className="space-y-2">
              <div className="text-sm text-muted-foreground">Market Trend</div>
              <div className="flex items-center gap-2">
                <motion.div 
                  className={`rounded-full p-1.5 ${
                    statsData.marketStatus === 'bullish' ? 'bg-green-500/20' : 
                    statsData.marketStatus === 'bearish' ? 'bg-red-500/20' : 'bg-orange-500/20'
                  }`}
                  animate={
                    hasLiveUpdates ? 
                    { 
                      scale: [1, 1.1, 1],
                      opacity: [0.7, 1, 0.7]
                    } : 
                    { scale: 1, opacity: 0.7 }
                  }
                  transition={{ duration: 0.8 }}
                >
                  <AnimatePresence mode="wait">
                    <motion.div
                      key={statsData.marketStatus}
                      initial={{ rotate: -10, opacity: 0 }}
                      animate={{ rotate: 0, opacity: 1 }}
                      exit={{ rotate: 10, opacity: 0 }}
                      transition={{ duration: 0.3 }}
                    >
                      {statsData.marketStatus === 'bullish' ? (
                        <ArrowUp className="h-5 w-5 text-green-500" />
                      ) : statsData.marketStatus === 'bearish' ? (
                        <ArrowDown className="h-5 w-5 text-red-500" />
                      ) : (
                        <ArrowRight className="h-5 w-5 text-orange-500" />
                      )}
                    </motion.div>
                  </AnimatePresence>
                </motion.div>
                
                <motion.span 
                  className="text-lg font-medium capitalize"
                  animate={{ 
                    opacity: hasLiveUpdates ? [0.8, 1, 0.8] : 0.8 
                  }}
                  transition={{ duration: 0.8 }}
                >
                  <AnimatePresence mode="wait">
                    <motion.span
                      key={statsData.marketStatus}
                      initial={{ y: 10, opacity: 0 }}
                      animate={{ y: 0, opacity: 1 }}
                      exit={{ y: -10, opacity: 0 }}
                      transition={{ duration: 0.3 }}
                    >
                      {statsData.marketStatus}
                    </motion.span>
                  </AnimatePresence>
                </motion.span>
              </div>
            </div>
            
            <div className="space-y-2">
              <div className="text-sm text-muted-foreground">Market Volatility</div>
              <div className="h-2 w-full rounded-full bg-secondary overflow-hidden">
                <motion.div 
                  className={`h-full rounded-full ${
                    statsData.volatilityIndex > 75 ? 'bg-red-500' :
                    statsData.volatilityIndex > 50 ? 'bg-orange-500' :
                    statsData.volatilityIndex > 25 ? 'bg-yellow-500' : 'bg-green-500'
                  }`}
                  initial={{ width: 0 }}
                  animate={{ 
                    width: `${statsData.volatilityIndex}%`,
                    opacity: hasLiveUpdates ? [0.8, 1, 0.8] : 0.8
                  }}
                  transition={{ 
                    duration: 0.8,
                    width: { type: "spring", stiffness: 100, damping: 15 }
                  }}
                />
              </div>
              <div className="flex items-center justify-between text-xs">
                <span>Low</span>
                <motion.span
                  className={`font-medium ${
                    statsData.volatilityIndex > 75 ? 'text-red-500' :
                    statsData.volatilityIndex > 50 ? 'text-orange-500' :
                    statsData.volatilityIndex > 25 ? 'text-yellow-500' : 'text-green-500'
                  }`}
                  animate={
                    hasLiveUpdates ? 
                    { scale: [1, 1.15, 1] } : 
                    { scale: 1 }
                  }
                  transition={{ duration: 0.5 }}
                >
                  <AnimatedValue 
                    value={statsData.volatilityIndex} 
                    precision={0} 
                    suffix="/100"
                    duration={0.8}
                  />
                </motion.span>
                <span>High</span>
              </div>
            </div>
            
            <div className="space-y-2">
              <div className="text-sm text-muted-foreground">Trading Opportunity</div>
              <div className="flex items-center gap-1">
                <motion.div
                  animate={
                    hasLiveUpdates ? 
                    { scale: [1, 1.1, 1], opacity: [0.8, 1, 0.8] } : 
                    { scale: 1, opacity: 0.8 }
                  }
                  transition={{ duration: 0.8 }}
                >
                  <Badge className="bg-primary/20 text-primary">HIGH</Badge>
                </motion.div>
                <span className="text-sm">Based on current conditions</span>
              </div>
              <motion.div 
                className="flex items-center gap-1 text-xs text-muted-foreground"
                whileHover={{ scale: 1.01, x: 2 }}
                transition={{ type: "spring", stiffness: 400, damping: 10 }}
              >
                <motion.div
                  animate={
                    hasLiveUpdates ? 
                    { x: [0, 2, 0], opacity: [0.7, 1, 0.7] } : 
                    { x: 0, opacity: 0.7 }
                  }
                  transition={{ duration: 0.6 }}
                >
                  <Link className="h-3 w-3" />
                </motion.div>
                <span>View recommended strategies</span>
              </motion.div>
            </div>
          </div>
        </CardContent>
      </Card>
      </motion.div>
    </div>
  );
}