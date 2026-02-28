import { create } from "zustand";

type AuthState = {
  isAuthenticated: boolean;
  phone?: string;
  isAuthModalOpen: boolean;
  setAuthenticated: (phone: string) => void;
  clearAuth: () => void;
  openAuthModal: () => void;
  closeAuthModal: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  phone: undefined,
  isAuthModalOpen: false,
  setAuthenticated: (phone) => set({ isAuthenticated: true, phone }),
  clearAuth: () => set({ isAuthenticated: false, phone: undefined }),
  openAuthModal: () => set({ isAuthModalOpen: true }),
  closeAuthModal: () => set({ isAuthModalOpen: false }),
}));
