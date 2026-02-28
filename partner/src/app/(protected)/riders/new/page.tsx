"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";

const schema = z.object({
  name: z.string().min(2, "Name is required"),
  phone: z.string().min(11, "Valid phone number required"),
  email: z.email().optional().or(z.literal("")),
  hubId: z.string().min(1, "Hub is required"),
  vehicleType: z.string().min(1, "Vehicle type required"),
  licensePlate: z.string().optional(),
});

type RiderValues = z.infer<typeof schema>;

export default function NewRiderPage() {
  const router = useRouter();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RiderValues>({ resolver: zodResolver(schema) });

  const onSubmit = async (_values: RiderValues) => {
    router.push("/riders");
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-lg font-semibold">Add Rider</h1>
      <form className="space-y-4 rounded-md border bg-white p-6" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Full Name</label>
            <Input {...register("name")} placeholder="Rider name" />
            <p className="mt-1 text-xs text-rose-600">{errors.name?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Phone</label>
            <Input {...register("phone")} placeholder="+8801XXXXXXXXX" />
            <p className="mt-1 text-xs text-rose-600">{errors.phone?.message}</p>
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Email (optional)</label>
            <Input {...register("email")} placeholder="rider@email.com" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Hub</label>
            <Select {...register("hubId")}>
              <option value="">Select hub</option>
              <option value="hub-1">Gulshan Hub</option>
              <option value="hub-2">Banani Hub</option>
              <option value="hub-3">Dhanmondi Hub</option>
              <option value="hub-4">Uttara Hub</option>
              <option value="hub-5">Mirpur Hub</option>
            </Select>
            <p className="mt-1 text-xs text-rose-600">{errors.hubId?.message}</p>
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium">Vehicle Type</label>
            <Select {...register("vehicleType")}>
              <option value="">Select vehicle</option>
              <option value="motorcycle">Motorcycle</option>
              <option value="bicycle">Bicycle</option>
              <option value="car">Car</option>
            </Select>
            <p className="mt-1 text-xs text-rose-600">{errors.vehicleType?.message}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">License Plate (optional)</label>
            <Input {...register("licensePlate")} placeholder="DHAKA-1234" />
          </div>
        </div>
        <div className="flex gap-2">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Creating..." : "Create Rider"}
          </Button>
          <Button type="button" className="bg-slate-200 text-slate-700 hover:bg-slate-300" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
