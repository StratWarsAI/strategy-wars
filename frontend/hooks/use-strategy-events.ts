
import { useEffect, useState } from 'react';
import { 
  TradeExecutedEvent, 
  TradeClosedEvent, 
  SimulationStatusEvent 
} from '@/types';
import { useWebSocket } from '@/lib/context/web-socket-context';

export function useStrategyEvents(strategyId: number) {
  const { lastMessage, isConnected } = useWebSocket();
  const [trades, setTrades] = useState<TradeExecutedEvent[]>([]);
  const [completedTrades, setCompletedTrades] = useState<TradeClosedEvent[]>([]);
  const [simulationStatus, setSimulationStatus] = useState<SimulationStatusEvent | null>(null);
  
  useEffect(() => {
    if (!lastMessage) return;
    
    // Filter messages for this strategy
    if (lastMessage.type === 'trade_executed' && lastMessage.strategy_id === strategyId) {
      setTrades(prev => [...prev, lastMessage as TradeExecutedEvent]);
    }
    
    if (lastMessage.type === 'trade_closed' && lastMessage.strategy_id === strategyId) {
      setCompletedTrades(prev => [...prev, lastMessage as TradeClosedEvent]);
    }
    
    if (lastMessage.type === 'simulation_status' && lastMessage.strategy_id === strategyId) {
      setSimulationStatus(lastMessage as SimulationStatusEvent);
    }
  }, [lastMessage, strategyId]);
  
  return {
    isConnected,
    trades,
    completedTrades,
    simulationStatus
  };
}