"use client";

import React, { useState } from "react";
import { Category, Product, ProductDetail, fetchClientApi } from "@/lib/api";
import { ProductModal } from "./product-modal";

type ProductGridProps = { 
  categories: (Category & { products: Product[] })[];
  restaurantId: string;
  restaurantName: string;
};

export function ProductGrid({ categories, restaurantId, restaurantName }: ProductGridProps) {
  const [selectedProduct, setSelectedProduct] = useState<ProductDetail | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  if (!categories || categories.length === 0) return null;

  const handleProductClick = async (product: Product) => {
    if (!product.is_available) return;
    
    setIsLoading(true);
    try {
      const detail = await fetchClientApi<ProductDetail>(`/restaurants/${restaurantId}/products/${product.id}`);
      setSelectedProduct(detail);
    } catch (error) {
      console.error("Failed to fetch product details:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-12 mt-8 pb-10 px-4 sm:px-0">
      {categories.map((cat) => (
        <section key={cat.id} id={`category-${cat.id}`} className="scroll-mt-[150px]">
          <h2 className="text-2xl font-bold text-neutral-900 mb-6">{cat.name}</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4 lg:gap-6">
            {cat.products?.map((product) => (
              <div 
                key={product.id} 
                onClick={() => handleProductClick(product)}
                className={`flex gap-4 p-4 rounded-2xl border transition-all duration-200 cursor-pointer group ${
                  product.is_available 
                    ? "bg-white border-neutral-100 hover:border-orange-200 hover:shadow-md hover:shadow-orange-100/50" 
                    : "bg-neutral-50/50 border-neutral-100 opacity-60 pointer-events-none"
                }`}
              >
                <div className="flex-1 flex flex-col justify-between">
                  <div>
                    <div className="flex items-start justify-between gap-2">
                       <h3 className="font-semibold text-neutral-900 leading-tight line-clamp-2 group-hover:text-orange-600 transition-colors uppercase">{product.name}</h3>
                       {!product.is_available && (
                         <span className="shrink-0 bg-neutral-200 text-neutral-500 text-[10px] font-bold px-2 py-0.5 rounded uppercase tracking-wider">Out of stock</span>
                       )}
                    </div>
                    {product.description && (
                      <p className="text-sm text-neutral-500 mt-2 line-clamp-2">{product.description}</p>
                    )}
                  </div>
                  
                  <div className="mt-4 flex items-center gap-3">
                    {product.has_discount ? (
                      <>
                        <span className="font-bold text-lg text-orange-600">${product.discount_price}</span>
                        <span className="text-sm text-neutral-400 line-through">${product.price}</span>
                      </>
                    ) : (
                      <span className="font-bold text-lg text-neutral-900">${product.price}</span>
                    )}
                  </div>
                </div>

                <div className="relative w-28 h-28 sm:w-32 sm:h-32 shrink-0 rounded-xl overflow-hidden bg-neutral-100 isolate">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img 
                    src={product.image_url || "https://placehold.co/400x400/png?text=Food"} 
                    alt={product.name}
                    className={`w-full h-full object-cover group-hover:scale-110 transition-transform duration-500 ${!product.is_available ? 'grayscale' : ''}`}
                    loading="lazy"
                  />
                  {product.is_available && (
                    <div className="absolute bottom-2 right-2 w-8 h-8 bg-white/90 backdrop-blur-sm shadow-sm text-orange-600 rounded-full flex items-center justify-center group-hover:bg-orange-500 group-hover:text-white transition-all duration-200">
                      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/><path d="M12 5v14"/></svg>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </section>
      ))}

      {selectedProduct && (
        <ProductModal 
          product={selectedProduct}
          restaurantId={restaurantId}
          restaurantName={restaurantName}
          isOpen={!!selectedProduct}
          onClose={() => setSelectedProduct(null)}
        />
      )}

      {isLoading && (
        <div className="fixed inset-0 z-[110] bg-black/10 backdrop-blur-[1px] flex items-center justify-center">
          <div className="w-12 h-12 border-4 border-orange-500 border-t-transparent rounded-full animate-spin" />
        </div>
      )}
    </div>
  );
}
