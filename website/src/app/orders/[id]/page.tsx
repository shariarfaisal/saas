"use client";

import React, { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { ProtectedRoute } from "@/components/auth/protected-route";
import { fetchClientApi, OrderDetail } from "@/lib/api";

const STATUS_CONFIG: Record<string, { label: string; icon: string; color: string; step: number }> = {
  "pending": { label: "Pending Payment", icon: "⏳", color: "bg-amber-100 text-amber-600", step: 0 },
  "created": { label: "Order Received", icon: "📥", color: "bg-blue-100 text-blue-600", step: 1 },
  "confirmed": { label: "Restaurant Confirmed", icon: "✅", color: "bg-emerald-100 text-emerald-600", step: 2 },
  "preparing": { label: "Preparing Food", icon: "🍳", color: "bg-orange-100 text-orange-600", step: 3 },
  "ready": { label: "Ready for Pickup", icon: "🥡", color: "bg-purple-100 text-purple-600", step: 4 },
  "picked": { label: "Out for Delivery", icon: "🛵", color: "bg-indigo-100 text-indigo-600", step: 5 },
  "delivered": { label: "Delivered", icon: "🎉", color: "bg-green-100 text-green-600", step: 6 },
  "cancelled": { label: "Cancelled", icon: "❌", color: "bg-red-100 text-red-600", step: -1 },
  "rejected": { label: "Rejected", icon: "🚫", color: "bg-red-100 text-red-600", step: -1 },
};

export default function OrderTrackingPage() {
  const router = useRouter();
  const { id } = useParams();
  const [orderDetail, setOrderDetail] = useState<OrderDetail | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Initial fetch
    fetchClientApi<OrderDetail>(`/orders/${id}`)
      .then(setOrderDetail)
      .catch(console.error)
      .finally(() => setLoading(false));

    // Handle real-time updates via EventSource
    const eventSource = new EventSource(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/orders/${id}/tracking`, {
        withCredentials: true
    });

    eventSource.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            setOrderDetail(data);
        } catch (e) {
            console.error("Failed to parse tracking data", e);
        }
    };

    eventSource.onerror = (err) => {
        console.error("EventSource failed:", err);
        eventSource.close();
    };

    return () => eventSource.close();
  }, [id]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="w-12 h-12 border-4 border-orange-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!orderDetail) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
         <div className="text-center">
            <h1 className="text-2xl font-bold text-neutral-900">Order not found</h1>
            <button onClick={() => router.push("/")} className="mt-4 text-orange-500 font-bold">Back to Home</button>
         </div>
      </div>
    );
  }

  const { order, items, timeline } = orderDetail;
  const statusInfo = STATUS_CONFIG[order.status] || { label: order.status, icon: "📦", color: "bg-neutral-100", step: 0 };

  return (
    <ProtectedRoute>
      <main className="min-h-screen bg-neutral-50 pb-20">
        <header className="bg-white border-b border-neutral-100 sticky top-0 z-40">
          <div className="max-w-4xl mx-auto px-4 h-16 flex items-center justify-between">
            <button onClick={() => router.push("/")} className="w-10 h-10 rounded-full bg-neutral-100 flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="m15 18-6-6 6-6"/></svg>
            </button>
            <h1 className="text-xl font-extrabold text-neutral-900 uppercase tracking-tighter">Order #{order.order_number}</h1>
            <div className="w-10" />
          </div>
        </header>

        <div className="max-w-4xl mx-auto px-4 py-8 space-y-8">
          {/* Status Banner */}
          <section className="bg-white rounded-3xl p-8 shadow-sm border border-neutral-100 flex flex-col items-center text-center">
            <div className={`w-20 h-20 rounded-full ${statusInfo.color} flex items-center justify-center text-4xl mb-4 animate-bounce`}>
               {statusInfo.icon}
            </div>
            <h2 className="text-3xl font-black text-neutral-900 mb-2 uppercase tracking-tight">{statusInfo.label}</h2>
            <p className="text-neutral-500 font-medium max-w-sm">
               {order.status === 'delivered' ? "Your food has been delivered. Enjoy your meal!" : "We'll keep you updated on your order progress."}
            </p>
          </section>

          {/* Stepper */}
          {statusInfo.step >= 0 && (
            <section className="bg-white rounded-3xl p-8 shadow-sm border border-neutral-100 overflow-x-auto">
               <div className="flex items-center justify-between min-w-[600px] px-4">
                  {[
                    { s: 1, label: "Received", icon: "📥" },
                    { s: 2, label: "Confirmed", icon: "✅" },
                    { s: 3, label: "Preparing", icon: "🍳" },
                    { s: 5, label: "On the way", icon: "🛵" },
                    { s: 6, label: "Delivered", icon: "🎉" }
                  ].map((step, idx) => {
                    const isActive = statusInfo.step >= step.s;
                    const isProcessing = statusInfo.step + 1 === step.s && order.status !== 'delivered';
                    return (
                      <React.Fragment key={step.s}>
                        <div className="flex flex-col items-center gap-2 relative z-10">
                           <div className={`w-12 h-12 rounded-2xl flex items-center justify-center text-xl transition-all duration-500 ${
                             isActive ? "bg-orange-500 text-white shadow-lg shadow-orange-200 scale-110" : 
                             isProcessing ? "bg-orange-100 text-orange-500 animate-pulse border-2 border-orange-200" : "bg-neutral-100 text-neutral-400"
                           }`}>
                             {step.icon}
                           </div>
                           <span className={`text-[10px] font-black uppercase tracking-widest ${isActive ? 'text-orange-500' : 'text-neutral-400'}`}>{step.label}</span>
                        </div>
                        {idx < 4 && (
                          <div className={`flex-1 h-1.5 rounded-full mx-2 -mt-6 transition-all duration-1000 ${
                            statusInfo.step > step.s ? "bg-orange-500" : "bg-neutral-100"
                          }`} />
                        )}
                      </React.Fragment>
                    );
                  })}
               </div>
            </section>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
             {/* Order Details */}
             <div className="space-y-8">
                <section className="bg-white rounded-3xl p-6 shadow-sm border border-neutral-100">
                  <h3 className="text-lg font-bold text-neutral-900 mb-6 uppercase tracking-tight flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-orange-500"><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"/><path d="M3 6h18"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>
                    Items Summary
                  </h3>
                  <div className="space-y-4">
                    {items.map(item => (
                      <div key={item.id} className="flex justify-between items-center bg-neutral-50 p-4 rounded-2xl border border-neutral-100">
                        <div>
                           <p className="font-bold text-neutral-900 text-sm uppercase">{item.product_name}</p>
                           <p className="text-xs text-neutral-500 font-medium">Qty: {item.quantity}</p>
                        </div>
                        <span className="font-bold text-neutral-900">${item.item_total.toFixed(2)}</span>
                      </div>
                    ))}
                    <div className="pt-4 border-t border-neutral-100 space-y-2">
                       <div className="flex justify-between text-sm font-medium text-neutral-500">
                          <span>Subtotal</span>
                          <span>${order.subtotal.toFixed(2)}</span>
                       </div>
                       <div className="flex justify-between text-sm font-medium text-neutral-500">
                          <span>Delivery Fee</span>
                          <span>${order.delivery_charge.toFixed(2)}</span>
                       </div>
                       {order.promo_discount_total > 0 && (
                          <div className="flex justify-between text-sm font-bold text-orange-600">
                             <span>Promo Discount</span>
                             <span>-${order.promo_discount_total.toFixed(2)}</span>
                          </div>
                       )}
                       <div className="flex justify-between text-lg font-black text-neutral-900 pt-2 border-t border-neutral-100">
                          <span>Total</span>
                          <span>${order.total_amount.toFixed(2)}</span>
                       </div>
                    </div>
                  </div>
                </section>
             </div>

             {/* Delivery & Timeline */}
             <div className="space-y-8">
                <section className="bg-white rounded-3xl p-6 shadow-sm border border-neutral-100">
                  <h3 className="text-lg font-bold text-neutral-900 mb-6 uppercase tracking-tight flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-orange-500"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
                    Delivery Info
                  </h3>
                  <div className="space-y-4">
                     <div>
                        <p className="text-[10px] font-black text-neutral-400 uppercase tracking-widest">Recipient</p>
                        <p className="font-bold text-neutral-900">{order.delivery_recipient_name} ({order.delivery_recipient_phone})</p>
                     </div>
                     <div>
                        <p className="text-[10px] font-black text-neutral-400 uppercase tracking-widest">Address</p>
                        <p className="font-bold text-neutral-900">{order.delivery_area}</p>
                        {JSON.stringify(order.delivery_address) !== '{}' && (
                           <p className="text-sm text-neutral-500 mt-1">{JSON.stringify(order.delivery_address)}</p>
                        )}
                     </div>
                     <div>
                        <p className="text-[10px] font-black text-neutral-400 uppercase tracking-widest">Payment</p>
                        <p className="font-bold text-neutral-900 uppercase">{order.payment_method} - {order.payment_status}</p>
                     </div>
                  </div>
                </section>

                <section className="bg-white rounded-3xl p-6 shadow-sm border border-neutral-100">
                   <h3 className="text-lg font-bold text-neutral-900 mb-6 uppercase tracking-tight flex items-center gap-2">
                     <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-orange-500"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                     Order Updates
                   </h3>
                   <div className="space-y-6 relative ml-2">
                      <div className="absolute left-0 top-2 bottom-2 w-[2px] bg-neutral-100" />
                      {timeline.map((event, idx) => (
                        <div key={event.id} className="relative pl-6">
                           <div className={`absolute left-[-5px] top-1.5 w-3 h-3 rounded-full border-2 border-white ${idx === 0 ? 'bg-orange-500 scale-125' : 'bg-neutral-300'}`} />
                           <p className={`text-sm font-bold ${idx === 0 ? 'text-neutral-900' : 'text-neutral-500'}`}>{event.description}</p>
                           <p className="text-[10px] text-neutral-400 font-medium">{new Date(event.created_at).toLocaleTimeString()}</p>
                        </div>
                      )).reverse()}
                   </div>
                </section>
             </div>
          </div>
        </div>
      </main>
    </ProtectedRoute>
  );
}
