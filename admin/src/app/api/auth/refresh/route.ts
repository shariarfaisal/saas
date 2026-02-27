import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function POST() {
  if (!(await cookies()).has("admin_access_token")) {
    return NextResponse.json({ message: "no session" }, { status: 401 });
  }

  const response = NextResponse.json({ ok: true });
  const token = `mock-${crypto.randomUUID()}`;
  response.cookies.set("admin_access_token", token, {
    httpOnly: true,
    sameSite: "lax",
    secure: true,
    path: "/",
    maxAge: 60 * 30,
  });
  return response;
}
