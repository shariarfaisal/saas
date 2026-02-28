import { Story } from "@/lib/api";

export function StoryStrip({ stories }: { stories: Story[] }) {
  if (!stories || stories.length === 0) return null;

  return (
    <div className="w-full">
      <h3 className="text-lg font-bold mb-4 px-1">Highlights</h3>
      <div className="flex gap-4 overflow-x-auto hide-scrollbar pb-4 overflow-y-visible px-1">
        {stories.map((story) => (
          <div key={story.id} className="flex-none flex flex-col items-center gap-2 group cursor-pointer w-[80px]">
            <div className="relative w-16 h-16 rounded-full p-[3px] bg-gradient-to-tr from-orange-400 to-orange-600 group-hover:scale-105 transition-transform">
              <div className="absolute inset-0 rounded-full border-2 border-white m-[2px] z-10 block pointer-events-none" />
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={story.thumbnail_url || story.media_url}
                alt={story.title || "Story"}
                className="w-full h-full object-cover rounded-full bg-neutral-100 relative z-0"
                loading="lazy"
              />
            </div>
            {story.title && (
              <span className="text-xs font-medium text-neutral-600 line-clamp-1 text-center w-full">
                {story.title}
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
