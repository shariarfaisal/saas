"use client";

import { useState } from "react";
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis, PieChart, Pie, Cell } from "recharts";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Card, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Download } from "lucide-react";
import { formatCurrency } from "@/lib/utils";

// Sales Data
const salesData = [
  { period: "Dec 1", orders: 42, revenue: 38200, commission: 5730, netPayable: 32470 },
  { period: "Dec 2", orders: 38, revenue: 34100, commission: 5115, netPayable: 28985 },
  { period: "Dec 3", orders: 51, revenue: 46800, commission: 7020, netPayable: 39780 },
  { period: "Dec 4", orders: 45, revenue: 41200, commission: 6180, netPayable: 35020 },
  { period: "Dec 5", orders: 58, revenue: 53400, commission: 8010, netPayable: 45390 },
  { period: "Dec 6", orders: 63, revenue: 57600, commission: 8640, netPayable: 48960 },
  { period: "Dec 7", orders: 55, revenue: 50100, commission: 7515, netPayable: 42585 },
];

// Top Products
const topProducts = [
  { name: "Kacchi Biryani (Full)", orders: 234, revenue: 152100, growth: "+12%" },
  { name: "Kacchi Biryani (Half)", orders: 189, revenue: 66150, growth: "+8%" },
  { name: "Tehari", orders: 156, revenue: 34320, growth: "+15%" },
  { name: "Chicken Reshmi Kebab", orders: 98, revenue: 27440, growth: "-3%" },
  { name: "Borhani", orders: 312, revenue: 18720, growth: "+22%" },
];

// Peak Hours Heatmap
const peakHoursData = [
  { day: "Mon", hours: [2, 3, 5, 8, 12, 18, 22, 25, 28, 24, 20, 15, 10, 6, 3] },
  { day: "Tue", hours: [1, 2, 4, 7, 11, 16, 20, 23, 26, 22, 18, 14, 9, 5, 2] },
  { day: "Wed", hours: [3, 4, 6, 9, 14, 20, 24, 28, 30, 26, 22, 16, 11, 7, 4] },
  { day: "Thu", hours: [2, 3, 5, 8, 13, 19, 23, 27, 29, 25, 21, 15, 10, 6, 3] },
  { day: "Fri", hours: [4, 5, 7, 10, 15, 22, 28, 32, 35, 30, 25, 18, 12, 8, 5] },
  { day: "Sat", hours: [5, 6, 8, 12, 18, 25, 30, 35, 38, 33, 28, 20, 14, 9, 6] },
  { day: "Sun", hours: [4, 5, 7, 11, 16, 23, 28, 33, 36, 31, 26, 19, 13, 8, 5] },
];
const hourLabels = ["8am", "9am", "10am", "11am", "12pm", "1pm", "2pm", "3pm", "4pm", "5pm", "6pm", "7pm", "8pm", "9pm", "10pm"];

// Order Status Breakdown
const orderStatusData = [
  { name: "Delivered", value: 328, color: "#059669" },
  { name: "Cancelled", value: 14, color: "#e11d48" },
  { name: "Rejected", value: 6, color: "#f59e0b" },
  { name: "In Progress", value: 5, color: "#3b82f6" },
];

// Rider Performance
const riderPerformance = [
  { name: "Karim Ahmed", deliveries: 156, avgTime: 24, rating: 4.8, completionRate: 98.5 },
  { name: "Sohel Rana", deliveries: 142, avgTime: 28, rating: 4.6, completionRate: 97.2 },
  { name: "Rafiq Islam", deliveries: 138, avgTime: 26, rating: 4.7, completionRate: 97.8 },
  { name: "Masud Parvez", deliveries: 121, avgTime: 30, rating: 4.5, completionRate: 96.1 },
];

// Customer Areas
const customerAreas = [
  { area: "Gulshan", orders: 89, revenue: 78200, percentage: 25.4 },
  { area: "Banani", orders: 67, revenue: 58400, percentage: 19.1 },
  { area: "Dhanmondi", orders: 54, revenue: 47200, percentage: 15.4 },
  { area: "Uttara", orders: 48, revenue: 41800, percentage: 13.7 },
  { area: "Mirpur", orders: 42, revenue: 36400, percentage: 12.0 },
  { area: "Others", orders: 50, revenue: 43600, percentage: 14.4 },
];

function getHeatColor(value: number): string {
  if (value >= 30) return "bg-emerald-600 text-white";
  if (value >= 20) return "bg-emerald-400 text-white";
  if (value >= 10) return "bg-emerald-200";
  if (value >= 5) return "bg-emerald-100";
  return "bg-slate-50";
}

export default function ReportsPage() {
  const [dateFrom, setDateFrom] = useState("2024-12-01");
  const [dateTo, setDateTo] = useState("2024-12-07");
  const [groupBy, setGroupBy] = useState("day");

  const handleExportCsv = () => {
    const headers = "Period,Orders,Revenue,Commission,Net Payable\n";
    const rows = salesData.map((d) => `${d.period},${d.orders},${d.revenue},${d.commission},${d.netPayable}`).join("\n");
    const blob = new Blob([headers + rows], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `sales-report-${dateFrom}-${dateTo}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Sales & Analytics</h1>
        <Button onClick={handleExportCsv}>
          <Download className="mr-1 h-4 w-4" />
          Export CSV
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-3 rounded-md border bg-white p-4">
        <div>
          <label className="mb-1 block text-xs text-slate-500">From</label>
          <Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} className="w-40" />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">To</label>
          <Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} className="w-40" />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">Group By</label>
          <Select value={groupBy} onChange={(e) => setGroupBy(e.target.value)} className="w-32">
            <option value="day">Day</option>
            <option value="week">Week</option>
            <option value="month">Month</option>
          </Select>
        </div>
      </div>

      {/* Sales Chart */}
      <Card>
        <CardTitle>Sales Report</CardTitle>
        <div className="mt-4 h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={salesData}>
              <XAxis dataKey="period" />
              <YAxis />
              <Tooltip formatter={(value: number | string | undefined) => formatCurrency(Number(value ?? 0))} />
              <Bar dataKey="revenue" fill="#0f172a" name="Revenue" />
              <Bar dataKey="netPayable" fill="#059669" name="Net Payable" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Top Selling Products */}
        <Card>
          <CardTitle>Top Selling Products</CardTitle>
          <table className="mt-3 w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th className="pb-2">Product</th>
                <th className="pb-2">Orders</th>
                <th className="pb-2">Revenue</th>
                <th className="pb-2">Growth</th>
              </tr>
            </thead>
            <tbody>
              {topProducts.map((p) => (
                <tr key={p.name} className="border-t">
                  <td className="py-2 font-medium">{p.name}</td>
                  <td className="py-2">{p.orders}</td>
                  <td className="py-2">{formatCurrency(p.revenue)}</td>
                  <td className="py-2">
                    <Badge variant={p.growth.startsWith("+") ? "success" : "danger"}>{p.growth}</Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>

        {/* Order Status Breakdown */}
        <Card>
          <CardTitle>Order Status Breakdown</CardTitle>
          <div className="mt-4 flex items-center justify-center gap-8">
            <div className="h-48 w-48">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={orderStatusData} cx="50%" cy="50%" innerRadius={40} outerRadius={70} dataKey="value" label={({ name, value }) => `${name}: ${value}`}>
                    {orderStatusData.map((entry) => (
                      <Cell key={entry.name} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="space-y-2">
              {orderStatusData.map((s) => (
                <div key={s.name} className="flex items-center gap-2 text-sm">
                  <span className="h-3 w-3 rounded" style={{ backgroundColor: s.color }}></span>
                  <span>{s.name}: {s.value}</span>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </div>

      {/* Peak Hours Heatmap */}
      <Card>
        <CardTitle>Peak Hours Heatmap</CardTitle>
        <div className="mt-4 overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr>
                <th className="pb-2 text-left">Day</th>
                {hourLabels.map((h) => (
                  <th key={h} className="pb-2 text-center">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {peakHoursData.map((row) => (
                <tr key={row.day}>
                  <td className="py-1 font-medium">{row.day}</td>
                  {row.hours.map((value, i) => (
                    <td key={i} className="py-1 text-center">
                      <span className={`inline-block h-7 w-7 rounded text-center leading-7 ${getHeatColor(value)}`}>
                        {value}
                      </span>
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Rider Performance */}
        <Card>
          <CardTitle>Rider Performance</CardTitle>
          <table className="mt-3 w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th className="pb-2">Rider</th>
                <th className="pb-2">Deliveries</th>
                <th className="pb-2">Avg Time</th>
                <th className="pb-2">Rating</th>
                <th className="pb-2">Completion</th>
              </tr>
            </thead>
            <tbody>
              {riderPerformance.map((r) => (
                <tr key={r.name} className="border-t">
                  <td className="py-2 font-medium">{r.name}</td>
                  <td className="py-2">{r.deliveries}</td>
                  <td className="py-2">{r.avgTime} min</td>
                  <td className="py-2">{r.rating}</td>
                  <td className="py-2">{r.completionRate}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>

        {/* Customer Area Distribution */}
        <Card>
          <CardTitle>Customer Area Distribution</CardTitle>
          <table className="mt-3 w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th className="pb-2">Area</th>
                <th className="pb-2">Orders</th>
                <th className="pb-2">Revenue</th>
                <th className="pb-2">%</th>
              </tr>
            </thead>
            <tbody>
              {customerAreas.map((a) => (
                <tr key={a.area} className="border-t">
                  <td className="py-2 font-medium">{a.area}</td>
                  <td className="py-2">{a.orders}</td>
                  <td className="py-2">{formatCurrency(a.revenue)}</td>
                  <td className="py-2">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-16 rounded-full bg-slate-200">
                        <div className="h-2 rounded-full bg-slate-900" style={{ width: `${a.percentage}%` }}></div>
                      </div>
                      <span className="text-xs">{a.percentage}%</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>
    </div>
  );
}
