'use client'

import { cn } from "@/lib/utils";
import Sidebar from "./sidebar";
import { useSidebar } from "@/hooks/use-sidebar";
import Header from "./header";

export default function BaseLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isMinimized } = useSidebar();
  return (
    <div className="h-screen flex overflow-hidden bg-background">
      <Sidebar />      
      <div 
        className={cn(
          "flex-1 flex flex-col h-screen transition-all duration-500",
          isMinimized ? "lg:ml-[72px]" : "lg:ml-60"
        )}
      >
        <Header />        
        <div className="flex-1 overflow-auto">
          {children}
        </div>
      </div>
    </div>
  );
} 