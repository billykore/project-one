import { proxyToBackend } from "@/lib/api-proxy";

export async function PUT(req: Request) {
  return proxyToBackend(req, "/notifications/read-all");
}
