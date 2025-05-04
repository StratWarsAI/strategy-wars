import { StrategyConfig } from "./strategy.type";

export interface SimulationRun {
    id: number;
    startTime: string;
    endTime: string;
    winnerStrategyId?: number;
    status: 'preparing' | 'running' | 'completed' | 'failed';
    simulationParameters: Record<string, any>;
    createdAt?: string;
    updatedAt?: string;
}

export interface SimulatedTrade {
    id: number;
    strategyId: number;
    tokenId: number;
    simulationRunId?: number;
    entryPrice: number;
    exitPrice?: number;
    entryTimestamp: number;
    exitTimestamp?: number;
    positionSize: number;
    profitLoss?: number;
    status: 'active' | 'completed' | 'canceled';
    exitReason?: string;
    entryUsdMarketCap: number;
    exitUsdMarketCap?: number;
    createdAt?: string;
    updatedAt?: string;
}

export interface SimulationEvent {
    id: number;
    strategyId: number;
    simulationRunId: number;
    eventType: string;
    eventData: Record<string, any>;
    timestamp: string;
    createdAt?: string;
}

export interface SimulationSummary {
    strategyId: number;
    strategyName: string;
    isRunning: boolean;
    startTime: number;
    executionTime: number;
    totalTrades: number;
    profitableTrades: number;
    losingTrades: number;
    winRate: number;
    totalProfit: number;
    totalLoss: number;
    avgProfit: number;
    avgLoss: number;
    maxDrawdown: number;
    netPnl: number;
    initialBalance: number;
    currentBalance: number;
    roi: number;
    activeTrades?: number;
    simConfig: StrategyConfig;
}