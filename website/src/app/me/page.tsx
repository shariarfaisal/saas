"use client";

import React, { useState } from "react";
import { useAuthStore } from "@/stores/auth-store";

export default function ProfilePage() {
  const { phone } = useAuthStore();
  
  const [name, setName] = useState("John Doe"); 
  const [email, setEmail] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    try {
      // Mock API call
      await new Promise(resolve => setTimeout(resolve, 800));
      alert("Profile updated successfully");
    } catch (err: Error | unknown) {
      alert((err as Error).message || "Failed to update profile");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="bg-white rounded-3xl p-6 sm:p-8 shadow-sm border border-neutral-100">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-bold text-neutral-900">My Profile</h2>
      </div>

      <form onSubmit={handleSave} className="space-y-6 max-w-lg">
        <div className="flex flex-col items-center sm:items-start sm:flex-row gap-6 mb-8">
          <div className="w-24 h-24 rounded-full bg-orange-100 border-4 border-white shadow flex flex-col items-center justify-center shrink-0">
             <span className="text-3xl text-orange-500 font-bold uppercase">{name?.[0] || 'U'}</span>
          </div>
          <div className="flex-1 space-y-2 mt-2">
             <button type="button" className="text-sm font-bold text-orange-600 bg-orange-50 px-4 py-2 rounded-xl transition-all hover:bg-orange-100">
               Change Photo
             </button>
             <p className="text-xs text-neutral-500 font-medium">JPEG or PNG, max 2MB.</p>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-bold text-neutral-700 uppercase tracking-wider ml-1">Full Name</label>
            <input 
              type="text" 
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-neutral-50 border border-neutral-100 rounded-2xl px-5 py-4 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:bg-white transition-all font-medium" 
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-bold text-neutral-700 uppercase tracking-wider ml-1">Phone Number</label>
            <input 
              type="tel" 
              value={phone || ""}
              disabled
              className="w-full bg-neutral-100/70 text-neutral-500 border border-neutral-100 rounded-2xl px-5 py-4 transition-all font-medium cursor-not-allowed" 
            />
            <p className="text-[10px] text-neutral-400 font-bold ml-1 uppercase">Phone cannot be changed</p>
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-bold text-neutral-700 uppercase tracking-wider ml-1">Email (Optional)</label>
          <input 
            type="email" 
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="e.g. hello@munchies.com"
            className="w-full bg-neutral-50 border border-neutral-100 rounded-2xl px-5 py-4 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:bg-white transition-all font-medium" 
          />
        </div>

        <div className="pt-4 border-t border-neutral-100 flex justify-end">
          <button 
            type="submit"
            disabled={isSaving}
            className="w-full sm:w-auto bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-4 px-8 rounded-2xl shadow-lg shadow-orange-100 transition-all active:scale-[0.98] disabled:opacity-50"
          >
            {isSaving ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </form>
    </div>
  );
}
