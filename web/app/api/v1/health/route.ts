import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  return NextResponse.json({
    status: 'HEALTHY',
    service: 'Anarva Cloud Platform API Gateway',
    version: '1.0.0',
    realityLabel: 'REAL VERCEL DEPLOYED SERVICE',
    timestamp: new Date().toISOString(),
  })
}
