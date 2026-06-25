import { proxyToBackend } from "@/lib/api-proxy";

export async function POST(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/posts/${id}/comments`);
}
