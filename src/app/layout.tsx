import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL("https://github.com/Udonburo/picca"),
  title: "Picca — Motion Scoring Build Log",
  description:
    "身体のキーポイントをスコアへ変換するモーション解析PoCの設計判断と学びを残した技術ログ。",
  icons: {
    icon: "/picca-mark.svg",
  },
  openGraph: {
    title: "Picca — Motion Scoring Build Log",
    description:
      "Next.js・Go・FastAPI・ONNX・Terraformを横断したモーション解析PoCの技術学習ログ。",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}
