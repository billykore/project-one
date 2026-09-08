import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  const url = new URL(req.url);
  const keys = url.searchParams.get("keys") || "";
  return proxyToBackend(req, `/feature-flags/evaluate?keys=${encodeURIComponent(keys)}`);
}
