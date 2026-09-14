import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  return proxyToBackend(req, `/posts${new URL(req.url).search}`);
}

export async function POST(req: Request) {
  return proxyToBackend(req, "/posts");
}
