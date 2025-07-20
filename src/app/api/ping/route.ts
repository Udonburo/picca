// app/api/ping/route.ts
import { NextResponse } from 'next/server';

const apiUrl = process.env.API_GO_URL ?? ''; // Cloud Run URL

export async function GET() {
  try {
    const r = await fetch(`${apiUrl}/v1/ping`);
    const data = await r.json();
    return NextResponse.json(data, { status: 200 });
  } catch (error) {
    // 変数 error を使ってログを出力
    console.error('Upstream ping failed:', error);
    return NextResponse.json(
      { error: 'upstream ping failed' },
      { status: 502 },
    );
  }
}
