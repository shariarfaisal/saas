"use client";

import { Badge } from "@/components/ui/badge";
import { formatCurrency, formatDate } from "@/lib/utils";

type Payment = {
  id: string;
  reference: string;
  amount: number;
  method: string;
  status: "completed" | "pending" | "failed";
  invoiceNumber: string;
  paidAt: string;
};

const mockPayments: Payment[] = [
  { id: "pay-1", reference: "TXN-2024-001234", amount: 158500, method: "Bank Transfer", status: "completed", invoiceNumber: "INV-2024-0024", paidAt: "2024-12-20" },
  { id: "pay-2", reference: "TXN-2024-001189", amount: 141255, method: "Bank Transfer", status: "completed", invoiceNumber: "INV-2024-0023", paidAt: "2024-12-05" },
  { id: "pay-3", reference: "TXN-2024-001145", amount: 131340, method: "bKash", status: "pending", invoiceNumber: "INV-2024-0022", paidAt: "2024-11-20" },
  { id: "pay-4", reference: "TXN-2024-001098", amount: 155600, method: "Bank Transfer", status: "completed", invoiceNumber: "INV-2024-0021", paidAt: "2024-11-05" },
];

const statusVariants = {
  completed: "success" as const,
  pending: "warning" as const,
  failed: "danger" as const,
};

export default function PaymentsPage() {
  return (
    <div className="space-y-4">
      <h1 className="text-lg font-semibold">Payment History</h1>

      <div className="rounded-md border bg-white p-4">
        <table className="w-full text-sm">
          <thead className="text-left text-slate-500">
            <tr>
              <th className="pb-2">Reference</th>
              <th className="pb-2">Invoice</th>
              <th className="pb-2">Amount</th>
              <th className="pb-2">Method</th>
              <th className="pb-2">Status</th>
              <th className="pb-2">Date</th>
            </tr>
          </thead>
          <tbody>
            {mockPayments.map((payment) => (
              <tr key={payment.id} className="border-t">
                <td className="py-2 font-mono text-xs">{payment.reference}</td>
                <td className="py-2">{payment.invoiceNumber}</td>
                <td className="py-2 font-semibold">{formatCurrency(payment.amount)}</td>
                <td className="py-2">{payment.method}</td>
                <td className="py-2">
                  <Badge variant={statusVariants[payment.status]}>{payment.status}</Badge>
                </td>
                <td className="py-2 text-xs">{formatDate(payment.paidAt)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
