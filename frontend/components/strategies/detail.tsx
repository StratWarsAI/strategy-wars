'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { startSimulation } from '@/lib/api';
import { Strategy } from '@/types';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import { ArrowLeft, PlayCircle, BarChart2, Activity, Zap, Clock, Info } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { toast } from 'sonner';

interface StrategyDetailProps {
    strategy: Strategy;
    initialSimulationRunning: boolean;
}

export function StrategyDetail({ 
    strategy, 
    initialSimulationRunning 
}: StrategyDetailProps) {
    const router = useRouter();
    const [simulating, setSimulating] = useState(false);
    const [isRunning, setIsRunning] = useState(initialSimulationRunning);
    
    const handleStartSimulation = async () => {
        try {
            setSimulating(true);
            await startSimulation(strategy.id);
            setIsRunning(true);
        } catch (error) {
            console.error('Failed to start simulation:', error);
        } finally {
            setSimulating(false);
        }
    };
    
    const handleBackClick = () => {
        router.back();
    };
    
    return (
        <div className="container mx-auto px-4 py-8">
            <div className="mb-6 flex items-center justify-between">
                <Button variant="outline" onClick={handleBackClick}>
                    <ArrowLeft className="mr-2 h-4 w-4" />
                    Back
                </Button>
            </div>
            
            <div className="mb-6">
                <h1 className="mb-2 text-2xl font-bold">{strategy.name}</h1>
                <div className="flex flex-wrap gap-2">
                    {strategy.tags?.map((tag, idx) => (
                        <Badge key={idx} variant="secondary">
                            {tag}
                        </Badge>
                    ))}
                    
                    {strategy.aiEnhanced && (
                        <Badge className="border-indigo-500/30 bg-indigo-500/20 text-indigo-500">
                            AI Enhanced
                        </Badge>
                    )}
                </div>
                <p className="mt-4 text-muted-foreground">{strategy.description}</p>
            </div>
            
            <Separator className="my-8" />
            
            {/* Strategy Metrics */}
            <div className="mb-10">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                    <BarChart2 className="h-5 w-5" />
                    Performance Metrics
                </h2>
                
                <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
                    {/* ROI */}
                    <Card className="overflow-hidden">
                        <div className="bg-primary/10 py-2 text-center">
                            <h3 className="text-sm font-medium">Return on Investment</h3>
                        </div>
                        <CardContent className="flex flex-col items-center justify-center pt-6">
                            <div className={`text-3xl font-bold ${(strategy.metrics?.roi || 0) > 0 ? 'text-green-500' : 'text-red-500'}`}>
                                {strategy.metrics?.roi ? (strategy.metrics.roi > 0 ? '+' : '') + strategy.metrics.roi.toFixed(2) + '%' : 'N/A'}
                            </div>
                            <div className="mt-2 text-sm text-muted-foreground">
                                {strategy.metrics?.initialBalance ? `From ${strategy.metrics.initialBalance} SOL` : ''}
                            </div>
                        </CardContent>
                    </Card>
                    
                    {/* Win Rate */}
                    <Card className="overflow-hidden">
                        <div className="bg-primary/10 py-2 text-center">
                            <h3 className="text-sm font-medium">Win Rate</h3>
                        </div>
                        <CardContent className="pt-6">
                            <div className="flex flex-col items-center justify-center">
                                <div className="text-3xl font-bold">
                                    {strategy.metrics?.winRate ? strategy.metrics.winRate.toFixed(1) + '%' : 'N/A'}
                                </div>
                                <div className="mt-4 w-full">
                                    <Progress 
                                        value={strategy.metrics?.winRate || 0} 
                                        className={`h-2 ${
                                            (strategy.metrics?.winRate || 0) > 70 ? 'bg-green-500' : 
                                            (strategy.metrics?.winRate || 0) > 50 ? 'bg-amber-500' : 
                                            'bg-red-500'
                                        }`} 
                                    />
                                </div>
                                <div className="mt-2 text-sm text-muted-foreground">
                                    {strategy.metrics?.winningTrades || 0} / {strategy.metrics?.totalTrades || 0} trades
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                    
                    {/* Total Trades */}
                    <Card className="overflow-hidden">
                        <div className="bg-primary/10 py-2 text-center">
                            <h3 className="text-sm font-medium">Trade Statistics</h3>
                        </div>
                        <CardContent className="flex flex-col gap-3 pt-6">
                            <div className="flex items-center justify-between">
                                <span className="text-sm">Total Trades:</span>
                                <span className="font-semibold">{strategy.metrics?.totalTrades || 0}</span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm">Avg. Profit:</span>
                                <span className={`font-semibold ${(strategy.metrics?.averageProfitPct || 0) > 0 ? 'text-green-500' : ''}`}>
                                    {strategy.metrics?.averageProfitPct ? strategy.metrics.averageProfitPct.toFixed(2) + '%' : 'N/A'}
                                </span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm">Avg. Loss:</span>
                                <span className="font-semibold text-red-500">
                                    {strategy.metrics?.averageLossPct ? strategy.metrics.averageLossPct.toFixed(2) + '%' : 'N/A'}
                                </span>
                            </div>
                        </CardContent>
                    </Card>
                    
                    {/* Performance Indicators */}
                    <Card className="overflow-hidden">
                        <div className="bg-primary/10 py-2 text-center">
                            <h3 className="text-sm font-medium">Performance Indicators</h3>
                        </div>
                        <CardContent className="flex flex-col gap-3 pt-6">
                            <div className="flex items-center justify-between">
                                <span className="text-sm">Profit Factor:</span>
                                <span className="font-semibold">
                                    {strategy.metrics?.profitFactor ? strategy.metrics.profitFactor.toFixed(2) : 'N/A'}
                                </span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm">Sharpe Ratio:</span>
                                <span className="font-semibold">
                                    {strategy.metrics?.sharpeRatio ? strategy.metrics.sharpeRatio.toFixed(2) : 'N/A'}
                                </span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm">Current Balance:</span>
                                <span className="font-semibold">
                                    {strategy.metrics?.balance ? strategy.metrics.balance.toFixed(2) + ' SOL' : 'N/A'}
                                </span>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
            
            <Separator className="my-8" />
            
            {/* Strategy Configuration */}
            <div className="mb-10">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                    <Zap className="h-5 w-5" />
                    Strategy Configuration
                </h2>
                
                <Card>
                    <CardContent className="grid grid-cols-1 gap-4 pt-6 md:grid-cols-2 lg:grid-cols-3">
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Risk Level</div>
                            <div className="text-base font-medium">
                                {strategy.config.riskLevel ? 
                                    strategy.config.riskLevel.charAt(0).toUpperCase() + strategy.config.riskLevel.slice(1) : 
                                    strategy.riskScore > 7 ? 'High' : strategy.riskScore > 4 ? 'Medium' : 'Low'
                                }
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Market Cap Threshold</div>
                            <div className="text-base font-medium">
                                ${strategy.config.marketCapThreshold?.toLocaleString()}
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Initial Balance</div>
                            <div className="text-base font-medium">
                                {strategy.config.initialBalance} SOL
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Take Profit</div>
                            <div className="text-base font-medium">
                                {strategy.config.takeProfitPct}%
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Stop Loss</div>
                            <div className="text-base font-medium">
                                {strategy.config.stopLossPct}%
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Position Size</div>
                            <div className="text-base font-medium">
                                {strategy.config.fixedPositionSizeSol} SOL
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Min Buys For Entry</div>
                            <div className="text-base font-medium">
                                {strategy.config.minBuysForEntry}
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Entry Time Window</div>
                            <div className="text-base font-medium">
                                {(strategy.config.entryTimeWindowSec / 60).toFixed(0)} minutes
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Max Hold Time</div>
                            <div className="text-base font-medium">
                                {(strategy.config.maxHoldTimeSec / 3600).toFixed(1)} hours
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
            
            {/* Strategy Info */}
            <div className="mb-10">
                <h2 className="mb-4 flex items-center gap-2 text-xl font-semibold">
                    <Info className="h-5 w-5" />
                    Additional Information
                </h2>
                
                <Card>
                    <CardContent className="grid grid-cols-1 gap-4 pt-6 md:grid-cols-2 lg:grid-cols-4">
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Created</div>
                            <div className="text-base font-medium">
                                {strategy.createdAt ? new Date(strategy.createdAt).toLocaleDateString() : 'N/A'}
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Last Win</div>
                            <div className="text-base font-medium">
                                {strategy.lastWinTime ? new Date(strategy.lastWinTime).toLocaleDateString() : 'Never'}
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Win Count</div>
                            <div className="text-base font-medium">
                                {strategy.winCount || 0}
                            </div>
                        </div>
                        
                        <div className="rounded-lg bg-secondary/10 p-3">
                            <div className="text-xs text-muted-foreground">Vote Count</div>
                            <div className="text-base font-medium">
                                {strategy.voteCount || 0}
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}