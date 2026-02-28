import { create } from "zustand";

type Restaurant = {
  id: string;
  name: string;
  logo?: string;
  isAvailable: boolean;
};

type AuthState = {
  email: string | null;
  expiresAt: number | null;
  restaurants: Restaurant[];
  activeRestaurantId: string | null;
  unreadNotifications: number;
  setSession: (email: string, expiresAt: number) => void;
  clearSession: () => void;
  setRestaurants: (restaurants: Restaurant[]) => void;
  setActiveRestaurant: (id: string) => void;
  setUnreadNotifications: (count: number) => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  email: null,
  expiresAt: null,
  restaurants: [],
  activeRestaurantId: null,
  unreadNotifications: 0,
  setSession: (email, expiresAt) => set({ email, expiresAt }),
  clearSession: () => set({ email: null, expiresAt: null, restaurants: [], activeRestaurantId: null }),
  setRestaurants: (restaurants) =>
    set((state) => ({
      restaurants,
      activeRestaurantId: state.activeRestaurantId ?? restaurants[0]?.id ?? null,
    })),
  setActiveRestaurant: (id) => set({ activeRestaurantId: id }),
  setUnreadNotifications: (count) => set({ unreadNotifications: count }),
}));
