"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

const schema = z.object({
  code: z.string().min(3, "Code must be at least 3 characters").max(20),
  type: z.enum(["percentage", "flat", "cashback"]),
  amount: z.number().min(1, "Amount is required"),
  cap: z.number().min(0).optional(),
  applyOn: z.enum(["order", "delivery", "product"]),
  minOrderAmount: z.number().min(0).optional(),
  maxUsage: z.number().min(1).optional(),
  perUserLimit: z.number().min(1).optional(),
  startsAt: z.string().min(1, "Start date required"),
  endsAt: z.string().min(1, "End date required"),
  cashbackAmount: z.number().min(0).optional(),
  description: z.string().optional(),
});

type PromoValues = z.infer<typeof schema>;

export default function NewPromotionPage() {
  const router = useRouter();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<PromoValues>({
    resolver: zodResolver(schema),
    defaultValues: { type: "percentage", applyOn: "order" },
  });

  const promoType = watch("type");

  const onSubmit = async () => {
    router.push("/promotions");
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-lg font-semibold">Create Promotion</h1>
      <form className="space-y-4 rounded-md border bg-white p-6" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Promo Code</label>
            <Input {...register("code")} placeholder="WELCOME20" className="uppercase" />
            <p className="mt-1 text-xs text-rose-600">{errors.code?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Type</label>
            <Select {...register("type")}>
              <option value="percentage">Percentage</option>
              <option value="flat">Flat Amount</option>
              <option value="cashback">Cashback</option>
            </Select>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">
              {promoType === "percentage" ? "Discount (%)" : "Discount Amount (৳)"}
            </label>
            <Input type="number" {...register("amount")} placeholder={promoType === "percentage" ? "20" : "50"} />
            <p className="mt-1 text-xs text-rose-600">{errors.amount?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Maximum Cap (৳)</label>
            <Input type="number" {...register("cap")} placeholder="200" />
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Apply On</label>
            <Select {...register("applyOn")}>
              <option value="order">Entire Order</option>
              <option value="delivery">Delivery Fee</option>
              <option value="product">Specific Product</option>
            </Select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Min Order Amount (৳)</label>
            <Input type="number" {...register("minOrderAmount")} placeholder="0" />
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium">Max Total Usage</label>
            <Input type="number" {...register("maxUsage")} placeholder="Unlimited" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Per-User Limit</label>
            <Input type="number" {...register("perUserLimit")} placeholder="1" />
          </div>
          {promoType === "cashback" && (
            <div>
              <label className="mb-1 block text-sm font-medium">Cashback Amount (৳)</label>
              <Input type="number" {...register("cashbackAmount")} placeholder="50" />
            </div>
          )}
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Start Date</label>
            <Input type="date" {...register("startsAt")} />
            <p className="mt-1 text-xs text-rose-600">{errors.startsAt?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">End Date</label>
            <Input type="date" {...register("endsAt")} />
            <p className="mt-1 text-xs text-rose-600">{errors.endsAt?.message}</p>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium">Description (optional)</label>
          <Textarea {...register("description")} rows={2} placeholder="Internal notes about this promotion" />
        </div>

        <div className="flex gap-2">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Creating..." : "Create Promotion"}
          </Button>
          <Button type="button" className="bg-slate-200 text-slate-700 hover:bg-slate-300" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
