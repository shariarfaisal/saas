import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  const body = (await request.json()) as { email?: string; password?: string };
  if (!body.email || !body.password) {
    return NextResponse.json({ message: "email and password required" }, { status: 400 });
  }

  // In production, this proxies to the backend auth endpoint.
  // For now, set a mock session token.
  const response = NextResponse.json({
    restaurants: [
      { id: "rest-1", name: "Main Branch", isAvailable: true },
      { id: "rest-2", name: "Downtown Branch", isAvailable: true },
    ],
  });

  const token = `mock-partner-${crypto.randomUUID()}`;
  response.cookies.set("partner_access_token", token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: 60 * 30,
  });

  return response;
}
