"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const schema = z
  .object({
    name: z.string().min(2, "Name is required"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type InviteValues = z.infer<typeof schema>;

export default function InvitePage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const token = searchParams.get("token");
  const email = searchParams.get("email");
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<InviteValues>({ resolver: zodResolver(schema) });

  const onSubmit = async (_values: InviteValues) => {
    setError(null);
    try {
      // In production, call backend to accept invitation
      // POST /api/v1/auth/accept-invite { token, name, password }
      router.push("/auth/login");
    } catch {
      setError("Failed to accept invitation. The link may have expired.");
    }
  };

  if (!token) {
    return (
      <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
        <h1 className="mb-4 text-xl font-semibold">Invalid Invitation</h1>
        <p className="text-sm text-slate-600">This invitation link is invalid or has expired.</p>
        <Link className="mt-4 inline-block text-sm text-slate-600 underline" href="/auth/login">
          Go to login
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
      <h1 className="mb-4 text-xl font-semibold">Accept Invitation</h1>
      {email && <p className="mb-4 text-sm text-slate-600">You&apos;ve been invited as {email}</p>}
      {error && <p className="mb-3 rounded bg-rose-50 p-2 text-sm text-rose-600">{error}</p>}
      <form className="space-y-4" onSubmit={handleSubmit(onSubmit)}>
        <div>
          <Input placeholder="Your name" {...register("name")} />
          <p className="mt-1 text-xs text-rose-600">{errors.name?.message}</p>
        </div>
        <div>
          <Input placeholder="Choose password" type="password" {...register("password")} />
          <p className="mt-1 text-xs text-rose-600">{errors.password?.message}</p>
        </div>
        <div>
          <Input placeholder="Confirm password" type="password" {...register("confirmPassword")} />
          <p className="mt-1 text-xs text-rose-600">{errors.confirmPassword?.message}</p>
        </div>
        <Button className="w-full" disabled={isSubmitting} type="submit">
          {isSubmitting ? "Setting up..." : "Set Password & Join"}
        </Button>
      </form>
    </div>
  );
}
