"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Clock, MapPin, User, X } from "lucide-react";
import { formatCurrency, timeAgo } from "@/lib/utils";

type OrderStatus = "new" | "confirmed" | "preparing" | "ready" | "picked";

type Order = {
  id: string;
  orderNumber: string;
  items: { name: string; qty: number; addons?: string[] }[];
  total: number;
  status: OrderStatus;
  customerName: string;
  customerArea: string;
  customerAddress: string;
  customerPhone: string;
  riderName?: string;
  riderPhone?: string;
  paymentMethod: string;
  paymentStatus: string;
  createdAt: string;
  timeline: { status: string; at: string }[];
};

const statusConfig: Record<OrderStatus, { label: string; variant: "warning" | "info" | "default" | "success" | "danger"; next?: { label: string; status: OrderStatus } }> = {
  new: { label: "New", variant: "warning", next: { label: "Confirm", status: "confirmed" } },
  confirmed: { label: "Confirmed", variant: "info", next: { label: "Start Preparing", status: "preparing" } },
  preparing: { label: "Preparing", variant: "default", next: { label: "Mark Ready", status: "ready" } },
  ready: { label: "Ready", variant: "success", next: { label: "Mark Picked", status: "picked" } },
  picked: { label: "Picked Up", variant: "success" },
};

const columns: OrderStatus[] = ["new", "confirmed", "preparing", "ready", "picked"];

const mockOrders: Order[] = [
  {
    id: "o-1", orderNumber: "KBC-001234", items: [{ name: "Kacchi Biryani (Full)", qty: 2, addons: ["Extra Raita"] }, { name: "Borhani", qty: 2 }],
    total: 1520, status: "new", customerName: "Arif Rahman", customerArea: "Gulshan", customerAddress: "House 12, Road 5, Gulshan 2",
    customerPhone: "+8801712345678", paymentMethod: "bKash", paymentStatus: "paid", createdAt: new Date(Date.now() - 120000).toISOString(),
    timeline: [{ status: "created", at: new Date(Date.now() - 120000).toISOString() }],
  },
  {
    id: "o-2", orderNumber: "KBC-001235", items: [{ name: "Tandoori Chicken", qty: 1 }, { name: "Naan", qty: 3 }],
    total: 640, status: "new", customerName: "Fatima Khan", customerArea: "Banani", customerAddress: "Apt 4B, Navana Tower",
    customerPhone: "+8801898765432", paymentMethod: "COD", paymentStatus: "pending", createdAt: new Date(Date.now() - 60000).toISOString(),
    timeline: [{ status: "created", at: new Date(Date.now() - 60000).toISOString() }],
  },
  {
    id: "o-3", orderNumber: "KBC-001232", items: [{ name: "Tehari", qty: 3 }],
    total: 660, status: "confirmed", customerName: "Hasan Ali", customerArea: "Dhanmondi", customerAddress: "Road 8, House 22",
    customerPhone: "+8801555555555", paymentMethod: "bKash", paymentStatus: "paid", createdAt: new Date(Date.now() - 300000).toISOString(),
    timeline: [{ status: "created", at: new Date(Date.now() - 300000).toISOString() }, { status: "confirmed", at: new Date(Date.now() - 240000).toISOString() }],
  },
  {
    id: "o-4", orderNumber: "KBC-001230", items: [{ name: "Shami Kebab", qty: 4 }, { name: "Paratha", qty: 4 }],
    total: 880, status: "preparing", customerName: "Nusrat Jahan", customerArea: "Uttara", customerAddress: "Sector 10, Road 5",
    customerPhone: "+8801666666666", riderName: "Karim", riderPhone: "+8801777777777", paymentMethod: "AamarPay", paymentStatus: "paid",
    createdAt: new Date(Date.now() - 600000).toISOString(),
    timeline: [
      { status: "created", at: new Date(Date.now() - 600000).toISOString() },
      { status: "confirmed", at: new Date(Date.now() - 540000).toISOString() },
      { status: "preparing", at: new Date(Date.now() - 420000).toISOString() },
    ],
  },
  {
    id: "o-5", orderNumber: "KBC-001228", items: [{ name: "Kacchi Biryani (Half)", qty: 1 }],
    total: 350, status: "ready", customerName: "Rahim Uddin", customerArea: "Mirpur", customerAddress: "Section 12, Block C",
    customerPhone: "+8801888888888", riderName: "Sohel", riderPhone: "+8801999999999", paymentMethod: "COD", paymentStatus: "pending",
    createdAt: new Date(Date.now() - 900000).toISOString(),
    timeline: [
      { status: "created", at: new Date(Date.now() - 900000).toISOString() },
      { status: "confirmed", at: new Date(Date.now() - 840000).toISOString() },
      { status: "preparing", at: new Date(Date.now() - 720000).toISOString() },
      { status: "ready", at: new Date(Date.now() - 300000).toISOString() },
    ],
  },
];

export default function OrdersPage() {
  const [orders, setOrders] = useState(mockOrders);
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);
  const [view, setView] = useState<"board" | "history">("board");
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const moveOrder = (orderId: string, newStatus: OrderStatus) => {
    setOrders((prev) =>
      prev.map((o) =>
        o.id === orderId
          ? {
              ...o,
              status: newStatus,
              timeline: [...o.timeline, { status: newStatus, at: new Date().toISOString() }],
            }
          : o,
      ),
    );
  };

  const filteredOrders = orders.filter((o) => {
    if (searchQuery && !o.orderNumber.toLowerCase().includes(searchQuery.toLowerCase()) && !o.customerName.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    if (statusFilter !== "all" && o.status !== statusFilter) return false;
    return true;
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Orders</h1>
        <div className="flex gap-2">
          <Button
            className={view === "board" ? "" : "bg-slate-200 text-slate-700 hover:bg-slate-300"}
            onClick={() => setView("board")}
          >
            Board
          </Button>
          <Button
            className={view === "history" ? "" : "bg-slate-200 text-slate-700 hover:bg-slate-300"}
            onClick={() => setView("history")}
          >
            History
          </Button>
        </div>
      </div>

      {view === "board" ? (
        <div className="grid grid-cols-5 gap-3">
          {columns.map((status) => {
            const colOrders = orders.filter((o) => o.status === status);
            const cfg = statusConfig[status];
            return (
              <div key={status} className="rounded-md border bg-white p-3">
                <div className="mb-3 flex items-center justify-between">
                  <Badge variant={cfg.variant}>{cfg.label}</Badge>
                  <span className="text-xs text-slate-500">{colOrders.length}</span>
                </div>
                <div className="space-y-2">
                  {colOrders.map((order) => (
                    <div
                      key={order.id}
                      className="cursor-pointer rounded border p-2 hover:bg-slate-50"
                      onClick={() => setSelectedOrder(order)}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-semibold">{order.orderNumber}</span>
                        <span className="text-xs text-slate-400">{timeAgo(order.createdAt)}</span>
                      </div>
                      <p className="mt-1 text-xs text-slate-600">
                        {order.items.map((i) => `${i.qty}x ${i.name}`).join(", ")}
                      </p>
                      <div className="mt-1 flex items-center justify-between">
                        <span className="flex items-center gap-1 text-xs text-slate-500">
                          <MapPin className="h-3 w-3" />
                          {order.customerArea}
                        </span>
                        <span className="text-xs font-semibold">{formatCurrency(order.total)}</span>
                      </div>
                      {cfg.next && (
                        <Button
                          className="mt-2 w-full text-xs"
                          onClick={(e) => {
                            e.stopPropagation();
                            moveOrder(order.id, cfg.next!.status);
                          }}
                        >
                          {cfg.next.label}
                        </Button>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        /* History Table */
        <div className="rounded-md border bg-white p-4">
          <div className="mb-4 flex gap-3">
            <Input
              placeholder="Search orders..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="max-w-xs"
            />
            <Select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="max-w-[160px]">
              <option value="all">All Status</option>
              {columns.map((s) => (
                <option key={s} value={s}>{statusConfig[s].label}</option>
              ))}
            </Select>
          </div>
          <table className="w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th className="pb-2">Order #</th>
                <th className="pb-2">Customer</th>
                <th className="pb-2">Items</th>
                <th className="pb-2">Total</th>
                <th className="pb-2">Status</th>
                <th className="pb-2">Time</th>
              </tr>
            </thead>
            <tbody>
              {filteredOrders.map((order) => (
                <tr key={order.id} className="cursor-pointer border-t hover:bg-slate-50" onClick={() => setSelectedOrder(order)}>
                  <td className="py-2 font-medium">{order.orderNumber}</td>
                  <td className="py-2">{order.customerName}</td>
                  <td className="py-2 text-xs">{order.items.map((i) => `${i.qty}x ${i.name}`).join(", ")}</td>
                  <td className="py-2 font-semibold">{formatCurrency(order.total)}</td>
                  <td className="py-2"><Badge variant={statusConfig[order.status].variant}>{statusConfig[order.status].label}</Badge></td>
                  <td className="py-2 text-xs text-slate-500">{timeAgo(order.createdAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Order Detail Drawer */}
      {selectedOrder && (
        <div className="fixed inset-0 z-50 flex justify-end bg-black/30" onClick={() => setSelectedOrder(null)}>
          <div className="w-full max-w-lg overflow-y-auto bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold">Order {selectedOrder.orderNumber}</h2>
              <button onClick={() => setSelectedOrder(null)}>
                <X className="h-5 w-5 text-slate-500" />
              </button>
            </div>

            <Badge variant={statusConfig[selectedOrder.status].variant} className="mb-4">
              {statusConfig[selectedOrder.status].label}
            </Badge>

            {/* Items */}
            <section className="mb-4">
              <h3 className="mb-2 text-sm font-semibold">Items</h3>
              <div className="space-y-1">
                {selectedOrder.items.map((item, i) => (
                  <div key={i} className="flex justify-between text-sm">
                    <span>
                      {item.qty}x {item.name}
                      {item.addons && <span className="text-xs text-slate-500"> + {item.addons.join(", ")}</span>}
                    </span>
                  </div>
                ))}
              </div>
              <div className="mt-2 flex justify-between border-t pt-2 font-semibold">
                <span>Total</span>
                <span>{formatCurrency(selectedOrder.total)}</span>
              </div>
            </section>

            {/* Customer */}
            <section className="mb-4">
              <h3 className="mb-2 text-sm font-semibold">Customer</h3>
              <div className="space-y-1 text-sm">
                <p className="flex items-center gap-1"><User className="h-3 w-3" /> {selectedOrder.customerName}</p>
                <p className="flex items-center gap-1"><MapPin className="h-3 w-3" /> {selectedOrder.customerAddress}</p>
                <p>📞 {selectedOrder.customerPhone}</p>
              </div>
            </section>

            {/* Rider */}
            {selectedOrder.riderName && (
              <section className="mb-4">
                <h3 className="mb-2 text-sm font-semibold">Rider</h3>
                <div className="space-y-1 text-sm">
                  <p>{selectedOrder.riderName}</p>
                  <p>📞 {selectedOrder.riderPhone}</p>
                </div>
              </section>
            )}

            {/* Payment */}
            <section className="mb-4">
              <h3 className="mb-2 text-sm font-semibold">Payment</h3>
              <div className="flex gap-2 text-sm">
                <Badge>{selectedOrder.paymentMethod}</Badge>
                <Badge variant={selectedOrder.paymentStatus === "paid" ? "success" : "warning"}>
                  {selectedOrder.paymentStatus}
                </Badge>
              </div>
            </section>

            {/* Timeline */}
            <section className="mb-4">
              <h3 className="mb-2 text-sm font-semibold">Timeline</h3>
              <div className="space-y-2">
                {selectedOrder.timeline.map((t, i) => (
                  <div key={i} className="flex items-center gap-2 text-sm">
                    <Clock className="h-3 w-3 text-slate-400" />
                    <span className="font-medium capitalize">{t.status}</span>
                    <span className="text-xs text-slate-500">{new Date(t.at).toLocaleTimeString()}</span>
                  </div>
                ))}
              </div>
            </section>

            {/* Actions */}
            {statusConfig[selectedOrder.status].next && (
              <Button
                className="w-full"
                onClick={() => {
                  moveOrder(selectedOrder.id, statusConfig[selectedOrder.status].next!.status);
                  setSelectedOrder(null);
                }}
              >
                {statusConfig[selectedOrder.status].next!.label}
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
