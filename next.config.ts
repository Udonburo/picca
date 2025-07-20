import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // Cloud Run 向けに軽量な standalone 出力
  output: 'standalone',

  // ↓ 既存オプションを追記するならここ
  // reactStrictMode: true,
  // experimental: { appDir: true },
};

export default nextConfig;
