import { Banner } from "@/lib/api";
import Image from "next/image";

export function HeroCarousel({ banners }: { banners: Banner[] }) {
  if (!banners || banners.length === 0) return null;

  return (
    <div className="relative w-full overflow-hidden rounded-2xl">
      <div className="flex snap-x snap-mandatory overflow-x-auto hide-scrollbar">
        {banners.map((banner) => (
          <div key={banner.id} className="min-w-full flex-none snap-center relative aspect-[21/9] md:aspect-[3/1]">
            <Image
              src={banner.image_url}
              alt={banner.title}
              fill
              className="object-cover"
              loading="lazy"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent flex items-end p-6">
              <div className="text-white">
                <h2 className="text-2xl font-bold">{banner.title}</h2>
                {banner.subtitle && <p className="text-neutral-200 mt-1">{banner.subtitle}</p>}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
