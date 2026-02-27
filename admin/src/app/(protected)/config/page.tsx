export default function ConfigPage() {
  return (
    <section className="space-y-4 rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">Platform Config & Feature Flags</h1>
      <div className="grid gap-2 md:grid-cols-2">
        <label className="flex items-center gap-2 rounded border p-2">
          <input type="checkbox" defaultChecked /> Enable rider auto-dispatch
        </label>
        <label className="flex items-center gap-2 rounded border p-2">
          <input type="checkbox" /> Enable global promos
        </label>
      </div>
      <div className="grid gap-2 md:grid-cols-3">
        <input className="rounded border px-3 py-2" placeholder="Payment API key" />
        <select className="rounded border px-3 py-2">
          <option>Live</option>
          <option>Test</option>
        </select>
        <input className="rounded border px-3 py-2" placeholder="SMS provider key" />
      </div>
      <div className="rounded border p-3">
        <p className="text-sm">Maintenance mode requires explicit confirmation.</p>
        <button className="mt-2 rounded bg-rose-600 px-3 py-2 text-sm text-white">Enable maintenance mode</button>
      </div>
      <p className="text-xs text-slate-500">
        All saves include audit metadata payload (`actor_id`, `reason`, `request_id`).
      </p>
    </section>
  );
}
