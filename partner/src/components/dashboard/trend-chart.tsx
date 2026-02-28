"use client";

import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

const data = [
  { day: "Mon", orders: 48, revenue: 52000 },
  { day: "Tue", orders: 55, revenue: 61000 },
  { day: "Wed", orders: 62, revenue: 68000 },
  { day: "Thu", orders: 58, revenue: 64000 },
  { day: "Fri", orders: 71, revenue: 78000 },
  { day: "Sat", orders: 85, revenue: 94000 },
  { day: "Sun", orders: 79, revenue: 87000 },
];

export function TrendChart() {
  return (
    <div className="h-72 w-full rounded-md border bg-white p-4">
      <h3 className="mb-3 text-sm font-semibold">7-Day Trend</h3>
      <ResponsiveContainer width="100%" height="90%">
        <LineChart data={data}>
          <XAxis dataKey="day" />
          <YAxis yAxisId="left" />
          <YAxis yAxisId="right" orientation="right" />
          <Tooltip
            formatter={(value: number | string | undefined, name: string) =>
              name === "revenue" ? `৳${Number(value ?? 0).toLocaleString()}` : value
            }
          />
          <Line yAxisId="left" dataKey="orders" stroke="#0f172a" strokeWidth={2} name="Orders" />
          <Line yAxisId="right" dataKey="revenue" stroke="#059669" strokeWidth={2} name="Revenue" />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
