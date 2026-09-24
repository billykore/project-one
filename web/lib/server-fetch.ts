import { cookies } from "next/headers";

export async function serverFetch(path: string, init?: RequestInit): Promise<Response> {
  const cookieStore = await cookies();
  const cookieString = cookieStore.toString();
  // Server components call this frontend's local route handlers. Production
  // mode does not imply TLS, so do not derive the internal scheme from NODE_ENV.
  const internalOrigin = process.env.INTERNAL_ORIGIN || `http://127.0.0.1:${process.env.PORT || "3000"}`;

  return fetch(new URL(path, internalOrigin), {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
      ...(cookieString ? { Cookie: cookieString } : {}),
    },
  });
}
