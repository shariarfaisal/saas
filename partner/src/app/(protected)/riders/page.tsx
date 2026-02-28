"use client";

import Link from "next/link";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, MapPin } from "lucide-react";

type Rider = {
  id: string;
  name: string;
  phone: string;
  hub: string;
  status: "active" | "inactive" | "on_delivery";
  location: string;
  todayOrders: number;
  isAvailable: boolean;
};

const mockRiders: Rider[] = [
  { id: "r-1", name: "Karim Ahmed", phone: "+8801712345678", hub: "Gulshan Hub", status: "active", location: "Gulshan 2", todayOrders: 12, isAvailable: true },
  { id: "r-2", name: "Sohel Rana", phone: "+8801898765432", hub: "Banani Hub", status: "on_delivery", location: "Banani 11", todayOrders: 8, isAvailable: true },
  { id: "r-3", name: "Rafiq Islam", phone: "+8801555555555", hub: "Dhanmondi Hub", status: "active", location: "Dhanmondi 8", todayOrders: 15, isAvailable: true },
  { id: "r-4", name: "Jahangir Alam", phone: "+8801666666666", hub: "Uttara Hub", status: "inactive", location: "Uttara Sec 10", todayOrders: 0, isAvailable: false },
  { id: "r-5", name: "Masud Parvez", phone: "+8801777777777", hub: "Mirpur Hub", status: "active", location: "Mirpur 12", todayOrders: 10, isAvailable: true },
];

const statusMap = {
  active: { label: "Active", variant: "success" as const },
  inactive: { label: "Inactive", variant: "danger" as const },
  on_delivery: { label: "On Delivery", variant: "info" as const },
};

export default function RidersPage() {
  const [riders, setRiders] = useState(mockRiders);
  const [search, setSearch] = useState("");

  const filtered = riders.filter(
    (r) => r.name.toLowerCase().includes(search.toLowerCase()) || r.hub.toLowerCase().includes(search.toLowerCase()),
  );

  const toggleAvailability = (id: string) => {
    setRiders((prev) => prev.map((r) => (r.id === id ? { ...r, isAvailable: !r.isAvailable } : r)));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Riders</h1>
        <Link href="/riders/new">
          <Button>
            <Plus className="mr-1 h-4 w-4" />
            Add Rider
          </Button>
        </Link>
      </div>

      <div className="rounded-md border bg-white p-4">
        <div className="mb-4">
          <Input
            placeholder="Search riders..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
        </div>

        <table className="w-full text-sm">
          <thead className="text-left text-slate-500">
            <tr>
              <th className="pb-2">Name</th>
              <th className="pb-2">Hub</th>
              <th className="pb-2">Status</th>
              <th className="pb-2">Location</th>
              <th className="pb-2">Today&apos;s Orders</th>
              <th className="pb-2">Available</th>
              <th className="pb-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((rider) => (
              <tr key={rider.id} className="border-t">
                <td className="py-2">
                  <Link href={`/riders/${rider.id}`} className="font-medium hover:underline">
                    {rider.name}
                  </Link>
                  <p className="text-xs text-slate-500">{rider.phone}</p>
                </td>
                <td className="py-2">{rider.hub}</td>
                <td className="py-2">
                  <Badge variant={statusMap[rider.status].variant}>{statusMap[rider.status].label}</Badge>
                </td>
                <td className="py-2">
                  <span className="flex items-center gap-1 text-xs">
                    <MapPin className="h-3 w-3" />
                    {rider.location}
                  </span>
                </td>
                <td className="py-2 text-center font-semibold">{rider.todayOrders}</td>
                <td className="py-2">
                  <label className="relative inline-flex cursor-pointer items-center">
                    <input
                      type="checkbox"
                      className="peer sr-only"
                      checked={rider.isAvailable}
                      onChange={() => toggleAvailability(rider.id)}
                    />
                    <div className="h-5 w-9 rounded-full bg-slate-200 after:absolute after:left-[2px] after:top-[2px] after:h-4 after:w-4 after:rounded-full after:bg-white after:transition-all peer-checked:bg-emerald-500 peer-checked:after:translate-x-full"></div>
                  </label>
                </td>
                <td className="py-2">
                  <Link href={`/riders/${rider.id}`}>
                    <Button className="text-xs bg-slate-600 hover:bg-slate-500">View</Button>
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
