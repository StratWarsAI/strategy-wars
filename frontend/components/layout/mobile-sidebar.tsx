'use client';
import { NavItems } from './nav-items';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { navigationItems } from './items'
import { MenuIcon } from 'lucide-react';
import { useState } from 'react';
import SidebarLogo from './logo';
import { Button } from '@/components/ui/button';

interface MobileSidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

export function MobileSidebar({ className }: MobileSidebarProps) {
  const [open, setOpen] = useState(false);
  
  return (
    <>
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild>
          <Button 
            variant="ghost" 
            size="icon" 
            className="md:hidden"
          >
            <MenuIcon className="h-6 w-6" />
            <span className="sr-only">Toggle menu</span>
          </Button>
        </SheetTrigger>
        
        <SheetContent 
          side="left" 
          className="flex w-80 flex-col p-0"
        >
          {/* Logo Section */}
          <div className="border-b border-border">
            <SidebarLogo isMinimized={false} />
          </div>

          {/* Navigation Items */}
          <div className="flex-1 overflow-y-auto">
            <nav className="flex flex-col gap-2 p-4">
              <NavItems
                items={navigationItems}
                isMobileNav={true}
                setOpen={setOpen}
              />
            </nav>
          </div>
        </SheetContent>
      </Sheet>
    </>
  );
}