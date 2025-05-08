'use client'

import { useState, useEffect } from 'react'
import { 
  RefreshCw, Wifi, WifiOff, BarChart, Clock 
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { motion, AnimatePresence } from 'framer-motion'

interface DashboardHeaderProps {
  selectedTimeframe: string
  onTimeframeChange: (timeframe: string) => void
  lastUpdated: Date
  isConnected: boolean
  isLoading: boolean
  hasLiveUpdates: boolean
  onRefresh: () => void
}

export function DashboardHeader({
  selectedTimeframe,
  onTimeframeChange,
  lastUpdated,
  isConnected,
  isLoading,
  hasLiveUpdates,
  onRefresh
}: DashboardHeaderProps) {
  // Format time
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className="space-y-2">
      {/* Timeframe Tabs & Refresh Button */}
      <div className="flex justify-between items-center">
        <Tabs 
          value={selectedTimeframe} 
          onValueChange={onTimeframeChange}
          className="w-auto"
        >
          <TabsList className="grid grid-cols-3 w-48">
            <TabsTrigger value="24h">24h</TabsTrigger>
            <TabsTrigger value="7d">7d</TabsTrigger>
            <TabsTrigger value="30d">30d</TabsTrigger>
          </TabsList>
        </Tabs>
        
        <div className="flex items-center gap-2">
          <AnimatePresence>
          {hasLiveUpdates && (
            <motion.div
              initial={{ opacity: 0, scale: 0.8, y: 5 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.8, y: -5 }}
              transition={{ duration: 0.3 }}
            >
              <Badge 
                variant="outline" 
                className="bg-green-500/10 text-green-500"
              >
                <motion.span
                  animate={{ opacity: [1, 0.6, 1] }}
                  transition={{ repeat: Infinity, duration: 1.5 }}
                >
                  Live update
                </motion.span>
              </Badge>
            </motion.div>
          )}
        </AnimatePresence>
        </div>
      </div>
      
      {/* Last updated info */}
      <div className="flex items-center justify-between text-xs text-muted-foreground">
        
        
      
      </div>
    </div>
  )
}