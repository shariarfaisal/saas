"use client";

import React, { useState } from "react";
import { Area } from "@/lib/api";

type AreaSelectorProps = {
  areas: Area[];
  currentAreaSlug?: string;
};

export function AreaSelector({ areas, currentAreaSlug }: AreaSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [loadingGps, setLoadingGps] = useState(false);

  const currentArea = areas.find((a) => a.slug === currentAreaSlug) || areas[0];

  const handleGpsDetect = () => {
    setLoadingGps(true);
    // Fake GPS detect
    setTimeout(() => {
      setLoadingGps(false);
      setIsOpen(false);
      // In real life, redirect or set state
      alert("Detected area: " + areas[0]?.name);
    }, 1500);
  };

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-2 text-sm font-medium bg-white/10 hover:bg-black/5 px-3 py-1.5 rounded-full transition-colors"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-orange-500"
        >
          <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z" />
          <circle cx="12" cy="10" r="3" />
        </svg>
        <span className="text-neutral-800 line-clamp-1 max-w-[120px]">
          {currentArea?.name || "Select delivery area"}
        </span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-neutral-400"
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm px-4">
          <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl animate-in fade-in zoom-in duration-200">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-neutral-900">Delivery Address</h2>
              <button
                onClick={() => setIsOpen(false)}
                className="text-neutral-400 hover:text-neutral-600 border border-neutral-100 p-1.5 rounded-full"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M18 6 6 18" />
                  <path d="m6 6 12 12" />
                </svg>
              </button>
            </div>

            <button
              onClick={handleGpsDetect}
              disabled={loadingGps}
              className="w-full flex items-center gap-3 p-3 bg-orange-50 text-orange-600 rounded-xl hover:bg-orange-100 transition-colors mb-4 border border-orange-100"
            >
              {loadingGps ? (
                <div className="w-5 h-5 rounded-full border-2 border-orange-200 border-t-orange-600 animate-spin" />
              ) : (
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 2v2" />
                  <path d="M12 20v2" />
                  <path d="M22 12h-2" />
                  <path d="M4 12H2" />
                  <circle cx="12" cy="12" r="4" />
                </svg>
              )}
              <div className="text-left flex-1 font-semibold text-sm">Use current location<span className="block text-xs font-normal opacity-80 mt-0.5">Using GPS to find you</span></div>
            </button>

            <div className="relative mb-5 flex items-center py-2">
              <div className="flex-grow border-t border-neutral-200"></div>
              <span className="flex-shrink-0 mx-4 text-xs font-medium text-neutral-400 uppercase tracking-widest">
                Or select area
              </span>
              <div className="flex-grow border-t border-neutral-200"></div>
            </div>

            <div className="max-h-60 overflow-y-auto space-y-1">
              {areas.length === 0 && (
                <p className="text-sm text-neutral-500 text-center py-4">No areas currently available</p>
              )}
              {areas.map((area) => (
                <button
                  key={area.id}
                  onClick={() => setIsOpen(false)}
                  className={`w-full flex items-center justify-between p-3 rounded-xl transition-colors hover:bg-neutral-50 border border-transparent hover:border-neutral-100 ${
                    currentArea?.id === area.id ? "bg-orange-50 border-orange-100 text-orange-900" : "text-neutral-700"
                  }`}
                >
                  <span className="font-medium">{area.name}</span>
                  {currentArea?.id === area.id && (
                    <svg className="w-5 h-5 text-orange-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
