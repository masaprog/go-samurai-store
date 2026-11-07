import type { NextConfig } from "next";

const isProd = process.env.NODE_ENV === "production";
const backendURL = isProd ? process.env.API_BASE_URL : "http://backend:8080";

const nextConfig: NextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      {
        // フロントエンドが/api/...をリクエストしたら
        source: "/api/:path*",
        // バックエンドの/api/...に転送する
        destination: `${backendURL}/api/:path*`,
      },
      {
        // フロントエンドが/uploads/xxx.jpgをリクエストしたら
        source: "/uploads/:path*",
        // バックエンドの/uploads/xxx.jpgに転送する
        destination: `${backendURL}/uploads/:path*`,
      },
    ];
  },
};

export default nextConfig;
