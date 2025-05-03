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