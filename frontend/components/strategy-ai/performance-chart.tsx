'use client';

import { useEffect, useState } from 'react';
import { 
  AreaChart, Area, XAxis, YAxis, CartesianGrid, 
  Tooltip, Legend, ResponsiveContainer, ReferenceLine 
} from 'recharts';
import { Skeleton } from "@/components/ui/skeleton";

interface DataPoint {
  time: string;
  strategy1: number;
  strategy2: number;
}

interface PerformanceChartProps {
  data: DataPoint[];
  height?: number;
  strategy1Name?: string;
  strategy2Name?: string;
}

export function PerformanceChart({ 
  data, 
  height = 300,
  strategy1Name = "Strategy 1",
  strategy2Name = "Strategy 2"
}: PerformanceChartProps) {
  const [mounted, setMounted] = useState(false);
  
  // Ensure chart only renders after mounting on client
  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div className="flex h-full w-full items-center justify-center">
        <Skeleton className="h-[300px] w-full rounded-lg" />
      </div>
    );
  }

  // Process data to ensure reasonable values
  const processedData = data && data.length > 0 ? data.map(point => ({
    ...point,
    // Cap ROI values to reasonable ranges (-100% to +1000%)
    strategy1: typeof point.strategy1 === 'number' ? Math.min(Math.max(point.strategy1, -100), 1000) : 0,
    strategy2: typeof point.strategy2 === 'number' ? Math.min(Math.max(point.strategy2, -100), 1000) : 0
  })) : [{ time: "00:00:00", strategy1: 0, strategy2: 0 }];

  // Calculate min and max for better Y axis scaling
  const allValues = processedData.flatMap(item => [item.strategy1, item.strategy2]);
  const minValue = Math.min(...allValues, 0); // Include 0 as a minimum
  const maxValue = Math.max(...allValues, 0); // Include 0 as a maximum

  // Calculate Y-axis domain with some padding
  const yDomain = [
    Math.floor(minValue - Math.abs(minValue * 0.1)), 
    Math.ceil(maxValue + Math.abs(maxValue * 0.1))
  ];

  // Get latest values for each strategy
  const latestPoint = processedData.length > 0 ? 
    processedData[processedData.length - 1] : 
    { strategy1: 0, strategy2: 0 };

  // Custom tooltip component
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="rounded-lg border border-border bg-card p-3 shadow-lg">
          <p className="mb-1 text-sm font-semibold">{label}</p>
          {payload.map((entry: any, index: number) => (
            <div 
              key={`item-${index}`} 
              className="flex items-center justify-between gap-4 mt-1"
            >
              <div className="flex items-center gap-2">
                <div 
                  className="h-3 w-3 rounded-full" 
                  style={{ backgroundColor: entry.color }}
                />
                <span className="text-sm">{entry.name}</span>
              </div>
              <span 
                className={`text-sm font-bold ${
                  entry.value > 0 ? 'text-green-500' : 'text-red-500'
                }`}
              >
                {entry.value > 0 ? '+' : ''}{entry.value.toFixed(2)}%
              </span>
            </div>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="w-full h-full">
      <div className="flex justify-between mb-2 text-sm">
        <div className={latestPoint.strategy1 >= 0 ? "text-green-500" : "text-red-500"}>
          {strategy1Name} ({latestPoint.strategy1 >= 0 ? "+" : ""}{latestPoint.strategy1.toFixed(1)}%)
        </div>
        <div className={latestPoint.strategy2 >= 0 ? "text-green-500" : "text-red-500"}>
          {strategy2Name} ({latestPoint.strategy2 >= 0 ? "+" : ""}{latestPoint.strategy2.toFixed(1)}%)
        </div>
      </div>
      
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart
          data={processedData}
          margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
        >
          <defs>
            <linearGradient id="colorStrategy1" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#99FF19" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#99FF19" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorStrategy2" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#7c93b5" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#7c93b5" stopOpacity={0} />
            </linearGradient>
          </defs>
          
          <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
          
          <XAxis 
            dataKey="time" 
            fontSize={12} 
            tick={{ fill: '#a0a0a0' }}
            tickLine={false}
          />
          
          <YAxis 
            fontSize={12}
            tick={{ fill: '#a0a0a0' }}
            tickLine={false}
            tickFormatter={(value) => `${value}%`}
            domain={yDomain}
          />
          
          <Tooltip content={<CustomTooltip />} />
          <ReferenceLine y={0} stroke="#666" strokeDasharray="3 3" />
          
          <Legend
            verticalAlign="top"
            height={36}
          />
          
          <Area 
            type="monotone" 
            dataKey="strategy1" 
            name={strategy1Name} 
            stroke="#99FF19" 
            fillOpacity={1}
            fill="url(#colorStrategy1)"
            strokeWidth={2}
            activeDot={{ r: 8 }}
          />
          
          <Area 
            type="monotone" 
            dataKey="strategy2" 
            name={strategy2Name} 
            stroke="#7c93b5" 
            fillOpacity={1}
            fill="url(#colorStrategy2)"
            strokeWidth={2}
            activeDot={{ r: 8 }}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}