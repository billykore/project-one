import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  return proxyToBackend(req, `/feeds${new URL(req.url).search}`);
}
