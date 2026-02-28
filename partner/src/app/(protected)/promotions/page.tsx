"use client";

import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardTitle } from "@/components/ui/card";
import { Plus } from "lucide-react";
import { formatCurrency } from "@/lib/utils";

type Promo = {
  id: string;
  code: string;
  type: "percentage" | "flat" | "cashback";
  amount: number;
  cap?: number;
  usageCount: number;
  status: "active" | "expired" | "draft";
  startsAt: string;
  endsAt: string;
  totalDiscountGiven: number;
  uniqueUsers: number;
};

const mockPromos: Promo[] = [
  { id: "promo-1", code: "WELCOME20", type: "percentage", amount: 20, cap: 200, usageCount: 456, status: "active", startsAt: "2024-12-01", endsAt: "2025-01-31", totalDiscountGiven: 45600, uniqueUsers: 312 },
  { id: "promo-2", code: "FLAT50", type: "flat", amount: 50, usageCount: 128, status: "active", startsAt: "2024-12-15", endsAt: "2025-01-15", totalDiscountGiven: 6400, uniqueUsers: 98 },
  { id: "promo-3", code: "CASHBACK10", type: "cashback", amount: 10, cap: 100, usageCount: 89, status: "active", startsAt: "2024-12-10", endsAt: "2025-02-28", totalDiscountGiven: 8900, uniqueUsers: 67 },
  { id: "promo-4", code: "SUMMER30", type: "percentage", amount: 30, cap: 300, usageCount: 1200, status: "expired", startsAt: "2024-06-01", endsAt: "2024-08-31", totalDiscountGiven: 186000, uniqueUsers: 890 },
];

const typeLabels = { percentage: "Percentage", flat: "Flat", cashback: "Cashback" };
const statusVariants = { active: "success" as const, expired: "danger" as const, draft: "default" as const };

export default function PromotionsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Promotions</h1>
        <Link href="/promotions/new">
          <Button>
            <Plus className="mr-1 h-4 w-4" />
            Create Promotion
          </Button>
        </Link>
      </div>

      {/* Performance Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <p className="text-xs text-slate-500">Active Promos</p>
          <p className="mt-1 text-2xl font-semibold">{mockPromos.filter((p) => p.status === "active").length}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Total Usage</p>
          <p className="mt-1 text-2xl font-semibold">{mockPromos.reduce((s, p) => s + p.usageCount, 0).toLocaleString()}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Total Discount Given</p>
          <p className="mt-1 text-2xl font-semibold">{formatCurrency(mockPromos.reduce((s, p) => s + p.totalDiscountGiven, 0))}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Unique Users</p>
          <p className="mt-1 text-2xl font-semibold">{mockPromos.reduce((s, p) => s + p.uniqueUsers, 0).toLocaleString()}</p>
        </Card>
      </div>

      {/* Promo Table */}
      <div className="rounded-md border bg-white p-4">
        <table className="w-full text-sm">
          <thead className="text-left text-slate-500">
            <tr>
              <th className="pb-2">Code</th>
              <th className="pb-2">Type</th>
              <th className="pb-2">Amount</th>
              <th className="pb-2">Cap</th>
              <th className="pb-2">Usage</th>
              <th className="pb-2">Status</th>
              <th className="pb-2">Period</th>
              <th className="pb-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            {mockPromos.map((promo) => (
              <tr key={promo.id} className="border-t">
                <td className="py-2 font-mono font-semibold">{promo.code}</td>
                <td className="py-2"><Badge>{typeLabels[promo.type]}</Badge></td>
                <td className="py-2">{promo.type === "percentage" ? `${promo.amount}%` : formatCurrency(promo.amount)}</td>
                <td className="py-2">{promo.cap ? formatCurrency(promo.cap) : "—"}</td>
                <td className="py-2 text-center">{promo.usageCount}</td>
                <td className="py-2"><Badge variant={statusVariants[promo.status]}>{promo.status}</Badge></td>
                <td className="py-2 text-xs">{promo.startsAt} → {promo.endsAt}</td>
                <td className="py-2">
                  <Link href={`/promotions/${promo.id}`}>
                    <Button className="text-xs bg-slate-600 hover:bg-slate-500">Edit</Button>
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
