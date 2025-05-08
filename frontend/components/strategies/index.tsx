'use client';

import { useState } from 'react';
import { Strategy } from '@/types';
import { StrategyGrid } from './strategy-grid';
import { FeaturedStrategyCard } from './featured-strategy-card';
import { EnhancedStrategyCard } from './enhanced-strategy-card';
import { Separator } from '@/components/ui/separator';
import { useRouter } from 'next/navigation';

interface StrategiesPageProps {
    initialStrategies: Strategy[];
    featuredStrategy: Strategy | null;
    topStrategies: Strategy[];
    error: string | null;
}

export function StrategiesPage({ 
    initialStrategies, 
    featuredStrategy, 
    topStrategies,
    error 
}: StrategiesPageProps) {
    const router = useRouter();
    
    const [strategies, setStrategies] = useState<Strategy[]>(initialStrategies);
    
    const handleStrategyClick = (strategy: Strategy) => {
        router.push(`/strategies/${strategy.id}`);
    };
    
    if (error) {
        return (
            <div className="mt-8 rounded-lg border border-red-200 bg-red-50 p-4 text-center dark:border-red-900 dark:bg-red-950/20">
                <p className="text-red-800 dark:text-red-400">{error}</p>
                <button
                    onClick={() => router.refresh()}
                    className="mt-4 rounded-md bg-red-100 px-4 py-2 text-sm font-medium text-red-900 hover:bg-red-200 dark:bg-red-900/20 dark:text-red-300 dark:hover:bg-red-900/30"
                >
                    Try Again
                </button>
            </div>
        );
    }
    
    // Show loading state if no strategies were found
    if (strategies.length === 0) {
        return (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                    <div key={i} className="h-64 animate-pulse rounded-xl bg-secondary/20"></div>
                ))}
            </div>
        );
    }
    
    return (
        <>
            {/* Featured Strategy */}
            {featuredStrategy && (
                <div className="mb-10">
                    <h2 className="mb-4 text-xl font-semibold">Featured Strategy</h2>
                    <FeaturedStrategyCard 
                        strategy={featuredStrategy} 
                        onClick={() => handleStrategyClick(featuredStrategy)}
                    />
                </div>
            )}
            
            <Separator className="my-8" />
            
            {/* Top Performing Strategies */}
            <div className="mb-10">
                <h2 className="mb-4 text-xl font-semibold">Top Performing Strategies</h2>
                <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                    {topStrategies.map(strategy => (
                        <EnhancedStrategyCard
                            key={strategy.id}
                            strategy={strategy}
                            onClick={() => handleStrategyClick(strategy)}
                        />
                    ))}
                </div>
            </div>
            
            <Separator className="my-8" />
            
            {/* All Strategies */}
            <StrategyGrid 
                strategies={strategies} 
                onStrategyClick={handleStrategyClick} 
            />
        </>
    );
}