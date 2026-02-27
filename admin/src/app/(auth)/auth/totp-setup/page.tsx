"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export default function TotpSetupPage() {
  const router = useRouter();
  const [code, setCode] = useState("");

  const verifySetup = async () => {
    await fetch("/api/auth/totp/setup-verify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
    });
    router.push("/");
  };

  return (
    <div className="mx-auto mt-12 max-w-md rounded-md border bg-white p-6">
      <h1 className="text-xl font-semibold">Set up 2FA (TOTP)</h1>
      <p className="mt-2 text-sm text-slate-600">Scan this QR seed in your authenticator app:</p>
      <div className="mt-3 rounded-md border bg-slate-100 p-4 text-xs">
        QR provisioning URI is provided securely by backend during setup.
      </div>
      <div className="mt-4">
        <Input onChange={(event) => setCode(event.target.value)} placeholder="Enter 6-digit TOTP" value={code} />
      </div>
      <Button className="mt-3 w-full" onClick={verifySetup}>
        Verify and Continue
      </Button>
    </div>
  );
}
