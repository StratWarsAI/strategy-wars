'use client';

import { cn } from '@/lib/utils';
import { MobileSidebar } from './mobile-sidebar';

export default function Header() {
  
  return (
    <header className="sticky top-0 z-10 h-[52px] w-full border-b bg-background">
      <div className="flex items-center justify-between h-full px-4">
        <div className={cn('block lg:!hidden')}>
          <MobileSidebar />
        </div>
        
      </div>
    </header>
  );
}