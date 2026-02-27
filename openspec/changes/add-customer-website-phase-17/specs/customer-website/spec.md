## ADDED Requirements
### Requirement: SSR-First Tenant-Aware Website Foundation
The system SHALL provide a Next.js App Router customer website that resolves tenant context from request host/subdomain on the server and serializes only tenant-safe configuration into client components.

#### Scenario: Tenant resolved for SSR page render
- **WHEN** a request is served for `[tenant-slug].platform.com`
- **THEN** the server resolves the tenant slug from the Host header and exposes tenant-safe config to the rendered page.

### Requirement: SEO Baseline for Customer Website
The system SHALL support SEO metadata and structured data output for customer-discovery pages.

#### Scenario: Homepage response includes metadata and JSON-LD
- **WHEN** a crawler requests a tenant homepage
- **THEN** the response includes Open Graph-capable metadata and baseline JSON-LD structured data.

### Requirement: Customer Purchase Journey
The system SHALL provide customer flows for OTP auth, discovery, restaurant browsing, cart, checkout, payment redirects, tracking, and account management.

#### Scenario: Authenticated customer places an order
- **WHEN** a customer authenticates with OTP and completes checkout
- **THEN** the system submits an idempotent order intent and routes to payment/tracking flows according to payment method.
