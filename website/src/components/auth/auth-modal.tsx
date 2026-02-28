"use client";

import React, { useState } from "react";
import { useAuthStore } from "@/stores/auth-store";

type AuthStep = "phone" | "otp" | "register";

export function AuthModal() {
  const { isAuthModalOpen, closeAuthModal, setAuthenticated } = useAuthStore();

  const [step, setStep] = useState<AuthStep>("phone");
  const [phone, setPhone] = useState("");
  const [otp, setOtp] = useState("");
  const [name, setName] = useState("");
  
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isAuthModalOpen) return null;

  const handleSendOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!phone) return;
    setIsLoading(true);
    setError(null);
    try {
      // Mock API call to send OTP
      await new Promise((resolve) => setTimeout(resolve, 800));
      setStep("otp");
    } catch (err: Error | unknown) {
      setError((err as Error).message || "Failed to send OTP");
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerifyOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!otp) return;
    setIsLoading(true);
    setError(null);
    try {
      // Mock API call to verify OTP
      await new Promise((resolve) => setTimeout(resolve, 800));
      
      // For demonstration, if phone ends with '0', we assume new user
      if (phone.endsWith("0")) {
        setStep("register");
      } else {
        await completeAuthentication("mock_jwt_token_existing_user");
      }
    } catch (err: Error | unknown) {
      setError((err as Error).message || "Failed to verify OTP");
    } finally {
      setIsLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) return;
    setIsLoading(true);
    setError(null);
    try {
      // Mock API call to register user
      await new Promise((resolve) => setTimeout(resolve, 800));
      await completeAuthentication("mock_jwt_token_new_user");
    } catch (err: Error | unknown) {
      setError((err as Error).message || "Failed to register");
    } finally {
      setIsLoading(false);
    }
  };

  const completeAuthentication = async (token: string) => {
    try {
      // Store token in httpOnly cookie via our Next route handler
      const res = await fetch("/api/auth/session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token }),
      });
      if (!res.ok) throw new Error("Failed to set session");
      
      setAuthenticated(phone);
      closeAuthModal();
      setStep("phone");
      setPhone("");
      setOtp("");
      setName("");
    } catch {
      setError("Failed to complete login");
    }
  };

  const handleClose = () => {
    closeAuthModal();
    // Re-set state after animation (simulate here by immediate reset)
    setTimeout(() => {
      setStep("phone");
      setError(null);
    }, 200);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm px-4">
      <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl animate-in fade-in zoom-in duration-200">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-neutral-900">
            {step === "phone" && "Welcome"}
            {step === "otp" && "Verify your number"}
            {step === "register" && "Complete your profile"}
          </h2>
          <button
            onClick={handleClose}
            className="text-neutral-500 hover:text-neutral-700 transition-colors"
            aria-label="Close"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
            </svg>
          </button>
        </div>

        {error && (
          <div className="mb-4 p-3 text-sm text-red-600 bg-red-50 rounded-lg">
            {error}
          </div>
        )}

        {step === "phone" && (
          <form onSubmit={handleSendOtp} className="space-y-4">
            <div>
              <label htmlFor="phone" className="block text-sm font-medium text-neutral-700 mb-1">
                Phone Number
              </label>
              <input
                id="phone"
                type="tel"
                required
                placeholder="+1 (555) 000-0000"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className="w-full rounded-lg border border-neutral-300 px-3 py-2 text-neutral-900 focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500"
                disabled={isLoading}
              />
            </div>
            <button
              type="submit"
              disabled={isLoading || !phone}
              className="w-full rounded-lg bg-orange-500 px-4 py-2 font-medium text-white hover:bg-orange-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isLoading ? "Sending..." : "Continue"}
            </button>
          </form>
        )}

        {step === "otp" && (
          <form onSubmit={handleVerifyOtp} className="space-y-4">
            <p className="text-sm text-neutral-600">
              Enter the 6-digit code sent to <span className="font-medium text-neutral-900">{phone}</span>
            </p>
            <div>
              <label htmlFor="otp" className="block text-sm font-medium text-neutral-700 mb-1">
                Security Code
              </label>
              <input
                id="otp"
                type="text"
                required
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                placeholder="000000"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                className="w-full rounded-lg border border-neutral-300 px-3 py-2 text-center text-lg tracking-widest text-neutral-900 focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500"
                disabled={isLoading}
              />
            </div>
            <button
              type="submit"
              disabled={isLoading || otp.length < 4}
              className="w-full rounded-lg bg-orange-500 px-4 py-2 font-medium text-white hover:bg-orange-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isLoading ? "Verifying..." : "Verify Code"}
            </button>
            <button
              type="button"
              onClick={() => setStep("phone")}
              className="w-full text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
              disabled={isLoading}
            >
              Use a different number
            </button>
          </form>
        )}

        {step === "register" && (
          <form onSubmit={handleRegister} className="space-y-4">
            <p className="text-sm text-neutral-600">
              Welcome to Munchies! Please tell us your name.
            </p>
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-neutral-700 mb-1">
                Full Name
              </label>
              <input
                id="name"
                type="text"
                required
                placeholder="John Doe"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full rounded-lg border border-neutral-300 px-3 py-2 text-neutral-900 focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500"
                disabled={isLoading}
              />
            </div>
            <button
              type="submit"
              disabled={isLoading || !name}
              className="w-full rounded-lg bg-orange-500 px-4 py-2 font-medium text-white hover:bg-orange-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isLoading ? "Saving..." : "Create Account"}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
