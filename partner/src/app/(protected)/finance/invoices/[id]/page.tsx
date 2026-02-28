"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatCurrency, formatDate } from "@/lib/utils";
import { Download } from "lucide-react";

const mockInvoice = {
  id: "inv-2",
  invoiceNumber: "INV-2024-0025",
  periodStart: "2024-12-16",
  periodEnd: "2024-12-31",
  issuedAt: "2025-01-01",
  status: "finalized" as const,
  restaurant: "Kacchi Bhai - Main Branch",
  tenant: "Kacchi Bhai Ltd.",
  breakdown: {
    grossSales: 245600,
    productDiscounts: 8200,
    netSales: 237400,
    commissionRate: 15,
    commissionAmount: 36840,
    vendorFundedPromos: 12300,
    platformFundedPromos: 4500,
    penalties: 500,
    adjustments: 1200,
    vat: 12780,
    deliveryRevenue: 18400,
    netPayable: 197260,
  },
  orderSummary: {
    totalOrders: 342,
    completedOrders: 328,
    cancelledOrders: 14,
    avgOrderValue: 718,
  },
};

export default function InvoiceDetailPage() {

  const handleDownloadPdf = () => {
    // In production, call GET /partner/finance/invoices/:id/pdf
    const element = document.createElement("a");
    element.setAttribute("href", "#");
    element.setAttribute("download", `${mockInvoice.invoiceNumber}.pdf`);
    element.click();
  };

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">{mockInvoice.invoiceNumber}</h1>
          <p className="text-sm text-slate-500">
            {formatDate(mockInvoice.periodStart)} – {formatDate(mockInvoice.periodEnd)}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="info">{mockInvoice.status}</Badge>
          <Button onClick={handleDownloadPdf}>
            <Download className="mr-1 h-4 w-4" />
            Download PDF
          </Button>
        </div>
      </div>

      {/* Invoice Header */}
      <Card>
        <div className="grid gap-4 md:grid-cols-2 text-sm">
          <div>
            <p className="text-xs text-slate-500">Restaurant</p>
            <p className="font-medium">{mockInvoice.restaurant}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Tenant</p>
            <p className="font-medium">{mockInvoice.tenant}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Issued Date</p>
            <p className="font-medium">{formatDate(mockInvoice.issuedAt)}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Invoice Period</p>
            <p className="font-medium">{formatDate(mockInvoice.periodStart)} – {formatDate(mockInvoice.periodEnd)}</p>
          </div>
        </div>
      </Card>

      {/* Order Summary */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <p className="text-xs text-slate-500">Total Orders</p>
          <p className="mt-1 text-xl font-semibold">{mockInvoice.orderSummary.totalOrders}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Completed</p>
          <p className="mt-1 text-xl font-semibold text-emerald-600">{mockInvoice.orderSummary.completedOrders}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Cancelled</p>
          <p className="mt-1 text-xl font-semibold text-rose-600">{mockInvoice.orderSummary.cancelledOrders}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Avg Order Value</p>
          <p className="mt-1 text-xl font-semibold">{formatCurrency(mockInvoice.orderSummary.avgOrderValue)}</p>
        </Card>
      </div>

      {/* Full Breakdown */}
      <Card>
        <h2 className="mb-4 text-sm font-semibold">Full Breakdown</h2>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between border-b pb-2">
            <span>Gross Food Sales</span>
            <span className="font-semibold">{formatCurrency(mockInvoice.breakdown.grossSales)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Product Discounts</span>
            <span className="text-rose-600">-{formatCurrency(mockInvoice.breakdown.productDiscounts)}</span>
          </div>
          <div className="flex justify-between border-b pb-2 font-medium">
            <span>Net Sales</span>
            <span>{formatCurrency(mockInvoice.breakdown.netSales)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Platform Commission ({mockInvoice.breakdown.commissionRate}%)</span>
            <span className="text-rose-600">-{formatCurrency(mockInvoice.breakdown.commissionAmount)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Vendor-Funded Promotions</span>
            <span className="text-rose-600">-{formatCurrency(mockInvoice.breakdown.vendorFundedPromos)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Platform-Funded Promotions</span>
            <span className="text-slate-400">{formatCurrency(mockInvoice.breakdown.platformFundedPromos)} (no deduction)</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Penalties</span>
            <span className="text-rose-600">-{formatCurrency(mockInvoice.breakdown.penalties)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Adjustments</span>
            <span className="text-emerald-600">+{formatCurrency(mockInvoice.breakdown.adjustments)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>VAT Collected</span>
            <span>{formatCurrency(mockInvoice.breakdown.vat)}</span>
          </div>
          <div className="flex justify-between border-b pb-2">
            <span>Delivery Revenue Share</span>
            <span>{formatCurrency(mockInvoice.breakdown.deliveryRevenue)}</span>
          </div>
          <div className="flex justify-between pt-3 text-base font-bold">
            <span>Net Payable to Restaurant</span>
            <span className="text-emerald-700">{formatCurrency(mockInvoice.breakdown.netPayable)}</span>
          </div>
        </div>
      </Card>
    </div>
  );
}
