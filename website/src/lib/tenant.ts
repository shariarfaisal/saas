import { headers } from "next/headers";

export type TenantConfig = {
  slug: string;
  displayName: string;
  primaryColor: string;
};

const DEFAULT_TENANT: TenantConfig = {
  slug: "munchies",
  displayName: "Munchies",
  primaryColor: "#f97316",
};

function resolveSlugFromHost(host: string | null): string {
  if (!host) {
    return DEFAULT_TENANT.slug;
  }

  const hostname = host.split(":")[0].toLowerCase();
  const segments = hostname.split(".");

  if (segments.length >= 3) {
    return segments[0];
  }

  return DEFAULT_TENANT.slug;
}

export async function resolveTenantConfig(): Promise<TenantConfig> {
  const requestHeaders = await headers();
  const slug = resolveSlugFromHost(requestHeaders.get("x-forwarded-host") ?? requestHeaders.get("host"));

  return {
    ...DEFAULT_TENANT,
    slug,
    displayName: slug.charAt(0).toUpperCase() + slug.slice(1),
  };
}
