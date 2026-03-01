"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { apiClient } from "@/lib/api-client";

const days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"] as const;

const DAY_MAP: Record<string, number> = {
  Sunday: 0,
  Monday: 1,
  Tuesday: 2,
  Wednesday: 3,
  Thursday: 4,
  Friday: 5,
  Saturday: 6,
};

const hourSchema = z.object({
  open: z.string(),
  close: z.string(),
  isClosed: z.boolean(),
});

const schema = z.object({
  name: z.string().min(2, "Name is required"),
  description: z.string().min(10, "Description must be at least 10 characters"),
  cuisines: z.string().min(1, "At least one cuisine required"),
  phone: z.string().optional(),
  email: z.string().email("Invalid email").optional().or(z.literal("")),
  address: z.string().min(5, "Address is required"),
  city: z.string().min(2, "City is required"),
  area: z.string().min(2, "Area is required"),
  logoUrl: z.string().url("Must be a valid URL").optional().or(z.literal("")),
  bannerImageUrl: z.string().url("Must be a valid URL").optional().or(z.literal("")),
  vatRate: z.number().min(0).max(100),
  isVatInclusive: z.boolean(),
  minOrderAmount: z.number().min(0),
  prepTime: z.number().min(5).max(120),
  autoAcceptOrders: z.boolean(),
  isAvailable: z.boolean(),
  hours: z.object(Object.fromEntries(days.map((d) => [d, hourSchema])) as Record<string, typeof hourSchema>),
});

type RestaurantValues = z.infer<typeof schema>;

const defaultHours = Object.fromEntries(
  days.map((d) => [d, { open: "10:00", close: "22:00", isClosed: false }]),
) as Record<string, { open: string; close: string; isClosed: boolean }>;

export default function NewRestaurantPage() {
  const router = useRouter();
  const [apiError, setApiError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<RestaurantValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      vatRate: 5,
      isVatInclusive: false,
      minOrderAmount: 0,
      prepTime: 25,
      autoAcceptOrders: false,
      isAvailable: true,
      hours: defaultHours,
    },
  });

  const hours = watch("hours");

  const onSubmit = async (data: RestaurantValues) => {
    setApiError(null);
    try {
      const cuisineArr = data.cuisines
        .split(",")
        .map((c) => c.trim())
        .filter(Boolean);

      const hoursPayload = days.map((day) => ({
        day_of_week: DAY_MAP[day],
        open_time: data.hours[day].open,
        close_time: data.hours[day].close,
        is_closed: data.hours[day].isClosed,
      }));

      const res = await apiClient.post("/partner/restaurants", {
        name: data.name,
        description: data.description,
        cuisines: cuisineArr,
        phone: data.phone || null,
        email: data.email || null,
        address_line1: data.address,
        city: data.city,
        area: data.area,
        logo_url: data.logoUrl || null,
        banner_image_url: data.bannerImageUrl || null,
        vat_rate: data.vatRate,
        is_vat_inclusive: data.isVatInclusive,
        min_order_amount: data.minOrderAmount,
        avg_prep_time_minutes: data.prepTime,
        auto_accept_orders: data.autoAcceptOrders,
        is_available: data.isAvailable,
        is_active: true,
      });

      const restaurantId: string = res.data?.id;
      if (restaurantId) {
        await apiClient.put(`/partner/restaurants/${restaurantId}/hours`, hoursPayload);
      }

      router.push("/restaurants");
    } catch (err: unknown) {
      const msg = (err as { message?: string })?.message ?? "Failed to create restaurant";
      setApiError(msg);
    }
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-lg font-semibold">Add Restaurant</h1>
      <form className="space-y-6 rounded-md border bg-white p-6" onSubmit={handleSubmit(onSubmit)}>
        {apiError && (
          <div className="rounded-md bg-rose-50 p-3 text-sm text-rose-700">{apiError}</div>
        )}

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Restaurant Name *</label>
            <Input {...register("name")} placeholder="My Restaurant" />
            <p className="mt-1 text-xs text-rose-600">{errors.name?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Cuisines *</label>
            <Input {...register("cuisines")} placeholder="Bangladeshi, Biryani, Mughlai" />
            <p className="mt-1 text-xs text-rose-600">{errors.cuisines?.message}</p>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium">Description *</label>
          <Textarea {...register("description")} rows={3} placeholder="Describe your restaurant..." />
          <p className="mt-1 text-xs text-rose-600">{errors.description?.message}</p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Phone</label>
            <Input {...register("phone")} placeholder="+880 1700 000000" />
            <p className="mt-1 text-xs text-rose-600">{errors.phone?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Email</label>
            <Input {...register("email")} type="email" placeholder="restaurant@example.com" />
            <p className="mt-1 text-xs text-rose-600">{errors.email?.message}</p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium">Address *</label>
            <Input {...register("address")} placeholder="Street address" />
            <p className="mt-1 text-xs text-rose-600">{errors.address?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">City *</label>
            <Input {...register("city")} placeholder="Dhaka" />
            <p className="mt-1 text-xs text-rose-600">{errors.city?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Area *</label>
            <Input {...register("area")} placeholder="Gulshan" />
            <p className="mt-1 text-xs text-rose-600">{errors.area?.message}</p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Logo URL</label>
            <Input {...register("logoUrl")} placeholder="https://..." />
            <p className="mt-1 text-xs text-rose-600">{errors.logoUrl?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Banner Image URL</label>
            <Input {...register("bannerImageUrl")} placeholder="https://..." />
            <p className="mt-1 text-xs text-rose-600">{errors.bannerImageUrl?.message}</p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium">VAT Rate (%)</label>
            <Input type="number" step="0.1" {...register("vatRate", { valueAsNumber: true })} />
            <p className="mt-1 text-xs text-rose-600">{errors.vatRate?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Min Order (৳)</label>
            <Input type="number" {...register("minOrderAmount", { valueAsNumber: true })} placeholder="0" />
            <p className="mt-1 text-xs text-rose-600">{errors.minOrderAmount?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Prep Time (minutes)</label>
            <Input type="number" {...register("prepTime", { valueAsNumber: true })} />
            <p className="mt-1 text-xs text-rose-600">{errors.prepTime?.message}</p>
          </div>
        </div>

        <div className="flex flex-wrap gap-6">
          <label className="flex items-center gap-2 text-sm font-medium">
            <input type="checkbox" {...register("isVatInclusive")} />
            VAT Inclusive
          </label>
          <label className="flex items-center gap-2 text-sm font-medium">
            <input type="checkbox" {...register("autoAcceptOrders")} />
            Auto-Accept Orders
          </label>
          <label className="flex items-center gap-2 text-sm font-medium">
            <input type="checkbox" {...register("isAvailable")} />
            Open / Available
          </label>
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
                    onChange={(e) =>
                      setValue(`hours.${day}.isClosed` as keyof RestaurantValues, (!e.target.checked) as never)
                    }
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
