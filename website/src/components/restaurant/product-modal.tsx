"use client";

import React, { useState, useEffect } from "react";
import Image from "next/image";
import { ModifierOption, ProductDetail, ModifierGroup } from "@/lib/api";
import { useCartStore } from "@/stores/cart-store";

type ProductModalProps = {
  product: ProductDetail;
  restaurantId: string;
  restaurantName: string;
  isOpen: boolean;
  onClose: () => void;
};

export function ProductModal({ 
  product, 
  restaurantId, 
  restaurantName, 
  isOpen, 
  onClose 
}: ProductModalProps) {
  const [quantity, setQuantity] = useState(1);
  const [selectedOptions, setSelectedOptions] = useState<Record<string, string[]>>({});
  const addItem = useCartStore((state) => state.addItem);

  useEffect(() => {
    if (isOpen) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setQuantity(1);
      // Initialize mandatory selections
      const initial: Record<string, string[]> = {};
      product.modifier_groups?.forEach(group => {
        if (group.min_required === 1 && group.max_allowed === 1) {
          initial[group.id] = [group.options[0].id];
        }
      });
      setSelectedOptions(initial);
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => { document.body.style.overflow = 'unset'; };
  }, [isOpen, product]);

  if (!isOpen) return null;

  const handleOptionToggle = (group: ModifierGroup, option: ModifierOption) => {
    const current = selectedOptions[group.id] || [];
    let next: string[];

    if (group.max_allowed === 1) {
      next = [option.id];
    } else {
      if (current.includes(option.id)) {
        next = current.filter(id => id !== option.id);
      } else {
        if (current.length < group.max_allowed) {
          next = [...current, option.id];
        } else {
          next = current;
        }
      }
    }
    setSelectedOptions({ ...selectedOptions, [group.id]: next });
  };

  const calculateTotalPrice = () => {
    let price = product.price;
    product.modifier_groups?.forEach(group => {
      const selectedIds = selectedOptions[group.id] || [];
      selectedIds.forEach(id => {
        const option = group.options.find(o => o.id === id);
        if (option) price += option.additional_price;
      });
    });
    return price * quantity;
  };

  const handleAddToCart = () => {
    // Generate unique ID based on product + modifiers
    const modifierKey = Object.values(selectedOptions).flat().sort().join("-");
    const cartItemId = `${product.id}-${modifierKey}`;

    const modifiers: {
      groupId: string;
      groupName: string;
      optionId: string;
      optionName: string;
      price: number;
    }[] = [];

    for (const groupId in selectedOptions) {
      const group = product.modifier_groups?.find((g) => g.id === groupId);
      if (!group) continue;
      selectedOptions[groupId].forEach((optId) => {
        const opt = group.options.find((o) => o.id === optId);
        if (opt) {
          modifiers.push({
            groupId: group.id,
            groupName: group.name,
            optionId: opt.id,
            optionName: opt.name,
            price: opt.additional_price,
          });
        }
      });
    }

    addItem({
      id: cartItemId,
      productId: product.id,
      name: product.name,
      price: product.price,
      quantity,
      image: product.image_url,
      modifiers,
      restaurantId,
      restaurantName,
    });

    onClose();
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      
      <div className="relative bg-white w-full max-w-2xl rounded-3xl overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-300 max-h-[90vh] flex flex-col">
        {/* Close Button */}
        <button 
          onClick={onClose}
          className="absolute top-4 right-4 z-10 w-10 h-10 bg-white/80 backdrop-blur-md rounded-full flex items-center justify-center text-neutral-900 shadow-md hover:bg-white transition-colors"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
        </button>

        <div className="overflow-y-auto flex-1 custom-scrollbar">
          {/* Hero Image */}
          <div className="relative h-64 sm:h-80 bg-neutral-100 isolate">
            <Image 
              src={product.image_url || "https://placehold.co/800x600/png?text=Product"}
              alt={product.name}
              fill
              className="object-cover"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent" />
          </div>

          <div className="p-6 sm:p-8 pt-0 -mt-10 relative z-10">
            <div className="bg-white rounded-2xl p-6 shadow-sm border border-neutral-100">
              <h2 className="text-3xl font-extrabold text-neutral-900">{product.name}</h2>
              {product.description && (
                <p className="text-neutral-500 mt-2 font-medium leading-relaxed">{product.description}</p>
              )}
              <div className="mt-4 flex items-center gap-3">
                 <span className="text-2xl font-bold text-neutral-900">${product.price}</span>
                 {product.has_discount && product.discount_price && (
                   <span className="text-sm bg-orange-100 text-orange-600 font-bold px-2.5 py-1 rounded-lg">
                     Save ${product.price - product.discount_price}
                   </span>
                 )}
              </div>
            </div>

            {/* Modifiers */}
            <div className="mt-8 space-y-8">
              {product.modifier_groups?.map(group => (
                <div key={group.id}>
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-bold text-neutral-900">{group.name}</h3>
                    <div className="flex items-center gap-2">
                       {group.min_required > 0 && (
                         <span className="text-[10px] font-bold bg-orange-100 text-orange-600 px-2 py-0.5 rounded-full uppercase tracking-wider">Required</span>
                       )}
                       <span className="text-xs font-semibold text-neutral-400">Select up to {group.max_allowed}</span>
                    </div>
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    {group.options.map(option => {
                      const isSelected = (selectedOptions[group.id] || []).includes(option.id);
                      return (
                        <button
                          key={option.id}
                          onClick={() => handleOptionToggle(group, option)}
                          disabled={!option.is_available}
                          className={`flex items-center justify-between p-4 rounded-xl border-2 transition-all duration-200 ${
                            isSelected 
                              ? "border-orange-500 bg-orange-50/50" 
                              : "border-neutral-100 hover:border-orange-200"
                          } ${!option.is_available ? 'opacity-50 cursor-not-allowed' : ''}`}
                        >
                          <div className="flex items-center gap-3">
                            <div className={`w-5 h-5 rounded-md border-2 flex items-center justify-center transition-colors ${
                              isSelected ? "bg-orange-500 border-orange-500" : "bg-white border-neutral-200"
                            }`}>
                              {isSelected && <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round"><path d="M20 6 9 17l-5-5"/></svg>}
                            </div>
                            <span className={`font-semibold ${isSelected ? 'text-orange-900' : 'text-neutral-700'}`}>{option.name}</span>
                          </div>
                          {option.additional_price > 0 && (
                            <span className={`text-sm font-bold ${isSelected ? 'text-orange-600' : 'text-neutral-400'}`}>+${option.additional_price}</span>
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Footer Actions */}
        <div className="p-6 bg-white border-t border-neutral-100 flex flex-col sm:flex-row items-center gap-4 shadow-[0_-4px_10px_rgba(0,0,0,0.03)]">
          <div className="flex items-center justify-between w-full sm:w-auto bg-neutral-100 rounded-2xl p-1 shrink-0">
            <button 
              onClick={() => setQuantity(Math.max(1, quantity - 1))}
              className="w-10 h-10 flex items-center justify-center text-neutral-900 hover:bg-white rounded-xl transition-all"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/></svg>
            </button>
            <span className="w-12 text-center font-bold text-lg text-neutral-900">{quantity}</span>
            <button 
              onClick={() => setQuantity(quantity + 1)}
              className="w-10 h-10 flex items-center justify-center text-neutral-900 hover:bg-white rounded-xl transition-all"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/><path d="M12 5v14"/></svg>
            </button>
          </div>

          <button 
            onClick={handleAddToCart}
            className="w-full bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-4 px-8 rounded-2xl shadow-lg shadow-orange-200 transition-all active:scale-[0.98] flex items-center justify-between"
          >
            <span>Add {quantity} to Cart</span>
            <span>${calculateTotalPrice().toFixed(2)}</span>
          </button>
        </div>
      </div>
    </div>
  );
}
