import { proxyToBackend } from "@/lib/api-proxy";

type Params = { params: Promise<{ key: string }> };

export async function GET(req: Request, { params }: Params) {
  const { key } = await params;
  const query = new URL(req.url).search;
  return proxyToBackend(req, `/admin/feature-flags/${encodeURIComponent(key)}/audit${query}`);
}
