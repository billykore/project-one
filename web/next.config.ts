import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/users/:path*",
        destination: "http://localhost:8080/users/:path*",
      },
      {
        source: "/posts/:path*",
        destination: "http://localhost:8080/posts/:path*",
      },
    ];
  },
};

export default nextConfig;
