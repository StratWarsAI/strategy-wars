export interface Strategy {
    id: number;
    name: string;
    description: string;
    config: StrategyConfig;
    isPublic: boolean;
    voteCount: number;
    winCount: number;
    lastWinTime?: string;
    tags: string[];
    complexityScore: number;
    riskScore: number;
    aiEnhanced: boolean;
    createdAt?: string;
    updatedAt?: string;
    metrics?: StrategyMetrics;
}

export interface StrategyConfig {
    riskLevel?: string; // "low", "medium", "high"    
    marketCapThreshold: number;
    minBuysForEntry: number;
    entryTimeWindowSec: number;
    takeProfitPct: number;
    stopLossPct: number;
    maxHoldTimeSec: number;
    fixedPositionSizeSol: number;
    initialBalance: number;
}

export interface StrategyEvent {
    id: string; // Unique ID for the event
    type: 'buy' | 'sell' | 'info' | 'analysis';
    token?: {
        id: number;
        symbol: string;
        name: string;
        imageUrl: string;
        twitterUrl?: string;
        websiteUrl?: string;
    };
    price?: number;
    entryPrice?: number;
    exitPrice?: number;
    amount?: number;
    profitLoss?: number;
    profitLossPct?: number;
    reason?: string;
    timestamp: number;
    message?: string;
    marketCap?: number;
}

export interface StrategyMetrics {
    totalTrades: number;
    winningTrades: number;
    losingTrades: number;
    winRate: number;
    averageProfitPct: number;
    averageLossPct: number;
    largestWinPct: number;
    largestLossPct: number;
    balance: number;
    initialBalance: number;
    roi: number;
    profitFactor: number;
    sharpeRatio?: number;
}