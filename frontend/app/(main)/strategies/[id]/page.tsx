import { getStrategy, getSimulationStatus } from '@/lib/api';
import { notFound } from 'next/navigation';
import { StrategyDetail } from '@/components/strategies/detail';
import { SearchParams } from '@/types';

export const dynamic = 'force-dynamic';

interface StrategyDetailPageProps {
  params: Promise<{ id: string }>;
  searchParams: Promise<SearchParams>;
}

export default async function StrategyDetailPage({
  params,
  searchParams
}: StrategyDetailPageProps) {
  try {
    const { id } = await params;
    const strategyId = parseInt(id);

    if (isNaN(strategyId)) {
      notFound();
    }

    const strategy = await getStrategy(strategyId);
    const simulationStatus = await getSimulationStatus(strategyId);

    if (!strategy.metrics) {
      return notFound();
    }

    return (
      <StrategyDetail
        strategy={strategy}
        initialSimulationRunning={simulationStatus.running}
      />
    );
  } catch (error) {
    notFound();
  }
}
