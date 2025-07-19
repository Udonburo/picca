import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json({
    db: process.env.DB_PASSWORD ?? null,
    key: !!process.env.API_KEY        // true/false だけ返す
  });
}
