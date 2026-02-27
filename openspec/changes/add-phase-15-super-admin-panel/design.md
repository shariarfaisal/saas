## Context
Phase 15 introduces a new frontend surface (`admin/`) while backend admin APIs already exist. We need to deliver broad page coverage quickly with consistent patterns for auth, API access, route protection, and mutation safety.

## Goals / Non-Goals
- Goals:
  - Provide complete admin information architecture and required workflows for TASK-071..079
  - Enforce secure session checks for all admin routes
  - Keep implementation modular and easy to extend against real backend payloads
- Non-Goals:
  - Re-implement backend admin APIs
  - Build production-perfect visual polish beyond required functionality
  - Add unrelated frontend apps (website/partner) in this change

## Decisions
- Decision: Use Next.js App Router route groups for authenticated shell and dedicated auth routes.
  - Rationale: Clear separation of protected vs public routes with middleware guard.
- Decision: Use Axios API client wrapper with request/response interceptors.
  - Rationale: Straightforward request ID injection and token refresh retry behavior.
- Decision: Use lightweight reusable UI blocks (cards/tables/forms/modals) and shadcn-style primitives.
  - Rationale: Minimizes implementation risk while satisfying broad page scope.
- Decision: Maintain server-authoritative session via cookies; mirror minimal client auth state in Zustand for UX only.
  - Rationale: Aligns with httpOnly-cookie requirement and secure route protection.

## Risks / Trade-offs
- Risk: Backend response shapes may differ from assumptions.
  - Mitigation: Centralized API normalizers and defensive optional rendering.
- Risk: Full feature breadth in one phase can create regression surface.
  - Mitigation: Keep pages feature-complete but structurally simple and test lint/build rigorously.

## Migration Plan
1. Scaffold admin app and shared providers.
2. Implement auth/session middleware and pages.
3. Add each Phase 15 page with required controls/actions.
4. Validate with lint/build and manual UI walkthrough.
5. Mark TASKS.md Phase 15 items complete.

## Open Questions
- Should QR setup be fetched from a dedicated `/admin/auth/2fa/setup` endpoint or included in login challenge payload?
- Is impersonation intended to use signed one-time URL token or backend-issued session exchange endpoint?
