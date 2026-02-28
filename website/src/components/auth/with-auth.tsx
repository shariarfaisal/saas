"use client";

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

export function withAuth<P extends object>(WrappedComponent: React.ComponentType<P>) {
  return function ProtectedRoute(props: P) {
    const { isAuthenticated, openAuthModal } = useAuthStore();
    const router = useRouter();
    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
      setIsMounted(true);
    }, []);

    useEffect(() => {
      if (isMounted && !isAuthenticated) {
        openAuthModal();
        router.push("/"); // Redirect to home or another public page
      }
    }, [isAuthenticated, isMounted, openAuthModal, router]);

    if (!isMounted || !isAuthenticated) {
      return null;
    }

    return <WrappedComponent {...props} />;
  };
}
