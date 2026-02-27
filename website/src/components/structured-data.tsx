import { JsonLdScript } from "next-seo";

import type { TenantConfig } from "@/lib/tenant";

export function StructuredData({ tenant }: Readonly<{ tenant: TenantConfig }>) {
  return (
    <JsonLdScript
      scriptKey="website-jsonld"
      data={{
        "@context": "https://schema.org",
        "@type": "WebSite",
        name: tenant.displayName,
        url: `https://${tenant.slug}.platform.com`,
        potentialAction: {
          "@type": "SearchAction",
          target: `https://${tenant.slug}.platform.com/search?q={search_term_string}`,
          "query-input": "required name=search_term_string",
        },
      }}
    />
  );
}
