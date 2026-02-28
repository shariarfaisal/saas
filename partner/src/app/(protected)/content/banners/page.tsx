"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Plus, GripVertical, Edit, Trash2, Image as ImageIcon } from "lucide-react";

type Banner = {
  id: string;
  title: string;
  imageUrl: string;
  linkType: "restaurant" | "product" | "category" | "url";
  linkValue: string;
  targetArea: string;
  startsAt: string;
  endsAt: string;
  sortOrder: number;
  isActive: boolean;
};

const mockBanners: Banner[] = [
  { id: "b-1", title: "Eid Special Offer", imageUrl: "/banners/eid.jpg", linkType: "restaurant", linkValue: "rest-1", targetArea: "Gulshan", startsAt: "2024-12-20", endsAt: "2025-01-10", sortOrder: 0, isActive: true },
  { id: "b-2", title: "Free Delivery Weekend", imageUrl: "/banners/delivery.jpg", linkType: "url", linkValue: "/promotions", targetArea: "All", startsAt: "2024-12-28", endsAt: "2024-12-30", sortOrder: 1, isActive: true },
  { id: "b-3", title: "New Menu Items", imageUrl: "/banners/menu.jpg", linkType: "category", linkValue: "cat-1", targetArea: "All", startsAt: "2024-12-15", endsAt: "2025-01-15", sortOrder: 2, isActive: false },
];

export default function BannersPage() {
  const [banners, setBanners] = useState(mockBanners);
  const [showForm, setShowForm] = useState(false);
  const [draggedId, setDraggedId] = useState<string | null>(null);

  const handleDragStart = (id: string) => setDraggedId(id);
  const handleDragOver = (e: React.DragEvent, targetId: string) => {
    e.preventDefault();
    if (!draggedId || draggedId === targetId) return;
    setBanners((prev) => {
      const items = [...prev];
      const dragIdx = items.findIndex((b) => b.id === draggedId);
      const targetIdx = items.findIndex((b) => b.id === targetId);
      const [removed] = items.splice(dragIdx, 1);
      items.splice(targetIdx, 0, removed);
      return items.map((b, i) => ({ ...b, sortOrder: i }));
    });
  };
  const handleDragEnd = () => setDraggedId(null);

  const toggleActive = (id: string) => {
    setBanners((prev) => prev.map((b) => (b.id === id ? { ...b, isActive: !b.isActive } : b)));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Banners</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-1 h-4 w-4" />
          Add Banner
        </Button>
      </div>

      <div className="space-y-2">
        {banners.map((banner) => (
          <div
            key={banner.id}
            draggable
            onDragStart={() => handleDragStart(banner.id)}
            onDragOver={(e) => handleDragOver(e, banner.id)}
            onDragEnd={handleDragEnd}
            className={`flex items-center gap-4 rounded-md border bg-white p-3 ${draggedId === banner.id ? "opacity-50" : ""}`}
          >
            <GripVertical className="h-4 w-4 cursor-grab text-slate-400" />
            <div className="flex h-16 w-24 items-center justify-center rounded bg-slate-100">
              <ImageIcon className="h-6 w-6 text-slate-400" />
            </div>
            <div className="flex-1">
              <p className="font-medium">{banner.title}</p>
              <div className="mt-1 flex gap-2 text-xs text-slate-500">
                <Badge>{banner.linkType}</Badge>
                <span>Area: {banner.targetArea}</span>
                <span>{banner.startsAt} → {banner.endsAt}</span>
              </div>
            </div>
            <Badge variant={banner.isActive ? "success" : "danger"}>
              {banner.isActive ? "Active" : "Inactive"}
            </Badge>
            <div className="flex gap-1">
              <button className="rounded p-1 hover:bg-slate-100" onClick={() => toggleActive(banner.id)}>
                <Edit className="h-4 w-4 text-slate-500" />
              </button>
              <button className="rounded p-1 hover:bg-slate-100">
                <Trash2 className="h-4 w-4 text-rose-500" />
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Banner Form Modal */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={() => setShowForm(false)}>
          <div className="w-full max-w-lg rounded-md bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">Add Banner</h2>
            <form className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium">Title</label>
                <Input placeholder="Banner title" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Image</label>
                <Input type="file" accept="image/*" />
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="mb-1 block text-sm font-medium">Link Type</label>
                  <Select>
                    <option value="restaurant">Restaurant</option>
                    <option value="product">Product</option>
                    <option value="category">Category</option>
                    <option value="url">URL</option>
                  </Select>
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium">Link Value</label>
                  <Input placeholder="ID or URL" />
                </div>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Target Area</label>
                <Input placeholder="All, Gulshan, Banani..." />
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="mb-1 block text-sm font-medium">Start Date</label>
                  <Input type="date" />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium">End Date</label>
                  <Input type="date" />
                </div>
              </div>
              <div className="flex gap-2">
                <Button type="button">Save Banner</Button>
                <Button type="button" className="bg-slate-200 text-slate-700 hover:bg-slate-300" onClick={() => setShowForm(false)}>
                  Cancel
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
