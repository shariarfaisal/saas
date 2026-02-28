"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";

const schema = z.object({
  code: z.string().min(3).max(20),
  type: z.enum(["percentage", "flat", "cashback"]),
  amount: z.number().min(1),
  cap: z.number().min(0).optional(),
  startsAt: z.string().min(1),
  endsAt: z.string().min(1),
});

type PromoValues = z.infer<typeof schema>;

const mockPromo = {
  code: "WELCOME20",
  type: "percentage" as const,
  amount: 20,
  cap: 200,
  startsAt: "2024-12-01",
  endsAt: "2025-01-31",
  status: "active" as const,
  usageCount: 456,
  totalDiscountGiven: 45600,
  uniqueUsers: 312,
};

export default function EditPromotionPage() {
  const router = useRouter();

  const {
    register,
    handleSubmit,
    formState: { isSubmitting },
  } = useForm<PromoValues>({
    resolver: zodResolver(schema),
    defaultValues: mockPromo,
  });

  const onSubmit = async () => {
    router.push("/promotions");
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Edit Promotion</h1>
        <Badge variant="success">{mockPromo.status}</Badge>
      </div>

      {/* Performance Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <p className="text-xs text-slate-500">Usage Count</p>
          <p className="mt-1 text-xl font-semibold">{mockPromo.usageCount}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Total Discount Given</p>
          <p className="mt-1 text-xl font-semibold">{formatCurrency(mockPromo.totalDiscountGiven)}</p>
        </Card>
        <Card>
          <p className="text-xs text-slate-500">Unique Users</p>
          <p className="mt-1 text-xl font-semibold">{mockPromo.uniqueUsers}</p>
        </Card>
      </div>

      <form className="space-y-4 rounded-md border bg-white p-6" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Promo Code</label>
            <Input {...register("code")} className="uppercase" />
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
            <label className="mb-1 block text-sm font-medium">Amount</label>
            <Input type="number" {...register("amount")} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Cap (৳)</label>
            <Input type="number" {...register("cap")} />
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Start Date</label>
            <Input type="date" {...register("startsAt")} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">End Date</label>
            <Input type="date" {...register("endsAt")} />
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
