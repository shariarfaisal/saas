import type { Metadata } from "next";
import "./globals.css";

import { AppProviders } from "@/providers/app-providers";

export const metadata: Metadata = {
  title: "Munchies Super Admin",
  description: "Cross-tenant super admin portal",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="bg-slate-50 text-slate-900 antialiased">
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
