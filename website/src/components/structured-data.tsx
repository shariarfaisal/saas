import { JsonLdScript } from "next-seo";

import { PLATFORM_DOMAIN } from "@/lib/site-config";
import type { TenantConfig } from "@/lib/tenant";

export function StructuredData({ tenant }: Readonly<{ tenant: TenantConfig }>) {
  return (
    <JsonLdScript
      scriptKey="website-jsonld"
      data={{
        "@context": "https://schema.org",
        "@type": "WebSite",
        name: tenant.displayName,
        url: `https://${tenant.slug}.${PLATFORM_DOMAIN}`,
        potentialAction: {
          "@type": "SearchAction",
          target: `https://${tenant.slug}.${PLATFORM_DOMAIN}/search?q={search_term_string}`,
          "query-input": "required name=search_term_string",
        },
      }}
    />
  );
}
