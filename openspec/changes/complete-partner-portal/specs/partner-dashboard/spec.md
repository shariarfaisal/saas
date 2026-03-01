## ADDED Requirements

### Requirement: Dashboard KPI Summary
The partner dashboard SHALL fetch live KPI data from `GET /partner/dashboard/summary` and display today's order count, today's revenue (BDT), pending order count, and average delivery time. Data SHALL refresh automatically every 60 seconds via TanStack Query's `staleTime` configuration.

#### Scenario: KPI cards load on dashboard
- **WHEN** a partner user navigates to the dashboard
- **THEN** the four KPI cards display real values fetched from the API
- **AND** a loading skeleton is shown while data is in flight

#### Scenario: KPI auto-refresh
- **WHEN** 60 seconds elapse since the last fetch
- **THEN** TanStack Query automatically refetches the summary and updates card values without a page reload

### Requirement: Live Incoming Order Panel via SSE
The dashboard SHALL maintain a Server-Sent Events connection to receive new order notifications in real time. The existing `use-sse.ts` hook SHALL be wired into the incoming-order panel component (`incoming-order-panel.tsx`), subscribing to the tenant+restaurant-scoped channel. A sound notification SHALL play on each new order event using the existing `use-audio-notification.ts` hook.

#### Scenario: New order arrives via SSE
- **WHEN** a customer places an order for the currently selected restaurant
- **THEN** the order card appears in the incoming-order panel within 2 seconds
- **AND** the audio notification plays once
- **AND** the pending order KPI count increments

#### Scenario: SSE reconnects after disconnect
- **WHEN** the SSE connection drops (network interruption)
- **THEN** the hook attempts reconnection with exponential backoff (max 30 s)
- **AND** a visual indicator shows "Reconnecting…" in the order panel

### Requirement: 7-Day Trend Charts from API
The dashboard trend charts (orders and revenue by day) SHALL be populated from `GET /partner/analytics/trends?days=7`. The existing Recharts `trend-chart.tsx` component SHALL accept real data props instead of hardcoded mock arrays.

#### Scenario: Trend chart displays real data
- **WHEN** the dashboard loads
- **THEN** the 7-day bar chart shows actual order and revenue figures per day from the API

### Requirement: Quick-Action Toggle Restaurant Availability
The dashboard quick-action "Toggle availability" button SHALL call `PATCH /partner/restaurants/:id/availability` and reflect the updated status in the button label and restaurant card without a full page reload.

#### Scenario: Partner toggles restaurant open/closed from dashboard
- **WHEN** the partner clicks "Toggle availability" on the dashboard
- **THEN** the API is called with the new availability value
- **AND** the button label changes to reflect the new state (Open / Closed)
- **AND** an error toast is shown if the API call fails
