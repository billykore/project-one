import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(_req, `/users/${username}`);
}
