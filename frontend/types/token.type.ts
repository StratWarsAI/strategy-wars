export interface Token {
    id: number;
    mintAddress: string;
    creatorAddress: string;
    name: string;
    symbol: string;
    imageUrl: string;
    twitterUrl: string;
    websiteUrl: string;
    telegramUrl: string;
    metadataUrl: string;
    createdTimestamp: number;
    marketCap: number;
    usdMarketCap: number;
    completed: boolean;
    kingOfTheHillTimeStamp: number;
    createdAt?: string;
}