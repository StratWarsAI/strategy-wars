import { SimulationSummary } from "@/types";
import { Strategy } from "@/types/strategy.type";
import { DashboardStats, DashboardCharts, CompleteDashboard } from "@/types/dashboard.type";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';

interface StrategyRaw {
  id: number;
  name: string;
  description: string;
  config: Record<string, any>;
  is_public: boolean;
  risk_score: number;
  created_at: string;
  updated_at: string;
}

/**
 * Fetch all public strategies
 */
export async function getStrategies(): Promise<Strategy[]> {
  const response = await fetch(`${API_BASE_URL}/strategies`);
  
  if (!response.ok) {
    throw new Error(`Failed to fetch strategies: ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Fetch a strategy by ID
 */
export async function getStrategy(id: number): Promise<Strategy> {
  const response = await fetch(`${API_BASE_URL}/strategies/${id}`);
  
  if (!response.ok) {
    throw new Error(`Failed to fetch strategy: ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Start a simulation for a strategy
 */
export async function startSimulation(strategyId: number): Promise<{ success: boolean; message: string }> {
  const response = await fetch(`${API_BASE_URL}/trigger/simulate/${strategyId}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({}),
  });
  
  if (!response.ok) {
    throw new Error(`Failed to start simulation: ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Stop a simulation for a strategy
 */
export async function stopSimulation(strategyId: number): Promise<{ success: boolean; message: string }> {
  const response = await fetch(`${API_BASE_URL}/trigger/stop/${strategyId}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({}),
  });
  
  if (!response.ok) {
    throw new Error(`Failed to stop simulation: ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Get simulation status for a strategy
 */
export async function getSimulationStatus(strategyId: number): Promise<{ success: boolean; running: boolean; message: string }> {
  const response = await fetch(`${API_BASE_URL}/trigger/status/${strategyId}`);
  
  if (!response.ok) {
    throw new Error(`Failed to get simulation status: ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Get simulation summary for a strategy
 */
export async function getSimulationSummary(strategyId: number): Promise<SimulationSummary | null> {
  const response = await fetch(`${API_BASE_URL}/simulations/summary/${strategyId}`);
  
  if (!response.ok) {
    throw new Error(`Failed to get simulation summary: ${response.statusText}`);
  }
  
  const data = await response.json();
  
  // If no simulation data available yet
  if (data.message && data.message.includes("No simulation")) {
    return null;
  }
  
  return data;
}

/**
 * Trigger AI performance analysis
 */
export async function triggerAnalysis(): Promise<{ success: boolean; message: string }> {
  const response = await fetch(`${API_BASE_URL}/trigger/analyze`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({}),
  });
  
  if (!response.ok) {
    throw new Error(`Failed to trigger analysis: ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Get WebSocket URL
 */
export function getWebSocketUrl(): string {
  return WS_URL;
}

/**
 * Fetch all running strategies
 */
export async function getRunningStrategies(): Promise<SimulationSummary[]> {
  const response = await fetch(`${API_BASE_URL}/simulations/running`);
  
  if (!response.ok) {
    throw new Error(`Failed to fetch running strategies: ${response.statusText}`);
  }
  
  return response.json();
}

/**
 * Fetch top performing strategies
 */
export async function fetchTopStrategies(): Promise<Strategy[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/strategies/top`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch top strategies: ${response.statusText}`);
    }
    
    return response.json();
  } catch (error) {
    console.error("Error fetching top strategies:", error);
    throw error;
  }
}

/**
 * Fetch dashboard statistics
 */
export async function getDashboardStats(timeframe: string = '24h'): Promise<DashboardStats> {
  try {
    const response = await fetch(`${API_BASE_URL}/dashboard/stats?timeframe=${timeframe}`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch dashboard stats: ${response.statusText}`);
    }
    
    return response.json();
  } catch (error) {
    console.error("Error fetching dashboard stats:", error);
    throw error;
  }
}

/**
 * Fetch dashboard chart data
 */
export async function getDashboardCharts(timeframe: string = '7d'): Promise<DashboardCharts> {
  try {
    const response = await fetch(`${API_BASE_URL}/dashboard/charts?timeframe=${timeframe}`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch dashboard charts: ${response.statusText}`);
    }
    
    return response.json();
  } catch (error) {
    console.error("Error fetching dashboard charts:", error);
    throw error;
  }
}

/**
 * Fetch complete dashboard data
 */
export async function getCompleteDashboard(timeframe: string = '24h'): Promise<CompleteDashboard> {
  try {
    const response = await fetch(`${API_BASE_URL}/dashboard/complete?timeframe=${timeframe}`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch complete dashboard: ${response.statusText}`);
    }
    
    return response.json();
  } catch (error) {
    console.error("Error fetching complete dashboard:", error);
    throw error;
  }
}