"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";

function PaymentSuccessContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const orderId = searchParams.get("order_id");

  useEffect(() => {
    if (orderId) {
      // Small delay just to let user read the success message briefly
      const timeout = setTimeout(() => {
        router.push(`/orders/${orderId}`);
      }, 3000);
      return () => clearTimeout(timeout);
    }
  }, [orderId, router]);

  return (
    <main className="min-h-screen flex items-center justify-center bg-neutral-50 px-4">
      <div className="bg-white rounded-3xl p-8 sm:p-12 shadow-sm border border-neutral-100 max-w-md w-full text-center">
        <div className="w-24 h-24 rounded-full bg-green-100 flex items-center justify-center text-5xl mx-auto mb-6">
          🎉
        </div>
        <h1 className="text-3xl font-black text-neutral-900 mb-2 tracking-tight">Payment Successful!</h1>
        <p className="text-neutral-500 mb-8 font-medium">
          Your payment was processed successfully. We&apos;re now preparing your order.
        </p>
        
        {orderId ? (
           <p className="text-sm text-neutral-400 font-bold animate-pulse">
             Redirecting you to the tracking page...
           </p>
        ) : (
           <button 
             onClick={() => router.push("/me/orders")}
             className="w-full bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-4 rounded-2xl shadow-xl shadow-orange-100 transition-all active:scale-[0.98]"
           >
             Go to My Orders
           </button>
        )}
      </div>
    </main>
  );
}

export default function PaymentSuccessPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><div className="w-12 h-12 border-4 border-orange-500 border-t-transparent rounded-full animate-spin" /></div>}>
      <PaymentSuccessContent />
    </Suspense>
  );
}
