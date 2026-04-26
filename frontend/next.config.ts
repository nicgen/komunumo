import type { NextConfig } from "next";

const apiInternalUrl = process.env.KOMUNUMO_API_INTERNAL_URL;
if (!apiInternalUrl) {
  throw new Error(
    "KOMUNUMO_API_INTERNAL_URL is required (e.g. http://backend:8080 in compose, http://localhost:8080 in solo dev).",
  );
}

const nextConfig: NextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${apiInternalUrl}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
