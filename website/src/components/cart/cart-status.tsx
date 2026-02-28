"use client";

import React, { useState } from "react";
import { useCartStore } from "@/stores/cart-store";
import { CartDrawer } from "./cart-drawer";

export function CartStatus() {
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const totalItems = useCartStore((state) => state.totalItems());
  const subtotal = useCartStore((state) => state.subtotal());

  if (totalItems === 0) return null;

  return (
    <>
      <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-[90] w-full max-w-md px-6 animate-in slide-in-from-bottom duration-500">
        <button 
          onClick={() => setIsDrawerOpen(true)}
          className="w-full bg-neutral-900 text-white rounded-2xl p-4 flex items-center justify-between shadow-2xl hover:scale-[1.02] active:scale-[0.98] transition-all"
        >
          <div className="flex items-center gap-3">
             <div className="bg-orange-500 w-8 h-8 rounded-lg flex items-center justify-center font-bold text-sm">
               {totalItems}
             </div>
             <span className="font-bold text-sm uppercase tracking-wider">Review Cart</span>
          </div>
          <span className="font-bold text-lg">${subtotal.toFixed(2)}</span>
        </button>
      </div>

      <CartDrawer isOpen={isDrawerOpen} onClose={() => setIsDrawerOpen(false)} />
    </>
  );
}
