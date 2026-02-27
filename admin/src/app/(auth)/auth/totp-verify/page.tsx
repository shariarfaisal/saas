"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export default function TotpVerifyPage() {
  const router = useRouter();
  const [code, setCode] = useState("");

  const verifyCode = async () => {
    await fetch("/api/auth/totp/verify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
    });
    router.push("/");
  };

  return (
    <div className="mx-auto mt-16 max-w-md rounded-md border bg-white p-6">
      <h1 className="mb-2 text-xl font-semibold">Verify TOTP</h1>
      <Input onChange={(event) => setCode(event.target.value)} placeholder="123456" value={code} />
      <Button className="mt-4 w-full" onClick={verifyCode}>
        Verify Login
      </Button>
    </div>
  );
}
