// src/app/api/predict/route.ts
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest) {
  try {
    const mlUrl = process.env.API_ML_URL!;
    const mlRes = await fetch(`${mlUrl}/predict`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(await req.json()),
    });
    if (!mlRes.ok) throw new Error('ML service error');
    const data = await mlRes.json();
    return NextResponse.json(data, { status: 200 });
  } catch {
    return NextResponse.json({ message: 'Upstream ML Error' }, { status: 502 });
  }
}