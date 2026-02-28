"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";

const days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"] as const;

const schema = z.object({
  name: z.string().min(2),
  description: z.string().min(10),
  cuisines: z.string().min(1),
  address: z.string().min(5),
  city: z.string().min(2),
  area: z.string().min(2),
  vatRate: z.number().min(0).max(100),
  prepTime: z.number().min(5).max(120),
});

type EditValues = z.infer<typeof schema>;

const mockRestaurant = {
  id: "rest-1",
  name: "Kacchi Bhai - Main Branch",
  description: "Authentic Kacchi Biryani & Bangladeshi Cuisine",
  cuisines: "Bangladeshi, Biryani, Mughlai",
  address: "House 12, Road 5, Gulshan 2, Dhaka",
  city: "Dhaka",
  area: "Gulshan",
  vatRate: 5,
  prepTime: 25,
  isAvailable: true,
  hours: Object.fromEntries(
    days.map((d) => [d, { open: "10:00", close: "22:00", isClosed: d === "Friday" }]),
  ),
};

export default function RestaurantDetailPage() {
  const router = useRouter();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<EditValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: mockRestaurant.name,
      description: mockRestaurant.description,
      cuisines: mockRestaurant.cuisines,
      address: mockRestaurant.address,
      city: mockRestaurant.city,
      area: mockRestaurant.area,
      vatRate: mockRestaurant.vatRate,
      prepTime: mockRestaurant.prepTime,
    },
  });

  const onSubmit = async () => {
    // PUT /partner/restaurants/:id
    router.push("/restaurants");
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Edit Restaurant</h1>
        <Badge variant={mockRestaurant.isAvailable ? "success" : "danger"}>
          {mockRestaurant.isAvailable ? "Available" : "Unavailable"}
        </Badge>
      </div>

      <form className="space-y-6 rounded-md border bg-white p-6" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Restaurant Name</label>
            <Input {...register("name")} />
            <p className="mt-1 text-xs text-rose-600">{errors.name?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Cuisines</label>
            <Input {...register("cuisines")} />
            <p className="mt-1 text-xs text-rose-600">{errors.cuisines?.message}</p>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium">Description</label>
          <Textarea {...register("description")} rows={3} />
          <p className="mt-1 text-xs text-rose-600">{errors.description?.message}</p>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium">Address</label>
            <Input {...register("address")} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">City</label>
            <Input {...register("city")} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Area</label>
            <Input {...register("area")} />
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">VAT Rate (%)</label>
            <Input type="number" step="0.1" {...register("vatRate", { valueAsNumber: true })} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Prep Time (minutes)</label>
            <Input type="number" {...register("prepTime", { valueAsNumber: true })} />
          </div>
        </div>

        <div>
          <h3 className="mb-3 text-sm font-semibold">Operating Hours</h3>
          <div className="space-y-2">
            {days.map((day) => {
              const h = mockRestaurant.hours[day];
              return (
                <div key={day} className="flex items-center gap-3">
                  <span className="w-24 text-sm font-medium">{day}</span>
                  <Badge variant={h.isClosed ? "danger" : "success"}>
                    {h.isClosed ? "Closed" : `${h.open} - ${h.close}`}
                  </Badge>
                </div>
              );
            })}
          </div>
        </div>

        <div className="flex gap-2">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Saving..." : "Save Changes"}
          </Button>
          <Button type="button" className="bg-slate-200 text-slate-700 hover:bg-slate-300" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
