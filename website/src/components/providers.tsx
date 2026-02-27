"use client";

import { QueryClientProvider } from "@tanstack/react-query";
import { createContext, useState } from "react";

import { makeQueryClient } from "@/lib/query-client";
import type { TenantConfig } from "@/lib/tenant";

export const TenantConfigContext = createContext<TenantConfig | null>(null);

export function AppProviders({
  tenant,
  children,
}: Readonly<{ tenant: TenantConfig; children: React.ReactNode }>) {
  const [queryClient] = useState(makeQueryClient);

  return (
    <TenantConfigContext.Provider value={tenant}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </TenantConfigContext.Provider>
  );
}
