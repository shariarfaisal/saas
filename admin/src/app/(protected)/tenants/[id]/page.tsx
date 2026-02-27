export default function TenantDetailPage() {
  return (
    <div className="space-y-4">
      <section className="rounded-md border bg-white p-4">
        <h1 className="text-lg font-semibold">Tenant Detail / Edit</h1>
        <div className="mt-2 grid gap-2 md:grid-cols-2">
          <input className="rounded border px-3 py-2" defaultValue="Kacchi Bhai" />
          <input className="rounded border px-3 py-2" defaultValue="12" />
        </div>
      </section>
      <section className="rounded-md border bg-white p-4">
        <h2 className="font-semibold">Suspend / Reinstate</h2>
        <textarea className="mt-2 w-full rounded border p-2" placeholder="Reason (mandatory)" />
        <div className="mt-2 flex gap-2">
          <button className="rounded bg-rose-600 px-3 py-2 text-sm text-white">Suspend</button>
          <button className="rounded bg-emerald-600 px-3 py-2 text-sm text-white">Reinstate</button>
          <a
            className="rounded bg-slate-900 px-3 py-2 text-sm text-white"
            href="https://partner.platform.com"
            rel="noreferrer"
            target="_blank"
          >
            Impersonate
          </a>
        </div>
      </section>
      <section className="rounded-md border bg-white p-4">
        <h2 className="font-semibold">Tenant Analytics Drill-down</h2>
        <p className="text-sm text-slate-600">Orders (30d): 7,834 · Revenue: ৳42,18,000 · Churn: 1.8%</p>
      </section>
    </div>
  );
}
