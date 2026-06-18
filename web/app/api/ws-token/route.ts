import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

/**
 * GET /api/ws-token
 *
 * Returns the access_token value from the HttpOnly cookie so that
 * the browser-side WebSocket client can pass it as a query parameter
 * during the upgrade handshake (browser WebSocket API does not allow
 * setting custom headers).
 *
 * This is intentionally a minimal endpoint — it only reflects the
 * caller's own cookie value back to them.  It does NOT bypass any
 * auth check; the backend WebSocket handler still validates the token
 * through the Authorize middleware.
 */
export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get('access_token')?.value;

  if (!token) {
    return NextResponse.json({ token: null }, { status: 401 });
  }

  return NextResponse.json({ token });
}
