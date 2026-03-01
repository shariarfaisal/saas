---
title: Requirements & Domain Reference
description: Quick reference for all Munchies requirements and domain knowledge
tags: [reference, requirements, domain]
---

# Requirements & Domain Reference

## Requirements Index

| Doc | Content | Read When |
|-----|---------|-----------|
| 01 | Vision & goals | Understanding business context |
| 02 | User stories | Implementing user-facing features |
| 03 | Multi-tenancy | Any tenant-scoped work |
| 04 | Domain model | Understanding entities & relationships |
| 05 | Feature requirements | Implementing any feature |
| 06 | Portals & roles | Auth, permissions, RBAC |
| 07 | Order lifecycle | Order flows, state machine |
| 08 | Pricing & financials | Commission, payments, invoices |
| 09 | Database schema | Any DB/migration work |
| 10 | API design | Any endpoint work |
| 11 | Notifications | Push, SMS, email, SSE |
| 12 | Analytics | Reports, dashboards, metrics |
| 13 | Infrastructure | Deployment, monitoring, scaling |

All docs at: `docs/requirements/{number}-{name}.md`

## Domain Entities (Key Relationships)

```
Platform
  └── Tenant (vendor business)
        ├── Restaurant(s)
        │     ├── Categories → Products (menu items)
        │     ├── Operating hours
        │     └── Inventory stocks
        ├── Orders
        │     ├── Order items (JSONB snapshot at order time)
        │     ├── Payment transactions
        │     └── Delivery assignment → Rider
        ├── Promotions (scoped to tenant)
        ├── Riders (tenant-owned fleet)
        ├── Invoices & payouts (settlement)
        └── Users (customers, staff, managers)
```

## Order State Machine

```
PENDING → CONFIRMED → PREPARING → READY → PICKED → DELIVERED
                                                  → CANCELLED (from any pre-DELIVERED state)
```

Critical: State transitions must be validated — no arbitrary jumps. See `redesign-order-system` proposal.

## Payment Methods

- bKash (mobile money — primary)
- AamarPay (gateway)
- SSLCommerz (gateway)
- Cash on Delivery (COD)
- Wallet (platform credit)

## Commission Model

Per-restaurant commission rate stored in DB. Applied per order, distributed proportionally across items for multi-restaurant orders. See `docs/requirements/08-pricing-and-financials.md`.

## Before Implementing Any Feature

1. Read the relevant requirement doc(s)
2. Check `openspec/changes/` for existing proposals
3. Verify multi-tenancy implications
4. Consider order lifecycle impact (if order-related)
5. Check pricing/financial impact (if money-related)
