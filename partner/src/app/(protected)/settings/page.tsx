"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardTitle } from "@/components/ui/card";

const notificationPreferences = [
  { key: "new_order", label: "New Orders", description: "Get notified when a new order comes in" },
  { key: "order_status", label: "Order Status Changes", description: "Notifications for order state transitions" },
  { key: "invoice_ready", label: "Invoice Ready", description: "When a new invoice is generated" },
  { key: "low_stock", label: "Low Stock Alerts", description: "When product stock is running low" },
  { key: "rider_update", label: "Rider Updates", description: "Rider availability and assignment updates" },
  { key: "promo_usage", label: "Promo Usage Alerts", description: "When promotions are nearing usage limits" },
  { key: "review_received", label: "New Reviews", description: "Customer review notifications" },
  { key: "payment_received", label: "Payment Received", description: "When a payout is processed" },
];

export default function SettingsPage() {
  const [preferences, setPreferences] = useState<Record<string, { push: boolean; email: boolean }>>(
    Object.fromEntries(notificationPreferences.map((p) => [p.key, { push: true, email: true }])),
  );

  const togglePref = (key: string, channel: "push" | "email") => {
    setPreferences((prev) => ({
      ...prev,
      [key]: { ...prev[key], [channel]: !prev[key][channel] },
    }));
  };

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-lg font-semibold">Settings</h1>

      {/* Vendor Profile */}
      <Card>
        <CardTitle>Vendor Profile</CardTitle>
        <form className="mt-4 space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium">Business Name</label>
              <Input defaultValue="Kacchi Bhai Ltd." />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Contact Email</label>
              <Input defaultValue="admin@kacchibhai.com" />
            </div>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium">Contact Phone</label>
              <Input defaultValue="+8801712345678" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Business Registration</label>
              <Input defaultValue="BIN-2024-12345" />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Business Address</label>
            <Textarea defaultValue="House 12, Road 5, Gulshan 2, Dhaka 1212" rows={2} />
          </div>
          <Button type="button">Save Profile</Button>
        </form>
      </Card>

      {/* Notification Preferences */}
      <Card>
        <CardTitle>Notification Preferences</CardTitle>
        <div className="mt-4">
          <div className="mb-2 grid grid-cols-[1fr_60px_60px] gap-2 text-xs font-semibold text-slate-500">
            <span>Event</span>
            <span className="text-center">Push</span>
            <span className="text-center">Email</span>
          </div>
          <div className="space-y-1">
            {notificationPreferences.map((pref) => (
              <div key={pref.key} className="grid grid-cols-[1fr_60px_60px] items-center gap-2 border-t py-2">
                <div>
                  <p className="text-sm font-medium">{pref.label}</p>
                  <p className="text-xs text-slate-500">{pref.description}</p>
                </div>
                <div className="flex justify-center">
                  <input
                    type="checkbox"
                    checked={preferences[pref.key]?.push}
                    onChange={() => togglePref(pref.key, "push")}
                    className="h-4 w-4"
                  />
                </div>
                <div className="flex justify-center">
                  <input
                    type="checkbox"
                    checked={preferences[pref.key]?.email}
                    onChange={() => togglePref(pref.key, "email")}
                    className="h-4 w-4"
                  />
                </div>
              </div>
            ))}
          </div>
          <Button type="button" className="mt-4">Save Preferences</Button>
        </div>
      </Card>

      {/* API Keys Placeholder */}
      <Card>
        <CardTitle>API Keys</CardTitle>
        <div className="mt-4 rounded-md border-2 border-dashed border-slate-200 bg-slate-50 p-6 text-center">
          <p className="text-sm text-slate-500">API keys management will be available in Phase 2.</p>
          <p className="mt-1 text-xs text-slate-400">Integrate with POS systems, delivery tracking, and more.</p>
        </div>
      </Card>
    </div>
  );
}
