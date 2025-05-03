'use client';

import { cn } from '@/lib/utils';
import { ChevronLeft } from 'lucide-react';
import { useSidebar } from '@/hooks/use-sidebar';
import SidebarLogo from './logo';
import { NavItems } from './nav-items';
import { navigationItems } from './items';

export default function Sidebar() {
  const { isMinimized, toggle } = useSidebar();

  return (
    <aside
      className={cn(
        'fixed top-0 left-0 z-20 h-screen border-r bg-card transition-[width] duration-500 hidden lg:block',
        isMinimized ? 'w-[72px]' : 'w-60'
      )}
    >
      <SidebarLogo isMinimized={isMinimized} />

      <ChevronLeft
        className={cn(
          'absolute -right-3 top-6 z-30 cursor-pointer rounded-full border bg-background p-1 text-foreground transition-transform',
          isMinimized && 'rotate-180'
        )}
        onClick={toggle}
      />
      
      <div className="px-4 overflow-y-auto h-[calc(100vh-60px)]">
        <NavItems items={navigationItems} />
      </div>
    </aside>
  );
}