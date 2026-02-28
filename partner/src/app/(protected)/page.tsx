"use client";

import { TrendChart } from "@/components/dashboard/trend-chart";
import { IncomingOrderPanel } from "@/components/dashboard/incoming-order-panel";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

const kpis = [
  { label: "Today's Orders", value: "127", change: "+12%" },
  { label: "Revenue", value: "৳84,320", change: "+8%" },
  { label: "Pending Orders", value: "5", change: "" },
  { label: "Avg Delivery Time", value: "32 min", change: "-3 min" },
];

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        {kpis.map(({ label, value, change }) => (
          <section className="rounded-md border bg-white p-4" key={label}>
            <p className="text-sm text-slate-500">{label}</p>
            <p className="mt-2 text-2xl font-semibold">{value}</p>
            {change && (
              <p className={`mt-1 text-xs ${change.startsWith("+") || change.startsWith("-") ? "text-emerald-600" : "text-slate-500"}`}>
                {change} vs yesterday
              </p>
            )}
          </section>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-[1fr_380px]">
        <TrendChart />
        <IncomingOrderPanel />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <section className="rounded-md border bg-white p-4">
          <h2 className="mb-3 text-sm font-semibold">Quick Actions</h2>
          <div className="flex flex-wrap gap-2">
            <Button className="bg-emerald-600 hover:bg-emerald-500">Toggle Restaurant Availability</Button>
            <Button className="bg-amber-600 hover:bg-amber-500">View Pending Issues</Button>
          </div>
        </section>

        <section className="rounded-md border bg-white p-4">
          <h2 className="mb-3 text-sm font-semibold">Recent Activity</h2>
          <div className="space-y-2">
            {[
              { text: "Order KBC-001230 delivered", time: "5 min ago", status: "success" },
              { text: "New rider application received", time: "12 min ago", status: "info" },
              { text: "Low stock alert: Borhani", time: "1 hr ago", status: "warning" },
            ].map(({ text, time, status }) => (
              <div key={text} className="flex items-center justify-between border-b pb-2 last:border-0">
                <div className="flex items-center gap-2">
                  <Badge variant={status as "success" | "info" | "warning"}>{status}</Badge>
                  <span className="text-sm">{text}</span>
                </div>
                <span className="text-xs text-slate-400">{time}</span>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}
