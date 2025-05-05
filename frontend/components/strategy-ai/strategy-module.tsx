'use client';

import React, { useState, useMemo, useEffect } from 'react';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { StrategyCard } from './strategy-card';
import { SimulationSummary } from '@/types';
import { useStrategyEvents } from '@/hooks/use-strategy-events';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { LineChart } from 'lucide-react';
import { PerformanceChart } from './performance-chart';
import { AICommentary } from './ai-commentary';

interface StrategyModuleProps {
  strategy1Data: SimulationSummary;
  strategy2Data: SimulationSummary;
}

export function StrategyModule({strategy1Data, strategy2Data}: StrategyModuleProps) {
  const [activeTab, setActiveTab] = useState("cards");
  const [chartData, setChartData] = useState<any[]>([]);

  const { 
    sortedTrades: strategy1Trades, 
    currentBalance: strategy1Balance, 
    currentRoi: strategy1Roi,
    totalTrades: strategy1TotalTrades,
    performanceData: strategy1Performance = []
  } = useStrategyEvents(strategy1Data?.strategyId);
  
  const { 
    sortedTrades: strategy2Trades, 
    currentBalance: strategy2Balance, 
    currentRoi: strategy2Roi,
    totalTrades: strategy2TotalTrades,
    performanceData: strategy2Performance = []
  } = useStrategyEvents(strategy2Data?.strategyId);
  
  
  
  // Generate combined performance data for the chart
  useEffect(() => {
    // Create initial data points
    const now = Math.floor(Date.now() / 1000);
    const initialData = [
      {
        time: new Date(now * 1000).toLocaleTimeString([], { 
          hour: '2-digit', 
          minute: '2-digit',
          second: '2-digit'
        }),
        strategy1: 0,
        strategy2: 0,
        timestamp: now
      }
    ];
    
    // If we don't have any data, use initial data
    if (strategy1Performance.length === 0 && strategy2Performance.length === 0) {
      setChartData(initialData);
      return;
    }
    
    // Combine all timestamps from both strategies
    const allTimestamps = new Set([
      ...strategy1Performance.map(p => p.timestamp),
      ...strategy2Performance.map(p => p.timestamp)
    ]);
    
    // Create sorted array of timestamps
    const sortedTimestamps = Array.from(allTimestamps).sort((a, b) => a - b);
    
    // If we have no timestamps after combining, use initial data
    if (sortedTimestamps.length === 0) {
      setChartData(initialData);
      return;
    }
    
    // Track the last known values
    let lastKnownStrategy1Roi = 0;
    let lastKnownStrategy2Roi = 0;
    
    // Map timestamps to data points
    const combined = sortedTimestamps.map(timestamp => {
      // Find data points for each strategy at this timestamp
      const s1Point = strategy1Performance.find(p => p.timestamp === timestamp);
      const s2Point = strategy2Performance.find(p => p.timestamp === timestamp);
      
      // Update last known values if we have data
      if (s1Point) lastKnownStrategy1Roi = s1Point.roi;
      if (s2Point) lastKnownStrategy2Roi = s2Point.roi;
      
      // Format time string
      const timeStr = new Date(timestamp * 1000).toLocaleTimeString([], { 
        hour: '2-digit', 
        minute: '2-digit',
        second: '2-digit'
      });
      
      return {
        time: timeStr,
        strategy1: s1Point ? s1Point.roi : lastKnownStrategy1Roi,
        strategy2: s2Point ? s2Point.roi : lastKnownStrategy2Roi,
        timestamp: timestamp
      };
    });
    
    console.log("Combined chart data:", combined);
    setChartData(combined);
  }, [strategy1Performance, strategy2Performance]);
  
  const updatedStrategy1 = {
    ...strategy1Data,
    currentBalance: strategy1Balance ?? strategy1Data.currentBalance,
    roi: strategy1Roi ?? strategy1Data.roi,
    totalTrades: strategy1TotalTrades ?? strategy1Data.totalTrades
  };
  
  const updatedStrategy2 = {
    ...strategy2Data,
    currentBalance: strategy2Balance ?? strategy2Data.currentBalance,
    roi: strategy2Roi ?? strategy2Data.roi,
    totalTrades: strategy2TotalTrades ?? strategy2Data.totalTrades
  };
  
  // Compare which strategy is currently winning based on latest ROI
  const strategy1IsWinning = (strategy1Roi ?? strategy1Data.roi) > (strategy2Roi ?? strategy2Data.roi);
  
  return (
    <div className="space-y-6">
        <AICommentary strategy1={updatedStrategy1} strategy2={updatedStrategy2} />
      
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-lg font-semibold sm:text-xl">Strategy Battle</h2>
          
          <Tabs 
            value={activeTab} 
            onValueChange={setActiveTab}
            className="mt-2 w-full sm:mt-0 sm:w-auto"
          >
            <TabsList className="grid w-full grid-cols-2 sm:w-[200px]">
              <TabsTrigger value="cards">Cards</TabsTrigger>
              <TabsTrigger value="chart">Chart</TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
        
        {activeTab === "cards" ? (
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <StrategyCard 
              strategy={updatedStrategy1}
              trades={strategy1Trades}
              isWinning={strategy1IsWinning}
            />            
            <StrategyCard 
              strategy={updatedStrategy2} 
              trades={strategy2Trades} 
              isWinning={!strategy1IsWinning}
            />
          </div>
        ) : (
          <Card className="overflow-hidden">
            <CardHeader className="bg-muted/30 pb-2 pt-3 sm:pb-3 sm:pt-4">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base sm:text-lg">Performance Comparison</CardTitle>
                <Badge className="bg-primary/20 text-primary">Live Data</Badge>
              </div>
            </CardHeader>
            <CardContent className="p-2 pt-4 sm:p-4 md:p-6">
              <div className="flex flex-wrap items-center justify-between gap-2 pb-4">
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 rounded-full bg-primary"></div>
                  <span className="text-sm font-medium">
                    Strategy #{updatedStrategy1.strategyId} ({updatedStrategy1.roi.toFixed(1)}%)
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 rounded-full bg-muted-foreground"></div>
                  <span className="text-sm font-medium">
                    Strategy #{updatedStrategy2.strategyId} ({updatedStrategy2.roi.toFixed(1)}%)
                  </span>
                </div>
              </div>
              <div className="h-[300px] sm:h-[350px] md:h-[400px]">
                <PerformanceChart 
                  data={chartData} 
                  strategy1Name={`Strategy #${updatedStrategy1.strategyId}`}
                  strategy2Name={`Strategy #${updatedStrategy2.strategyId}`}
                />
              </div>
              <div className="mt-4 rounded-md bg-muted/30 p-3 text-xs text-muted-foreground sm:text-sm">
                <div className="flex items-center gap-1">
                  <LineChart className="h-3 w-3" />
                  <span>Chart shows ROI% growth over simulation time</span>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}