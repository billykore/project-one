import { loginRequestBodySchema } from "./schema";

export async function POST(req: Request) {
  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return Response.json({ error: "Invalid JSON body" }, { status: 400 });
  }

  const parsed = loginRequestBodySchema.safeParse(body);
  if (!parsed.success) {
    return Response.json(
      { error: parsed.error.issues[0]?.message || "Invalid request" },
      { status: 400 }
    );
  }

  const backendUrl = `${process.env.API_URL || "http://localhost:8080"}/api/v1/auth/login`;

  let backend: Response;
  try {
    backend = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(parsed.data),
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

  return Response.json(data, { status: backend.status });
}