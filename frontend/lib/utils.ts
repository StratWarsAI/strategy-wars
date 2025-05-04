import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const formatNumber = (num: number | undefined) => {
  if (num === undefined) return 'N/A';
  
  if (num >= 1000000) {
    return (num / 1000000).toFixed(2) + 'M';
  } else if (num >= 1000) {
    return (num / 1000).toFixed(2) + 'K';
  } else {
    return num.toFixed(2);
  }
};

export const formatExitReason = (reason: string | undefined) => {
  if (!reason) return 'Unknown';
  return reason.charAt(0).toUpperCase() + reason.slice(1).replace('_', ' ');
};