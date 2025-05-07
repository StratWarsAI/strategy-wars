'use client'

import { useEffect, useState } from 'react';
import { 
  WebSocketMessage,
  TradeExecutedEvent, 
  TradeClosedEvent, 
  SimulationStatusEvent,
  SimulationStartedEvent,
  SimulationCompletedEvent
} from '@/types/websocket.type';
import { useWebSocket } from '@/lib/context/web-socket-context';
import { StrategyEvent } from '@/types';

export interface PerformanceDataPoint {
  time: string;
  roi: number;
  balance: number;
  winRate?: number;
  timestamp: number;
}

export function useStrategyEvents(strategyId: number | null) {
  const { lastMessage, isConnected } = useWebSocket();
  const [events, setEvents] = useState<StrategyEvent[]>([]);
  const [trades, setTrades] = useState<TradeExecutedEvent[]>([]);
  const [completedTrades, setCompletedTrades] = useState<TradeClosedEvent[]>([]);
  const [simulationStatus, setSimulationStatus] = useState<SimulationStatusEvent | null>(null);
  const [isRunning, setIsRunning] = useState(false);
  
  const [currentBalance, setCurrentBalance] = useState<number | null>(null);
  const [currentRoi, setCurrentRoi] = useState<number | null>(null);
  const [totalTrades, setTotalTrades] = useState<number | null>(null);
  const [activeTrades, setActiveTrades] = useState<number | null>(null);
  const [winRate, setWinRate] = useState<number | null>(null);
  
  // Performance tracking
  const [performanceData, setPerformanceData] = useState<PerformanceDataPoint[]>([]);
  
  // Subscribe to a strategy when component mounts
  useEffect(() => {
    // Clear events when strategy changes
    setEvents([]);
    setTrades([]);
    setCompletedTrades([]);
    setSimulationStatus(null);
    setIsRunning(false);
    setCurrentBalance(null);
    setCurrentRoi(null);
    setTotalTrades(null);
    setActiveTrades(null);
    setWinRate(null);
    setPerformanceData([]);
    
    // Add initial data point with 0 ROI
    if (strategyId) {
      const initialPoint: PerformanceDataPoint = {
        time: new Date().toLocaleTimeString([], { 
          hour: '2-digit', 
          minute: '2-digit',
          second: '2-digit'
        }),
        roi: 0,
        balance: 0,
        winRate: 0,
        timestamp: Math.floor(Date.now() / 1000)
      };
      setPerformanceData([initialPoint]);
    }
  }, [strategyId]);

  // Process new messages
  useEffect(() => {
    if (!lastMessage || !strategyId) return;

    // Only process messages for the selected strategy
    if (lastMessage.strategy_id !== strategyId) return;

    processWebSocketMessage(lastMessage);
  }, [lastMessage, strategyId]);
  
  // Debug log performance data
  useEffect(() => {
    if (strategyId && performanceData.length > 0) {
      console.log(`Performance data for strategy ${strategyId}:`, performanceData);
    }
  }, [performanceData, strategyId]);

  const processWebSocketMessage = (message: WebSocketMessage) => {
    // For trades, we keep the original processing logic
    if (message.type === 'trade_executed' && message.strategy_id === strategyId) {
      setTrades(prev => {
        const exists = prev.some(t => 
          t.tokenId === (message as TradeExecutedEvent).tokenId && 
          t.timestamp === message.timestamp
        );
        return exists ? prev : [...prev, message as TradeExecutedEvent];
      });
    }
    
    if (message.type === 'trade_closed' && message.strategy_id === strategyId) {
      setCompletedTrades(prev => {
        const exists = prev.some(t => 
          t.tokenId === (message as TradeClosedEvent).tokenId && 
          t.timestamp === message.timestamp
        );
        return exists ? prev : [...prev, message as TradeClosedEvent];
      });
    }
    
    if (message.type === 'simulation_status' && message.strategy_id === strategyId) {
      const statusEvent = message as SimulationStatusEvent;
      setSimulationStatus(statusEvent);
      
      // Update state values
      setCurrentBalance(statusEvent.currentBalance);
      setCurrentRoi(statusEvent.roi);
      setTotalTrades(statusEvent.totalTrades);
      setActiveTrades(statusEvent.activeTrades);
      setWinRate(statusEvent.winRate);
      
      // Format time for display
      const timeStr = new Date(statusEvent.timestamp * 1000).toLocaleTimeString([], { 
        hour: '2-digit', 
        minute: '2-digit',
        second: '2-digit'
      });
      
      // Add new data point for performance chart
      setPerformanceData(prev => {
        // Check if we already have a data point with this timestamp
        const exists = prev.some(p => p.timestamp === statusEvent.timestamp);
        if (exists) return prev;
        
        // Add new data point and ensure they are sorted by timestamp
        const updatedData = [
          ...prev,
          {
            time: timeStr,
            roi: statusEvent.roi,
            balance: statusEvent.currentBalance,
            winRate: statusEvent.winRate,
            timestamp: statusEvent.timestamp
          }
        ].sort((a, b) => a.timestamp - b.timestamp);
        
        return updatedData;
      });
    }

    // Enhanced event processing for the unified events array
    switch (message.type) {
      case 'simulation_started':
        const startEvent = message as SimulationStartedEvent;
        setIsRunning(true);
        addEvent({
          id: `start-${startEvent.timestamp}`,
          type: 'info',
          timestamp: startEvent.timestamp,
          message: 'Simulation started',
        });
        break;

      case 'simulation_completed':
        const completeEvent = message as SimulationCompletedEvent;
        setIsRunning(false);
        addEvent({
          id: `complete-${completeEvent.timestamp}`,
          type: 'info',
          timestamp: completeEvent.timestamp,
          message: `Simulation completed. Evaluated tokens in ${completeEvent.executionTimeSec.toFixed(1)}s`,
        });
        break;

      case 'trade_executed':
        const tradeEvent = message as TradeExecutedEvent;
        addEvent({
          id: `buy-${tradeEvent.tokenId}-${tradeEvent.timestamp}`,
          type: 'buy',
          token: {
            id: tradeEvent.tokenId,
            symbol: tradeEvent.tokenSymbol || '',
            name: tradeEvent.tokenName || '',
            imageUrl: tradeEvent.imageUrl || '',
            websiteUrl: tradeEvent.websiteUrl || '',
            twitterUrl: tradeEvent.twitterUrl || '',
          },
          price: tradeEvent.price,
          amount: tradeEvent.amount,
          timestamp: tradeEvent.timestamp,
          marketCap: tradeEvent.entryMarketCap || tradeEvent.usdMarketCap,
        });
        break;

      case 'trade_closed':
        const closeEvent = message as TradeClosedEvent;
        addEvent({
          id: `sell-${closeEvent.tokenId}-${closeEvent.timestamp}`,
          type: 'sell',
          token: {
            id: closeEvent.tokenId,
            symbol: closeEvent.tokenSymbol || '',
            name: closeEvent.tokenName || '',  
            imageUrl: closeEvent.imageUrl || '',   
            websiteUrl: closeEvent.websiteUrl || '',
            twitterUrl: closeEvent.twitterUrl || '',  
          },
          entryPrice: closeEvent.entryPrice,
          exitPrice: closeEvent.exitPrice,
          profitLoss: closeEvent.profitLoss,
          profitLossPct: closeEvent.profitLossPct,
          reason: closeEvent.exitReason,
          timestamp: closeEvent.timestamp,
          marketCap: closeEvent.exitMarketCap || closeEvent.usdMarketCap,
        });
        break;
    }
  };

  const addEvent = (event: StrategyEvent) => {
    setEvents(prev => [event, ...prev]);
  };
  
  // Get sorted trades
  const getSortedTrades = () => {
    return [...trades, ...completedTrades]
      .sort((a, b) => b.timestamp - a.timestamp); 
  };
  
  return {
    isConnected,
    trades,
    completedTrades,
    simulationStatus,
    events,
    isRunning,
    clearEvents: () => setEvents([]),
    sortedTrades: getSortedTrades(),
    currentBalance,
    currentRoi,
    totalTrades,
    activeTrades,
    winRate,
    performanceData 
  };
}