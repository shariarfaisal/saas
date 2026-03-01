---
title: Spec Agent
description: Creates and reviews OpenSpec change proposals for architectural decisions
tags: [planning, specification, architecture]
capabilities:
  - Create well-formed OpenSpec proposals
  - Review impact on existing modules
  - Identify dependencies between proposals
  - Reference requirements documentation
---

# Spec Agent

## When to Invoke

- New feature that spans multiple modules
- Breaking API changes
- Database schema changes affecting existing tables
- Architecture shifts (e.g., adding a new service boundary)
- Performance or security overhauls

## Proposal Structure

All proposals live in `openspec/changes/{proposal-name}/`:

```
{proposal-name}/
├── design.md       # What & why, architectural decisions
├── specs/          # Detailed spec changes (ADDED/MODIFIED/REMOVED)
└── tasks.md        # Implementation checklist with TASK-XXX numbers
```

## Active Proposals to Be Aware Of

| Proposal | Status | Blocks |
|----------|--------|--------|
| `redesign-order-system` | Pending | Finance system |
| `complete-partner-portal` | Pending | Partner portal launch |
| `add-customer-website-phase-17` | 1/10 done | Customer launch |
| `add-auth-and-tenancy` | Not started | Everything |
| `inventory-promos-orders` | Not started | Order flow |
| `phase-9-14-finance-notifications-analytics` | Proposal only | Business ops |

## Critical Path

```
auth-and-tenancy → inventory-promos-orders → redesign-order-system → finance
                                           → complete-partner-portal
                                           → customer-website
```

## Before Creating a New Proposal

1. Read `openspec/AGENTS.md` for format guidelines
2. Check if an existing proposal covers the scope
3. Read relevant `docs/requirements/` documents
4. Identify which existing modules are affected
5. Check the critical path — does this block or depend on other work?

## Reference
- `openspec/AGENTS.md` — proposal format and workflow
- `openspec/project.md` — project context and conventions
