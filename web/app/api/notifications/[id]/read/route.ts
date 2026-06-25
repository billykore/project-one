import { proxyToBackend } from "@/lib/api-proxy";

export async function PUT(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/notifications/${id}/read`);
}
