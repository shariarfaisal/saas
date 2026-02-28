"use client";

import { useEffect, useCallback } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { apiClient } from "@/lib/api-client";

export function useNotificationPolling(intervalMs = 30000) {
  const setUnreadNotifications = useAuthStore((s) => s.setUnreadNotifications);

  const fetchCount = useCallback(async () => {
    try {
      const res = await apiClient.get<{ count: number }>("/partner/notifications/unread-count");
      setUnreadNotifications(res.data.count);
    } catch {
      // Silently ignore polling errors
    }
  }, [setUnreadNotifications]);

  useEffect(() => {
    fetchCount();
    const timer = setInterval(fetchCount, intervalMs);
    return () => clearInterval(timer);
  }, [fetchCount, intervalMs]);
}
