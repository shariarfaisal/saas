"use client";

import Link from "next/link";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth-store";
import { MapPin, Clock, Plus } from "lucide-react";

type Restaurant = {
  id: string;
  name: string;
  description: string;
  address: string;
  cuisines: string[];
  isAvailable: boolean;
  rating: number;
  totalOrders: number;
  prepTime: number;
};

const mockRestaurants: Restaurant[] = [
  {
    id: "rest-1",
    name: "Kacchi Bhai - Main Branch",
    description: "Authentic Kacchi Biryani & Bangladeshi Cuisine",
    address: "House 12, Road 5, Gulshan 2, Dhaka",
    cuisines: ["Bangladeshi", "Biryani", "Mughlai"],
    isAvailable: true,
    rating: 4.5,
    totalOrders: 12540,
    prepTime: 25,
  },
  {
    id: "rest-2",
    name: "Kacchi Bhai - Downtown",
    description: "Authentic Kacchi Biryani & Bangladeshi Cuisine",
    address: "Shop 3, Tower Plaza, Motijheel, Dhaka",
    cuisines: ["Bangladeshi", "Biryani"],
    isAvailable: false,
    rating: 4.3,
    totalOrders: 8320,
    prepTime: 30,
  },
];

export default function RestaurantsPage() {
  const [restaurants, setRestaurants] = useState(mockRestaurants);
  const setActiveRestaurant = useAuthStore((s) => s.setActiveRestaurant);

  const toggleAvailability = (id: string) => {
    setRestaurants((prev) =>
      prev.map((r) => (r.id === id ? { ...r, isAvailable: !r.isAvailable } : r)),
    );
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Restaurants</h1>
        <Link href="/restaurants/new">
          <Button>
            <Plus className="mr-1 h-4 w-4" />
            Add Restaurant
          </Button>
        </Link>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {restaurants.map((restaurant) => (
          <div key={restaurant.id} className="rounded-md border bg-white p-4">
            <div className="flex items-start justify-between">
              <div>
                <Link
                  href={`/restaurants/${restaurant.id}`}
                  className="text-base font-semibold hover:underline"
                  onClick={() => setActiveRestaurant(restaurant.id)}
                >
                  {restaurant.name}
                </Link>
                <p className="mt-1 text-sm text-slate-500">{restaurant.description}</p>
              </div>
              <label className="relative inline-flex cursor-pointer items-center">
                <input
                  type="checkbox"
                  className="peer sr-only"
                  checked={restaurant.isAvailable}
                  onChange={() => toggleAvailability(restaurant.id)}
                />
                <div className="h-6 w-11 rounded-full bg-slate-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:bg-white after:transition-all peer-checked:bg-emerald-500 peer-checked:after:translate-x-full"></div>
              </label>
            </div>

            <div className="mt-3 flex flex-wrap gap-1">
              {restaurant.cuisines.map((c) => (
                <Badge key={c}>{c}</Badge>
              ))}
            </div>

            <div className="mt-3 flex items-center gap-4 text-xs text-slate-500">
              <span className="flex items-center gap-1">
                <MapPin className="h-3 w-3" />
                {restaurant.address}
              </span>
            </div>

            <div className="mt-3 grid grid-cols-4 gap-2 text-center">
              <div>
                <p className="text-lg font-semibold">{restaurant.rating}</p>
                <p className="text-xs text-slate-500">Rating</p>
              </div>
              <div>
                <p className="text-lg font-semibold">{restaurant.totalOrders.toLocaleString()}</p>
                <p className="text-xs text-slate-500">Orders</p>
              </div>
              <div>
                <p className="text-lg font-semibold">{restaurant.prepTime}m</p>
                <p className="text-xs text-slate-500">Prep Time</p>
              </div>
              <div>
                <Badge variant={restaurant.isAvailable ? "success" : "danger"}>
                  {restaurant.isAvailable ? "Open" : "Closed"}
                </Badge>
              </div>
            </div>

            <div className="mt-3 flex gap-2">
              <Link href={`/restaurants/${restaurant.id}`} className="flex-1">
                <Button className="w-full bg-slate-600 hover:bg-slate-500">
                  <Clock className="mr-1 h-3 w-3" />
                  Edit
                </Button>
              </Link>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
