"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";

type CommissionRow = {
  id: string;
  tenant: string;
  restaurant: string;
  grossSales: number;
  commissionRate: number;
  commissionAmount: number;
  vatCollected: number;
  netPayable: number;
  period: string;
  status: "draft" | "finalized" | "paid";
};

const mockRows: CommissionRow[] = [
  {
    id: "inv-1",
    tenant: "Kacchi Bhai",
    restaurant: "Kacchi Bhai - Main Branch",
    grossSales: 281000,
    commissionRate: 15,
    commissionAmount: 42150,
    vatCollected: 14050,
    netPayable: 224800,
    period: "2026-02",
    status: "finalized",
  },
  {
    id: "inv-2",
    tenant: "Kacchi Bhai",
    restaurant: "Kacchi Bhai - Downtown",
    grossSales: 158000,
    commissionRate: 15,
    commissionAmount: 23700,
    vatCollected: 7900,
    netPayable: 126400,
    period: "2026-02",
    status: "paid",
  },
  {
    id: "inv-3",
    tenant: "Pizza Hub",
    restaurant: "Pizza Hub - Gulshan",
    grossSales: 340000,
    commissionRate: 18,
    commissionAmount: 61200,
    vatCollected: 17000,
    netPayable: 261800,
    period: "2026-02",
    status: "draft",
  },
];

const STATUS_VARIANT: Record<string, "success" | "warning" | "info" | "danger" | "default"> = {
  paid: "success",
  finalized: "info",
  draft: "warning",
};

export default function CommissionsPage() {
  const [period, setPeriod] = useState("2026-02");

  const filtered = mockRows.filter((r) => !period || r.period === period);
  const totalCommission = filtered.reduce((s, r) => s + r.commissionAmount, 0);
  const totalGross = filtered.reduce((s, r) => s + r.grossSales, 0);

  return (
    <section className="space-y-4 rounded-md border bg-white p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Commission Ledger</h1>
        <div className="flex gap-2">
          <input
            className="rounded border px-3 py-2 text-sm"
            placeholder="Period (YYYY-MM)"
            value={period}
            onChange={(e) => setPeriod(e.target.value)}
          />
          <button className="rounded bg-slate-900 px-3 py-2 text-sm text-white">Filter</button>
          <button className="rounded bg-slate-700 px-3 py-2 text-sm text-white">Export CSV</button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <div className="rounded-md border p-3">
          <p className="text-xs text-slate-500">Gross Sales</p>
          <p className="mt-1 text-lg font-semibold">৳{totalGross.toLocaleString()}</p>
        </div>
        <div className="rounded-md border p-3">
          <p className="text-xs text-slate-500">Total Commission</p>
          <p className="mt-1 text-lg font-semibold text-emerald-700">৳{totalCommission.toLocaleString()}</p>
        </div>
        <div className="rounded-md border p-3">
          <p className="text-xs text-slate-500">Finalized</p>
          <p className="mt-1 text-lg font-semibold">{filtered.filter((r) => r.status === "finalized").length}</p>
        </div>
        <div className="rounded-md border p-3">
          <p className="text-xs text-slate-500">Pending (Draft)</p>
          <p className="mt-1 text-lg font-semibold">{filtered.filter((r) => r.status === "draft").length}</p>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="border-b bg-slate-50">
            <tr>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Tenant</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Restaurant</th>
              <th className="px-3 py-2 text-right font-medium text-slate-600">Gross Sales</th>
              <th className="px-3 py-2 text-right font-medium text-slate-600">Rate</th>
              <th className="px-3 py-2 text-right font-medium text-slate-600">Commission</th>
              <th className="px-3 py-2 text-right font-medium text-slate-600">VAT</th>
              <th className="px-3 py-2 text-right font-medium text-slate-600">Net Payable</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Period</th>
              <th className="px-3 py-2 text-left font-medium text-slate-600">Status</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((row) => (
              <tr key={row.id} className="border-b last:border-0 hover:bg-slate-50">
                <td className="px-3 py-2">{row.tenant}</td>
                <td className="px-3 py-2 text-slate-600">{row.restaurant}</td>
                <td className="px-3 py-2 text-right">৳{row.grossSales.toLocaleString()}</td>
                <td className="px-3 py-2 text-right">{row.commissionRate}%</td>
                <td className="px-3 py-2 text-right font-medium text-emerald-700">৳{row.commissionAmount.toLocaleString()}</td>
                <td className="px-3 py-2 text-right">৳{row.vatCollected.toLocaleString()}</td>
                <td className="px-3 py-2 text-right font-semibold">৳{row.netPayable.toLocaleString()}</td>
                <td className="px-3 py-2">{row.period}</td>
                <td className="px-3 py-2">
                  <Badge variant={STATUS_VARIANT[row.status]}>{row.status}</Badge>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
