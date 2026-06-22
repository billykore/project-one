import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// ponytail: presence check only; backend enforces actual token validity on every API call.
// Add jose + RSA public key env var when you need Edge-side expiry/signature checks.
const PUBLIC_ROUTES = new Set(["/login", "/register", "/error"]);

// ponytail: proxy replaces deprecated middleware convention (Next.js 16)
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // ponytail: allow public pages, static assets, and API calls through
  if (PUBLIC_ROUTES.has(pathname) || pathname.startsWith("/_next") || pathname.startsWith("/api") || pathname.startsWith("/static")) {
    return NextResponse.next();
  }

  if (!request.cookies.has("access_token")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|.*\\.png$).*)"],
};
