import { HeroCarousel } from "@/components/home/hero-carousel";
import { StoryStrip } from "@/components/home/story-strip";
import { AreaSelector } from "@/components/home/area-selector";
import { RestaurantGrid } from "@/components/home/restaurant-grid";
import { StructuredData } from "@/components/structured-data";
import { UserHeaderActions } from "@/components/home/user-header-actions";
import { fetchServerApi } from "@/lib/server-api";
import { Banner, Story, Area, PagedResponse, Restaurant } from "@/lib/api";
import { resolveTenantConfig } from "@/lib/tenant";

export const revalidate = 60; // SSR with ISR every 60s

export default async function Home() {
  const tenant = await resolveTenantConfig();

  // Parallel fetch SSR data
  const [banners, stories, areas, restaurantsRes] = await Promise.all([
    fetchServerApi<Banner[]>("/storefront/banners").catch(() => []),
    fetchServerApi<Story[]>("/storefront/stories").catch(() => []),
    fetchServerApi<Area[]>("/storefront/areas").catch(() => []),
    fetchServerApi<PagedResponse<Restaurant>>("/storefront/restaurants?page=1&per_page=20").catch(() => ({
      data: [],
      meta: { total: 0, page: 1, per_page: 20 },
    })),
  ]);

  return (
    <main className="mx-auto flex min-h-screen max-w-5xl flex-col bg-white">
      <StructuredData tenant={tenant} />
      
      {/* Top Header Navigation for Storefront */}
      <header className="sticky top-0 z-40 bg-white/80 backdrop-blur-md border-b border-neutral-100 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-full bg-orange-500 flex items-center justify-center text-white font-bold pb-0.5">M</div>
          <span className="font-bold text-lg hidden sm:block tracking-tight text-neutral-900">{tenant.displayName}</span>
        </div>
        
        {/* Responsive Area Selector inside header */}
        <div className="flex-1 flex justify-center max-w-sm mx-4">
           {areas.length > 0 && <AreaSelector areas={areas} />}
        </div>
        
        <UserHeaderActions />
      </header>

      <div className="flex flex-col gap-8 px-4 sm:px-6 py-6 pb-20">
        <HeroCarousel banners={banners} />
        
        <StoryStrip stories={stories} />
        
        {/* Banner ad slot / Section simulation */}
        <div className="w-full bg-orange-50 rounded-2xl p-6 flex flex-col sm:flex-row items-center justify-between gap-6 border border-orange-100">
          <div>
            <h3 className="text-xl font-bold text-orange-900 mb-1">Cravings sorted in 30 minutes 🍔</h3>
            <p className="text-orange-700">Explore local favorities and top-rated restaurants near you.</p>
          </div>
          <button className="whitespace-nowrap bg-orange-500 text-white px-6 py-2.5 rounded-full font-medium hover:bg-orange-600 transition-colors shadow-sm">
            View offers
          </button>
        </div>

        <RestaurantGrid initialData={restaurantsRes} />
      </div>
    </main>
  );
}
