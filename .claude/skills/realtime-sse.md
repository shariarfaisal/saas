---
title: Real-time SSE Implementation
description: Guide for implementing Server-Sent Events for live updates
tags: [sse, realtime, websocket]
---

# Real-time Features (SSE & WebSocket)

## Architecture

- **SSE (Server-Sent Events):** Order status updates, new order notifications, partner dashboard
- **WebSocket:** Rider location tracking only
- **Redis Pub/Sub:** Event fanout across multiple API nodes

## Backend: Publishing Events

```go
// Publish to Redis when order status changes
func (s *Service) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID,
    newStatus sqlc.OrderStatus) error {
    // ... update DB ...

    // Publish event to Redis
    event := map[string]interface{}{
        "type":      "order_status_changed",
        "order_id":  orderID.String(),
        "status":    newStatus,
        "timestamp": time.Now().UTC(),
    }
    payload, _ := json.Marshal(event)
    s.redis.Publish(ctx, fmt.Sprintf("tenant:%s:orders", tenantID), payload)
    return nil
}
```

## Backend: SSE Endpoint

```go
func (h *Handler) StreamOrderUpdates(w http.ResponseWriter, r *http.Request) {
    t := tenant.FromContext(r.Context())
    flusher, ok := w.(http.Flusher)
    if !ok {
        respond.Error(w, apperror.Internal("streaming not supported", nil))
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    ch := fmt.Sprintf("tenant:%s:orders", t.ID)
    sub := h.redis.Subscribe(r.Context(), ch)
    defer sub.Close()

    for msg := range sub.Channel() {
        fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
        flusher.Flush()
    }
}
```

## Frontend: useSSE Hook

```tsx
"use client";

import { useEffect, useRef, useCallback, useState } from "react";

type SSEOptions = {
  url: string;
  onMessage: (event: MessageEvent) => void;
  enabled?: boolean;
};

export function useSSE({ url, onMessage, enabled = true }: SSEOptions) {
  const [connected, setConnected] = useState(false);
  const esRef = useRef<EventSource | null>(null);
  const retryRef = useRef(0);

  const connect = useCallback(() => {
    if (!enabled) return;
    const es = new EventSource(url, { withCredentials: true });
    esRef.current = es;

    es.onopen = () => { setConnected(true); retryRef.current = 0; };
    es.onmessage = onMessage;
    es.onerror = () => {
      setConnected(false);
      es.close();
      if (retryRef.current < 5) {
        const delay = Math.min(1000 * Math.pow(2, retryRef.current), 30000);
        retryRef.current += 1;
        setTimeout(connect, delay);
      }
    };
  }, [url, onMessage, enabled]);

  useEffect(() => {
    connect();
    return () => { esRef.current?.close(); };
  }, [connect]);

  return { connected };
}
```

## Usage in Partner Dashboard

```tsx
const { connected } = useSSE({
  url: `${API_BASE}/sse/orders`,
  onMessage: (event) => {
    const data = JSON.parse(event.data);
    if (data.type === "order_status_changed") {
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    }
    if (data.type === "new_order") {
      // Play notification sound, show toast
    }
  },
});
```

## Background Jobs (asynq)

Redis-backed job queue for async processing:
- `order:auto_confirm` — every 1 min
- `order:auto_cancel` — every 5 min
- `rider:auto_assign` — event-driven
- `invoice:generate` — daily
- `notification:send` — event-driven
