import { StrategyModule } from "@/components/strategy-ai/strategy-module";
import Container from "@/components/layout/container";
import { getRunningStrategies } from "@/lib/api";
import { QueryKeys } from "@/lib/query/query-keys";
import { dehydrate, HydrationBoundary, QueryClient } from "@tanstack/react-query";

export default async function Page() {
    const queryClient = new QueryClient();
    
    const runningStrategies = await getRunningStrategies();
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