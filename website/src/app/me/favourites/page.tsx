"use client";

import React, { useState } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";

export default function FavouritesPage() {
  const router = useRouter();
  const [favourites, setFavourites] = useState([
    { id: "1", type: "restaurant", name: "Burger King", subtitle: "Fast Food • American", image: "https://placehold.co/400x300/png?text=Burger+King", rating: 4.5, time: "25-35 min" },
    { id: "2", type: "restaurant", name: "Starbucks", subtitle: "Coffee • Bakery", image: "https://placehold.co/400x300/png?text=Starbucks", rating: 4.8, time: "15-20 min" },
  ]);

  const removeFavourite = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setFavourites(favourites.filter(f => f.id !== id));
  };

  return (
    <div className="bg-white rounded-3xl p-6 sm:p-8 shadow-sm border border-neutral-100">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-bold text-neutral-900">Favourites</h2>
      </div>

      {favourites.length === 0 ? (
         <div className="text-center py-16 border-2 border-dashed border-neutral-100 rounded-3xl">
            <div className="w-20 h-20 bg-neutral-50 rounded-full flex mx-auto items-center justify-center text-4xl mb-4">❤️</div>
            <p className="text-neutral-900 font-bold mb-2">No favourites yet</p>
            <p className="text-neutral-500 font-medium text-sm mb-6">Save your favorite restaurants to access them quickly.</p>
            <button onClick={() => router.push("/")} className="bg-orange-500 text-white font-extrabold py-3 px-8 rounded-xl shadow-lg shadow-orange-100 active:scale-95 transition-transform">
              Browse Restaurants
            </button>
         </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {favourites.map(fav => (
            <div 
              key={fav.id} 
              onClick={() => router.push(`/restaurants/mock-slug-${fav.id}`)}
              className="bg-white border-2 border-neutral-100 rounded-2xl overflow-hidden hover:border-orange-200 hover:shadow-lg transition-all cursor-pointer group flex flex-col relative"
            >
              <div className="relative w-full h-32 bg-neutral-100">
                <Image src={fav.image} alt={fav.name} fill className="object-cover group-hover:scale-105 transition-transform duration-500" />
                <button 
                  onClick={(e) => removeFavourite(fav.id, e)}
                  className="absolute top-3 right-3 w-8 h-8 bg-white/90 backdrop-blur rounded-full flex items-center justify-center text-red-500 shadow-sm hover:scale-110 hover:bg-white transition-all z-10"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/></svg>
                </button>
              </div>
              <div className="p-4 flex-1 flex flex-col">
                <div className="flex justify-between items-start mb-1">
                   <h3 className="font-bold text-neutral-900 leading-tight">{fav.name}</h3>
                   <span className="flex items-center gap-1 text-xs font-bold bg-neutral-100 px-2 py-1 rounded-md text-neutral-700 shrink-0">
                     <span className="text-orange-500">★</span> {fav.rating}
                   </span>
                </div>
                <p className="text-xs text-neutral-500 font-medium mb-3">{fav.subtitle}</p>
                <div className="mt-auto flex items-center gap-1 text-xs text-neutral-500 font-bold bg-neutral-50 w-max px-2 py-1 rounded-md">
                   ⏱️ {fav.time}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
