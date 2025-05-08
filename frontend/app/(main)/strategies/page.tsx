import { getStrategies } from '@/lib/api';
import { Separator } from '@/components/ui/separator';
import Container from '@/components/layout/container';
import { StrategiesPage } from '@/components/strategies';
import { SearchParams, Strategy } from '@/types';

export const dynamic = 'force-dynamic';

interface StrategiesPageProps {
  params: Promise<{}>;
  searchParams: Promise<SearchParams>;
}

export default async function Page({
  params,
  searchParams
}: StrategiesPageProps) {
    // Fetch strategies data on the server
    let strategies: Strategy[] = [];
    let error: string | null = null;
    
    try {
        strategies = await getStrategies();
        
        if (strategies.length > 0) {
            console.log('First strategy data structure:', JSON.stringify(strategies[0], null, 2));
        }
    } catch (err) {
        console.error("Failed to fetch strategies:", err);
        // Don't fail the build, just show an error message to the user
        error = "Failed to load strategies. Please try again later.";
        // Return empty strategies array to prevent build failures
        strategies = [];
    }
    
    // Select the featured strategy (highest ROI or most votes)
    let featuredStrategy: Strategy | null = null;
    if (strategies.length > 0) {
        const sorted = [...strategies].sort((a, b) => {
            // Sort by ROI first if available
            if (a.metrics?.roi && b.metrics?.roi) {
                return b.metrics.roi - a.metrics.roi;
            }
            // Then by vote count
            return (b.voteCount || 0) - (a.voteCount || 0);
        });
        featuredStrategy = sorted[0];
    }
    
    // Get top performing strategies
    const topStrategies: Strategy[] = strategies
        .filter(s => s.metrics?.roi && s.metrics.roi > 0)
        .sort((a, b) => (b.metrics?.roi || 0) - (a.metrics?.roi || 0))
        .slice(0, 3);
    
    
    // If there are no strategies with metrics, just use the available strategies
    if (topStrategies.length === 0 && strategies.length > 0) {
        // Just take the first 3 strategies (or fewer if there aren't 3)
        const availableStrategies: Strategy[] = [...strategies].slice(0, 3);
        
        if (availableStrategies.length > 0 && !featuredStrategy) {
            featuredStrategy = availableStrategies[0];
        }
    }
    
    // Render the client component with the server-fetched data
    return (
        <Container>
            <div className="px-4 py-8">            
                {/* Pass all data to client component which handles interactivity */}
                <StrategiesPage 
                    initialStrategies={strategies} 
                    featuredStrategy={featuredStrategy || (strategies.length > 0 ? strategies[0] : null)} 
                    topStrategies={topStrategies.length > 0 ? topStrategies : strategies.slice(0, 3)}
                    error={error}
                />
            </div>
        </Container>
    );
}