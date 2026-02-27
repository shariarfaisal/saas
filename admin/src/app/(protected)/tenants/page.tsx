import Link from "next/link";

export default function TenantsPage() {
  return (
    <section className="space-y-4 rounded-md border bg-white p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Tenants</h1>
        <Link className="rounded bg-slate-900 px-3 py-2 text-sm text-white" href="/tenants/new">
          Create tenant
        </Link>
      </div>
      <table className="w-full text-sm">
        <thead className="text-left text-slate-500">
          <tr>
            <th>Name</th>
            <th>Plan</th>
            <th>Status</th>
            <th>Orders</th>
            <th>Commission</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          {["Kacchi Bhai", "Pizza Hub"].map((name) => (
            <tr className="border-t" key={name}>
              <td className="py-2">
                <Link className="underline" href="/tenants/tenant-1">
                  {name}
                </Link>
              </td>
              <td>Growth</td>
              <td>active</td>
              <td>320</td>
              <td>12%</td>
              <td>2026-02-21</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
