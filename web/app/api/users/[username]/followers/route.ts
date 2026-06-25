import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(_req, `/users/${username}/followers`);
}

export async function POST(req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(req, `/users/${username}/followers`);
}

export async function DELETE(req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(req, `/users/${username}/followers`);
}
