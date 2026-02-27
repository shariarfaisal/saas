import { create } from "zustand";

type AuthState = {
  email: string | null;
  expiresAt: number | null;
  setSession: (email: string, expiresAt: number) => void;
  clearSession: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  email: null,
  expiresAt: null,
  setSession: (email, expiresAt) => set({ email, expiresAt }),
  clearSession: () => set({ email: null, expiresAt: null }),
}));
