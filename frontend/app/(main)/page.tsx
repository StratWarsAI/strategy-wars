import { Suspense } from 'react';
import Container from '@/components/layout/container';
import { getCompleteDashboard } from '@/lib/api';
import Dashboard from '@/components/dashboard/dashboard';
import { SearchParams } from '@/types';

// Disable static generation for this page
export const dynamic = 'force-dynamic';


export default async function HomePage({
  searchParams,
}: { searchParams: Promise<SearchParams> }) {
  const params = await searchParams;
  // Get timeframe from URL query or default to 24h
  const timeframe = typeof params.timeframe === 'string' ? params.timeframe : '24h';
  
  // Fetch initial dashboard data server-side
  let initialData;
  try {
    initialData = await getCompleteDashboard(timeframe);
  } catch (error) {
    console.error("Failed to fetch dashboard data:", error);
    initialData = {
      stats: {
        total_balance: 0,
        balance_change: 0,
        balance_change_percent: 0,
        total_profits: 0,
        total_trades: 0,
        winning_trades: 0,
        losing_trades: 0,
        win_rate: 0,
        active_trade_count: 0,
        avg_hold_time: "0h 0m",
        top_performer: {
          id: 0,
          name: "N/A",
          roi: 0,
          trades: 0
        },
        market_status: "neutral",
        volatility_index: 0,
        last_updated: new Date().toISOString()
      },
      charts: {
        performance_data: [],
        strategy_data: [],
        recent_stats: []
      }
    };
  }

  return (
    <Container>
      <Suspense fallback={<div>Loading dashboard...</div>}>
        <Dashboard initialData={initialData} defaultTimeframe={timeframe} />
      </Suspense>
    </Container>
  );
}