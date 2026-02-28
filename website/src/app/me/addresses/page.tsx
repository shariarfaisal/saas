"use client";

import React, { useState } from "react";
import { Address } from "@/lib/api";

export default function AddressesPage() {
  const [addresses, setAddresses] = useState<Address[]>([
    {
      id: "addr_1",
      name: "Home",
      address_line: "Apt 4B, 123 Main St",
      area: "gulshan",
      is_default: true,
    }
  ]);
  const [isAdding, setIsAdding] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  // Form State
  const [name, setName] = useState("");
  const [addressLine, setAddressLine] = useState("");

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !addressLine) return;

    setIsSaving(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 800));
      const newAddr: Address = {
        id: `addr_${Date.now()}`,
        name,
        address_line: addressLine,
        area: "unspecified",
        is_default: addresses.length === 0,
      };
      setAddresses([...addresses, newAddr]);
      setIsAdding(false);
      setName("");
      setAddressLine("");
    } finally {
      setIsSaving(false);
    }
  };

  const removeAddress = (id: string) => {
    if (confirm("Are you sure?")) {
      setAddresses(addresses.filter(a => a.id !== id));
    }
  };

  const setAsDefault = (id: string) => {
    setAddresses(addresses.map(a => ({
      ...a,
      is_default: a.id === id,
    })));
  };

  return (
    <div className="bg-white rounded-3xl p-6 sm:p-8 shadow-sm border border-neutral-100">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-bold text-neutral-900">Saved Addresses</h2>
        {!isAdding && (
          <button 
            onClick={() => setIsAdding(true)}
            className="bg-neutral-900 hover:bg-neutral-800 text-white font-bold py-2.5 px-5 rounded-xl text-sm transition-colors flex items-center gap-2"
          >
            <span className="text-lg">+</span>
            Add New
          </button>
        )}
      </div>

      {isAdding ? (
        <form onSubmit={handleSave} className="bg-neutral-50 rounded-2xl p-6 border border-neutral-100 mb-8 max-w-lg">
           <h3 className="font-bold text-neutral-900 mb-4 text-lg">Add New Address</h3>
           <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-xs font-bold text-neutral-700 uppercase tracking-wider ml-1">Label</label>
                <input 
                  type="text" 
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Home, Office"
                  className="w-full bg-white border border-neutral-200 rounded-xl px-4 py-3 focus:outline-none focus:ring-2 focus:ring-orange-500/20 font-medium" 
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-bold text-neutral-700 uppercase tracking-wider ml-1">Full Address</label>
                <textarea 
                  value={addressLine}
                  onChange={(e) => setAddressLine(e.target.value)}
                  placeholder="e.g. Block C, Road 2, Flat 4A"
                  className="w-full bg-white border border-neutral-200 rounded-xl px-4 py-3 focus:outline-none focus:ring-2 focus:ring-orange-500/20 font-medium resize-none" 
                  rows={3}
                />
              </div>

              <div className="pt-2 flex gap-3">
                <button 
                  type="submit"
                  disabled={isSaving || !name || !addressLine}
                  className="flex-1 bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-3 rounded-xl transition-all disabled:opacity-50"
                >
                  {isSaving ? "Saving..." : "Save Address"}
                </button>
                <button 
                  type="button"
                  onClick={() => setIsAdding(false)}
                  className="flex-1 bg-white hover:bg-neutral-50 border border-neutral-200 text-neutral-700 font-extrabold py-3 rounded-xl transition-all"
                >
                  Cancel
                </button>
              </div>
           </div>
        </form>
      ) : null}

      {addresses.length === 0 && !isAdding ? (
         <div className="text-center py-12 border-2 border-dashed border-neutral-100 rounded-2xl">
            <div className="w-16 h-16 bg-neutral-100 rounded-full flex mx-auto items-center justify-center text-3xl mb-4">📍</div>
            <p className="text-neutral-500 font-medium">No saved addresses yet.</p>
         </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {addresses.map(addr => (
            <div key={addr.id} className={`p-5 rounded-2xl border-2 flex flex-col items-start transition-all ${addr.is_default ? 'border-orange-500 bg-orange-50/30' : 'border-neutral-100 hover:border-orange-200'}`}>
               <div className="flex w-full items-center justify-between mb-2">
                 <div className="flex items-center gap-2">
                    <span className="font-extrabold text-neutral-900">{addr.name}</span>
                    {addr.is_default && (
                      <span className="bg-orange-100 text-orange-600 text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded-full">Default</span>
                    )}
                 </div>
               </div>
               <p className="text-sm text-neutral-500 font-medium mb-6 line-clamp-2">
                  {addr.address_line}
               </p>

               <div className="mt-auto w-full flex items-center justify-between pt-4 border-t border-neutral-100/60">
                 {!addr.is_default && (
                   <button 
                     onClick={() => setAsDefault(addr.id)}
                     className="text-xs font-bold text-neutral-500 hover:text-orange-600 transition-colors"
                   >
                     Set as default
                   </button>
                 )}
                 <div className="ml-auto flex gap-3">
                   <button 
                     onClick={() => removeAddress(addr.id)}
                     className="text-xs font-bold text-red-500 hover:text-red-600 transition-colors bg-red-50 p-2 rounded-lg"
                   >
                     Remove
                   </button>
                 </div>
               </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
