import { proxyToBackend } from "@/lib/api-proxy";

type Params = { params: Promise<{ key: string }> };

export async function GET(req: Request, { params }: Params) {
  const { key } = await params;
  return proxyToBackend(req, `/admin/feature-flags/${encodeURIComponent(key)}`);
}

export async function PUT(req: Request, { params }: Params) {
  const { key } = await params;
  return proxyToBackend(req, `/admin/feature-flags/${encodeURIComponent(key)}`);
}
