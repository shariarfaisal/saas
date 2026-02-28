import { notFound } from "next/navigation";
import Image from "next/image";
import { fetchServerApi } from "@/lib/server-api";
import { Restaurant, Product, Category } from "@/lib/api";
import { StickyNav } from "@/components/restaurant/sticky-nav";
import { ProductGrid } from "@/components/restaurant/product-grid";
import { UserHeaderActions } from "@/components/home/user-header-actions";

type RestaurantWithMenu = Restaurant & {
  categories: (Category & {
    products: Product[];
  })[];
};

export const revalidate = 60; // ISR

export default async function RestaurantPage({ params }: { params: { slug: string } }) {
  // Fetch from the backend exactly what the slug matches 
  let restaurant: RestaurantWithMenu | null = null;
  try {
    // We assume backend returns the full payload with categories & products
    // Alternatively, if the endpoint is separated, we would fetch them in parallel
    restaurant = await fetchServerApi<RestaurantWithMenu>(`/restaurants/${params.slug}`);
  } catch {
    // 404 handled gracefully
  }

  if (!restaurant) {
    notFound();
  }

  // Fallback map cuisines
  const cuisines = restaurant.cuisines?.join(" • ") || "Cafe • Comfort Food";

  return (
    <main className="mx-auto flex min-h-screen max-w-5xl flex-col bg-neutral-50 pb-24">
      {/* Search Header fallback/Nav */}
      <nav className="sticky top-0 z-50 bg-white/90 backdrop-blur-md border-b border-neutral-100 flex items-center justify-between p-4 px-4 sm:px-6">
         <div className="flex items-center gap-3">
            <button className="bg-neutral-100 w-10 h-10 rounded-full flex items-center justify-center hover:bg-neutral-200">
               <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            </button>
            <h1 className="font-bold text-lg hidden sm:block">{restaurant.name}</h1>
         </div>
         <div className="flex-1 max-w-md ml-4 mr-2 sm:mx-6 rounded-full bg-neutral-100 px-4 py-2 flex items-center gap-2">
           <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-neutral-400"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
           <input type="text" placeholder="Search menu..." className="bg-transparent border-none outline-none w-full text-sm font-medium placeholder:text-neutral-500" />
         </div>
          <UserHeaderActions />
      </nav>

      {/* Header Info */}
      <div className="relative w-full h-48 md:h-64 lg:h-80 bg-neutral-200">
        <Image
          src={restaurant.cover_image || "https://placehold.co/1200x600/png?text=Cover"}
          alt={restaurant.name}
          fill
          className="object-cover"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent" />
      </div>
      
      <div className="bg-white -mt-6 sm:-mt-10 mx-0 sm:mx-6 rounded-t-3xl sm:rounded-3xl relative z-10 p-6 shadow-sm border border-neutral-100">
         <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
           <div>
             <h1 className="text-3xl font-extrabold text-neutral-900">{restaurant.name}</h1>
             <p className="text-neutral-500 font-medium mt-1">{cuisines}</p>
             <p className="text-sm text-neutral-400 mt-2">{restaurant.description || "Fresh and delicious food delivered straight to your door."}</p>
           </div>
           
           <div className="flex items-center gap-3 bg-neutral-50 rounded-xl p-3 border border-neutral-100 self-stretch sm:self-auto">
             <div className="flex flex-col items-center justify-center px-2">
                <span className="flex items-center gap-1 font-bold text-neutral-900">
                   <svg className="w-4 h-4 text-orange-400" fill="currentColor" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" /></svg>
                   {restaurant.rating?.toFixed(1) || "4.8"}
                </span>
                <span className="text-xs text-neutral-500 font-medium mt-0.5">Ratings</span>
             </div>
             <div className="w-[1px] h-8 bg-neutral-200 mx-1"></div>
             <div className="flex flex-col items-center justify-center px-2">
                <span className="font-bold text-neutral-900">{restaurant.delivery_time_mins || 30}</span>
                <span className="text-xs text-neutral-500 font-medium mt-0.5">Mins</span>
             </div>
             <div className="w-[1px] h-8 bg-neutral-200 mx-1"></div>
             <div className="flex flex-col items-center justify-center px-2">
                <span className="font-bold text-orange-600">Free</span>
                <span className="text-xs text-neutral-500 font-medium mt-0.5">Delivery</span>
             </div>
           </div>
         </div>
      </div>

      <div className="px-0 sm:px-6 mt-4">
         <StickyNav categories={restaurant.categories || []} />
         <ProductGrid 
            categories={restaurant.categories || []} 
            restaurantId={restaurant.id}
            restaurantName={restaurant.name}
          />
      </div>
    </main>
  );
}
