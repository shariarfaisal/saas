import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  const body = (await request.json()) as { email?: string; password?: string };
  if (!body.email || !body.password) {
    return NextResponse.json({ message: "email and password required" }, { status: 400 });
  }

  const hasEnrollment = (await cookies()).has("admin_2fa_enrolled");
  return NextResponse.json({ next: hasEnrollment ? "verify" : "setup" });
}
