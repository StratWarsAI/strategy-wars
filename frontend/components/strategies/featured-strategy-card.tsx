'use client';

import { Strategy } from '@/types';
import { ArrowUpRight, TrendingUp, Users, Clock, Star } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Separator } from '@/components/ui/separator';

interface FeaturedStrategyCardProps {
  strategy: Strategy;
  onClick?: () => void;
}

export function FeaturedStrategyCard({ strategy, onClick }: FeaturedStrategyCardProps) {
  const {
    name,
    description,
    metrics,
    voteCount,
    winCount,
    lastWinTime,
    tags,
    aiEnhanced
  } = strategy;

  // Format date
  const lastWin = lastWinTime ? new Date(lastWinTime) : null;
  const formattedLastWin = lastWin ? 
    `${lastWin.toLocaleDateString()} at ${lastWin.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}` : 
    'No wins yet';

  // Calculate time since creation
  const createdDate = strategy.createdAt ? new Date(strategy.createdAt) : new Date();
  const now = new Date();
  const daysSinceCreation = Math.floor((now.getTime() - createdDate.getTime()) / (1000 * 60 * 60 * 24));
  
  return (
    <Card 
      className={`group relative overflow-hidden border-2 border-primary/50 bg-gradient-to-br from-card to-secondary/10 transition-all duration-300 hover:shadow-lg hover:shadow-primary/10 ${onClick ? 'cursor-pointer' : ''}`}
      onClick={onClick}
    >
      {/* Background decorative elements */}
      <div className="absolute right-0 top-0 -z-10 h-40 w-40 translate-x-1/4 -translate-y-1/4 rounded-full bg-primary/5 blur-3xl"></div>
      <div className="absolute bottom-0 left-0 -z-10 h-40 w-40 -translate-x-1/4 translate-y-1/4 rounded-full bg-primary/5 blur-3xl"></div>
      
      {/* Featured badge */}
      <div className="absolute -right-12 -top-3 rotate-45 bg-primary px-12 py-1 text-xs font-medium text-primary-foreground shadow-md">
        FEATURED
      </div>
      
      <CardHeader className="pb-2 pt-4">
        <div className="flex items-start gap-3">
          <div className="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <Star className="h-6 w-6" />
          </div>
          
          <div className="flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <CardTitle className="text-xl md:text-2xl">{name}</CardTitle>
              {aiEnhanced && (
                <Badge className="border-indigo-500/30 bg-indigo-500/20 text-xs text-indigo-500">
                  AI ENHANCED
                </Badge>
              )}
            </div>
            <CardDescription className="mt-1 line-clamp-2">{description}</CardDescription>
          </div>
        </div>
      </CardHeader>
      
      <CardContent>
        <div className="mt-4 grid grid-cols-1 gap-6 md:grid-cols-2">
          {/* Performance metrics */}
          <div className="space-y-4 rounded-xl bg-card/50 p-4 backdrop-blur-sm">
            <h3 className="flex items-center gap-1.5 text-sm font-medium">
              <TrendingUp className="h-4 w-4 text-primary" /> Performance Metrics
            </h3>
            
            <div className="grid grid-cols-2 gap-3">
              <div>
                <div className="text-xs text-muted-foreground">ROI</div>
                <div className="flex items-center gap-1 text-xl font-bold text-green-500">
                  <ArrowUpRight className="h-4 w-4" />
                  +{metrics?.roi?.toFixed(2)}%
                </div>
              </div>
              
              <div>
                <div className="text-xs text-muted-foreground">Win Rate</div>
                <div className="text-xl font-bold">{metrics?.winRate?.toFixed(1)}%</div>
                <Progress 
                  value={metrics?.winRate || 0} 
                  className={`mt-1 h-1.5 ${
                    (metrics?.winRate || 0) > 70 ? 'bg-green-500' : (metrics?.winRate || 0) > 50 ? 'bg-amber-500' : 'bg-red-500'
                  }`} 
                />
              </div>
              
              <div>
                <div className="text-xs text-muted-foreground">Total Trades</div>
                <div className="text-xl font-bold">{metrics?.totalTrades || 0}</div>
              </div>
              
              <div>
                <div className="text-xs text-muted-foreground">Profit Factor</div>
                <div className="text-xl font-bold">{metrics?.profitFactor?.toFixed(2) || 0}</div>
              </div>
            </div>
            
            <Separator className="bg-primary/10" />
            
            <div className="grid grid-cols-2 gap-y-1 text-sm">
              <div className="text-muted-foreground">Initial Balance:</div>
              <div className="font-medium">{metrics?.initialBalance || 0} SOL</div>
              
              <div className="text-muted-foreground">Current Balance:</div>
              <div className="font-medium">{metrics?.balance || 0} SOL</div>
              
              <div className="text-muted-foreground">Sharpe Ratio:</div>
              <div className="font-medium">{metrics?.sharpeRatio?.toFixed(2) || 'N/A'}</div>
            </div>
          </div>
          
          {/* Community stats */}
          <div className="space-y-4 rounded-xl bg-card/50 p-4 backdrop-blur-sm">
            <h3 className="flex items-center gap-1.5 text-sm font-medium">
              <Users className="h-4 w-4 text-primary" /> Community Stats
            </h3>
            
            <div className="grid grid-cols-2 gap-4">
              <div className="rounded-lg bg-background/50 p-3">
                <div className="text-xs text-muted-foreground">Votes</div>
                <div className="text-xl font-bold">{voteCount}</div>
              </div>
              
              <div className="rounded-lg bg-background/50 p-3">
                <div className="text-xs text-muted-foreground">Wins</div>
                <div className="text-xl font-bold">{winCount}</div>
              </div>
            </div>
            
            <div className="rounded-lg bg-background/50 p-3">
              <div className="text-xs text-muted-foreground">Last Win</div>
              <div className="flex items-center gap-1.5 text-sm">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                {formattedLastWin}
              </div>
            </div>
            
            <div className="rounded-lg bg-background/50 p-3">
              <div className="text-xs text-muted-foreground">Strategy Age</div>
              <div className="flex items-center gap-1.5 text-sm">
                <Calendar className="h-3.5 w-3.5 text-muted-foreground" />
                {daysSinceCreation} days
              </div>
            </div>
            
            {tags && tags.length > 0 && (
              <div>
                <div className="mb-1.5 text-xs text-muted-foreground">Tags</div>
                <div className="flex flex-wrap gap-1.5">
                  {tags.map((tag, index) => (
                    <span 
                      key={index}
                      className="rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
        
        {/* Call to action */}
        <div className="mt-6 flex items-center justify-center">
          <button 
            className="flex items-center gap-1 rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
            onClick={(e) => {
              e.stopPropagation();
              onClick && onClick();
            }}
          >
            View Strategy Details
            <ArrowUpRight className="h-4 w-4" />
          </button>
        </div>
      </CardContent>
      
      {/* Bottom glowing border animation */}
      <div className="absolute bottom-0 left-0 h-0.5 w-full overflow-hidden">
        <div className="animate-shimmer h-full w-1/3 -translate-x-full bg-gradient-to-r from-transparent via-primary to-transparent"></div>
      </div>
    </Card>
  );
}

// Calendar icon component for the strategy age
function Calendar(props: React.ComponentProps<typeof Clock>) {
  return <Clock {...props} />;
}