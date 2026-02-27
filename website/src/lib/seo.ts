import type { Metadata } from "next";

import { PLATFORM_DOMAIN } from "@/lib/site-config";
import type { TenantConfig } from "@/lib/tenant";

type SeoConfig = {
  titleTemplate: string;
  defaultTitle: string;
  description: string;
  canonical: string;
  openGraph: NonNullable<Metadata["openGraph"]>;
};

export function buildSeoConfig(tenant: TenantConfig): SeoConfig {
  const canonical = `https://${tenant.slug}.${PLATFORM_DOMAIN}`;

  return {
    titleTemplate: `%s | ${tenant.displayName}`,
    defaultTitle: `${tenant.displayName} Food Delivery`,
    description: `Order food from ${tenant.displayName} restaurants with fast delivery.`,
    canonical,
    openGraph: {
      type: "website",
      url: canonical,
      title: `${tenant.displayName} Food Delivery`,
      description: `Order food from ${tenant.displayName} restaurants with fast delivery.`,
      siteName: tenant.displayName,
    },
  };
}

export function toMetadata(config: SeoConfig): Metadata {
  return {
    metadataBase: new URL(config.canonical),
    title: {
      default: config.defaultTitle,
      template: config.titleTemplate,
    },
    description: config.description,
    alternates: { canonical: config.canonical },
    openGraph: config.openGraph,
  };
}
