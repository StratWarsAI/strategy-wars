'use client';

import { useState, useEffect } from 'react';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { Clock } from 'lucide-react';
import { useWebSocket } from '@/lib/context/web-socket-context';
import { 
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger 
} from '@/components/ui/tooltip';

const PAGE_TITLES: Record<string, { title: string }> = {
  '/': { 
    title: 'Dashboard'
  },
  '/strategy-ai': { 
    title: 'Strategy AI'
  }
};

export function TopNavbar() {
  const pathname = usePathname();
  const { isConnected } = useWebSocket();
  const [sessionTime, setSessionTime] = useState('59:59');
  const [checkTime, setCheckTime] = useState('09:59');
  const [sessionProgress, setSessionProgress] = useState(100);
  const [checkProgress, setCheckProgress] = useState(100);
  
  const currentPage = PAGE_TITLES[pathname] || { title: 'Strategy Wars' };

  useEffect(() => {
    let session = 3600; // 60 minutes
    let check = 600; // 10 minutes
    const initialSession = 3600;
    const initialCheck = 600;
    
    const interval = setInterval(() => {
      session--;
      check--;
      
      // Calculate progress percentages
      const sessionPct = Math.round((session / initialSession) * 100);
      const checkPct = Math.round((check / initialCheck) * 100);
      
      if (session <= 0) {
        session = initialSession;
      }
      
      if (check <= 0) {
        check = initialCheck;
      }
      
      const formatTime = (seconds: number) => {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
      };
      
      setSessionTime(formatTime(session));
      setCheckTime(formatTime(check));
      setSessionProgress(sessionPct);
      setCheckProgress(checkPct);
    }, 1000);
    
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="h-full flex flex-col justify-between w-full">
      <div className="flex items-center justify-between h-full px-2 sm:px-4 py-2 w-full">
        <div className="flex items-center">
          <h1 className="text-lg sm:text-xl font-semibold">{currentPage.title}</h1>
        </div>
        
        <div className="flex items-center space-x-2 sm:space-x-3 ml-auto">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex items-center gap-2 sm:gap-3">
                  <div className="flex items-center gap-1.5">
                    <div className="relative h-6 w-6 sm:h-8 sm:w-8 flex items-center justify-center">
                      <svg className="absolute h-full w-full" viewBox="0 0 32 32">
                        <circle 
                          cx="16" 
                          cy="16" 
                          r="14"
                          className="fill-none stroke-muted stroke-[1]"
                        />
                        <circle 
                          cx="16" 
                          cy="16" 
                          r="14"
                          className="fill-none stroke-primary stroke-[2]"
                          strokeDasharray="88"
                          strokeDashoffset={88 - (88 * sessionProgress / 100)}
                          transform="rotate(-90 16 16)"
                        />
                      </svg>
                      <Clock className="h-3 w-3 sm:h-4 sm:w-4 text-muted-foreground" />
                    </div>
                    <span className="text-xs font-mono hidden xs:inline">{sessionTime}</span>
                  </div>
                  
                  <div className="hidden sm:flex items-center gap-1.5">
                    <div className="relative h-5 w-5 sm:h-6 sm:w-6 flex items-center justify-center">
                      <svg className="absolute h-full w-full" viewBox="0 0 24 24">
                        <circle 
                          cx="12" 
                          cy="12" 
                          r="10"
                          className="fill-none stroke-muted stroke-[1]"
                        />
                        <circle 
                          cx="12" 
                          cy="12" 
                          r="10"
                          className="fill-none stroke-accent stroke-[2]"
                          strokeDasharray="63"
                          strokeDashoffset={63 - (63 * checkProgress / 100)}
                          transform="rotate(-90 12 12)"
                        />
                      </svg>
                    </div>
                    <span className="text-xs font-mono text-muted-foreground">{checkTime}</span>
                  </div>
                </div>
              </TooltipTrigger>
              <TooltipContent side="bottom">
                <div className="text-xs">
                  <p>Session time: {sessionTime}</p>
                  <p>Next check: {checkTime}</p>
                </div>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <div className="flex items-center gap-1.5">
            <div className={cn(
              "h-2 w-2 rounded-full transition-colors duration-300",
              isConnected ? "bg-green-500" : "bg-red-500"
            )} />
            <span className="text-xs hidden sm:inline">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
        </div>
      </div>
      
      <div className="h-px bg-border/70 w-full" />
    </div>
  );
}