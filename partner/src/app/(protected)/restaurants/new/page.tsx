"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Select } from "@/components/ui/select";

const days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"] as const;

const hourSchema = z.object({
  open: z.string(),
  close: z.string(),
  isClosed: z.boolean(),
});

const schema = z.object({
  name: z.string().min(2, "Name is required"),
  description: z.string().min(10, "Description must be at least 10 characters"),
  cuisines: z.string().min(1, "At least one cuisine required"),
  address: z.string().min(5, "Address is required"),
  city: z.string().min(2, "City is required"),
  area: z.string().min(2, "Area is required"),
  vatRate: z.coerce.number().min(0).max(100),
  prepTime: z.coerce.number().min(5).max(120),
  hours: z.object(Object.fromEntries(days.map((d) => [d, hourSchema])) as Record<string, typeof hourSchema>),
});

type RestaurantValues = z.infer<typeof schema>;

const defaultHours = Object.fromEntries(
  days.map((d) => [d, { open: "10:00", close: "22:00", isClosed: false }]),
) as Record<string, { open: string; close: string; isClosed: boolean }>;

export default function NewRestaurantPage() {
  const router = useRouter();

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<RestaurantValues>({
    resolver: zodResolver(schema),
    defaultValues: { vatRate: 5, prepTime: 25, hours: defaultHours },
  });

  const hours = watch("hours");

  const onSubmit = async (_values: RestaurantValues) => {
    // In production, POST to /partner/restaurants
    router.push("/restaurants");
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-lg font-semibold">Add Restaurant</h1>
      <form className="space-y-6 rounded-md border bg-white p-6" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Restaurant Name</label>
            <Input {...register("name")} placeholder="My Restaurant" />
            <p className="mt-1 text-xs text-rose-600">{errors.name?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Cuisines</label>
            <Input {...register("cuisines")} placeholder="Bangladeshi, Biryani, Mughlai" />
            <p className="mt-1 text-xs text-rose-600">{errors.cuisines?.message}</p>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium">Description</label>
          <Textarea {...register("description")} rows={3} placeholder="Describe your restaurant..." />
          <p className="mt-1 text-xs text-rose-600">{errors.description?.message}</p>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium">Address</label>
            <Input {...register("address")} placeholder="Street address" />
            <p className="mt-1 text-xs text-rose-600">{errors.address?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">City</label>
            <Input {...register("city")} placeholder="Dhaka" />
            <p className="mt-1 text-xs text-rose-600">{errors.city?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Area</label>
            <Input {...register("area")} placeholder="Gulshan" />
            <p className="mt-1 text-xs text-rose-600">{errors.area?.message}</p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">VAT Rate (%)</label>
            <Input type="number" step="0.1" {...register("vatRate")} />
            <p className="mt-1 text-xs text-rose-600">{errors.vatRate?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Prep Time (minutes)</label>
            <Input type="number" {...register("prepTime")} />
            <p className="mt-1 text-xs text-rose-600">{errors.prepTime?.message}</p>
          </div>
        </div>

        <div>
          <h3 className="mb-3 text-sm font-semibold">Operating Hours</h3>
          <div className="space-y-2">
            {days.map((day) => (
              <div key={day} className="flex items-center gap-3">
                <span className="w-24 text-sm font-medium">{day}</span>
                <label className="flex items-center gap-1 text-sm">
                  <input
                    type="checkbox"
                    checked={!hours?.[day]?.isClosed}
                    onChange={(e) => setValue(`hours.${day}.isClosed` as keyof RestaurantValues, !e.target.checked)}
                  />
                  Open
                </label>
                {!hours?.[day]?.isClosed && (
                  <>
                    <Input
                      type="time"
                      className="w-32"
                      {...register(`hours.${day}.open` as keyof RestaurantValues)}
                    />
                    <span className="text-sm">to</span>
                    <Input
                      type="time"
                      className="w-32"
                      {...register(`hours.${day}.close` as keyof RestaurantValues)}
                    />
                  </>
                )}
              </div>
            ))}
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium">Images</label>
          <Input type="file" accept="image/*" multiple />
          <p className="mt-1 text-xs text-slate-500">Upload restaurant logo and cover images</p>
        </div>

        <div className="flex gap-2">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Creating..." : "Create Restaurant"}
          </Button>
          <Button type="button" className="bg-slate-200 text-slate-700 hover:bg-slate-300" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
