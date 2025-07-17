import { NextResponse } from 'next/server';

export async function GET() {
  const url   = process.env.API_GO_URL ?? '';   // Cloud Run URL を env に
  const resp  = await fetch(`${url}/v1/ping`);
  const data  = await resp.json();
  return NextResponse.json(data);
}
