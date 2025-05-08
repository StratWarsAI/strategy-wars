'use client';

import { ArrowUpRight, ArrowDownRight, Twitter, Globe, TrendingUp, BarChart2, Award, AlertTriangle, Clock, Activity } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import Image from 'next/image';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { SimulationSummary, Strategy, TradeClosedEvent, TradeExecutedEvent } from '@/types';
import { formatExitReason, formatNumber } from '@/lib/utils';

interface EnhancedStrategyCardProps {
  strategy: Strategy | SimulationSummary;
  trades?: (TradeExecutedEvent | TradeClosedEvent)[];
  isWinning?: boolean;
  isSimulation?: boolean;
  showActions?: boolean;
  onClick?: () => void;
}

// Type guards
const isStrategy = (strategy: Strategy | SimulationSummary): strategy is Strategy => {
  return 'name' in strategy && 'description' in strategy;
};

const isSimulationSummary = (strategy: Strategy | SimulationSummary): strategy is SimulationSummary => {
  return 'strategyName' in strategy && 'currentBalance' in strategy;
};

export function EnhancedStrategyCard({ 
  strategy, 
  trades = [], 
  isWinning = false, 
  isSimulation = false,
  showActions = true,
  onClick
}: EnhancedStrategyCardProps) {
  
  // Handle both Strategy and SimulationSummary interfaces
  const strategyName = isStrategy(strategy) ? strategy.name : strategy.strategyName;
  const strategyId = isStrategy(strategy) ? strategy.id : strategy.strategyId;
  const description = isStrategy(strategy) ? strategy.description : '';
  
  // Metrics data
  const metrics = isStrategy(strategy) && strategy.metrics ? strategy.metrics : strategy;
  const winRate = isSimulationSummary(strategy) ? strategy.winRate : 
                 (metrics && 'winRate' in metrics ? metrics.winRate : 0);
  const totalTrades = isSimulationSummary(strategy) ? strategy.totalTrades : 
                     (metrics && 'totalTrades' in metrics ? metrics.totalTrades : 0);
  const balance = isSimulationSummary(strategy) ? strategy.currentBalance : 
                 (metrics && 'balance' in metrics ? metrics.balance : 0);
  const initialBalance = isSimulationSummary(strategy) ? strategy.initialBalance : 
                        (metrics && 'initialBalance' in metrics ? metrics.initialBalance : 0);
  const roi = isSimulationSummary(strategy) ? strategy.roi : 
             (metrics && 'roi' in metrics ? metrics.roi : 0);
  
  // Card border and status styling
  const getCardStatus = () => {
    if (isWinning) return 'border-primary shadow-lg shadow-primary/20';
    if (roi > 15) return 'border-green-500 shadow-lg shadow-green-500/20';
    if (roi > 0) return 'border-emerald-400 shadow-md shadow-emerald-400/10';
    if (roi < 0) return 'border-red-500';
    return 'border-border';
  };

  // Strategy risk/complexity styling
  const getRiskBadge = () => {
    if (isStrategy(strategy) && strategy.riskScore !== undefined) {
      const { riskScore } = strategy;
      if (riskScore > 7) return { color: 'border-red-500 bg-red-500/10 text-red-500', label: 'High Risk' };
      if (riskScore > 4) return { color: 'border-yellow-500 bg-yellow-500/10 text-yellow-500', label: 'Medium Risk' };
      return { color: 'border-green-500 bg-green-500/10 text-green-500', label: 'Low Risk' };
    }
    return null;
  };

  const getComplexityBadge = () => {
    if (isStrategy(strategy) && strategy.complexityScore !== undefined) {
      const { complexityScore } = strategy;
      if (complexityScore > 7) return { color: 'border-purple-500 bg-purple-500/10 text-purple-500', label: 'Complex' };
      if (complexityScore > 4) return { color: 'border-blue-500 bg-blue-500/10 text-blue-500', label: 'Moderate' };
      return { color: 'border-green-500 bg-green-500/10 text-green-500', label: 'Simple' };
    }
    return null;
  };

  return (
    <Card 
      className={`h-full transition-all duration-300 hover:scale-[1.01] ${getCardStatus()} ${onClick ? 'cursor-pointer' : ''}`}
      onClick={onClick}
    >
      <CardHeader className="pb-2 pt-3 sm:pb-3 sm:pt-4">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center">
              <CardTitle className="text-base sm:text-lg md:text-xl">
                {isSimulation ? `Strategy #${strategyId}` : strategyName}
              </CardTitle>
              {isWinning && (
                <Badge className="ml-2 border-primary/30 bg-primary/20 text-xs text-primary hover:bg-primary/30 sm:text-sm">
                  <Award className="mr-0.5 h-3 w-3" />
                  WINNING
                </Badge>
              )}
              {isStrategy(strategy) && strategy.aiEnhanced && (
                <Badge className="ml-2 border-indigo-500/30 bg-indigo-500/20 text-xs text-indigo-500 hover:bg-indigo-500/30 sm:text-sm">
                  AI ENHANCED
                </Badge>
              )}
            </div>
            {!isSimulation && isStrategy(strategy) && (
              <CardDescription className="mt-1 line-clamp-2">{description}</CardDescription>
            )}
            {isSimulation && (
              <p className="mt-1 text-xs text-muted-foreground sm:text-sm">{strategyName}</p>
            )}
          </div>
          
          {/* Show badges for risk and complexity */}
          {!isSimulation && isStrategy(strategy) && strategy.riskScore !== undefined && (
            <div className="flex flex-col gap-1">
              {getRiskBadge() && (
                <div className={`rounded-md px-2 py-0.5 text-xs ${getRiskBadge()?.color}`}>
                  {getRiskBadge()?.label}
                </div>
              )}
              {getComplexityBadge() && (
                <div className={`rounded-md px-2 py-0.5 text-xs ${getComplexityBadge()?.color}`}>
                  {getComplexityBadge()?.label}
                </div>
              )}
            </div>
          )}
        </div>
      </CardHeader>
      
      <CardContent className="pt-0">
        {/* Key Metrics */}
        <div className="mb-3 grid grid-cols-2 gap-2 rounded-xl bg-gradient-to-br from-secondary/5 to-secondary/20 p-2 sm:mb-4 sm:gap-4">
          {/* Balance */}
          <div className="rounded-lg bg-card/80 p-2 backdrop-blur-sm">
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <BarChart2 className="h-3.5 w-3.5" />
              <span>Balance</span>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm font-semibold sm:text-base">{balance?.toFixed(2)} SOL</div>
              <div className={`flex items-center text-xs ${roi > 0 ? 'text-green-500' : 'text-red-500'}`}>
                {roi > 0 ? (
                  <ArrowUpRight className="mr-0.5 h-3 w-3" />
                ) : (
                  <ArrowDownRight className="mr-0.5 h-3 w-3" />
                )}
                {roi > 0 ? '+' : ''}{roi.toFixed(1)}%
              </div>
            </div>
            <div className="mt-1 text-xs text-muted-foreground">
              Initial: {initialBalance.toFixed(2)} SOL
            </div>
          </div>
          
          {/* Performance */}
          <div className="rounded-lg bg-card/80 p-2 backdrop-blur-sm">
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Activity className="h-3.5 w-3.5" />
              <span>Performance</span>
            </div>
            <div className="text-sm font-semibold sm:text-base">{winRate.toFixed(1)}% Win Rate</div>
            <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-muted">
              <div className="relative h-full w-full">
                <Progress 
                  value={winRate} 
                  className={`h-full ${
                    winRate > 70 ? 'bg-green-500' : winRate > 50 ? 'bg-amber-500' : 'bg-red-500'
                  }`} 
                />
              </div>
            </div>
          </div>
          
          {/* Trades */}
          <div className="rounded-lg bg-card/80 p-2 backdrop-blur-sm">
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <TrendingUp className="h-3.5 w-3.5" />
              <span>Trades</span>
            </div>
            <div className="text-sm font-semibold sm:text-base">{totalTrades} Total</div>
            <div className="mt-1 text-xs text-muted-foreground">
              {trades.filter(t => t.action === 'buy').length} Active
            </div>
          </div>
          
          {/* ROI with visual indicator */}
          <div className="rounded-lg bg-card/80 p-2 backdrop-blur-sm">
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <TrendingUp className="h-3.5 w-3.5" />
              <span>ROI</span>
            </div>
            <div className={`text-sm font-semibold sm:text-base ${roi > 0 ? 'text-green-500' : roi < 0 ? 'text-red-500' : 'text-muted-foreground'}`}>
              {roi > 0 ? '+' : ''}{roi.toFixed(1)}%
            </div>
            <div className="mt-2 flex w-full items-center gap-1">
              <div className="h-1 flex-1 rounded-full bg-red-500/20"></div>
              <div className="h-1 flex-1 rounded-full bg-amber-500/20"></div>
              <div className="h-1 flex-1 rounded-full bg-green-500/20"></div>
              <div 
                className={`absolute h-3 w-3 rounded-full border-2 ${roi > 0 ? 'border-green-500 bg-green-500/20' : 'border-red-500 bg-red-500/20'}`}
                style={{ 
                  marginLeft: `${Math.min(Math.max((roi + 50) * 100 / 100, 0), 100)}%`, 
                  transform: 'translateX(-50%)' 
                }}
              ></div>
            </div>
          </div>
        </div>
        
        {trades.length > 0 && (
          <div className="mt-3 sm:mt-4">
            <h3 className="mb-2 flex items-center gap-1 text-xs font-medium sm:text-sm">
              <Clock className="h-3.5 w-3.5" /> Recent Trades
            </h3>
            <ScrollArea className="h-44 pr-2 sm:h-56 md:h-[260px]">
              <div className="space-y-2 sm:space-y-3">
                {trades.map(trade => (
                  <div 
                    key={`${trade.tokenId}-${trade.timestamp}`}
                    className={`group relative rounded-lg border p-2 transition-all duration-300 hover:shadow-md ${
                      'action' in trade && trade.action === 'buy' 
                        ? 'border-l-2 border-l-green-600 bg-green-950/10' 
                        : 'profit' in trade && trade.profit && trade.profit > 0 
                          ? 'border-l-2 border-l-green-600 bg-green-950/10' 
                          : 'border-l-2 border-l-red-600 bg-red-950/10'
                    } sm:p-2.5`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        {trade.imageUrl ? (
                          <div className="relative h-6 w-6 overflow-hidden rounded-full bg-card ring-1 ring-primary/10">
                            <Image 
                              src={trade.imageUrl} 
                              alt={trade.tokenName} 
                              width={24} 
                              height={24} 
                              className="h-full w-full object-cover"
                            />
                          </div>
                        ) : (
                          <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                            {trade.tokenName?.substring(0, 2)}
                          </div>
                        )}
                        <span className="text-xs font-semibold sm:text-sm">
                          {trade.tokenSymbol ? `${trade.tokenName} (${trade.tokenSymbol})` : trade.tokenName}
                        </span>
                        <div className="flex gap-1">
                          <TooltipProvider>
                            {trade.twitterUrl && (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <a 
                                    href={trade.twitterUrl} 
                                    target="_blank" 
                                    rel="noopener noreferrer" 
                                    className="text-blue-400 opacity-70 transition-opacity hover:opacity-100 hover:text-blue-500"
                                  >
                                    <Twitter className="h-3.5 w-3.5" />
                                  </a>
                                </TooltipTrigger>
                                <TooltipContent>
                                  <p className="text-xs">Twitter</p>
                                </TooltipContent>
                              </Tooltip>
                            )}
                            
                            {trade.websiteUrl && (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <a 
                                    href={trade.websiteUrl} 
                                    target="_blank" 
                                    rel="noopener noreferrer" 
                                    className="text-muted-foreground opacity-70 transition-opacity hover:opacity-100 hover:text-primary"
                                  >
                                    <Globe className="h-3.5 w-3.5" />
                                  </a>
                                </TooltipTrigger>
                                <TooltipContent>
                                  <p className="text-xs">Website</p>
                                </TooltipContent>
                              </Tooltip>
                            )}
                          </TooltipProvider>
                        </div>
                      </div>
                      
                      <Badge 
                        variant={'action' in trade && trade.action === 'buy' ? 'success' : 'profit' in trade && trade.profit && trade.profit > 0 ? 'success' : 'destructive'} 
                        className="text-[10px] sm:text-xs"
                      >
                        {'action' in trade ? trade.action.toUpperCase() : 'TRADE'}
                      </Badge>
                    </div>
                    
                    <div className="mt-1 flex justify-between text-xs text-muted-foreground">
                      <span>Time: {new Date(trade.timestamp * 1000).toLocaleTimeString()}</span>
                    </div>
                    
                    {'action' in trade && trade.action === 'buy' ? (
                      <div className="mt-1 grid grid-cols-2 gap-x-2 gap-y-0.5">
                        <div className="text-xs">Price: {trade.price?.toFixed(6)} SOL</div>
                        <div className="text-xs">Amount: {trade.amount} SOL</div>
                        {'entryMarketCap' in trade && trade.entryMarketCap && (
                          <div className="col-span-2 text-xs">Market Cap: ${formatNumber(trade.entryMarketCap)}</div>
                        )}
                      </div>
                    ) : (
                      <div className="mt-1 grid grid-cols-2 gap-x-2 gap-y-0.5">
                        {'entryPrice' in trade && <div className="text-xs">Entry: {trade.entryPrice?.toFixed(6)} SOL</div>}  
                        {'exitPrice' in trade && <div className="text-xs">Exit: {trade.exitPrice?.toFixed(6)} SOL</div>}
                        {'profitLoss' in trade && trade.profitLoss !== undefined && 'profitLossPct' in trade && (
                          <div className={`text-xs ${trade.profitLoss > 0 ? 'text-green-500' : 'text-red-500'}`}>
                            {trade.profitLoss > 0 ? '+' : ''}{trade.profitLossPct.toFixed(1)}%
                          </div>
                        )}
                        {'exitReason' in trade && trade.exitReason && (
                          <div className="text-xs">Reason: {formatExitReason(trade.exitReason)}</div>
                        )}
                      </div>
                    )}
                    
                    {/* Hover effect with animated border */}
                    <div className="absolute inset-0 -z-10 rounded-lg opacity-0 transition-opacity duration-300 group-hover:opacity-100">
                      <div className="absolute inset-0 rounded-lg bg-gradient-to-r from-primary/20 via-primary/10 to-primary/20 blur-sm"></div>
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </div>
        )}
        
        {/* Strategy Tags */}
        {isStrategy(strategy) && strategy.tags && strategy.tags.length > 0 && (
          <div className="mt-3">
            <div className="flex flex-wrap gap-1">
              {strategy.tags.map((tag, index) => (
                <span 
                  key={index} 
                  className="rounded-md bg-secondary/40 px-1.5 py-0.5 text-[10px]"
                >
                  {tag}
                </span>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}