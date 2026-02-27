import type { Metadata } from "next";
import "./globals.css";

import { AppProviders } from "@/components/providers";
import { buildSeoConfig, toMetadata } from "@/lib/seo";
import { resolveTenantConfig } from "@/lib/tenant";

export async function generateMetadata(): Promise<Metadata> {
  const tenant = await resolveTenantConfig();
  return toMetadata(buildSeoConfig(tenant));
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const tenant = await resolveTenantConfig();

  return (
    <html lang="en">
      <body className="antialiased">
        <AppProviders tenant={tenant}>{children}</AppProviders>
      </body>
    </html>
  );
}
