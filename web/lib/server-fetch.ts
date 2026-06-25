import { cookies, headers } from "next/headers";

export async function serverFetch(path: string, init?: RequestInit): Promise<Response> {
  const [cookieStore, headersList] = await Promise.all([cookies(), headers()]);
  const cookieString = cookieStore.toString();
  const host = headersList.get("host") || "localhost:3000";
  const protocol = process.env.NODE_ENV === "production" ? "https" : "http";
  const baseUrl = `${protocol}://${host}`;

  return fetch(`${baseUrl}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
      ...(cookieString ? { Cookie: cookieString } : {}),
    },
  });
}
