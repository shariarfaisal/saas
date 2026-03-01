"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";

type Payout = {
  id: string;
  tenant: string;
  restaurant: string;
  amount: number;
  period: string;
  status: "pending" | "processing" | "paid" | "failed";
  dueDate: string;
  paymentRef?: string;
};

const initialPayouts: Payout[] = [
  {
    id: "pay-1",
    tenant: "Pizza Hub",
    restaurant: "Pizza Hub - Gulshan",
    amount: 261800,
    period: "2026-02",
    status: "processing",
    dueDate: "2026-03-07",
  },
  {
    id: "pay-2",
    tenant: "Kacchi Bhai",
    restaurant: "Kacchi Bhai - Downtown",
    amount: 126400,
    period: "2026-02",
    status: "paid",
    dueDate: "2026-03-07",
    paymentRef: "TXN-20260307-001",
  },
  {
    id: "pay-3",
    tenant: "Kacchi Bhai",
    restaurant: "Kacchi Bhai - Main Branch",
    amount: 224800,
    period: "2026-02",
    status: "pending",
    dueDate: "2026-03-07",
  },
];

const STATUS_VARIANT: Record<string, "success" | "warning" | "info" | "danger" | "default"> = {
  paid: "success",
  processing: "info",
  pending: "warning",
  failed: "danger",
};

export default function PayoutsPage() {
  const [payouts, setPayouts] = useState(initialPayouts);
  const [markingId, setMarkingId] = useState<string | null>(null);

  const markAsPaid = (id: string) => {
    const ref = `TXN-${Date.now()}`;
    setMarkingId(id);
    setTimeout(() => {
      setPayouts((prev) =>
        prev.map((p) => (p.id === id ? { ...p, status: "paid", paymentRef: ref } : p)),
      );
      setMarkingId(null);
    }, 800);
  };

  const totalPending = payouts
    .filter((p) => p.status === "pending" || p.status === "processing")
    .reduce((s, p) => s + p.amount, 0);

  return (
    <section className="space-y-4 rounded-md border bg-white p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Payout Tracking</h1>
        <div className="rounded-md bg-amber-50 px-3 py-1.5 text-sm font-medium text-amber-700">
          Pending: ৳{totalPending.toLocaleString()}
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="border-b bg-slate-50">
            <tr>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Tenant</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Restaurant</th>
              <th className="px-3 py-2 text-right font-medium text-slate-600">Amount</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Period</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Due Date</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Status</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Ref</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Action</th>
            </tr>
          </thead>
          <tbody>
            {payouts.map((p) => (
              <tr key={p.id} className="border-b last:border-0 hover:bg-slate-50">
                <td className="px-3 py-2 font-medium">{p.tenant}</td>
                <td className="px-3 py-2 text-slate-600">{p.restaurant}</td>
                <td className="px-3 py-2 text-right font-semibold">৳{p.amount.toLocaleString()}</td>
                <td className="px-3 py-2">{p.period}</td>
                <td className="px-3 py-2">{p.dueDate}</td>
                <td className="px-3 py-2">
                  <Badge variant={STATUS_VARIANT[p.status]}>{p.status}</Badge>
                </td>
                <td className="px-3 py-2 text-xs text-slate-500">{p.paymentRef ?? "—"}</td>
                <td className="px-3 py-2">
                  {p.status !== "paid" && (
                    <button
                      disabled={markingId === p.id}
                      onClick={() => markAsPaid(p.id)}
                      className="rounded bg-emerald-600 px-2 py-1 text-xs text-white hover:bg-emerald-500 disabled:opacity-50"
                    >
                      {markingId === p.id ? "Saving…" : "Mark Paid"}
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
