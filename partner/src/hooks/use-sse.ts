"use client";

import { useEffect, useRef, useCallback, useState } from "react";

type SSEOptions = {
  url: string;
  onMessage: (event: MessageEvent) => void;
  enabled?: boolean;
};

export function useSSE({ url, onMessage, enabled = true }: SSEOptions) {
  const [connected, setConnected] = useState(false);
  const eventSourceRef = useRef<EventSource | null>(null);
  const retryCountRef = useRef(0);
  const maxRetries = 5;

  const connect = useCallback(() => {
    if (!enabled) return;

    const eventSource = new EventSource(url, { withCredentials: true });
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setConnected(true);
      retryCountRef.current = 0;
    };

    eventSource.onmessage = onMessage;

    eventSource.onerror = () => {
      setConnected(false);
      eventSource.close();
      eventSourceRef.current = null;

      if (retryCountRef.current < maxRetries) {
        const delay = Math.min(1000 * Math.pow(2, retryCountRef.current), 30000);
        retryCountRef.current += 1;
        setTimeout(connect, delay);
      }
    };
  }, [url, onMessage, enabled]);

  useEffect(() => {
    connect();
    return () => {
      eventSourceRef.current?.close();
      eventSourceRef.current = null;
    };
  }, [connect]);

  return { connected };
}
