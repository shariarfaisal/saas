"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const schema = z.object({
  email: z.email(),
});

type ForgotValues = z.infer<typeof schema>;

export default function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotValues>({ resolver: zodResolver(schema) });

  const onSubmit = async () => {
    // In production, call the backend password reset endpoint
    setSent(true);
  };

  if (sent) {
    return (
      <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
        <h1 className="mb-4 text-xl font-semibold">Check Your Email</h1>
        <p className="text-sm text-slate-600">
          If an account exists with that email, we&apos;ve sent a password reset link. Please check your inbox.
        </p>
        <Link className="mt-4 inline-block text-sm text-slate-600 underline" href="/auth/login">
          Back to login
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
      <h1 className="mb-4 text-xl font-semibold">Forgot Password</h1>
      <p className="mb-4 text-sm text-slate-600">Enter your email address and we&apos;ll send you a reset link.</p>
      <form className="space-y-4" onSubmit={handleSubmit(onSubmit)}>
        <div>
          <Input placeholder="you@restaurant.com" {...register("email")} />
          <p className="mt-1 text-xs text-rose-600">{errors.email?.message}</p>
        </div>
        <Button className="w-full" disabled={isSubmitting} type="submit">
          {isSubmitting ? "Sending..." : "Send Reset Link"}
        </Button>
      </form>
      <Link className="mt-3 inline-block text-xs text-slate-600 underline" href="/auth/login">
        Back to login
      </Link>
    </div>
  );
}
