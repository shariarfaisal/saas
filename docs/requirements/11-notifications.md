# 11 — Notifications & Real-time

## 11.1 Notification Channels

| Channel | Use Case | Provider |
|---------|----------|----------|
| Push (FCM) | Order updates, new orders for restaurants, rider assignments | Firebase Cloud Messaging |
| SMS | OTP verification, order confirmation | SSL Wireless (BD) / Twilio |
| Email | Invoice ready, registration, password reset | SendGrid / SES |
| In-app | Notification center in all portals | Stored in `notifications` table |
| SSE | Real-time order updates in portal dashboards | Server-Sent Events |
| WebSocket | Rider location streaming | WebSocket |

---

## 11.2 Push Notification Events

| Event | Recipient | Title | Body |
|-------|-----------|-------|------|
| `order.created` | Restaurant manager(s) | "New Order 🛵" | "Order #KBC-001 — ৳450" |
| `order.confirmed` | Customer | "Order Confirmed ✅" | "Your order is being prepared!" |
| `order.preparing` | Customer | "Order is Being Prepared 👨‍🍳" | "Estimated delivery: 30 mins" |
| `order.ready` | Rider | "Order Ready for Pickup 📦" | "Pick up from [Restaurant Name]" |
| `order.picked` | Customer | "Your Food is on the Way! 🛵" | "Rider [Name] is heading to you" |
| `order.delivered` | Customer | "Order Delivered 🎉" | "Enjoy your meal! Rate your experience" |
| `order.cancelled` | Customer + Restaurant | "Order Cancelled" | "Your order has been cancelled" |
| `order.rejected` | Customer | "Order Rejected" | "Sorry, [Restaurant] rejected your order" |
| `rider.assigned` | Rider | "New Delivery Assigned!" | "Pick up from [Restaurant], deliver to [Area]" |
| `invoice.ready` | Restaurant manager | "Invoice Ready 📄" | "Your [period] invoice is ready for review" |
| `issue.resolved` | Customer + Restaurant | "Issue Resolved" | "Your order issue has been resolved" |
| `refund.processed` | Customer | "Refund Processed 💸" | "৳[amount] refunded to your [payment method]" |

---

## 11.3 SSE Architecture

Server-Sent Events provide real-time updates to browser clients.

### Connection Endpoint
```
GET /api/v1/events/subscribe
    Authorization: Bearer <token>
    Accept: text/event-stream
```

### Event Format
```
event: order.created
data: {"order_id": "...", "order_number": "KBC-001", "total": "450.00", ...}

event: order.status_changed
data: {"order_id": "...", "old_status": "confirmed", "new_status": "preparing"}
```

### Routing SSE Events
```
Redis Pub/Sub channels:
  tenant:{tenant_id}:orders          → Broadcast to all partner dashboard connections for this tenant
  user:{user_id}:orders              → Broadcast to a specific user (customer order tracking)
  restaurant:{restaurant_id}:orders  → Broadcast to restaurant's SSE connections
```

**Flow:**
1. Go service generates an event (e.g., order status changed)
2. Service publishes to appropriate Redis channel
3. SSE handler for each connected client is subscribed to relevant channels
4. Handler receives Redis pub/sub message and writes to SSE stream

**Reconnection:** Clients send `Last-Event-ID` header on reconnect; server replays last 5 minutes of events from Redis cache.

---

## 11.4 Rider WebSocket

Real-time rider location streaming:

```
WS /api/v1/rider/ws
   Authorization: Bearer <rider_token>
```

**Rider → Server:**
```json
{"type": "location", "lat": 23.7945, "lng": 90.4051, "timestamp": "..."}
{"type": "status", "status": "in_hub"}
{"type": "picked", "order_id": "...", "restaurant_id": "..."}
```

**Server → Rider:**
```json
{"type": "order_assigned", "order": {...}}
{"type": "ping"}
```

**Admin/Partner → Rider Locations:**
Partner portal requests rider locations via polling:
```
GET /api/v1/partner/riders/tracking
→ Returns all active riders' last known location
```
Or SSE subscription for live tracking (Phase 2).

---

## 11.5 Notification Storage

All notifications are persisted in the `notifications` table for the in-app notification center:
- Unread count badge in UI
- Mark as read (single or all)
- Notifications paginated (newest first)
- Retained for 90 days then purged

---

## 11.6 SMS Templates (Bangladesh)

| Event | Template |
|-------|----------|
| OTP | "Your [BrandName] verification code is [OTP]. Valid for 2 minutes. Do not share this code." |
| Order Confirmed | "Your order [ORDER_NUMBER] has been confirmed! Estimated delivery: [ETA] mins. Track: [URL]" |
| Order Delivered | "Your order [ORDER_NUMBER] has been delivered. Enjoy! Rate your experience: [URL]" |

SMS provider must support Bangla Unicode for future localisation.

---

## 11.7 Email Templates

| Template | Trigger |
|----------|---------|
| `welcome` | New customer registration |
| `invoice_ready` | Monthly/daily invoice generated |
| `order_confirmation` | Order placed (online payment) |
| `refund_processed` | Refund completed |
| `password_reset` | Password reset request |
| `vendor_invitation` | New vendor team member invited |
| `tenant_suspended` | Tenant suspended by admin |

Email service must support HTML templates with dynamic variables.
