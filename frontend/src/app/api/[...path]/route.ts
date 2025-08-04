import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

// Handle all HTTP methods
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  return handleRequest(request, path, 'GET');
}

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  return handleRequest(request, path, 'POST');
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  return handleRequest(request, path, 'PUT');
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  return handleRequest(request, path, 'DELETE');
}

export async function OPTIONS(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  return handleRequest(request, path, 'OPTIONS');
}

async function handleRequest(
  request: NextRequest,
  pathSegments: string[],
  method: string
) {
  const path = pathSegments.join('/');
  const backendUrl = `${BACKEND_URL}/${path}`;

  try {
    // Get the request body if present
    let body: string | null = null;
    if (request.body && ['POST', 'PUT', 'PATCH'].includes(method)) {
      body = await request.text();
    }

    // Forward relevant headers
    const headers: HeadersInit = {
      'Content-Type': request.headers.get('Content-Type') || 'application/json',
    };

    // Forward ConnectRPC specific headers (excluding encoding headers)
    const connectHeaders = [
      'connect-protocol-version',
      'connect-timeout-ms',
    ];
    
    connectHeaders.forEach(header => {
      const value = request.headers.get(header);
      if (value) {
        headers[header] = value;
      }
    });

    // Make the request to the backend
    const response = await fetch(backendUrl, {
      method,
      headers,
      body,
    });

    // Get response body
    const responseBody = await response.text();

    // Create response with the same status and headers
    const nextResponse = new NextResponse(responseBody, {
      status: response.status,
      statusText: response.statusText,
    });

    // Forward only safe response headers
    const safeHeaders = [
      'content-type',
      'connect-protocol-version',
      'connect-timeout-ms',
    ];
    
    safeHeaders.forEach(header => {
      const value = response.headers.get(header);
      if (value && header !== 'content-encoding') {
        nextResponse.headers.set(header, value);
      }
    });

    // Add CORS headers for development
    if (process.env.NODE_ENV === 'development') {
      nextResponse.headers.set('Access-Control-Allow-Origin', '*');
      nextResponse.headers.set('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
      nextResponse.headers.set('Access-Control-Allow-Headers', 'Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms');
    }

    return nextResponse;
  } catch (error) {
    console.error('Proxy error:', error);
    return NextResponse.json(
      { error: 'Failed to proxy request' },
      { status: 500 }
    );
  }
}