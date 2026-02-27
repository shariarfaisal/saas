export default function InvoiceDetailPage() {
  return (
    <section className="space-y-3 rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">Invoice Detail</h1>
      <p className="text-sm text-slate-600">Approve/finalize/mark-paid actions</p>
      <div className="flex gap-2">
        <button className="rounded bg-amber-600 px-3 py-2 text-sm text-white">Approve</button>
        <button className="rounded bg-slate-900 px-3 py-2 text-sm text-white">Finalize</button>
        <button className="rounded bg-emerald-600 px-3 py-2 text-sm text-white">Mark paid</button>
      </div>
    </section>
  );
}
