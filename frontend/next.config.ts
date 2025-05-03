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
      }
    ]
  }
};

export default nextConfig;
