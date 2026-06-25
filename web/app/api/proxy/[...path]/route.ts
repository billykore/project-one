import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL || "http://localhost:8080";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  return proxy(req, await params);
}

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  return proxy(req, await params);
}

export async function PUT(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  return proxy(req, await params);
}

export async function DELETE(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  return proxy(req, await params);
}

async function proxy(req: NextRequest, { path }: { path: string[] }) {
  const endpoint = `/${path.join("/")}`;
  const backendUrl = `${API_URL}${endpoint}`;

  // Forward cookies from the incoming request so the backend can authenticate
  const cookieHeader = req.cookies.toString();

  let body: string | undefined;
  const contentType = req.headers.get("content-type") || "";
  if (contentType.includes("application/json")) {
    try {
      body = await req.text();
    } catch {
      // No body
    }
  }

  try {
    const backend = await fetch(backendUrl, {
      method: req.method,
      headers: {
        "Content-Type": contentType || "application/json",
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
      ...(body ? { body } : {}),
    });

    const data = await backend.text();
    const responseHeaders: Record<string, string> = {};

    // Forward Content-Type so the client parses JSON responses correctly
    const backendContentType = backend.headers.get("content-type");
    if (backendContentType) {
      responseHeaders["Content-Type"] = backendContentType;
    }

    // Forward Set-Cookie headers so auth cookies reach the browser
    const setCookie = backend.headers.get("set-cookie");
    if (setCookie) {
      responseHeaders["Set-Cookie"] = setCookie;
    }

    return new NextResponse(data, {
      status: backend.status,
      headers: responseHeaders,
    });
  } catch {
    return NextResponse.json(
      { error: "Unable to connect to the server" },
      { status: 502 },
    );
  }
}
