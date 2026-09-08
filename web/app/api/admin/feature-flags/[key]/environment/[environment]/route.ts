import { proxyToBackend } from "@/lib/api-proxy";

type Params = { params: Promise<{ key: string; environment: string }> };

export async function PATCH(req: Request, { params }: Params) {
  const { key, environment } = await params;
  return proxyToBackend(req, `/admin/feature-flags/${encodeURIComponent(key)}/environment/${encodeURIComponent(environment)}`);
}
