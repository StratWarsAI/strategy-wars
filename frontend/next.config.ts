import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "ipfs.io",
        pathname: "/ipfs/**"
      },
      {
        protocol: "https",
        hostname: "metadata.pumplify.eu",
        pathname: "/**"
      },
      {
        protocol: "https",
        hostname: "s2.coinmarketcap.com",
        pathname: "/**"
      },
      {
        protocol: "https",
        hostname: "pbs.twimg.com",
        pathname: "/**"
      },
      {
        protocol: "http",
        hostname: "93.205.14.16",
        pathname: "/images/**"
      },
      {
        protocol: "https",
        hostname: "faggotnuked.com",
        pathname: "/**"
      },
      {
        protocol: "https",
        hostname: "metadata.pumployer.fun",
        pathname: "/**"
      },
      {
        protocol: "https",
        hostname: "mainnet-boop.s3.us-east-1.amazonaws.com",
        pathname: "/**"
      },
      {
        protocol: "http",
        hostname: "149.51.224.102",
        pathname: "/images/**"
      },
      {
        protocol: "http",
        hostname: "159.65.198.85",
        pathname: "/**"
      },
      {
        protocol: "https",
        hostname: "pumpbucket.party",
        pathname: "/**"
      },
      {
        protocol: "https",
        hostname: "gateway.pinata.cloud",
        pathname: "/ipfs/**"
      },
      {
        protocol: "https",
        hostname: "wsrv.nl",
        pathname: "/**"
      },
      {
        protocol: "http",
        hostname: "172.86.93.253:3000",
        pathname: "/**"
      }
    ]
  }
};

export default nextConfig;
