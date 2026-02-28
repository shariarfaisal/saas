"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  LayoutDashboard,
  Store,
  UtensilsCrossed,
  ShoppingBag,
  Bike,
  Tag,
  Receipt,
  BarChart3,
  Image,
  Users,
  Settings,
  Bell,
  ChevronDown,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth-store";
import { useNotificationPolling } from "@/hooks/use-notification-polling";

const nav = [
  { label: "Dashboard", href: "/", icon: LayoutDashboard },
  { label: "Restaurants", href: "/restaurants", icon: Store },
  { label: "Menu", href: "/menu", icon: UtensilsCrossed },
  { label: "Orders", href: "/orders", icon: ShoppingBag },
  { label: "Riders", href: "/riders", icon: Bike },
  { label: "Promotions", href: "/promotions", icon: Tag },
  { label: "Finance", href: "/finance", icon: Receipt },
  { label: "Reports", href: "/reports", icon: BarChart3 },
  { label: "Content", href: "/content/banners", icon: Image },
  { label: "Team", href: "/team", icon: Users },
  { label: "Settings", href: "/settings", icon: Settings },
] as const;

export function PartnerShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const clearSession = useAuthStore((s) => s.clearSession);
  const restaurants = useAuthStore((s) => s.restaurants);
  const activeRestaurantId = useAuthStore((s) => s.activeRestaurantId);
  const setActiveRestaurant = useAuthStore((s) => s.setActiveRestaurant);
  const unreadNotifications = useAuthStore((s) => s.unreadNotifications);

  useNotificationPolling();

  const logout = async () => {
    await fetch("/api/auth/logout", { method: "POST" });
    clearSession();
    router.push("/auth/login");
  };

  const activeRestaurant = restaurants.find((r) => r.id === activeRestaurantId);

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="border-b bg-white">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
          <div className="flex items-center gap-4">
            <h1 className="text-lg font-semibold">Munchies Partner</h1>
            {restaurants.length > 1 && (
              <div className="relative">
                <select
                  className="appearance-none rounded-md border border-slate-200 bg-slate-50 py-1.5 pl-3 pr-8 text-sm font-medium"
                  value={activeRestaurantId ?? ""}
                  onChange={(e) => setActiveRestaurant(e.target.value)}
                >
                  {restaurants.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name}
                    </option>
                  ))}
                </select>
                <ChevronDown className="pointer-events-none absolute right-2 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              </div>
            )}
          </div>
          <div className="flex items-center gap-3">
            <button className="relative rounded-md p-2 hover:bg-slate-100" title="Notifications">
              <Bell className="h-5 w-5 text-slate-600" />
              {unreadNotifications > 0 && (
                <span className="absolute -right-0.5 -top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-rose-500 text-[10px] font-bold text-white">
                  {unreadNotifications > 9 ? "9+" : unreadNotifications}
                </span>
              )}
            </button>
            <Button onClick={logout} className="bg-slate-600 hover:bg-slate-500">
              Logout
            </Button>
          </div>
        </div>
      </header>
      <div className="mx-auto grid max-w-7xl grid-cols-1 gap-6 px-6 py-6 md:grid-cols-[220px_1fr]">
        <aside className="rounded-md border bg-white p-2">
          {activeRestaurant && (
            <div className="mb-3 rounded-md bg-slate-50 px-3 py-2">
              <p className="text-xs text-slate-500">Active Branch</p>
              <p className="text-sm font-medium">{activeRestaurant.name}</p>
            </div>
          )}
          {nav.map(({ label, href, icon: Icon }) => {
            const active = href === "/" ? pathname === "/" : pathname.startsWith(href);
            return (
              <Link
                className={`flex items-center gap-2 rounded px-3 py-2 text-sm ${active ? "bg-slate-900 text-white" : "text-slate-700 hover:bg-slate-100"}`}
                href={href}
                key={href}
              >
                <Icon className="h-4 w-4" />
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
