"use client";

import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useAudioNotification } from "@/hooks/use-audio-notification";

type IncomingOrder = {
  id: string;
  orderNumber: string;
  items: string;
  total: number;
  customerArea: string;
  createdAt: string;
};

const mockOrders: IncomingOrder[] = [
  {
    id: "ord-1",
    orderNumber: "KBC-001234",
    items: "2x Kacchi Biryani, 1x Borhani",
    total: 1450,
    customerArea: "Gulshan",
    createdAt: new Date(Date.now() - 60000).toISOString(),
  },
  {
    id: "ord-2",
    orderNumber: "KBC-001235",
    items: "1x Tandoori Chicken, 2x Naan",
    total: 890,
    customerArea: "Banani",
    createdAt: new Date(Date.now() - 30000).toISOString(),
  },
];

function CountdownTimer({ expiresAt, onExpire }: { expiresAt: number; onExpire: () => void }) {
  const [remaining, setRemaining] = useState(Math.max(0, Math.floor((expiresAt - Date.now()) / 1000)));

  useEffect(() => {
    const timer = setInterval(() => {
      const left = Math.max(0, Math.floor((expiresAt - Date.now()) / 1000));
      setRemaining(left);
      if (left <= 0) {
        onExpire();
        clearInterval(timer);
      }
    }, 1000);
    return () => clearInterval(timer);
  }, [expiresAt, onExpire]);

  const minutes = Math.floor(remaining / 60);
  const seconds = remaining % 60;
  const isUrgent = remaining < 60;

  return (
    <span className={`text-xs font-mono font-semibold ${isUrgent ? "text-rose-600" : "text-slate-600"}`}>
      {minutes}:{seconds.toString().padStart(2, "0")}
    </span>
  );
}

export function IncomingOrderPanel() {
  const [orders, setOrders] = useState<IncomingOrder[]>(mockOrders);
  const { enable, play } = useAudioNotification();

  // Enable audio on first interaction
  useEffect(() => {
    const handler = () => {
      enable();
      document.removeEventListener("click", handler);
    };
    document.addEventListener("click", handler);
    return () => document.removeEventListener("click", handler);
  }, [enable]);

  // SSE would connect here in production
  // useSSE({ url: "/api/v1/partner/orders/stream", onMessage: ... });

  const handleAccept = useCallback(
    (orderId: string) => {
      setOrders((prev) => prev.filter((o) => o.id !== orderId));
      play();
    },
    [play],
  );

  const handleReject = useCallback((orderId: string) => {
    setOrders((prev) => prev.filter((o) => o.id !== orderId));
  }, []);

  const handleExpire = useCallback((orderId: string) => {
    setOrders((prev) => prev.filter((o) => o.id !== orderId));
  }, []);

  if (orders.length === 0) {
    return (
      <div className="rounded-md border bg-white p-4">
        <h3 className="mb-2 text-sm font-semibold">Incoming Orders</h3>
        <p className="text-sm text-slate-500">No pending orders right now.</p>
      </div>
    );
  }

  return (
    <div className="rounded-md border bg-white p-4">
      <div className="mb-3 flex items-center gap-2">
        <h3 className="text-sm font-semibold">Incoming Orders</h3>
        <Badge variant="danger">{orders.length} new</Badge>
      </div>
      <div className="space-y-3">
        {orders.map((order) => (
          <div key={order.id} className="rounded-md border border-amber-200 bg-amber-50 p-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-semibold">{order.orderNumber}</span>
              <CountdownTimer
                expiresAt={new Date(order.createdAt).getTime() + 180000}
                onExpire={() => handleExpire(order.id)}
              />
            </div>
            <p className="mt-1 text-xs text-slate-600">{order.items}</p>
            <div className="mt-1 flex items-center justify-between">
              <span className="text-xs text-slate-500">{order.customerArea}</span>
              <span className="text-sm font-semibold">৳{order.total.toLocaleString()}</span>
            </div>
            <div className="mt-2 flex gap-2">
              <Button
                className="flex-1 bg-emerald-600 hover:bg-emerald-500"
                onClick={() => handleAccept(order.id)}
              >
                Accept
              </Button>
              <Button
                className="flex-1 bg-rose-600 hover:bg-rose-500"
                onClick={() => handleReject(order.id)}
              >
                Reject
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
