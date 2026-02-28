"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Plus, GripVertical, Edit, Trash2 } from "lucide-react";

type HomepageSection = {
  id: string;
  title: string;
  type: "featured_restaurants" | "popular_products" | "cuisine_collection" | "promotional" | "custom";
  displayOrder: number;
  isActive: boolean;
  itemCount: number;
};

const mockSections: HomepageSection[] = [
  { id: "sec-1", title: "Featured Restaurants", type: "featured_restaurants", displayOrder: 0, isActive: true, itemCount: 6 },
  { id: "sec-2", title: "Popular This Week", type: "popular_products", displayOrder: 1, isActive: true, itemCount: 12 },
  { id: "sec-3", title: "Biryani Collection", type: "cuisine_collection", displayOrder: 2, isActive: true, itemCount: 8 },
  { id: "sec-4", title: "Special Offers", type: "promotional", displayOrder: 3, isActive: false, itemCount: 4 },
];

const typeLabels: Record<string, string> = {
  featured_restaurants: "Featured Restaurants",
  popular_products: "Popular Products",
  cuisine_collection: "Cuisine Collection",
  promotional: "Promotional",
  custom: "Custom",
};

export default function SectionsPage() {
  const [sections, setSections] = useState(mockSections);
  const [showForm, setShowForm] = useState(false);
  const [draggedId, setDraggedId] = useState<string | null>(null);

  const handleDragStart = (id: string) => setDraggedId(id);
  const handleDragOver = (e: React.DragEvent, targetId: string) => {
    e.preventDefault();
    if (!draggedId || draggedId === targetId) return;
    setSections((prev) => {
      const items = [...prev];
      const dragIdx = items.findIndex((s) => s.id === draggedId);
      const targetIdx = items.findIndex((s) => s.id === targetId);
      const [removed] = items.splice(dragIdx, 1);
      items.splice(targetIdx, 0, removed);
      return items.map((s, i) => ({ ...s, displayOrder: i }));
    });
  };
  const handleDragEnd = () => setDraggedId(null);

  const toggleActive = (id: string) => {
    setSections((prev) => prev.map((s) => (s.id === id ? { ...s, isActive: !s.isActive } : s)));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Homepage Sections</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-1 h-4 w-4" />
          Add Section
        </Button>
      </div>

      <div className="space-y-2">
        {sections.map((section) => (
          <div
            key={section.id}
            draggable
            onDragStart={() => handleDragStart(section.id)}
            onDragOver={(e) => handleDragOver(e, section.id)}
            onDragEnd={handleDragEnd}
            className={`flex items-center gap-4 rounded-md border bg-white p-4 ${draggedId === section.id ? "opacity-50" : ""}`}
          >
            <GripVertical className="h-4 w-4 cursor-grab text-slate-400" />
            <div className="flex-1">
              <p className="font-medium">{section.title}</p>
              <div className="mt-1 flex gap-2 text-xs">
                <Badge>{typeLabels[section.type]}</Badge>
                <span className="text-slate-500">{section.itemCount} items</span>
              </div>
            </div>
            <Badge variant={section.isActive ? "success" : "danger"}>
              {section.isActive ? "Active" : "Inactive"}
            </Badge>
            <div className="flex gap-1">
              <button className="rounded p-1 hover:bg-slate-100" onClick={() => toggleActive(section.id)}>
                <Edit className="h-4 w-4 text-slate-500" />
              </button>
              <button className="rounded p-1 hover:bg-slate-100">
                <Trash2 className="h-4 w-4 text-rose-500" />
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Section Form Modal */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={() => setShowForm(false)}>
          <div className="w-full max-w-lg rounded-md bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">Add Section</h2>
            <form className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium">Section Title</label>
                <Input placeholder="e.g. Featured Restaurants" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Type</label>
                <Select>
                  <option value="featured_restaurants">Featured Restaurants</option>
                  <option value="popular_products">Popular Products</option>
                  <option value="cuisine_collection">Cuisine Collection</option>
                  <option value="promotional">Promotional</option>
                  <option value="custom">Custom</option>
                </Select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Items</label>
                <p className="text-xs text-slate-500">Select restaurants or products to include in this section</p>
                <Input placeholder="Search and add items..." className="mt-1" />
              </div>
              <div className="flex gap-2">
                <Button type="button">Save Section</Button>
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
