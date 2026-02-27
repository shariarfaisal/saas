export default function CommissionsPage() {
  return (
    <section className="space-y-3 rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">Commission Ledger</h1>
      <div className="flex gap-2">
        <input className="rounded border px-3 py-2" placeholder="Period (YYYY-MM)" />
        <button className="rounded bg-slate-900 px-3 py-2 text-sm text-white">Filter</button>
        <button className="rounded bg-slate-700 px-3 py-2 text-sm text-white">Export CSV</button>
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr>
            <th className="text-left">Tenant</th>
            <th className="text-left">Commission</th>
            <th className="text-left">Period</th>
          </tr>
        </thead>
        <tbody>
          <tr className="border-t">
            <td className="py-2">Kacchi Bhai</td>
            <td>৳42,120</td>
            <td>2026-02</td>
          </tr>
        </tbody>
      </table>
    </section>
  );
}
