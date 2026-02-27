# 14 — Gap Analysis & Remediation

## 14.1 Audit Scope

This audit reviewed the full requirements set (`01` to `13`) against world-class SaaS expectations for:
- Product completeness
- Multi-tenant safety
- Financial correctness
- API resilience
- Reliability/operations readiness
- Governance and release discipline

---

## 14.2 Executive Summary

### What is already strong
- Clear modular-monolith architecture (correctly avoids premature microservices)
- Strong baseline domain coverage for food commerce
- Detailed order lifecycle and portal-level feature mapping
- Practical Bangladesh-first payment and maps constraints
- Good initial PostgreSQL + SQLC schema direction

### High-impact gaps found
1. Idempotency and exactly-once guarantees were implied but not explicitly standardized.
2. Tenant isolation relied mainly on application filters without DB-level defense-in-depth.
3. Financial model lacked explicit immutable ledger standards.
4. Reliability objectives (SLO/RPO/RTO) and DR drills were not fully specified.
5. API evolution and webhook security contracts needed hard requirements.

---

## 14.3 Gap Matrix

| Area | Current State | Gap | Risk if Ignored | Priority |
|------|---------------|-----|-----------------|----------|
| Order creation reliability | Good flow defined | No hard idempotency contract | Duplicate orders/charges | P0 |
| Payment callbacks | Callback flows defined | No explicit dedupe model | Double-capture/refund inconsistencies | P0 |
| Tenant isolation | App-layer tenant filter | No RLS baseline | Cross-tenant data leak from query bug | P0 |
| Finance correctness | Invoice formulas defined | No immutable ledger standard | Reconciliation disputes, audit failures | P0 |
| Webhook trust | Events listed | No signature/replay protection | Spoofed events, fraud surface | P1 |
| API lifecycle | Versioning mentioned | No deprecation contract | Breaking clients in production | P1 |
| DR readiness | Backups mentioned | No RPO/RTO commitment | Unbounded downtime/data loss | P1 |
| Release governance | CI examples present | No gated release policy | Regressions in production | P1 |
| Fraud controls | Lightly covered | Missing abuse prevention controls | Promo abuse/COD fraud | P1 |
| Cross-vertical flexibility | Mentioned as goal | Missing explicit abstraction gates | Rework when expanding beyond food | P2 |

---

## 14.4 Remediation Plan (Prioritized)

## P0 — Must-have before production launch
- Enforce idempotency for checkout/payment/refunds.
- Implement callback dedupe by gateway transaction ID.
- Add DB-level tenant protection baseline (RLS or equivalent guard).
- Define immutable ledger or equivalent append-only finance event model.

## P1 — Must-have within early production window
- Signed webhook contract + replay protection.
- SLO/SLI targets and alert routing ownership.
- RPO/RTO commitments and restore drill cadence.
- API deprecation and backward-compatibility policy.
- Fraud-abuse controls for OTP/promo/COD risk.

## P2 — Scale optimization
- Advanced support tooling and operator runbooks.
- Cross-vertical policy engine maturity.
- Automated anomaly detection in analytics pipelines.

---

## 14.5 Non-Negotiable Launch Checklist

- [ ] No endpoint creates duplicate financial side effects on retry.
- [ ] Tenant boundary cannot be bypassed by missing WHERE clause bugs.
- [ ] Finance totals can be recomputed from immutable event history.
- [ ] Payment reconciliation dashboard has daily auto-check and alerts.
- [ ] Incident response runbook tested with real drill.
- [ ] Release rollback tested in staging with realistic load.

---

## 14.6 Open Decisions to Finalize

1. Ledger approach: full double-entry now vs phased adoption.
2. RLS rollout strategy: strict from day 1 vs phased with compatibility layer.
3. Payout cadence defaults: daily vs weekly for low-volume tenants.
4. Multi-region DR timing: immediate launch vs post-MVP milestone.

These decisions should be locked before implementation kickoff to avoid architectural churn.
