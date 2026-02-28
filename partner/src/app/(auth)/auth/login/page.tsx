"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/auth-store";

const schema = z.object({
  email: z.email(),
  password: z.string().min(8),
});

type LoginValues = z.infer<typeof schema>;

export default function LoginPage() {
  const router = useRouter();
  const setSession = useAuthStore((s) => s.setSession);
  const setRestaurants = useAuthStore((s) => s.setRestaurants);
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({ resolver: zodResolver(schema) });

  const onSubmit = async (values: LoginValues) => {
    setError(null);
    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(values),
      });

      if (!response.ok) {
        const data = (await response.json()) as { message?: string };
        setError(data.message ?? "Login failed");
        return;
      }

      const data = (await response.json()) as {
        restaurants: Array<{ id: string; name: string; isAvailable: boolean }>;
      };

      setSession(values.email, Date.now() + 30 * 60 * 1000);
      setRestaurants(data.restaurants);

      if (data.restaurants.length > 1) {
        router.push("/auth/login?picker=1");
        return;
      }

      router.push("/");
    } catch {
      setError("Network error. Please try again.");
    }
  };

  return (
    <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
      <h1 className="mb-4 text-xl font-semibold">Partner Login</h1>
      {error && <p className="mb-3 rounded bg-rose-50 p-2 text-sm text-rose-600">{error}</p>}
      <form className="space-y-4" onSubmit={handleSubmit(onSubmit)}>
        <div>
          <Input placeholder="you@restaurant.com" {...register("email")} />
          <p className="mt-1 text-xs text-rose-600">{errors.email?.message}</p>
        </div>
        <div>
          <Input placeholder="********" type="password" {...register("password")} />
          <p className="mt-1 text-xs text-rose-600">{errors.password?.message}</p>
        </div>
        <Button className="w-full" disabled={isSubmitting} type="submit">
          {isSubmitting ? "Signing in..." : "Sign In"}
        </Button>
      </form>
      <a className="mt-3 inline-block text-xs text-slate-600 underline" href="/auth/forgot-password">
        Forgot password?
      </a>
    </div>
  );
}
