"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const tenantSchema = z.object({
  name: z.string().min(2),
  slug: z.string().min(2),
  plan: z.string().min(2),
  commissionRate: z.coerce.number().min(0).max(100),
});

export default function CreateTenantPage() {
  const { register, handleSubmit } = useForm({
    resolver: zodResolver(tenantSchema),
  });

  return (
    <form className="grid max-w-xl gap-3 rounded-md border bg-white p-4" onSubmit={handleSubmit(() => undefined)}>
      <h1 className="text-lg font-semibold">Create Tenant</h1>
      <Input placeholder="Tenant name" {...register("name")} />
      <Input placeholder="slug" {...register("slug")} />
      <Input placeholder="plan" {...register("plan")} />
      <Input placeholder="commission %" type="number" {...register("commissionRate")} />
      <Button type="submit">Create</Button>
    </form>
  );
}
