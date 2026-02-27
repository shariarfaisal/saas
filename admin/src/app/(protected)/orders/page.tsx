import Link from "next/link";

export default function OrdersPage() {
  return (
    <div className="space-y-4 rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">Order Management</h1>
      <div className="grid gap-2 md:grid-cols-5">
        <input className="rounded border px-3 py-2" placeholder="Tenant" />
        <input className="rounded border px-3 py-2" placeholder="Status" />
        <input className="rounded border px-3 py-2" type="date" />
      </div>
      <section className="rounded border p-3">
        <h2 className="font-medium">Order Detail Modal</h2>
        <p className="text-sm text-slate-600">Timeline · payment events · rider history · audit entries</p>
        <div className="mt-2 grid gap-2 md:grid-cols-3">
          <select className="rounded border px-3 py-2">
            <option>Force status override</option>
            <option>confirmed</option>
            <option>preparing</option>
            <option>delivered</option>
          </select>
          <input className="rounded border px-3 py-2" placeholder="Mandatory reason" />
          <button className="rounded bg-slate-900 px-3 py-2 text-sm text-white">Apply override</button>
        </div>
        <Link className="mt-2 inline-block text-sm underline" href="/issues/issue-1">
          Link to issue resolution
        </Link>
      </section>
    </div>
  );
}
