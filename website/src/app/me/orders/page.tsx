"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Order, fetchMyOrders } from "@/lib/api";

const STATUS_BADGE: Record<string, string> = {
  "pending": "bg-amber-100 text-amber-600",
  "created": "bg-blue-100 text-blue-600",
  "confirmed": "bg-emerald-100 text-emerald-600",
  "preparing": "bg-orange-100 text-orange-600",
  "ready": "bg-purple-100 text-purple-600",
  "picked": "bg-indigo-100 text-indigo-600",
  "delivered": "bg-green-100 text-green-600",
  "cancelled": "bg-red-100 text-red-600",
  "rejected": "bg-red-100 text-red-600",
};

export default function OrderHistoryPage() {
  const router = useRouter();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMyOrders()
      .then(res => setOrders(res.data))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="bg-white rounded-3xl p-6 shadow-sm border border-neutral-100">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-neutral-900">My Orders</h2>
      </div>

      <div>
          {loading ? (
            <div className="flex flex-col gap-4">
               {[1, 2, 3].map(i => (
                 <div key={i} className="h-32 bg-white rounded-3xl animate-pulse" />
               ))}
            </div>
          ) : orders.length === 0 ? (
            <div className="text-center py-20 bg-white rounded-3xl border border-neutral-100 shadow-sm">
               <div className="w-20 h-20 bg-neutral-50 rounded-full flex items-center justify-center mx-auto mb-6">
                 <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#d1d5db" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"/><path d="M3 6h18"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>
               </div>
               <h2 className="text-xl font-bold text-neutral-900">No orders yet</h2>
               <p className="text-neutral-500 mt-2">Hungry? Order something delicious!</p>
               <button onClick={() => router.push("/")} className="mt-8 bg-orange-500 text-white font-bold py-3 px-8 rounded-2xl">Browse Restaurants</button>
            </div>
          ) : (
            <div className="space-y-4">
              {orders.map(order => (
                <button 
                  key={order.id}
                  onClick={() => router.push(`/orders/${order.id}`)}
                  className="w-full bg-white rounded-3xl p-6 shadow-sm border border-neutral-100 text-left hover:border-orange-200 hover:shadow-md transition-all group"
                >
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <h3 className="font-black text-neutral-900 text-lg uppercase tracking-tighter">Order #{order.order_number}</h3>
                      <p className="text-xs text-neutral-400 font-bold mt-0.5">{new Date(order.created_at).toLocaleDateString()} • {new Date(order.created_at).toLocaleTimeString()}</p>
                    </div>
                    <span className={`px-4 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest ${STATUS_BADGE[order.status] || 'bg-neutral-100'}`}>
                      {order.status}
                    </span>
                  </div>

                  <div className="flex items-center justify-between pt-4 border-t border-neutral-50">
                    <div className="flex items-center gap-2">
                       <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-neutral-400 group-hover:text-orange-500 transition-colors"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
                       <span className="text-sm font-bold text-neutral-600">{order.delivery_area}</span>
                    </div>
                    <span className="text-lg font-black text-neutral-900">${order.total_amount.toFixed(2)}</span>
                  </div>
                </button>
              ))}
            </div>
          )}
      </div>
    </div>
  );
}
