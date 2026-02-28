import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { PartnerShell } from "@/components/layout/partner-shell";

export default async function ProtectedLayout({ children }: { children: React.ReactNode }) {
  const cookieStore = await cookies();

  if (!cookieStore.has("partner_access_token")) {
    redirect("/auth/login?expired=1");
  }

  return <PartnerShell>{children}</PartnerShell>;
}
