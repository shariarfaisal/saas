"use client";

import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/lib/api-client";

const days = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"] as const;

const hourSchema = z.object({
  open: z.string(),
  close: z.string(),
  isClosed: z.boolean(),
});

const schema = z.object({
  name: z.string().min(2),
  description: z.string().min(10),
  cuisines: z.string().min(1),
  phone: z.string().optional(),
  email: z.string().email("Invalid email").optional().or(z.literal("")),
  address: z.string().min(5),
  city: z.string().min(2),
  area: z.string().min(2),
  logoUrl: z.string().optional(),
  bannerImageUrl: z.string().optional(),
  vatRate: z.number().min(0).max(100),
  isVatInclusive: z.boolean(),
  minOrderAmount: z.number().min(0),
  prepTime: z.number().min(5).max(120),
  autoAcceptOrders: z.boolean(),
  hours: z.object(Object.fromEntries(days.map((d) => [d, hourSchema])) as Record<string, typeof hourSchema>),
});

type EditValues = z.infer<typeof schema>;

type OperatingHour = {
  day_of_week: number;
  open_time: string;
  close_time: string;
  is_closed: boolean;
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function mapRestaurantToForm(r: any): Partial<EditValues> {
  return {
    name: r.name ?? "",
    description: r.description ?? "",
    cuisines: (r.cuisines ?? []).join(", "),
    phone: r.phone ?? "",
    email: r.email ?? "",
    address: r.address_line1 ?? "",
    city: r.city ?? "",
    area: r.area ?? "",
    logoUrl: r.logo_url ?? "",
    bannerImageUrl: r.banner_image_url ?? "",
    vatRate: parseFloat(r.vat_rate ?? "5"),
    isVatInclusive: r.is_vat_inclusive ?? false,
    minOrderAmount: parseFloat(r.min_order_amount ?? "0"),
    prepTime: r.avg_prep_time_minutes ?? 25,
    autoAcceptOrders: r.auto_accept_orders ?? false,
  };
}

function hoursFromApi(apiHours: OperatingHour[]): EditValues["hours"] {
  const defaultH = { open: "10:00", close: "22:00", isClosed: false };
  const result = Object.fromEntries(days.map((d) => [d, { ...defaultH }])) as EditValues["hours"];
  for (const h of apiHours) {
    const day = days[h.day_of_week];
    if (day) {
      result[day] = {
        open: (h.open_time ?? "10:00").slice(0, 5),   // DB stores HH:MM:SS; time input needs HH:MM
        close: (h.close_time ?? "22:00").slice(0, 5), // DB stores HH:MM:SS; time input needs HH:MM
        isClosed: h.is_closed,
      };
    }
  }
  return result;
}

export default function RestaurantDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [restaurant, setRestaurant] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [apiError, setApiError] = useState<string | null>(null);
  const [availLoading, setAvailLoading] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<EditValues>({
    resolver: zodResolver(schema),
  });

  const hours = watch("hours");

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const [resRes, hoursRes] = await Promise.all([
          apiClient.get(`/partner/restaurants/${id}`),
          apiClient.get(`/partner/restaurants/${id}/hours`),
        ]);
        if (!mounted) return;
        const r = resRes.data;
        setRestaurant(r);
        reset({
          ...mapRestaurantToForm(r),
          hours: hoursFromApi(hoursRes.data ?? []),
        });
      } catch {
        if (mounted) setApiError("Failed to load restaurant data");
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, [id, reset]);

  const toggleAvailability = async () => {
    if (!restaurant) return;
    setAvailLoading(true);
    try {
      const res = await apiClient.patch(`/partner/restaurants/${id}/availability`, {
        is_available: !restaurant.is_available,
      });
      setRestaurant(res.data);
    } catch {
      setApiError("Failed to update availability");
    } finally {
      setAvailLoading(false);
    }
  };

  const onSubmit = async (data: EditValues) => {
    setApiError(null);
    try {
      const cuisineArr = data.cuisines
        .split(",")
        .map((c) => c.trim())
        .filter(Boolean);

      const hoursPayload = days.map((day, idx) => ({
        day_of_week: idx,
        open_time: data.hours[day].open,
        close_time: data.hours[day].close,
        is_closed: data.hours[day].isClosed,
      }));

      await Promise.all([
        apiClient.put(`/partner/restaurants/${id}`, {
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
        }),
        apiClient.put(`/partner/restaurants/${id}/hours`, hoursPayload),
      ]);

      router.push("/restaurants");
    } catch (err: unknown) {
      const msg = (err as { message?: string })?.message ?? "Failed to save changes";
      setApiError(msg);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20 text-sm text-slate-500">
        Loading restaurant…
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Edit Restaurant</h1>
        <div className="flex items-center gap-3">
          <Badge variant={restaurant?.is_available ? "success" : "danger"}>
            {restaurant?.is_available ? "Open" : "Closed"}
          </Badge>
          <button
            type="button"
            disabled={availLoading}
            onClick={toggleAvailability}
            className="rounded-md border px-3 py-1.5 text-sm font-medium hover:bg-slate-50 disabled:opacity-50"
          >
            {availLoading ? "Updating…" : restaurant?.is_available ? "Close Restaurant" : "Open Restaurant"}
          </button>
        </div>
      </div>

      <form className="space-y-6 rounded-md border bg-white p-6" onSubmit={handleSubmit(onSubmit)}>
        {apiError && (
          <div className="rounded-md bg-rose-50 p-3 text-sm text-rose-700">{apiError}</div>
        )}

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Restaurant Name</label>
            <Input {...register("name")} />
            <p className="mt-1 text-xs text-rose-600">{errors.name?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Cuisines</label>
            <Input {...register("cuisines")} placeholder="Bangladeshi, Biryani" />
            <p className="mt-1 text-xs text-rose-600">{errors.cuisines?.message}</p>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium">Description</label>
          <Textarea {...register("description")} rows={3} />
          <p className="mt-1 text-xs text-rose-600">{errors.description?.message}</p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Phone</label>
            <Input {...register("phone")} placeholder="+880 1700 000000" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Email</label>
            <Input {...register("email")} type="email" placeholder="restaurant@example.com" />
            <p className="mt-1 text-xs text-rose-600">{errors.email?.message}</p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium">Address</label>
            <Input {...register("address")} />
            <p className="mt-1 text-xs text-rose-600">{errors.address?.message}</p>
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
            <label className="mb-1 block text-sm font-medium">Logo URL</label>
            <Input {...register("logoUrl")} placeholder="https://..." />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Banner Image URL</label>
            <Input {...register("bannerImageUrl")} placeholder="https://..." />
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium">VAT Rate (%)</label>
            <Input type="number" step="0.1" {...register("vatRate", { valueAsNumber: true })} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Min Order (৳)</label>
            <Input type="number" {...register("minOrderAmount", { valueAsNumber: true })} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Prep Time (min)</label>
            <Input type="number" {...register("prepTime", { valueAsNumber: true })} />
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
        </div>

        <div>
          <h3 className="mb-3 text-sm font-semibold">Operating Hours</h3>
          <div className="space-y-2">
            {hours &&
              days.map((day) => (
                <div key={day} className="flex items-center gap-3">
                  <span className="w-24 text-sm font-medium">{day}</span>
                  <label className="flex items-center gap-1 text-sm">
                    <input
                      type="checkbox"
                      checked={!hours?.[day]?.isClosed}
                      onChange={(e) =>
                        setValue(
                          `hours.${day}.isClosed` as keyof EditValues,
                          (!e.target.checked) as never,
                        )
                      }
                    />
                    Open
                  </label>
                  {!hours?.[day]?.isClosed && (
                    <>
                      <Input
                        type="time"
                        className="w-32"
                        {...register(`hours.${day}.open` as keyof EditValues)}
                      />
                      <span className="text-sm">to</span>
                      <Input
                        type="time"
                        className="w-32"
                        {...register(`hours.${day}.close` as keyof EditValues)}
                      />
                    </>
                  )}
                </div>
              ))}
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
