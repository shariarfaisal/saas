"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const schema = z.object({
  email: z.email(),
  password: z.string().min(8),
});

type LoginValues = z.infer<typeof schema>;

export default function LoginPage() {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({ resolver: zodResolver(schema) });

  const onSubmit = async (values: LoginValues) => {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(values),
    });
    const result = (await response.json()) as { next: "setup" | "verify" };
    router.push(result.next === "setup" ? "/auth/totp-setup" : "/auth/totp-verify");
  };

  return (
    <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
      <h1 className="mb-4 text-xl font-semibold">Admin Login</h1>
      <form className="space-y-4" onSubmit={handleSubmit(onSubmit)}>
        <div>
          <Input placeholder="admin@platform.com" {...register("email")} />
          <p className="mt-1 text-xs text-rose-600">{errors.email?.message}</p>
        </div>
        <div>
          <Input placeholder="********" type="password" {...register("password")} />
          <p className="mt-1 text-xs text-rose-600">{errors.password?.message}</p>
        </div>
        <Button className="w-full" disabled={isSubmitting} type="submit">
          {isSubmitting ? "Signing in..." : "Continue"}
        </Button>
      </form>
      <p className="mt-3 text-xs text-slate-500">
        First login redirects to QR setup. Subsequent logins require TOTP verification.
      </p>
      <Link className="mt-2 inline-block text-xs text-slate-600 underline" href="/auth/totp-verify">
        Already have QR? Go to verify
      </Link>
    </div>
  );
}
