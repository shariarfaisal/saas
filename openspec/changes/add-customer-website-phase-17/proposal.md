# Change: Add Phase 17 Customer Website (TASK-092 through TASK-101)

## Why
The platform currently lacks the customer-facing website application (`website/`) required for browsing restaurants, placing orders, tracking deliveries, and managing customer accounts. Phase 17 establishes the full customer experience while preserving tenant isolation and SEO requirements.

## What Changes
- Create `website/` Next.js 14 App Router project with SSR-first setup, tenant-aware server resolution, SEO baseline, and Cloudflare-friendly headers (TASK-092).
- Implement customer phone OTP authentication flow with first-time profile completion, cookie-backed session route handlers, and protected route support (TASK-093).
- Implement homepage discovery UX: banners, stories, cuisine filters, area selection, infinite restaurants list, and sort controls (TASK-094).
- Implement restaurant menu page with SSR, sticky category navigation, product search, and availability-aware product listing (TASK-095).
- Implement product detail experience with variant/addon selection, quantity stepper, and validated dynamic pricing (TASK-096).
- Implement persistent multi-restaurant cart and session sync behavior after login (TASK-097).
- Implement protected checkout with address management, promo validation, payment selection, and idempotent order placement (TASK-098).
- Implement bKash and AamarPay redirect handling with success/fail/cancel recovery pages and pending-state safety (TASK-099).
- Implement real-time order tracking via SSE with timeline progression, rider card, cancellation gates, and ETA countdown (TASK-100).
- Implement account pages for profile, addresses, order history, wallet, favourites, and notifications (TASK-101).

## Impact
- Affected specs: customer-website
- Affected code: `website/**`, `openspec/changes/add-customer-website-phase-17/**`
- Integration dependencies: backend storefront/auth/orders/payments/me APIs under `backend/internal/modules/**`
- Risks: scope size is large; phased implementation will prioritize foundational SSR and tenant-safe architecture before feature depth
