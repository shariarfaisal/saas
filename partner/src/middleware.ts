import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const PUBLIC_PATHS = ["/auth/login", "/auth/forgot-password", "/auth/reset-password", "/auth/invite"];

export function middleware(request: NextRequest) {
  const path = request.nextUrl.pathname;
  if (path.startsWith("/api") || path.startsWith("/_next") || path === "/favicon.ico" || PUBLIC_PATHS.includes(path)) {
    return NextResponse.next();
  }

  if (!request.cookies.has("partner_access_token")) {
    return NextResponse.redirect(new URL("/auth/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image).*)"],
};
