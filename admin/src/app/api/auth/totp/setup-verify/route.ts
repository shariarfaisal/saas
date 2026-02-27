import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  const body = (await request.json()) as { code?: string };
  if (!body.code || body.code.length < 6) {
    return NextResponse.json({ message: "invalid code" }, { status: 400 });
  }

  const response = NextResponse.json({ ok: true });
  response.cookies.set("admin_2fa_enrolled", "1", { httpOnly: true, sameSite: "lax", secure: true, path: "/" });
  response.cookies.set("admin_access_token", "session-token", {
    httpOnly: true,
    sameSite: "lax",
    secure: true,
    path: "/",
    maxAge: 60 * 30,
  });
  return response;
}
