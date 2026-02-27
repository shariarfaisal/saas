import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { AdminShell } from "@/components/layout/admin-shell";

export default async function ProtectedLayout({ children }: { children: React.ReactNode }) {
  const cookieStore = await cookies();

  if (!cookieStore.has("admin_access_token")) {
    redirect("/auth/login?expired=1");
  }

  return <AdminShell>{children}</AdminShell>;
}
