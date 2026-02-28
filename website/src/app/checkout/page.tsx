"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { ProtectedRoute } from "@/components/auth/protected-route";
import { useCartStore } from "@/stores/cart-store";
import { Area, fetchClientApi, calculateCharges, createOrder, ChargeBreakdown } from "@/lib/api";
import { useAuthStore } from "@/stores/auth-store";

export default function CheckoutPage() {
  const router = useRouter();
  const { items, subtotal, clearCart } = useCartStore();
  const { phone } = useAuthStore();
  
  const [areas, setAreas] = useState<Area[]>([]);
  const [loading, setLoading] = useState(false);
  const [calculating, setCalculating] = useState(false);
  const [breakdown, setBreakdown] = useState<ChargeBreakdown | null>(null);
  
  // Form State
  const [recipientName, setRecipientName] = useState("");
  const [recipientPhone, setRecipientPhone] = useState(phone || "");
  const [selectedArea, setSelectedArea] = useState("");
  const [addressDetails, setAddressDetails] = useState("");
  const [paymentMethod, setPaymentMethod] = useState("cod");
  const [promoCode, setPromoCode] = useState("");
  const [appliedPromo, setAppliedPromo] = useState("");
  const [promoError, setPromoError] = useState("");

  useEffect(() => {
    // Fetch areas for selection
    fetchClientApi<Area[]>("/areas").then(setAreas).catch(console.error);
  }, []);

  useEffect(() => {
    if (items.length === 0) {
      router.push("/");
      return;
    }
    
    // Recalculate charges whenever items, area or promo changes
    if (selectedArea) {
      handleCalculateCharges(appliedPromo);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [items, selectedArea]);

  const handleCalculateCharges = async (code?: string) => {
    setCalculating(true);
    setPromoError("");
    try {
      const res = await calculateCharges({
        items: items.map(item => ({
          product_id: item.productId,
          restaurant_id: item.restaurantId,
          quantity: item.quantity,
          unit_price: item.price.toString(),
          modifier_price: (item.modifiers?.reduce((s, m) => s + m.price, 0) || 0).toString(),
          item_discount: "0",
          item_vat: "0",
          product_name: item.name
        })),
        delivery_area: selectedArea,
        promo_code: code
      });
      setBreakdown(res);
      if (code && !res.promo_result?.valid) {
          setPromoError(res.promo_result?.error_message || "Invalid promo code");
          setAppliedPromo("");
      } else if (code && res.promo_result?.valid) {
          setAppliedPromo(code);
      }
    } catch (error) {
      console.error("Calculation failed", error);
    } finally {
      setCalculating(false);
    }
  };

  const handleApplyPromo = () => {
    handleCalculateCharges(promoCode);
  };

  const handlePlaceOrder = async () => {
    if (!selectedArea || !recipientName || !recipientPhone) {
      alert("Please fill in all required fields");
      return;
    }

    setLoading(true);
    try {
      const orderReq = {
        items: items.map(item => ({
          product_id: item.productId,
          restaurant_id: item.restaurantId,
          quantity: item.quantity,
          unit_price: item.price.toString(),
          modifier_price: (item.modifiers?.reduce((s, m) => s + m.price, 0) || 0).toString(),
          product_name: item.name,
          product_snapshot: {},
          selected_modifiers: item.modifiers || [],
          item_discount: "0",
          item_vat: "0"
        })),
        promo_code: appliedPromo,
        payment_method: paymentMethod,
        delivery_recipient_name: recipientName,
        delivery_recipient_phone: recipientPhone,
        delivery_area: selectedArea,
        customer_note: addressDetails
      };

      const res = await createOrder(orderReq);
      clearCart();
      
      if (res.payment_url) {
        window.location.href = res.payment_url;
      } else {
        router.push(`/orders/${res.order?.id || res.order?.ID || res.order?.order_number || ''}`);
      }
    } catch (error: Error | unknown) {
      alert((error as Error).message || "Failed to place order");
    } finally {
      setLoading(false);
    }
  };

  return (
    <ProtectedRoute>
      <main className="min-h-screen bg-neutral-50 pb-20">
        <header className="bg-white border-b border-neutral-100 sticky top-0 z-40">
          <div className="max-w-5xl mx-auto px-4 h-16 flex items-center justify-between">
            <button onClick={() => router.back()} className="w-10 h-10 rounded-full bg-neutral-100 flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            </button>
            <h1 className="text-xl font-extrabold text-neutral-900">Checkout</h1>
            <div className="w-10" />
          </div>
        </header>

        <div className="max-w-5xl mx-auto px-4 py-8 grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column: Form */}
          <div className="lg:col-span-2 space-y-8">
            {/* Delivery Details */}
            <section className="bg-white rounded-3xl p-6 sm:p-8 shadow-sm border border-neutral-100">
              <div className="flex items-center gap-3 mb-6">
                <div className="w-10 h-10 rounded-xl bg-orange-100 flex items-center justify-center text-orange-600">
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
                </div>
                <h2 className="text-2xl font-bold text-neutral-900">Delivery Details</h2>
              </div>

              <div className="space-y-6">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-sm font-bold text-neutral-700 uppercase tracking-wider ml-1">Recipient Name *</label>
                    <input 
                      type="text" 
                      value={recipientName}
                      onChange={(e) => setRecipientName(e.target.value)}
                      placeholder="e.g. John Doe" 
                      className="w-full bg-neutral-50 border border-neutral-100 rounded-2xl px-5 py-4 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:bg-white transition-all font-medium" 
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-bold text-neutral-700 uppercase tracking-wider ml-1">Phone Number *</label>
                    <input 
                      type="tel" 
                      value={recipientPhone}
                      onChange={(e) => setRecipientPhone(e.target.value)}
                      placeholder="e.g. 01712345678" 
                      className="w-full bg-neutral-50 border border-neutral-100 rounded-2xl px-5 py-4 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:bg-white transition-all font-medium" 
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-bold text-neutral-700 uppercase tracking-wider ml-1">Delivery Area *</label>
                  <select 
                    value={selectedArea}
                    onChange={(e) => setSelectedArea(e.target.value)}
                    className="w-full bg-neutral-50 border border-neutral-100 rounded-2xl px-5 py-4 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:bg-white transition-all font-medium appearance-none cursor-pointer"
                  >
                    <option value="">Select your area</option>
                    {areas.map(area => (
                      <option key={area.id} value={area.slug}>{area.name}</option>
                    ))}
                  </select>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-bold text-neutral-700 uppercase tracking-wider ml-1">Address Details (Apartment, Floor, etc.)</label>
                  <textarea 
                    rows={3}
                    value={addressDetails}
                    onChange={(e) => setAddressDetails(e.target.value)}
                    placeholder="e.g. House 12, Road 4, Flat B2" 
                    className="w-full bg-neutral-50 border border-neutral-100 rounded-2xl px-5 py-4 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:bg-white transition-all font-medium resize-none" 
                  />
                </div>
              </div>
            </section>

            {/* Payment Method */}
            <section className="bg-white rounded-3xl p-6 sm:p-8 shadow-sm border border-neutral-100">
              <div className="flex items-center gap-3 mb-6">
                <div className="w-10 h-10 rounded-xl bg-orange-100 flex items-center justify-center text-orange-600">
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>
                </div>
                <h2 className="text-2xl font-bold text-neutral-900">Payment Method</h2>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                {[
                  { id: "cod", name: "Cash on Delivery", icon: "💵" },
                  { id: "bkash", name: "bKash", icon: "🇧" },
                  { id: "aamarpay", name: "Card / Online", icon: "💳" }
                ].map((method) => (
                  <button
                    key={method.id}
                    onClick={() => setPaymentMethod(method.id)}
                    className={`p-5 rounded-2xl border-2 flex flex-col items-center gap-3 transition-all ${
                      paymentMethod === method.id 
                        ? "border-orange-500 bg-orange-50/30" 
                        : "border-neutral-100 hover:border-orange-200"
                    }`}
                  >
                    <span className="text-2xl">{method.icon}</span>
                    <span className="font-bold text-sm text-neutral-900">{method.name}</span>
                    <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center mt-auto ${
                      paymentMethod === method.id ? "border-orange-500" : "border-neutral-200"
                    }`}>
                      {paymentMethod === method.id && <div className="w-2.5 h-2.5 bg-orange-500 rounded-full" />}
                    </div>
                  </button>
                ))}
              </div>
            </section>
          </div>

          {/* Right Column: Order Summary */}
          <div className="space-y-6">
            <section className="bg-white rounded-3xl p-6 shadow-sm border border-neutral-100 sticky top-24">
              <h3 className="text-xl font-bold text-neutral-900 mb-6">Order Summary</h3>
              
              <div className="space-y-4 mb-8 max-h-60 overflow-y-auto pr-2 custom-scrollbar">
                {items.map((item) => (
                  <div key={item.id} className="flex gap-3">
                    <div className="relative w-12 h-12 rounded-lg overflow-hidden bg-neutral-100 shrink-0">
                      <Image src={item.image || "https://placehold.co/100x100/png?text=Food"} alt={item.name} fill className="object-cover" />
                    </div>
                    <div className="flex-1 min-w-0">
                       <p className="font-bold text-neutral-900 text-sm truncate uppercase">{item.name}</p>
                       <p className="text-xs text-neutral-500 italic mt-0.5">Qty: {item.quantity}</p>
                    </div>
                    <span className="font-bold text-neutral-900 text-sm">
                      ${((item.price + (item.modifiers?.reduce((s, m) => s + m.price, 0) || 0)) * item.quantity).toFixed(2)}
                    </span>
                  </div>
                ))}
              </div>

              {/* Promo Code */}
              <div className="mb-8">
                 <div className="flex gap-2">
                    <input 
                      type="text" 
                      placeholder="Promo Code" 
                      value={promoCode}
                      onChange={(e) => setPromoCode(e.target.value)}
                      className="flex-1 bg-neutral-50 border border-neutral-100 rounded-xl px-4 py-2 text-sm font-bold focus:outline-none focus:ring-2 focus:ring-orange-500/20 uppercase"
                    />
                    <button 
                      onClick={handleApplyPromo}
                      disabled={calculating || !promoCode}
                      className="bg-neutral-900 text-white text-xs font-black uppercase px-4 py-2 rounded-xl hover:bg-neutral-800 transition-colors disabled:opacity-50"
                    >
                      Apply
                    </button>
                 </div>
                 {promoError && <p className="text-red-500 text-xs mt-2 font-bold">{promoError}</p>}
                 {appliedPromo && <p className="text-green-600 text-xs mt-2 font-bold flex items-center gap-1">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><path d="M20 6 9 17l-5-5"/></svg>
                    Code {appliedPromo} applied
                 </p>}
              </div>

              {/* Price Breakdown */}
              <div className="space-y-3 pt-6 border-t border-neutral-100">
                <div className="flex justify-between text-neutral-500 font-medium">
                  <span>Subtotal</span>
                  <span className="text-neutral-900">${(breakdown?.subtotal || subtotal()).toFixed(2)}</span>
                </div>
                <div className="flex justify-between text-neutral-500 font-medium">
                  <span>Delivery Fee</span>
                  <span className="text-neutral-900 text-green-600">
                    {breakdown?.delivery_charge ? `$${breakdown.delivery_charge.toFixed(2)}` : "Select area"}
                  </span>
                </div>
                {breakdown && breakdown.promo_discount_total > 0 && (
                   <div className="flex justify-between text-orange-600 font-bold">
                    <span>Discount</span>
                    <span>-${breakdown.promo_discount_total.toFixed(2)}</span>
                  </div>
                )}
                <div className="flex justify-between text-xl font-black text-neutral-900 pt-3 border-t border-neutral-100">
                  <span>Total</span>
                  <span>${(breakdown?.total_amount || subtotal()).toFixed(2)}</span>
                </div>
              </div>

              <button 
                onClick={handlePlaceOrder}
                disabled={loading || calculating || !selectedArea}
                className="w-full bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-5 rounded-2xl shadow-xl shadow-orange-100 transition-all active:scale-[0.98] mt-8 flex items-center justify-center gap-3 disabled:opacity-50"
              >
                {loading ? <div className="w-6 h-6 border-3 border-white border-t-transparent rounded-full animate-spin" /> : (
                  <>
                    <span>Place Order</span>
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
                  </>
                )}
              </button>
            </section>
          </div>
        </div>
      </main>
    </ProtectedRoute>
  );
}
