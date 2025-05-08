'use client';

import { useState } from 'react';
import { EnhancedStrategyCard } from './enhanced-strategy-card';
import { Strategy } from '@/types';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { TrendingUp, Calendar, Award, Flame } from 'lucide-react';

interface StrategyGridProps {
  strategies: Strategy[];
  onStrategyClick?: (strategy: Strategy) => void;
}

export function StrategyGrid({ strategies, onStrategyClick }: StrategyGridProps) {
  const [filter, setFilter] = useState<'all' | 'trending' | 'winning' | 'newest'>('all');
  
  // Mock data for demonstration (replace with actual data in production)
  const mockTrades = []; // Empty for simplicity, would be filled with actual trade data
  
  const filteredStrategies = () => {
    switch (filter) {
      case 'trending':
        return [...strategies].sort((a, b) => (b.metrics?.roi || 0) - (a.metrics?.roi || 0));
      case 'winning':
        return strategies.filter(s => (s.metrics?.roi || 0) > 0)
          .sort((a, b) => (b.winCount || 0) - (a.winCount || 0));
      case 'newest':
        return [...strategies].sort((a, b) => 
          new Date(b.createdAt || '').getTime() - new Date(a.createdAt || '').getTime()
        );
      default:
        return strategies;
    }
  };

  return (
    <div className="space-y-4">
      <Tabs defaultValue="all" className="w-full">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold md:text-xl">Trading Strategies</h2>
          <TabsList className="grid w-auto grid-cols-4 h-9">
            <TabsTrigger value="all" onClick={() => setFilter('all')}>
              All
            </TabsTrigger>
            <TabsTrigger value="trending" onClick={() => setFilter('trending')}>
              <Flame className="mr-1 h-3.5 w-3.5" />
              Trending
            </TabsTrigger>
            <TabsTrigger value="winning" onClick={() => setFilter('winning')}>
              <Award className="mr-1 h-3.5 w-3.5" />
              Winning
            </TabsTrigger>
            <TabsTrigger value="newest" onClick={() => setFilter('newest')}>
              <Calendar className="mr-1 h-3.5 w-3.5" />
              Newest
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="all" className="mt-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filteredStrategies().map((strategy) => (
              <EnhancedStrategyCard
                key={strategy.id}
                strategy={strategy}
                isWinning={strategy.winCount > 0}
                onClick={() => onStrategyClick && onStrategyClick(strategy)}
              />
            ))}
          </div>
        </TabsContent>
        
        <TabsContent value="trending" className="mt-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filteredStrategies().map((strategy) => (
              <EnhancedStrategyCard
                key={strategy.id}
                strategy={strategy}
                isWinning={strategy.winCount > 0}
                onClick={() => onStrategyClick && onStrategyClick(strategy)}
              />
            ))}
          </div>
        </TabsContent>
        
        <TabsContent value="winning" className="mt-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filteredStrategies().map((strategy) => (
              <EnhancedStrategyCard
                key={strategy.id}
                strategy={strategy}
                isWinning={true}
                onClick={() => onStrategyClick && onStrategyClick(strategy)}
              />
            ))}
          </div>
        </TabsContent>
        
        <TabsContent value="newest" className="mt-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filteredStrategies().map((strategy) => (
              <EnhancedStrategyCard
                key={strategy.id}
                strategy={strategy}
                isWinning={strategy.winCount > 0}
                onClick={() => onStrategyClick && onStrategyClick(strategy)}
              />
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}