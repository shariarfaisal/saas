"use client";

import React, { useState, useEffect } from "react";
import Image from "next/image";
import { useCartStore } from "@/stores/cart-store";
import { useAuthStore } from "@/stores/auth-store";

export function CartDrawer({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) {
  const { items, updateQuantity, removeItem, subtotal } = useCartStore();
  const { isAuthenticated, openAuthModal } = useAuthStore();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMounted(true);
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "unset";
    }
    return () => { document.body.style.overflow = "unset"; };
  }, [isOpen]);

  if (!mounted) return null;
  if (!isOpen) return null;

  // Group items by restaurant
  const groupedItems = items.reduce((acc, item) => {
    if (!acc[item.restaurantId]) {
      acc[item.restaurantId] = {
        name: item.restaurantName,
        items: [],
      };
    }
    acc[item.restaurantId].items.push(item);
    return acc;
  }, {} as Record<string, { name: string; items: typeof items }>);

  const total = subtotal();

  return (
    <div className="fixed inset-0 z-[150] flex justify-end">
      <div className="absolute inset-0 bg-black/40 backdrop-blur-sm animate-in fade-in duration-300" onClick={onClose} />
      
      <div className="relative w-full max-w-md bg-white h-full shadow-2xl flex flex-col animate-in slide-in-from-right duration-500">
        {/* Header */}
        <div className="p-6 border-b border-neutral-100 flex items-center justify-between bg-white sticky top-0 z-10">
          <div>
            <h2 className="text-2xl font-extrabold text-neutral-900">Your Cart</h2>
            <p className="text-sm text-neutral-500 font-medium">{items.length} items from {Object.keys(groupedItems).length} restaurants</p>
          </div>
          <button 
            onClick={onClose}
            className="w-10 h-10 bg-neutral-100 rounded-full flex items-center justify-center text-neutral-900 hover:bg-neutral-200 transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-8 custom-scrollbar">
          {items.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center">
              <div className="w-24 h-24 bg-neutral-50 rounded-full flex items-center justify-center mb-6">
                <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#d1d5db" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/></svg>
              </div>
              <h3 className="text-xl font-bold text-neutral-900">Your cart is empty</h3>
              <p className="text-neutral-500 mt-2">Add some delicious items from your favorite restaurants!</p>
              <button 
                onClick={onClose}
                className="mt-8 bg-orange-500 text-white font-bold py-3 px-8 rounded-2xl hover:bg-orange-600 transition-all shadow-lg shadow-orange-100"
              >
                Browse Restaurants
              </button>
            </div>
          ) : (
            Object.entries(groupedItems).map(([resId, group]) => (
              <div key={resId} className="space-y-4">
                <div className="flex items-center gap-2 border-b border-neutral-100 pb-2">
                   <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-orange-500"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
                   <h3 className="font-bold text-neutral-900 uppercase tracking-tight">{group.name}</h3>
                </div>
                
                <div className="space-y-6">
                  {group.items.map((item) => (
                    <div key={item.id} className="flex gap-4">
                      <div className="relative w-20 h-20 rounded-xl overflow-hidden bg-neutral-100 shrink-0">
                        <Image src={item.image || "https://placehold.co/200x200/png?text=Food"} alt={item.name} fill className="object-cover" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex justify-between items-start gap-2">
                          <h4 className="font-bold text-neutral-900 leading-tight uppercase truncate">{item.name}</h4>
                          <button 
                            onClick={() => removeItem(item.id)}
                            className="text-neutral-300 hover:text-red-500 transition-colors"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/></svg>
                          </button>
                        </div>
                        
                        {item.modifiers && item.modifiers.length > 0 && (
                          <p className="text-xs text-neutral-400 mt-1 line-clamp-1 italic">
                            {item.modifiers.map(m => m.optionName).join(", ")}
                          </p>
                        )}
                        
                        <div className="mt-3 flex items-center justify-between">
                          <span className="font-bold text-neutral-900">
                            ${((item.price + (item.modifiers?.reduce((s, m) => s + m.price, 0) || 0)) * item.quantity).toFixed(2)}
                          </span>
                          
                          <div className="flex items-center bg-neutral-50 rounded-lg p-0.5 border border-neutral-100">
                            <button 
                              onClick={() => updateQuantity(item.id, item.quantity - 1)}
                              className="w-7 h-7 flex items-center justify-center text-neutral-900 hover:bg-white rounded-md transition-all"
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/></svg>
                            </button>
                            <span className="w-8 text-center text-sm font-bold">{item.quantity}</span>
                            <button 
                              onClick={() => updateQuantity(item.id, item.quantity + 1)}
                              className="w-7 h-7 flex items-center justify-center text-neutral-900 hover:bg-white rounded-md transition-all"
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/><path d="M12 5v14"/></svg>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))
          )}
        </div>

        {/* Footer */}
        {items.length > 0 && (
          <div className="p-6 bg-white border-t border-neutral-100 space-y-4 shadow-[0_-10px_20px_rgba(0,0,0,0.02)]">
            <div className="flex items-center justify-between text-neutral-500">
              <span className="font-medium">Subtotal</span>
              <span className="font-bold text-neutral-900 text-xl">${total.toFixed(2)}</span>
            </div>
            
            <p className="text-[10px] text-neutral-400 text-center uppercase tracking-widest font-bold">Taxes and delivery calculated at checkout</p>

            <button 
              onClick={() => {
                if (!isAuthenticated) {
                  openAuthModal();
                } else {
                  // Navigate to checkout
                  window.location.href = "/checkout";
                }
              }}
              className="w-full bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-5 rounded-2xl shadow-xl shadow-orange-100 transition-all active:scale-[0.98] flex items-center justify-center gap-3"
            >
              <span>Go to Checkout</span>
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
