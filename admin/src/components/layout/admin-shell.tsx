"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth-store";

const nav = [
  ["Dashboard", "/"],
  ["Tenants", "/tenants"],
  ["Users", "/users"],
  ["Orders", "/orders"],
  ["Commissions", "/finance/commissions"],
  ["Invoices", "/finance/invoices"],
  ["Payouts", "/finance/payouts"],
  ["Issues", "/issues"],
  ["Config", "/config"],
] as const;

export function AdminShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const clearSession = useAuthStore((state) => state.clearSession);

  const logout = async () => {
    await fetch("/api/auth/logout", { method: "POST" });
    clearSession();
    router.push("/auth/login");
  };

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="border-b bg-white">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <h1 className="text-lg font-semibold">Munchies Admin</h1>
          <Button onClick={logout}>Logout</Button>
        </div>
      </header>
      <div className="mx-auto grid max-w-7xl grid-cols-1 gap-6 px-6 py-6 md:grid-cols-[220px_1fr]">
        <aside className="rounded-md border bg-white p-2">
          {nav.map(([label, href]) => {
            const active = pathname === href || pathname.startsWith(`${href}/`);
            return (
              <Link
                className={`block rounded px-3 py-2 text-sm ${active ? "bg-slate-900 text-white" : "text-slate-700 hover:bg-slate-100"}`}
                href={href}
                key={href}
              >
                {label}
              </Link>
            );
          })}
        </aside>
        <main>{children}</main>
      </div>
    </div>
  );
}
