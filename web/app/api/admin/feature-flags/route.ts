import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  return proxyToBackend(req, "/admin/feature-flags");
}

export async function POST(req: Request) {
  return proxyToBackend(req, "/admin/feature-flags");
}
