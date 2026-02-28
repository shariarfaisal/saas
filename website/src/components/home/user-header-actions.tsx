"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

export function UserHeaderActions() {
  const { isAuthenticated, openAuthModal } = useAuthStore();
  const router = useRouter();

  return (
    <div className="flex items-center gap-3">
      {isAuthenticated ? (
         <button 
           onClick={() => router.push("/me/orders")}
           className="h-10 px-4 rounded-full bg-neutral-100 flex items-center justify-center hover:bg-neutral-200 transition text-neutral-900 font-bold text-sm gap-2"
         >
           <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"/><path d="M3 6h18"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>
           <span className="hidden sm:inline">My Orders</span>
         </button>
      ) : (
        <button 
          onClick={openAuthModal}
          className="h-10 px-6 rounded-full bg-neutral-900 text-white flex items-center justify-center hover:bg-neutral-800 transition font-bold text-sm"
        >
          Sign In
        </button>
      )}
    </div>
  );
}
