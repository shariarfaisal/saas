import { StructuredData } from "@/components/structured-data";
import { PLATFORM_DOMAIN } from "@/lib/site-config";
import { resolveTenantConfig } from "@/lib/tenant";

export const revalidate = 60;

export default async function Home() {
  const tenant = await resolveTenantConfig();

  return (
    <main className="mx-auto flex min-h-screen max-w-5xl flex-col gap-6 px-6 py-10">
      <StructuredData tenant={tenant} />
      <section className="rounded-2xl bg-orange-500 px-6 py-10 text-white">
        <p className="text-sm uppercase tracking-wide">
          {tenant.slug}.{PLATFORM_DOMAIN}
        </p>
        <h1 className="text-3xl font-semibold">{tenant.displayName} website foundation is ready</h1>
        <p className="mt-2 text-orange-100">SSR tenant resolution, SEO metadata, and structured data are configured.</p>
      </section>
    </main>
  );
}
