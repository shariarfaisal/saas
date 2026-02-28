"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Plus, Image as ImageIcon, Trash2 } from "lucide-react";
import { formatDate } from "@/lib/utils";

type Story = {
  id: string;
  title: string;
  mediaUrl: string;
  mediaType: "image" | "video";
  restaurantName: string;
  restaurantId: string;
  expiresAt: string;
  isActive: boolean;
  views: number;
};

const mockStories: Story[] = [
  { id: "s-1", title: "Behind the Kitchen", mediaUrl: "/stories/kitchen.jpg", mediaType: "image", restaurantName: "Kacchi Bhai - Main", restaurantId: "rest-1", expiresAt: "2025-01-05", isActive: true, views: 1240 },
  { id: "s-2", title: "Chef Special Today", mediaUrl: "/stories/chef.mp4", mediaType: "video", restaurantName: "Kacchi Bhai - Main", restaurantId: "rest-1", expiresAt: "2024-12-31", isActive: true, views: 890 },
  { id: "s-3", title: "New Year Menu Preview", mediaUrl: "/stories/newyear.jpg", mediaType: "image", restaurantName: "Kacchi Bhai - Downtown", restaurantId: "rest-2", expiresAt: "2025-01-01", isActive: false, views: 456 },
];

export default function StoriesPage() {
  const [stories, setStories] = useState(mockStories);
  const [showForm, setShowForm] = useState(false);

  const toggleActive = (id: string) => {
    setStories((prev) => prev.map((s) => (s.id === id ? { ...s, isActive: !s.isActive } : s)));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Stories</h1>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-1 h-4 w-4" />
          Add Story
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {stories.map((story) => (
          <div key={story.id} className="rounded-md border bg-white overflow-hidden">
            <div className="flex h-40 items-center justify-center bg-slate-100">
              <ImageIcon className="h-10 w-10 text-slate-300" />
            </div>
            <div className="p-3">
              <div className="flex items-center justify-between">
                <p className="font-medium">{story.title}</p>
                <Badge variant={story.isActive ? "success" : "danger"}>
                  {story.isActive ? "Active" : "Inactive"}
                </Badge>
              </div>
              <p className="mt-1 text-xs text-slate-500">{story.restaurantName}</p>
              <div className="mt-2 flex items-center justify-between text-xs text-slate-500">
                <span>Expires: {formatDate(story.expiresAt)}</span>
                <span>{story.views.toLocaleString()} views</span>
              </div>
              <div className="mt-2 flex gap-1">
                <Badge>{story.mediaType}</Badge>
              </div>
              <div className="mt-3 flex gap-2">
                <Button className="flex-1 text-xs bg-slate-600 hover:bg-slate-500" onClick={() => toggleActive(story.id)}>
                  {story.isActive ? "Deactivate" : "Activate"}
                </Button>
                <button className="rounded p-2 hover:bg-slate-100">
                  <Trash2 className="h-4 w-4 text-rose-500" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Story Form Modal */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={() => setShowForm(false)}>
          <div className="w-full max-w-lg rounded-md bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">Add Story</h2>
            <form className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium">Title</label>
                <Input placeholder="Story title" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Media</label>
                <Input type="file" accept="image/*,video/*" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Restaurant</label>
                <select className="w-full rounded-md border px-3 py-2 text-sm">
                  <option value="rest-1">Kacchi Bhai - Main Branch</option>
                  <option value="rest-2">Kacchi Bhai - Downtown</option>
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Expiry Date</label>
                <Input type="date" />
              </div>
              <div className="flex gap-2">
                <Button type="button">Save Story</Button>
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
