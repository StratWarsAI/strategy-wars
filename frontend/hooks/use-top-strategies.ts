'use client';

import { useEffect, useState } from 'react';
import { useWebSocket } from '@/lib/context/web-socket-context';
import { Strategy } from '@/types/strategy.type';
import { SimulationStatusEvent } from '@/types/websocket.type';
import { fetchTopStrategies } from '@/lib/api';

/**
 * Custom hook for managing top strategies with real-time updates
 */
export function useTopStrategies() {
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const { lastMessage, isConnected } = useWebSocket();

  // Fetch initial strategies
  useEffect(() => {
    const loadStrategies = async () => {
      try {
        setIsLoading(true);
        const data = await fetchTopStrategies();
        setStrategies(data);
        setError(null);
      } catch (err) {
        console.error('Error fetching strategies:', err);
        setError(err instanceof Error ? err : new Error('Failed to fetch strategies'));
        // Hata durumunda da loading'i kapatıyoruz
        setIsLoading(false);
      } finally {
        setIsLoading(false);
      }
    };

    loadStrategies();

    // Set up a refresh interval (every 10 seconds)
    const interval = setInterval(loadStrategies, 10000);
    return () => clearInterval(interval);
  }, []);

  // Listen for WebSocket updates
  useEffect(() => {
    if (lastMessage && lastMessage.type === 'simulation_status') {
      const event = lastMessage as SimulationStatusEvent;
      
      setStrategies((prevStrategies) => 
        prevStrategies.map((strategy) => {
          if (strategy.id === event.strategy_id) {
            // Create a sound effect for significant changes
            if (Math.abs((strategy.metrics?.roi || 0) - event.roi) > 0.02) {
              playUpdateSound(event.roi > (strategy.metrics?.roi || 0));
            }
            
            return {
              ...strategy,
              metrics: {
                ...strategy.metrics,
                totalTrades: event.totalTrades,
                winningTrades: event.profitableTrades,
                losingTrades: event.losingTrades,
                winRate: event.winRate, // Backend değerini doğrudan alıyoruz
                roi: event.roi, // Backend değerini doğrudan alıyoruz
                balance: event.currentBalance,
                initialBalance: event.initialBalance
              }
            };
          }
          return strategy;
        })
      );
    }
  }, [lastMessage]);

  // Play a sound effect for updates
  const playUpdateSound = (isPositive: boolean) => {
    try {
      const audio = new Audio(isPositive 
        ? '/sounds/positive.mp3' 
        : '/sounds/negative.mp3');
      audio.volume = 0.2;
      audio.play().catch(err => console.log('Audio play prevented:', err));
    } catch (err) {
      console.log('Audio playback error:', err);
    }
  };

  // Get strategies sorted by ROI
  const getSortedStrategies = () => {
    return [...strategies].sort((a, b) => 
      (b.metrics?.roi || 0) - (a.metrics?.roi || 0)
    );
  };

  return {
    strategies: getSortedStrategies(),
    isLoading,
    error,
    isConnected,
    refresh: async () => {
      try {
        setIsLoading(true);
        const data = await fetchTopStrategies();
        setStrategies(data);
        setError(null);
        console.log("Strategies refreshed successfully:", data);
      } catch (err) {
        console.error("Error refreshing strategies:", err);
        setError(err instanceof Error ? err : new Error('Failed to refresh strategies'));
      } finally {
        setIsLoading(false);
      }
    }
  };
}