"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const schema = z
  .object({
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type ResetValues = z.infer<typeof schema>;

function ResetPasswordForm() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const [success, setSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetValues>({ resolver: zodResolver(schema) });

  const onSubmit = async () => {
    // In production, call backend with token + new password
    setSuccess(true);
  };

  if (!token) {
    return (
      <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
        <h1 className="mb-4 text-xl font-semibold">Invalid Link</h1>
        <p className="text-sm text-slate-600">This password reset link is invalid or has expired.</p>
        <Link className="mt-4 inline-block text-sm text-slate-600 underline" href="/auth/forgot-password">
          Request a new reset link
        </Link>
      </div>
    );
  }

  if (success) {
    return (
      <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
        <h1 className="mb-4 text-xl font-semibold">Password Reset</h1>
        <p className="text-sm text-slate-600">Your password has been reset successfully. You can now log in.</p>
        <Link className="mt-4 inline-block text-sm text-slate-600 underline" href="/auth/login">
          Go to login
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
      <h1 className="mb-4 text-xl font-semibold">Set New Password</h1>
      <form className="space-y-4" onSubmit={handleSubmit(onSubmit)}>
        <div>
          <Input placeholder="New password" type="password" {...register("password")} />
          <p className="mt-1 text-xs text-rose-600">{errors.password?.message}</p>
        </div>
        <div>
          <Input placeholder="Confirm password" type="password" {...register("confirmPassword")} />
          <p className="mt-1 text-xs text-rose-600">{errors.confirmPassword?.message}</p>
        </div>
        <Button className="w-full" disabled={isSubmitting} type="submit">
          {isSubmitting ? "Resetting..." : "Reset Password"}
        </Button>
      </form>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={<div className="mx-auto mt-16 max-w-md p-6 text-center text-slate-500">Loading...</div>}>
      <ResetPasswordForm />
    </Suspense>
  );
}
