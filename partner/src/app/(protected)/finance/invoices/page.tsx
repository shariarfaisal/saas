"use client";

import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatCurrency, formatDate } from "@/lib/utils";
import { Download } from "lucide-react";

type Invoice = {
  id: string;
  invoiceNumber: string;
  periodStart: string;
  periodEnd: string;
  grossSales: number;
  commission: number;
  promoDeductions: number;
  penalties: number;
  adjustments: number;
  netPayable: number;
  status: "draft" | "finalized" | "paid" | "overdue";
  issuedAt: string;
};

const mockInvoices: Invoice[] = [
  {
    id: "inv-1", invoiceNumber: "INV-2024-0024", periodStart: "2024-12-01", periodEnd: "2024-12-15",
    grossSales: 198000, commission: 29700, promoDeductions: 9800, penalties: 0, adjustments: 0, netPayable: 158500,
    status: "paid", issuedAt: "2024-12-16",
  },
  {
    id: "inv-2", invoiceNumber: "INV-2024-0025", periodStart: "2024-12-16", periodEnd: "2024-12-31",
    grossSales: 245600, commission: 36840, promoDeductions: 12300, penalties: 500, adjustments: 1200, netPayable: 197260,
    status: "finalized", issuedAt: "2025-01-01",
  },
  {
    id: "inv-3", invoiceNumber: "INV-2024-0023", periodStart: "2024-11-16", periodEnd: "2024-11-30",
    grossSales: 176300, commission: 26445, promoDeductions: 8400, penalties: 200, adjustments: 0, netPayable: 141255,
    status: "paid", issuedAt: "2024-12-01",
  },
  {
    id: "inv-4", invoiceNumber: "INV-2024-0022", periodStart: "2024-11-01", periodEnd: "2024-11-15",
    grossSales: 162400, commission: 24360, promoDeductions: 7200, penalties: 0, adjustments: 500, netPayable: 131340,
    status: "overdue", issuedAt: "2024-11-16",
  },
];

const statusVariants = {
  draft: "default" as const,
  finalized: "info" as const,
  paid: "success" as const,
  overdue: "danger" as const,
};

export default function InvoicesPage() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Invoices</h1>
      </div>

      <div className="rounded-md border bg-white p-4">
        <table className="w-full text-sm">
          <thead className="text-left text-slate-500">
            <tr>
              <th className="pb-2">Invoice #</th>
              <th className="pb-2">Period</th>
              <th className="pb-2">Gross Sales</th>
              <th className="pb-2">Commission</th>
              <th className="pb-2">Net Payable</th>
              <th className="pb-2">Status</th>
              <th className="pb-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            {mockInvoices.map((inv) => (
              <tr key={inv.id} className="border-t">
                <td className="py-2">
                  <Link href={`/finance/invoices/${inv.id}`} className="font-medium hover:underline">
                    {inv.invoiceNumber}
                  </Link>
                </td>
                <td className="py-2 text-xs">
                  {formatDate(inv.periodStart)} – {formatDate(inv.periodEnd)}
                </td>
                <td className="py-2">{formatCurrency(inv.grossSales)}</td>
                <td className="py-2 text-rose-600">{formatCurrency(inv.commission)}</td>
                <td className="py-2 font-semibold">{formatCurrency(inv.netPayable)}</td>
                <td className="py-2">
                  <Badge variant={statusVariants[inv.status]}>{inv.status}</Badge>
                </td>
                <td className="py-2">
                  <div className="flex gap-1">
                    <Link href={`/finance/invoices/${inv.id}`}>
                      <Button className="text-xs bg-slate-600 hover:bg-slate-500">View</Button>
                    </Link>
                    <Button className="text-xs bg-slate-200 text-slate-700 hover:bg-slate-300" title="Download PDF">
                      <Download className="h-3 w-3" />
                    </Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
