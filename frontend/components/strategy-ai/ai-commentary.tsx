'use client';

import { useState, useEffect } from 'react';
import { Brain, LineChart, Zap, Settings, ArrowRightLeft, Layers } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { SimulationSummary } from '@/types';
import { AIAnalysisEvent } from '@/types/websocket.type';

interface AICommentaryProps {
  strategy1?: SimulationSummary;
  strategy2?: SimulationSummary;
  aiAnalysis?: AIAnalysisEvent | null;
}

export function AICommentary({ strategy1, strategy2, aiAnalysis }: AICommentaryProps) {
  const [generatedStrategies, setGeneratedStrategies] = useState(() => {
    return [
      {
        id: strategy1?.strategyId || 4,
        name: strategy1?.strategyName || ".",
        description: "",
        params: {
          marketCapThreshold: strategy1?.simConfig?.marketCapThreshold || 4000,
          minBuysForEntry: strategy1?.simConfig?.minBuysForEntry || 1,
          entryTimeWindowSec: strategy1?.simConfig?.entryTimeWindowSec || 120,
          takeProfitPct: strategy1?.simConfig?.takeProfitPct || 25,
          stopLossPct: strategy1?.simConfig?.stopLossPct || 10,
          maxHoldTimeSec: strategy1?.simConfig?.maxHoldTimeSec || 30,
          fixedPositionSizeSol: strategy1?.simConfig?.fixedPositionSizeSol || 0.8,
          initialBalance: strategy1?.initialBalance || 50
        },
        performance: {
          roi: strategy1?.roi || 11.09,
          winRate: strategy1?.winRate || 51.5,
          trades: strategy1?.totalTrades || 332,
          activeTrades: strategy1?.activeTrades || 8
        }
      },
      {
        id: strategy2?.strategyId || 7,
        name: strategy2?.strategyName || "",
        description: "",
        params: {
          marketCapThreshold: strategy2?.simConfig?.marketCapThreshold || 7000,
          minBuysForEntry: strategy2?.simConfig?.minBuysForEntry || 2,
          entryTimeWindowSec: strategy2?.simConfig?.entryTimeWindowSec || 120,
          takeProfitPct: strategy2?.simConfig?.takeProfitPct || 65,
          stopLossPct: strategy2?.simConfig?.stopLossPct || 10,
          maxHoldTimeSec: strategy2?.simConfig?.maxHoldTimeSec || 90,
          fixedPositionSizeSol: strategy2?.simConfig?.fixedPositionSizeSol || 1,
          initialBalance: strategy2?.initialBalance || 50
        },
        performance: {
          roi: strategy2?.roi || 1.33,
          winRate: strategy2?.winRate || 65.08,
          trades: strategy2?.totalTrades || 63,
          activeTrades: strategy2?.activeTrades || 4
        }
      }
    ];
  });

  // Generate initial commentary based on provided data
  const [currentCommentary, setCurrentCommentary] = useState(() => {
    const s1Id = strategy1?.strategyId || 4;
    const s2Id = strategy2?.strategyId || 7;
    const s1Roi = strategy1?.roi || 11.09;
    const s2Roi = strategy2?.roi || 1.33;
    const roiDiff = Math.abs(s1Roi - s2Roi).toFixed(1);
    const s1WinRate = strategy1?.winRate || 51.5;
    const s2WinRate = strategy2?.winRate || 65.08;
    const betterRoi = s1Roi > s2Roi ? s1Id : s2Id;
    const betterWinRate = s1WinRate > s2WinRate ? s1Id : s2Id;
    
    let analysisText = '';
    if (betterRoi === betterWinRate) {
      analysisText = `Strategy #${betterRoi} is outperforming in both ROI (${betterRoi === s1Id ? s1Roi : s2Roi}% vs ${betterRoi === s1Id ? s2Roi : s1Roi}%) and win rate.`;
    } else {
      analysisText = `Strategy #${betterRoi} has better ROI (${betterRoi === s1Id ? s1Roi : s2Roi}%), while Strategy #${betterWinRate} has better win rate (${betterWinRate === s1Id ? s1WinRate : s2WinRate}%).`;
    }
    
    if (strategy1 && strategy2) {
      if (strategy1.totalTrades > strategy2.totalTrades * 3) {
        analysisText += ` Strategy #${s1Id} is much more active with ${strategy1.totalTrades} trades vs only ${strategy2.totalTrades} trades.`;
      } else if (strategy2.totalTrades > strategy1.totalTrades * 3) {
        analysisText += ` Strategy #${s2Id} is much more active with ${strategy2.totalTrades} trades vs only ${strategy1.totalTrades} trades.`;
      }
    }
    
    return {
      id: 1,
      timestamp: Date.now() - 600000,
      title: "10 Minute Analysis",
      comment: analysisText
    };
  });
  
  const [nextUpdate, setNextUpdate] = useState(600); 
  
  // Timer simulation
  useEffect(() => {
    const interval = setInterval(() => {
      setNextUpdate(prev => {
        if (prev <= 0) {
          return 600;
        }
        return prev - 1;
      });
    }, 1000);
    
    return () => clearInterval(interval);
  }, []);

  // Update the state when strategies change or aiAnalysis is received
  useEffect(() => {
    if (strategy1 && strategy2) {
      // Update generated strategies
      setGeneratedStrategies([
        {
          id: strategy1.strategyId || 4,
          name: strategy1.strategyName || "Crypto Rocket",
          description: "This strategy focuses on tokens that show a sudden burst in trading volume with aggressive entry and exit parameters",
          params: {
            marketCapThreshold: strategy1.simConfig?.marketCapThreshold || 4000,
            minBuysForEntry: strategy1.simConfig?.minBuysForEntry || 1,
            entryTimeWindowSec: strategy1.simConfig?.entryTimeWindowSec || 120,
            takeProfitPct: strategy1.simConfig?.takeProfitPct || 25,
            stopLossPct: strategy1.simConfig?.stopLossPct || 10,
            maxHoldTimeSec: strategy1.simConfig?.maxHoldTimeSec || 30,
            fixedPositionSizeSol: strategy1.simConfig?.fixedPositionSizeSol || 0.8,
            initialBalance: strategy1.initialBalance || 50
          },
          performance: {
            roi: strategy1.roi || 0,
            winRate: strategy1.winRate || 0,
            trades: strategy1.totalTrades || 0,
            activeTrades: strategy1.activeTrades || 0
          }
        },
        {
          id: strategy2.strategyId || 7,
          name: strategy2.strategyName || "Quick Hit Blitz",
          description: "A more aggressive approach focusing on quick profits with higher take profit targets",
          params: {
            marketCapThreshold: strategy2.simConfig?.marketCapThreshold || 7000,
            minBuysForEntry: strategy2.simConfig?.minBuysForEntry || 2,
            entryTimeWindowSec: strategy2.simConfig?.entryTimeWindowSec || 120,
            takeProfitPct: strategy2.simConfig?.takeProfitPct || 65,
            stopLossPct: strategy2.simConfig?.stopLossPct || 10,
            maxHoldTimeSec: strategy2.simConfig?.maxHoldTimeSec || 90,
            fixedPositionSizeSol: strategy2.simConfig?.fixedPositionSizeSol || 1,
            initialBalance: strategy2.initialBalance || 50
          },
          performance: {
            roi: strategy2.roi || 0,
            winRate: strategy2.winRate || 0,
            trades: strategy2.totalTrades || 0,
            activeTrades: strategy2.activeTrades || 0
          }
        }
      ]);
      
      // Generate new commentary based on strategies if we don't have AI analysis
      if (!aiAnalysis) {
        const s1Id = strategy1.strategyId;
        const s2Id = strategy2.strategyId;
        const s1Roi = strategy1.roi;
        const s2Roi = strategy2.roi;
        const roiDiff = Math.abs(s1Roi - s2Roi).toFixed(1);
        const s1WinRate = strategy1.winRate;
        const s2WinRate = strategy2.winRate;
        const betterRoi = s1Roi > s2Roi ? s1Id : s2Id;
        const betterWinRate = s1WinRate > s2WinRate ? s1Id : s2Id;
        
        let analysisText = '';
        if (betterRoi === betterWinRate) {
          analysisText = `Strategy #${betterRoi} is outperforming in both ROI (${betterRoi === s1Id ? s1Roi : s2Roi}% vs ${betterRoi === s1Id ? s2Roi : s1Roi}%) and win rate.`;
        } else {
          analysisText = `Strategy #${betterRoi} has better ROI (${betterRoi === s1Id ? s1Roi : s2Roi}%), while Strategy #${betterWinRate} has better win rate (${betterWinRate === s1Id ? s1WinRate : s2WinRate}%).`;
        }
        
        if (strategy1.totalTrades > strategy2.totalTrades * 3) {
          analysisText += ` Strategy #${s1Id} is much more active with ${strategy1.totalTrades} trades vs only ${strategy2.totalTrades} trades.`;
        } else if (strategy2.totalTrades > strategy1.totalTrades * 3) {
          analysisText += ` Strategy #${s2Id} is much more active with ${strategy2.totalTrades} trades vs only ${strategy1.totalTrades} trades.`;
        }
        
        // Update the commentary with current timestamp
        setCurrentCommentary({
          id: Date.now(),
          timestamp: Date.now(),
          title: "Updated Analysis",
          comment: analysisText
        });
      }
      
      setNextUpdate(600);
    }
  }, [strategy1, strategy2]); // Run effect when strategy data changes
  
  // Effect for AI analysis updates
  useEffect(() => {
    if (aiAnalysis) {
      console.log("Received AI analysis:", aiAnalysis);
      
      setCurrentCommentary({
        id: aiAnalysis.timestamp,
        timestamp: aiAnalysis.timestamp * 1000, // Convert from Unix timestamp
        title: "AI Strategy Analysis",
        comment: aiAnalysis.analysis
      });
      
      // If the analysis is for the currently displayed strategy, update that strategy
      if (strategy1 && aiAnalysis.strategy_id === strategy1.strategyId) {
        setGeneratedStrategies(prev => {
          const updated = [...prev];
          if (updated[0]) {
            updated[0].performance = {
              ...updated[0].performance,
              roi: aiAnalysis.roi, // Changed from aiAnalysis.performance_metrics.roi
              winRate: aiAnalysis.win_rate // Changed from aiAnalysis.performance_metrics.win_rate
            };
          }
          return updated;
        });
      } else if (strategy2 && aiAnalysis.strategy_id === strategy2.strategyId) {
        setGeneratedStrategies(prev => {
          const updated = [...prev];
          if (updated[1]) {
            updated[1].performance = {
              ...updated[1].performance,
              roi: aiAnalysis.roi, // Changed from aiAnalysis.performance_metrics.roi
              winRate: aiAnalysis.win_rate // Changed from aiAnalysis.performance_metrics.win_rate
            };
          }
          return updated;
        });
      }
      
      setNextUpdate(600);
    }
  }, [aiAnalysis, strategy1, strategy2]);

  // Format time
  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };
  
  // Format remaining time
  const formatRemainingTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  // Format parameter values
  const formatParamValue = (key: string, value: any) => {
    switch (key) {
      case 'marketCapThreshold':
        return `$${value.toLocaleString()}`;
      case 'takeProfitPct':
      case 'stopLossPct':
        return `${value}%`;
      case 'entryTimeWindowSec':
      case 'maxHoldTimeSec':
        const minutes = Math.floor(value / 60);
        const seconds = value % 60;
        return `${minutes}m ${seconds > 0 ? `${seconds}s` : ''}`;
      case 'fixedPositionSizeSol':
      case 'initialBalance':
        return `${value} SOL`;
      default:
        return value;
    }
  };

  // Convert camelCase to readable text
  const formatParamName = (name: string) => {
    return name
      .replace(/([A-Z])/g, ' $1')
      .replace(/^./, (str) => str.toUpperCase())
      .replace(/Pct/, '%')
      .replace(/Sec/, ' (seconds)')
      .replace(/Sol/, ' (SOL)');
  };

  return (
    <Card className="border-primary/30">
      <CardHeader className="bg-primary/5 border-b space-y-1 px-4 py-3 sm:px-6 sm:py-4">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <div className="rounded-full bg-primary/20 p-1.5">
              <Brain className="h-4 w-4 text-primary sm:h-5 sm:w-5" />
            </div>
            <div>
              <CardTitle className="text-base sm:text-lg md:text-xl">AI Strategy Analysis</CardTitle>
              <CardDescription className="text-xs sm:text-sm">Real-time AI recommendations based on performance data</CardDescription>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="outline" className="flex items-center gap-1 text-xs">
              <Zap className="h-3 w-3" />
              <span className="hidden sm:inline">Auto-updating</span>
              <span className="inline sm:hidden">Auto</span>
            </Badge>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="px-4 py-3 sm:p-6">
        <Tabs defaultValue="analysis">
          <TabsList className="mb-4 w-full">
            <TabsTrigger value="analysis" className="flex-1 text-xs sm:text-sm">Analysis</TabsTrigger>
            <TabsTrigger value="strategies" className="flex-1 text-xs sm:text-sm">Strategy Details</TabsTrigger>
          </TabsList>
          
          <TabsContent value="analysis" className="space-y-3 sm:space-y-4">
            <div className="flex items-center gap-2 text-xs text-muted-foreground sm:text-sm">
              <LineChart className="h-3 w-3 sm:h-4 sm:w-4" />
              <span>{currentCommentary.title} • Generated at {formatTime(currentCommentary.timestamp)}</span>
            </div>
            
            <div className="text-sm sm:text-base md:text-lg">{currentCommentary.comment}</div>
          </TabsContent>
          
          <TabsContent value="strategies">
            <div className="space-y-4">
              {generatedStrategies.map(strategy => (
                <Card key={strategy.id} className={`${strategy.performance.roi > 10 ? 'border-primary' : ''}`}>
                  <CardHeader className="p-4 pb-2">
                    <div className="flex justify-between">
                      <div>
                        <CardTitle className="text-base flex items-center gap-2">
                          Strategy #{strategy.id}
                          {strategy.performance.roi > 10 && (
                            <Badge color="success">Top Performer</Badge>
                          )}
                        </CardTitle>
                        <CardDescription className="mt-1">{strategy.name}</CardDescription>
                      </div>
                      <Badge className={`${strategy.performance.roi > 10 ? 'bg-primary/20 text-primary' : 'bg-muted/50'}`}>
                        +{strategy.performance.roi.toFixed(1)}% ROI
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1">{strategy.description}</p>
                  </CardHeader>
                  
                  <CardContent className="p-4 pt-1">
                    <Accordion type="single" collapsible className="w-full">
                      <AccordionItem value="params">
                        <AccordionTrigger className="text-sm py-2">
                          <div className="flex items-center gap-2">
                            <Settings className="h-4 w-4" />
                            <span>Strategy Parameters</span>
                          </div>
                        </AccordionTrigger>
                        <AccordionContent>
                          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                            {Object.entries(strategy.params).map(([key, value]) => (
                              <div key={key} className="flex justify-between p-2 bg-muted/30 rounded-md">
                                <div className="text-xs">{formatParamName(key)}</div>
                                <div className="text-xs font-medium">{formatParamValue(key, value)}</div>
                              </div>
                            ))}
                          </div>
                        </AccordionContent>
                      </AccordionItem>
                      
                      <AccordionItem value="details">
                        <AccordionTrigger className="text-sm py-2">
                          <div className="flex items-center gap-2">
                            <Layers className="h-4 w-4" />
                            <span>Strategy Logic</span>
                          </div>
                        </AccordionTrigger>
                        <AccordionContent>
                          <div className="space-y-3 text-sm">
                            <div className="rounded-md bg-muted/30 p-3">
                              <h4 className="font-medium mb-2 flex items-center gap-2">
                                <Badge variant="outline" className="text-xs">Entry</Badge>
                                <span>When to buy tokens</span>
                              </h4>
                              <ul className="pl-5 space-y-1 list-disc">
                                <li>Market cap at least ${strategy.params.marketCapThreshold.toLocaleString()}</li>
                                <li>At least {strategy.params.minBuysForEntry} buy transactions within {formatParamValue('entryTimeWindowSec', strategy.params.entryTimeWindowSec)}</li>
                              </ul>
                            </div>
                            
                            <div className="rounded-md bg-muted/30 p-3">
                              <h4 className="font-medium mb-2 flex items-center gap-2">
                                <Badge variant="outline" className="text-xs">Exit</Badge>
                                <span>When to sell tokens</span>
                              </h4>
                              <ul className="pl-5 space-y-1 list-disc">
                                <li>Take profit at {strategy.params.takeProfitPct}% price increase</li>
                                <li>Stop loss at {strategy.params.stopLossPct}% price decrease</li>
                                <li>Max hold time: {formatParamValue('maxHoldTimeSec', strategy.params.maxHoldTimeSec)}</li>
                              </ul>
                            </div>
                            
                            <div className="rounded-md bg-muted/30 p-3">
                              <h4 className="font-medium mb-2 flex items-center gap-2">
                                <Badge variant="outline" className="text-xs">Position</Badge>
                                <span>Trading size</span>
                              </h4>
                              <ul className="pl-5 space-y-1 list-disc">
                                <li>{strategy.params.fixedPositionSizeSol} SOL per trade</li>
                                <li>Initial balance: {strategy.params.initialBalance} SOL</li>
                              </ul>
                            </div>
                          </div>
                        </AccordionContent>
                      </AccordionItem>
                      
                      <AccordionItem value="performance">
                        <AccordionTrigger className="text-sm py-2">
                          <div className="flex items-center gap-2">
                            <ArrowRightLeft className="h-4 w-4" />
                            <span>Performance Metrics</span>
                          </div>
                        </AccordionTrigger>
                        <AccordionContent>
                          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                            <div className="rounded-md bg-muted/30 p-3">
                              <div className="text-xs text-muted-foreground">ROI</div>
                              <div className="text-base font-medium text-primary">+{strategy.performance.roi.toFixed(1)}%</div>
                            </div>
                            <div className="rounded-md bg-muted/30 p-3">
                              <div className="text-xs text-muted-foreground">Win Rate</div>
                              <div className="text-base font-medium">{strategy.performance.winRate.toFixed(1)}%</div>
                            </div>
                            <div className="rounded-md bg-muted/30 p-3">
                              <div className="text-xs text-muted-foreground">Total Trades</div>
                              <div className="text-base font-medium">{strategy.performance.trades}</div>
                            </div>
                            <div className="rounded-md bg-muted/30 p-3">
                              <div className="text-xs text-muted-foreground">Active Trades</div>
                              <div className="text-base font-medium">{strategy.performance.activeTrades}</div>
                            </div>
                          </div>
                        </AccordionContent>
                      </AccordionItem>
                    </Accordion>
                  </CardContent>
                </Card>
              ))}
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}