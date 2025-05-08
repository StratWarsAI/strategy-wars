import { StrategyModule } from "@/components/strategy-ai/strategy-module";
import Container from "@/components/layout/container";
import { getRunningStrategies } from "@/lib/api";
import { QueryKeys } from "@/lib/query/query-keys";
import { dehydrate, HydrationBoundary, QueryClient } from "@tanstack/react-query";
import { SearchParams, SimulationSummary } from "@/types";

export const dynamic = 'force-dynamic';

interface StrategyAIPageProps {
  params: Promise<{}>;
  searchParams: Promise<SearchParams>;
}

export default async function Page({
  params,
  searchParams
}: StrategyAIPageProps) {
    const queryClient = new QueryClient();
    
    let runningStrategies: SimulationSummary[] = [];
    try {
      runningStrategies = await getRunningStrategies();
    } catch (error) {
      console.error("Failed to fetch running strategies:", error);
      runningStrategies = [];
    }
    
    const runningStrategiesKeys = new QueryKeys('runningStrategies');
    await queryClient.prefetchQuery({
      queryKey: runningStrategiesKeys.all(),
      queryFn: () => runningStrategies,
    });
    
    console.log("Running strategies:", runningStrategies);

    return (
        <Container>
            <HydrationBoundary state={dehydrate(queryClient)}>
                <StrategyModule strategy1Data={runningStrategies[0]} strategy2Data={runningStrategies[1]}   />
            </HydrationBoundary>
        </Container>
    )
}