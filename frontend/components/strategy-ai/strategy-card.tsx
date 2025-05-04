'use client';

import { ArrowUpRight, ArrowDownRight, Twitter, Globe } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import Image from 'next/image';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { SimulationSummary, Strategy, TradeClosedEvent, TradeExecutedEvent } from '@/types';
import { formatExitReason, formatNumber } from '@/lib/utils';


interface StrategyCardProps {
  strategy: SimulationSummary;
  trades: (TradeExecutedEvent | TradeClosedEvent)[];
  isWinning: boolean;
}

export function StrategyCard({ strategy, trades, isWinning }: StrategyCardProps) {

  return (
    <Card className={`h-full ${isWinning ? 'border-primary' : 'border-border'}`}>
      <CardHeader className="pb-2 pt-3 sm:pb-3 sm:pt-4">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center">
              <CardTitle className="text-base sm:text-lg md:text-xl">Strategy #{strategy.strategyId}</CardTitle>
              {isWinning && (
                <Badge className="ml-2 border-primary/30 bg-primary/20 text-xs text-primary hover:bg-primary/30 sm:text-sm">
                  WINNING
                </Badge>
              )}
            </div>
            <p className="mt-1 text-xs text-muted-foreground sm:text-sm">{strategy.strategyName}</p>
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground sm:line-clamp-none">{'Description'}</p>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="pt-0">
        <div className="mb-3 grid grid-cols-2 gap-2 sm:mb-4 sm:gap-4">
          <div className="rounded-lg bg-secondary/20 p-2">
            <div className="text-xs text-muted-foreground">Balance</div>
            <div className="flex items-center justify-between">
              <div className="text-xs font-semibold sm:text-sm">{strategy.currentBalance.toFixed(2)} SOL</div>
              <div className="flex items-center text-xs text-primary">
                <ArrowUpRight className="mr-0.5 h-3 w-3" />
                +{strategy.roi.toFixed(1)}%
              </div>
            </div>
            <div className="mt-1 text-xs text-muted-foreground">
            Initial: {strategy.initialBalance.toFixed(2)} SOL
            </div>
          </div>
          
          <div className="rounded-lg bg-secondary/20 p-2">
            <div className="text-xs text-muted-foreground">Performance</div>
            <div className="text-xs font-semibold sm:text-sm">{strategy.winRate.toFixed(2)}% Win Rate</div>
            <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-muted">
              <Progress value={strategy.winRate} className="h-full" />
            </div>
          </div>
          
          <div className="rounded-lg bg-secondary/20 p-2">
            <div className="text-xs text-muted-foreground">Trades</div>
            <div className="text-xs font-semibold sm:text-sm">{strategy.totalTrades} Total</div>
            <div className="mt-1 text-xs text-muted-foreground">
              {'Active Trades'} Active
            </div>
          </div>
          
          <div className="rounded-lg bg-secondary/20 p-2">
            <div className="text-xs text-muted-foreground">ROI</div>
            <div className="text-xs font-semibold text-primary sm:text-sm">+{strategy.roi.toFixed(1)}%</div>
            <div className="mt-1 text-xs text-muted-foreground">
              From Initial Balance
            </div>
          </div>
        </div>
        
        <div className="mt-3 sm:mt-4">
          <h3 className="mb-2 text-xs font-medium sm:text-sm">Recent Trades</h3>
          <ScrollArea className="h-40 pr-2 sm:h-56 md:h-[260px]">
            <div className="space-y-2 sm:space-y-3">
              {trades.map(trade => (
                <div 
                 key={`${trade.tokenId}-${trade.timestamp}`}
                  className={`rounded-lg border p-2 ${
                    trade.action === 'buy' 
                      ? 'border-l-2 border-l-green-600 bg-green-950/10' 
                      : trade.profit && trade.profit > 0 
                        ? 'border-l-2 border-l-green-600 bg-green-950/10' 
                        : 'border-l-2 border-l-red-600 bg-red-950/10'
                  } sm:p-2.5`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {trade.imageUrl ? (
                        <div className="relative h-6 w-6 overflow-hidden rounded-full bg-secondary/50">
                          <Image 
                            src={trade.imageUrl} 
                            alt={trade.tokenName} 
                            width={24} 
                            height={24} 
                            className="h-full w-full object-cover"
                          />
                        </div>
                      ) : (
                        <div className="flex h-6 w-6 items-center justify-center rounded-full bg-secondary/50 text-xs font-medium">
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
                                  className="text-blue-400 hover:text-blue-500"
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
                                  className="text-muted-foreground hover:text-primary"
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
                    
                    <Badge color={trade.action === 'buy' ? 'success' : trade.profit && trade.profit > 0 ? 'success' : 'destructive'} className="text-[10px] sm:text-xs">
                      {trade.action.toUpperCase()}
                    </Badge>
                  </div>
                  
                  <div className="mt-0.5 flex justify-between text-xs text-muted-foreground">
                    <span>Time: {new Date(trade.timestamp * 1000).toLocaleTimeString()}</span>
                  </div>
                  
                  {trade.action === 'buy' ? (
                    <>
                      <div className="mt-0.5 text-xs">Price: {trade.price?.toFixed(6)} SOL</div>
                      <div className="mt-0.5 text-xs">Amount: {trade.amount} SOL</div>
                      {trade.entryMarketCap && (
                        <div className="mt-0.5 text-xs">Market Cap: ${formatNumber(trade.entryMarketCap)}</div>
                      )}
                    </>
                  ) : (
                    <>
                      <div className="mt-0.5 flex flex-wrap gap-1 text-xs">
                        <span>Entry: {trade.entryPrice?.toFixed(6)} SOL</span>
                        <span>/ Exit: {trade.exitPrice?.toFixed(6)} SOL</span>
                      </div>
                      {trade.profitLoss !== undefined && (
                        <div className={`mt-0.5 text-xs ${trade.profitLoss > 0 ? 'text-green-500' : 'text-red-500'}`}>
                          {trade.profitLoss > 0 ? '+' : ''}{trade.profitLossPct.toFixed(1)}%
                        </div>
                      )}
                      {trade.exitReason && (
                        <div className="mt-0.5 text-xs">Reason: {formatExitReason(trade.exitReason)}</div>
                      )}
                    </>
                  )}
                </div>
              ))}
            </div>
          </ScrollArea>
        </div>
      </CardContent>
    </Card>
  );
}