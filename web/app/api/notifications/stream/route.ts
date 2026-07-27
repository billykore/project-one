const BACKEND_URL = process.env.API_URL || 'http://localhost:8080';

export async function GET(req: Request) {
  const cookieHeader = req.headers.get('cookie') || '';

  try {
    const backend = await fetch(`${BACKEND_URL}/notifications/stream`, {
      method: 'GET',
      headers: {
        Accept: 'text/event-stream',
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
    });

    const headers = new Headers();
    backend.headers.forEach((value, key) => {
      if (key.toLowerCase() !== 'content-length') {
        headers.set(key, value);
      }
    });

    if (!headers.has('content-type')) {
      headers.set('content-type', 'text/event-stream');
    }
    if (!headers.has('cache-control')) {
      headers.set('cache-control', 'no-cache');
    }

    return new Response(backend.body, {
      status: backend.status,
      headers,
    });
  } catch {
    return new Response('Unable to connect to the server', {
      status: 502,
      headers: { 'content-type': 'text/plain; charset=utf-8' },
    });
  }
}
