"use client";

import React, { useEffect, useState } from "react";
import { Category } from "@/lib/api";

export function StickyNav({ categories }: { categories: Category[] }) {
  const [activeId, setActiveId] = useState<string>(categories[0]?.id || "");

  useEffect(() => {
    // Intersection observer logic to highlight active section
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
          }
        });
      },
      { rootMargin: "-20% 0px -80% 0px", threshold: 0 }
    );

    categories.forEach((cat) => {
      const el = document.getElementById(`category-${cat.id}`);
      if (el) observer.observe(el);
    });

    return () => observer.disconnect();
  }, [categories]);

  const scrollToCat = (id: string) => {
    setActiveId(id);
    const el = document.getElementById(`category-${id}`);
    if (el) {
      // Offset scrolling by height of sticky nav + some margin
      const y = el.getBoundingClientRect().top + window.scrollY - 140;
      window.scrollTo({ top: y, behavior: "smooth" });
    }
  };

  if (categories.length === 0) return null;

  return (
    <div className="sticky top-[73px] sm:top-[73px] z-40 bg-white/95 backdrop-blur-md border border-neutral-100 rounded-b-2xl sm:rounded-2xl px-4 py-2 mt-2 -mx-4 sm:mx-0 shadow-sm flex items-center gap-3 overflow-x-auto hide-scrollbar">
      {categories.map((cat) => (
        <button
          key={cat.id}
          onClick={() => scrollToCat(cat.id)}
          className={`px-4 py-2 rounded-xl text-sm font-semibold whitespace-nowrap transition-colors duration-200 ${
            activeId === `category-${cat.id}`
              ? "bg-orange-600 text-white shadow-md shadow-orange-500/20"
              : "bg-transparent text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900"
          }`}
        >
          {cat.name}
        </button>
      ))}
    </div>
  );
}
