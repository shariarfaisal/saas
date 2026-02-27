import { create } from "zustand";

type AuthState = {
  isAuthenticated: boolean;
  phone?: string;
  setAuthenticated: (phone: string) => void;
  clearAuth: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  phone: undefined,
  setAuthenticated: (phone) => set({ isAuthenticated: true, phone }),
  clearAuth: () => set({ isAuthenticated: false, phone: undefined }),
}));
