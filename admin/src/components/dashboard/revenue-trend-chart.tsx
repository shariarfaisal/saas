"use client";

import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

const data = [
  { day: "Mon", revenue: 120000 },
  { day: "Tue", revenue: 132000 },
  { day: "Wed", revenue: 141000 },
  { day: "Thu", revenue: 158000 },
  { day: "Fri", revenue: 170000 },
  { day: "Sat", revenue: 179500 },
  { day: "Sun", revenue: 189000 },
];

export function RevenueTrendChart() {
  return (
    <div className="h-72 w-full rounded-md border bg-white p-4">
      <h3 className="mb-3 text-sm font-semibold">Revenue Trend (7 days)</h3>
      <ResponsiveContainer width="100%" height="90%">
        <LineChart data={data}>
          <XAxis dataKey="day" />
          <YAxis />
          <Tooltip formatter={(value: number | string | undefined) => `৳${Number(value ?? 0).toLocaleString()}`} />
          <Line dataKey="revenue" stroke="#0f172a" strokeWidth={2} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
