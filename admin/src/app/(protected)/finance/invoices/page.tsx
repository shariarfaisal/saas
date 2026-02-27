import Link from "next/link";

export default function InvoicesPage() {
  return (
    <section className="space-y-3 rounded-md border bg-white p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Invoices</h1>
        <button className="rounded bg-slate-900 px-3 py-2 text-sm text-white">Generate invoice</button>
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr>
            <th className="text-left">Invoice</th>
            <th className="text-left">Tenant</th>
            <th className="text-left">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr className="border-t">
            <td className="py-2">
              <Link className="underline" href="/finance/invoices/inv-1">
                INV-001
              </Link>
            </td>
            <td>Kacchi Bhai</td>
            <td>draft</td>
          </tr>
        </tbody>
      </table>
    </section>
  );
}
