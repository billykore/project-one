import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  const url = new URL(req.url);
  const q = url.searchParams.get("q") || "";
  const cursor = url.searchParams.get("cursor") || "";
  const limit = url.searchParams.get("limit") || "20";

  const params = new URLSearchParams({ q, limit });
  if (cursor) params.set("cursor", cursor);

  return proxyToBackend(req, `/users/search?${params.toString()}`);
}
