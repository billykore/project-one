import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/notifications",
        destination: "http://localhost:8080/notifications",
      },
      {
        source: "/notifications/:path*",
        destination: "http://localhost:8080/notifications/:path*",
      },
    ];
  },
};

export default nextConfig;
