export interface Trade {
    id: number;
    token: string;
    type: string;
    price?: number;
    entryPrice?: number;
    exitPrice?: number;
    amount: number;
    time: string;
    marketCap?: number;
    profit?: number;
    reason?: string;
    imageUrl?: string;
    twitterUrl?: string;
    websiteUrl?: string;
    symbol?: string;
  }