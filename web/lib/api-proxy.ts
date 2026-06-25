const BACKEND_URL = process.env.API_URL || "http://localhost:8080";

export async function proxyToBackend(req: Request, backendPath: string): Promise<Response> {
  let body: string | undefined;
  const contentType = req.headers.get("content-type") || "";

  if (contentType.includes("application/json")) {
    try {
      body = await req.text();
    } catch {
      // no body
    }
  }

  const cookieHeader = req.headers.get("cookie") || "";

  let backend: Response;
  try {
    backend = await fetch(`${BACKEND_URL}${backendPath}`, {
      method: req.method,
      headers: {
        "Content-Type": contentType || "application/json",
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
      ...(body ? { body } : {}),
    });
  } catch {
    return Response.json({ error: "Unable to connect to the server" }, { status: 502 });
  }

  let data: unknown;
  try {
    data = await backend.json();
  } catch {
    data = { error: "Invalid response from server" };
  }

  const responseHeaders: Record<string, string> = {};
  const setCookie = backend.headers.get("set-cookie");
  if (setCookie) {
    responseHeaders["Set-Cookie"] = setCookie;
  }

  return Response.json(data, { status: backend.status, headers: responseHeaders });
}
