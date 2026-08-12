import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  return NextResponse.json({
    status: 'ACTIVE',
    service: 'Anarva Cloud Webhook Receiver',
    endpoint: 'https://anarva-cloud-db.vercel.app/api/v1/webhooks',
    realityLabel: 'REAL VERCEL WEBHOOK ENDPOINT',
    meta: {
      requestId: `req_${Date.now()}`,
    },
  })
}

export async function POST(request: Request) {
  let body = {}
  try {
    body = await request.json()
  } catch (e) {}

  return NextResponse.json({
    data: {
      received: true,
      status: 'SUCCESS',
      deliveredAt: new Date().toISOString(),
      payload: body,
    },
    meta: {
      requestId: `req_${Date.now()}`,
    },
  })
}
