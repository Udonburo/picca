import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // Keep framework-generated AI guidance out of this archival repository.
  agentRules: false,

  // Cloud Run 向けに軽量な standalone 出力
  output: 'standalone',

  // ↓ 既存オプションを追記するならここ
  // reactStrictMode: true,
  // experimental: { appDir: true },
};

export default nextConfig;
