'use client';

import React, { useState } from 'react';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { StrategyCard } from './strategy-card';
import { SimulationSummary } from '@/types';
import { useStrategyEvents } from '@/hooks/use-strategy-events';

interface StrategyModuleProps {
    strategy1Data: SimulationSummary;
    strategy2Data: SimulationSummary;
}

export function StrategyModule({strategy1Data, strategy2Data}: StrategyModuleProps) {
  const [activeTab, setActiveTab] = useState("cards");

  const { 
    sortedTrades: strategy1Trades, 
    currentBalance: strategy1Balance, 
    currentRoi: strategy1Roi,
    totalTrades: strategy1TotalTrades
  } = useStrategyEvents(strategy1Data.strategyId);
  
  const { 
    sortedTrades: strategy2Trades, 
    currentBalance: strategy2Balance, 
    currentRoi: strategy2Roi,
    totalTrades: strategy2TotalTrades
  } = useStrategyEvents(strategy2Data.strategyId);
  
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
  
  const strategy1IsWinning = (strategy1Roi ?? strategy1Data.roi) > (strategy2Roi ?? strategy2Data.roi);
  
  return (
    <div className="space-y-6">
      {/* <AICommentary /> */}
      
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
          <div></div>
        )}
      </div>
    </div>
  );
}