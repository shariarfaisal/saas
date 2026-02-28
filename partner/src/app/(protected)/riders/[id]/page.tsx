"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardTitle } from "@/components/ui/card";
import { MapPin, Phone, Calendar } from "lucide-react";
import { formatCurrency } from "@/lib/utils";

const mockRider = {
  id: "r-1",
  name: "Karim Ahmed",
  phone: "+8801712345678",
  email: "karim@email.com",
  hub: "Gulshan Hub",
  status: "active" as const,
  vehicleType: "Motorcycle",
  licensePlate: "DHAKA-1234",
  joinedAt: "2024-06-15",
  stats: {
    totalDeliveries: 1247,
    thisMonthDeliveries: 156,
    avgDeliveryTime: 28,
    rating: 4.7,
    completionRate: 97.2,
  },
  earnings: {
    today: 1850,
    thisWeek: 12400,
    thisMonth: 48600,
  },
  penalties: [
    { id: 1, reason: "Late delivery (>15 min)", amount: 50, date: "2024-12-20" },
    { id: 2, reason: "Customer complaint", amount: 100, date: "2024-12-15" },
  ],
  attendance: {
    present: 24,
    absent: 2,
    late: 3,
    totalDays: 29,
  },
};

const attendanceCalendar = Array.from({ length: 30 }, (_, i) => ({
  day: i + 1,
  status: i === 5 || i === 12 ? "absent" : i === 7 || i === 15 || i === 22 ? "late" : "present",
}));

export default function RiderDetailPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">{mockRider.name}</h1>
          <div className="mt-1 flex items-center gap-3 text-sm text-slate-500">
            <span className="flex items-center gap-1"><Phone className="h-3 w-3" /> {mockRider.phone}</span>
            <span className="flex items-center gap-1"><MapPin className="h-3 w-3" /> {mockRider.hub}</span>
          </div>
        </div>
        <Badge variant="success">Active</Badge>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-5">
        {[
          { label: "Total Deliveries", value: mockRider.stats.totalDeliveries.toLocaleString() },
          { label: "This Month", value: mockRider.stats.thisMonthDeliveries.toString() },
          { label: "Avg Delivery Time", value: `${mockRider.stats.avgDeliveryTime} min` },
          { label: "Rating", value: mockRider.stats.rating.toString() },
          { label: "Completion Rate", value: `${mockRider.stats.completionRate}%` },
        ].map(({ label, value }) => (
          <Card key={label}>
            <p className="text-xs text-slate-500">{label}</p>
            <p className="mt-1 text-xl font-semibold">{value}</p>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Earnings */}
        <Card>
          <CardTitle>Earnings</CardTitle>
          <div className="mt-3 space-y-2">
            <div className="flex justify-between border-b pb-2">
              <span className="text-sm">Today</span>
              <span className="font-semibold">{formatCurrency(mockRider.earnings.today)}</span>
            </div>
            <div className="flex justify-between border-b pb-2">
              <span className="text-sm">This Week</span>
              <span className="font-semibold">{formatCurrency(mockRider.earnings.thisWeek)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm">This Month</span>
              <span className="font-semibold">{formatCurrency(mockRider.earnings.thisMonth)}</span>
            </div>
          </div>
        </Card>

        {/* Penalties */}
        <Card>
          <CardTitle>Penalties</CardTitle>
          {mockRider.penalties.length === 0 ? (
            <p className="mt-3 text-sm text-slate-500">No penalties.</p>
          ) : (
            <div className="mt-3 space-y-2">
              {mockRider.penalties.map((p) => (
                <div key={p.id} className="flex items-center justify-between border-b pb-2 last:border-0">
                  <div>
                    <p className="text-sm">{p.reason}</p>
                    <p className="text-xs text-slate-400">{p.date}</p>
                  </div>
                  <span className="font-semibold text-rose-600">-{formatCurrency(p.amount)}</span>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>

      {/* Attendance Calendar */}
      <Card>
        <CardTitle>
          <div className="flex items-center gap-1">
            <Calendar className="h-4 w-4" />
            Attendance Calendar
          </div>
        </CardTitle>
        <div className="mt-3 flex items-center gap-3 text-xs">
          <span className="flex items-center gap-1"><span className="h-3 w-3 rounded bg-emerald-400"></span> Present ({mockRider.attendance.present})</span>
          <span className="flex items-center gap-1"><span className="h-3 w-3 rounded bg-rose-400"></span> Absent ({mockRider.attendance.absent})</span>
          <span className="flex items-center gap-1"><span className="h-3 w-3 rounded bg-amber-400"></span> Late ({mockRider.attendance.late})</span>
        </div>
        <div className="mt-3 grid grid-cols-7 gap-1">
          {attendanceCalendar.map(({ day, status }) => (
            <div
              key={day}
              className={`flex h-8 items-center justify-center rounded text-xs font-medium ${
                status === "present" ? "bg-emerald-100 text-emerald-700" :
                status === "absent" ? "bg-rose-100 text-rose-700" :
                "bg-amber-100 text-amber-700"
              }`}
            >
              {day}
            </div>
          ))}
        </div>
      </Card>

      {/* Live Location Placeholder */}
      <Card>
        <CardTitle>Live Location</CardTitle>
        <div className="mt-3 flex h-48 items-center justify-center rounded-md border-2 border-dashed border-slate-200 bg-slate-50">
          <p className="text-sm text-slate-400">Map integration (Google Maps / Mapbox) — Phase 2</p>
        </div>
      </Card>
    </div>
  );
}
