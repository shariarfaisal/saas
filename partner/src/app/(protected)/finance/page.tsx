"use client";

import Link from "next/link";
import { Card } from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";
import { AlertTriangle } from "lucide-react";

const summary = {
  currentPeriod: {
    label: "Dec 16 – Dec 31, 2024",
    grossSales: 245600,
    commission: 36840,
    promoDeductions: 12300,
    penalties: 500,
    adjustments: 1200,
    netPayable: 197260,
  },
  ytd: {
    grossSales: 4890000,
    commission: 733500,
    promoDeductions: 245000,
    netPayable: 3911500,
  },
  outstandingBalance: 42300,
};

export default function FinancePage() {
  return (
    <div className="space-y-6">
      <h1 className="text-lg font-semibold">Finance</h1>

      {/* Outstanding Balance Alert */}
      {summary.outstandingBalance > 0 && (
        <div className="flex items-center gap-3 rounded-md border border-amber-200 bg-amber-50 p-4">
          <AlertTriangle className="h-5 w-5 text-amber-600" />
          <div>
            <p className="text-sm font-semibold text-amber-800">Outstanding Balance</p>
            <p className="text-sm text-amber-700">
              You have {formatCurrency(summary.outstandingBalance)} in outstanding invoices.
            </p>
          </div>
        </div>
      )}

      {/* Current Period Summary */}
      <Card>
        <h2 className="mb-3 text-sm font-semibold">Current Period — {summary.currentPeriod.label}</h2>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between border-b pb-2">
            <span>Gross Sales</span>
            <span className="font-semibold">{formatCurrency(summary.currentPeriod.grossSales)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Platform Commission (15%)</span>
            <span className="font-semibold text-rose-600">-{formatCurrency(summary.currentPeriod.commission)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Vendor-Funded Promos</span>
            <span className="font-semibold text-rose-600">-{formatCurrency(summary.currentPeriod.promoDeductions)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Penalties</span>
            <span className="font-semibold text-rose-600">-{formatCurrency(summary.currentPeriod.penalties)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Adjustments</span>
            <span className="font-semibold text-emerald-600">+{formatCurrency(summary.currentPeriod.adjustments)}</span>
          </div>
          <div className="flex justify-between pt-2 text-base font-bold">
            <span>Net Payable</span>
            <span className="text-emerald-700">{formatCurrency(summary.currentPeriod.netPayable)}</span>
          </div>
        </div>
      </Card>

      {/* YTD Summary */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <p className="text-xs text-slate-500">YTD Gross Sales</p>
          <p className="mt-1 text-2xl font-semibold">{formatCurrency(summary.ytd.grossSales)}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">YTD Commission</p>
          <p className="mt-1 text-2xl font-semibold">{formatCurrency(summary.ytd.commission)}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">YTD Promo Deductions</p>
          <p className="mt-1 text-2xl font-semibold">{formatCurrency(summary.ytd.promoDeductions)}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">YTD Net Payable</p>
          <p className="mt-1 text-2xl font-semibold text-emerald-700">{formatCurrency(summary.ytd.netPayable)}</p>
        </Card>
      </div>

      {/* Quick Links */}
      <div className="grid gap-4 md:grid-cols-2">
        <Link href="/finance/invoices" className="rounded-md border bg-white p-4 hover:bg-slate-50">
          <h3 className="font-semibold">Invoices</h3>
          <p className="mt-1 text-sm text-slate-500">View and download your invoices</p>
        </Link>
        <Link href="/finance/payments" className="rounded-md border bg-white p-4 hover:bg-slate-50">
          <h3 className="font-semibold">Payment History</h3>
          <p className="mt-1 text-sm text-slate-500">Track received payments</p>
        </Link>
      </div>
    </div>
  );
}
