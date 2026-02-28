"use client";

import { Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";

function PaymentFailedContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const orderId = searchParams.get("order_id");
  const message = searchParams.get("message") || "Your payment could not be processed.";

  return (
    <main className="min-h-screen flex items-center justify-center bg-neutral-50 px-4">
      <div className="bg-white rounded-3xl p-8 sm:p-12 shadow-sm border border-neutral-100 max-w-md w-full text-center">
        <div className="w-24 h-24 rounded-full bg-red-100 flex items-center justify-center text-5xl mx-auto mb-6">
          ⚠️
        </div>
        <h1 className="text-3xl font-black text-neutral-900 mb-2 tracking-tight">Payment Failed</h1>
        <p className="text-neutral-500 mb-8 font-medium">
          {message}
        </p>
        
        <div className="space-y-4">
          {orderId ? (
            <button 
              onClick={() => router.push(`/orders/${orderId}`)}
              className="w-full bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-4 rounded-2xl shadow-xl shadow-orange-100 transition-all active:scale-[0.98]"
            >
              Try Again
            </button>
          ) : (
            <button 
              onClick={() => router.push("/checkout")}
              className="w-full bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-4 rounded-2xl shadow-xl shadow-orange-100 transition-all active:scale-[0.98]"
            >
              Return to Checkout
            </button>
          )}
          
          <button 
            onClick={() => router.push("/")}
            className="w-full bg-neutral-100 hover:bg-neutral-200 text-neutral-900 font-extrabold py-4 rounded-2xl transition-all"
          >
            Go to Home
          </button>
        </div>
      </div>
    </main>
  );
}

export default function PaymentFailPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><div className="w-12 h-12 border-4 border-orange-500 border-t-transparent rounded-full animate-spin" /></div>}>
      <PaymentFailedContent />
    </Suspense>
  );
}
