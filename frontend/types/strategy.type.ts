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