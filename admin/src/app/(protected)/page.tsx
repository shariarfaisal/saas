import { RevenueTrendChart } from "@/components/dashboard/revenue-trend-chart";

const kpis = [
  ["Total orders today", "3,214"],
  ["Total commission", "৳128,430"],
  ["Active tenants", "86"],
  ["Active riders", "1,209"],
];

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        {kpis.map(([label, value]) => (
          <section className="rounded-md border bg-white p-4" key={label}>
            <p className="text-sm text-slate-500">{label}</p>
            <p className="mt-2 text-2xl font-semibold">{value}</p>
          </section>
        ))}
      </div>

      <RevenueTrendChart />

      <div className="grid gap-4 lg:grid-cols-2">
        <section className="rounded-md border bg-white p-4">
          <h2 className="mb-2 text-sm font-semibold">Active Tenants</h2>
          <table className="w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th>Name</th>
                <th>Status</th>
                <th>Orders</th>
              </tr>
            </thead>
            <tbody>
              {[
                ["Kacchi Bhai", "active", "421"],
                ["Pizza Hub", "active", "233"],
                ["Biryani Point", "suspended", "0"],
              ].map(([name, status, orders]) => (
                <tr className="border-t" key={name}>
                  <td className="py-2">{name}</td>
                  <td className="py-2">
                    <span
                      className={`rounded px-2 py-1 text-xs ${status === "active" ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}
                    >
                      {status}
                    </span>
                  </td>
                  <td className="py-2">{orders}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>

        <section className="rounded-md border bg-white p-4">
          <h2 className="mb-2 text-sm font-semibold">System Health</h2>
          <dl className="space-y-2 text-sm">
            <div className="flex justify-between border-b pb-2">
              <dt>API latency p95</dt>
              <dd className="font-semibold">182 ms</dd>
            </div>
            <div className="flex justify-between border-b pb-2">
              <dt>Error rate (5m)</dt>
              <dd className="font-semibold">0.42%</dd>
            </div>
            <div className="flex justify-between">
              <dt>Queue depth</dt>
              <dd className="font-semibold">37</dd>
            </div>
          </dl>
        </section>
      </div>
    </div>
  );
}
