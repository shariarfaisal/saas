"use client";

import React from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { ProtectedRoute } from "@/components/auth/protected-route";

const ACCOUNT_LINKS = [
  { href: "/me", label: "Profile", icon: "👤" },
  { href: "/me/addresses", label: "Saved Addresses", icon: "📍" },
  { href: "/me/orders", label: "My Orders", icon: "🛍️" },
  { href: "/me/wallet", label: "Wallet", icon: "💳" },
  { href: "/me/favourites", label: "Favourites", icon: "❤️" },
  { href: "/me/notifications", label: "Notifications", icon: "🔔" },
];

export default function AccountLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();

  return (
    <ProtectedRoute>
      <main className="min-h-screen bg-neutral-50 pb-20">
        <header className="bg-white border-b border-neutral-100 sticky top-0 z-40 hidden md:block">
          <div className="max-w-6xl mx-auto px-4 h-16 flex items-center justify-between">
            <button onClick={() => router.push("/")} className="w-10 h-10 rounded-full bg-neutral-100 flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            </button>
            <h1 className="text-xl font-extrabold text-neutral-900">My Account</h1>
            <div className="w-10" />
          </div>
        </header>

        <div className="max-w-6xl mx-auto px-4 py-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            {/* Sidebar Navigation */}
            <div className="md:col-span-1 border-b md:border-b-0 md:border-r border-neutral-100 pb-4 md:pb-0 md:pr-4 overflow-x-auto">
              <nav className="flex md:flex-col gap-2 min-w-max md:min-w-0">
                {ACCOUNT_LINKS.map(link => {
                  const isActive = pathname === link.href;
                  return (
                    <Link
                      key={link.href}
                      href={link.href}
                      className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-bold text-sm ${
                        isActive ? "bg-orange-50 text-orange-600 shadow-sm" : "text-neutral-600 hover:bg-neutral-100"
                      }`}
                    >
                      <span className="text-lg">{link.icon}</span>
                      {link.label}
                    </Link>
                  )
                })}
              </nav>
            </div>

            {/* Main Content Area */}
            <div className="md:col-span-3">
              {children}
            </div>
          </div>
        </div>
      </main>
    </ProtectedRoute>
  );
}
