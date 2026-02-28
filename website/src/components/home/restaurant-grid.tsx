"use client";

import React, { useState } from "react";
import Link from "next/link";
import { Restaurant, PagedResponse } from "@/lib/api";

type RestaurantGridProps = {
  initialData: PagedResponse<Restaurant>;
  areaSlug?: string;
};

export function RestaurantGrid({ initialData }: RestaurantGridProps) {
  const [cuisineFilter, setCuisineFilter] = useState("all");
  // In a real app we'd fetch via internal API route (e.g. /api/restaurants) that proxies
  // to backend so we don't expose backend to browser. For MVP we'll just display initial data
  // or use a simple load more. Let's rely on initialData for SSR and assume simple layout here.

  const restaurants = initialData?.data || [];
  
  // Fake client-side filter for now
  const filtered = restaurants.filter(r => 
    cuisineFilter === "all" ? true : r.cuisines?.includes(cuisineFilter)
  );

  return (
    <div className="w-full">
      <div className="flex flex-col sm:flex-row gap-4 mb-6 justify-between items-center">
        <h2 className="text-2xl font-bold">Local Favorites</h2>
        <div className="flex items-center gap-3 w-full sm:w-auto overflow-x-auto pb-1 sm:pb-0 hide-scrollbar">
          <button 
             onClick={() => setCuisineFilter("all")} 
             className={`px-4 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${cuisineFilter === 'all' ? 'bg-orange-500 text-white' : 'bg-neutral-100 text-neutral-700 hover:bg-neutral-200'}`}>
            All Cuisines
          </button>
          <button 
             onClick={() => setCuisineFilter("Burgers")} 
             className={`px-4 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${cuisineFilter === 'Burgers' ? 'bg-orange-500 text-white' : 'bg-neutral-100 text-neutral-700 hover:bg-neutral-200'}`}>
            Burgers
          </button>
          <button 
             onClick={() => setCuisineFilter("Pizza")} 
             className={`px-4 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${cuisineFilter === 'Pizza' ? 'bg-orange-500 text-white' : 'bg-neutral-100 text-neutral-700 hover:bg-neutral-200'}`}>
            Pizza
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {filtered.length === 0 && (
          <p className="col-span-full py-10 text-center text-neutral-500">No restaurants found for the selected filter.</p>
        )}
        {filtered.map((restaurant) => (
          <Link key={restaurant.id} href={`/restaurants/${restaurant.slug}`} className="group block">
            <div className="relative w-full aspect-[4/3] rounded-2xl overflow-hidden bg-neutral-100 mb-3">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={restaurant.cover_image || "https://placehold.co/600x400/png?text=Restaurant"}
                alt={restaurant.name}
                className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                loading="lazy"
              />
              <div className="absolute top-4 left-4 flex flex-col gap-2">
                {!restaurant.is_open && (
                  <span className="bg-red-500/90 backdrop-blur-sm text-white text-xs font-bold px-2.5 py-1 rounded-lg">
                    CLOSED
                  </span>
                )}
                {restaurant.has_discount && (
                  <span className="bg-orange-500/90 backdrop-blur-sm text-white text-xs font-bold px-2.5 py-1 rounded-lg">
                    {restaurant.discount_price || "Offer"}
                  </span>
                )}
              </div>
              <div className="absolute top-4 right-4 bg-white/90 backdrop-blur-sm rounded-lg px-2 py-1 flex items-center gap-1 shadow-sm text-sm font-semibold">
                <svg className="w-4 h-4 text-orange-400" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
                {restaurant.rating?.toFixed(1) || "4.5"}
              </div>
            </div>
            
            <div className="flex justify-between items-start pt-1">
              <div>
                <h3 className="font-semibold text-lg text-neutral-900 group-hover:text-orange-600 transition-colors line-clamp-1">{restaurant.name}</h3>
                <p className="text-sm text-neutral-500 line-clamp-1">
                  {restaurant.cuisines?.join(" • ") || "Munchies Signature • Fast Food"}
                </p>
              </div>
              <div className="bg-neutral-100 rounded-lg px-2.5 py-1 text-xs font-medium text-neutral-700 whitespace-nowrap shadow-sm mt-0.5 border border-neutral-200/60">
                {restaurant.delivery_time_mins || 30} min
              </div>
            </div>
          </Link>
        ))}
      </div>
      
      {initialData?.meta && initialData.meta.total > filtered.length && (
         <div className="mt-8 flex justify-center">
             <button className="px-6 py-2.5 border border-neutral-200 shadow-sm rounded-xl font-medium text-neutral-700 hover:bg-neutral-50 transition-colors">
                Load more restaurants
             </button>
         </div>
      )}
    </div>
  );
}
