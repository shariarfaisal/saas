---
title: Code Reviewer
description: Reviews code changes for correctness, security, and adherence to project patterns
tags: [review, quality, security]
capabilities:
  - Review Go backend code for pattern compliance
  - Review Next.js frontend code for best practices
  - Check security vulnerabilities
  - Verify multi-tenancy compliance
---

# Code Reviewer Agent

## Backend Review Checklist

### Layer Discipline
- [ ] Handlers only parse HTTP + call service + respond
- [ ] Services contain business logic, never touch HTTP
- [ ] No direct DB calls from handlers (must go through service/repository)

### Error Handling
- [ ] All errors converted via `toAppError()` or typed `apperror.*`
- [ ] `pgx.ErrNoRows` handled as `apperror.NotFound()`
- [ ] Internal errors wrapped with `apperror.Internal()` (hides details)
- [ ] No panic in production code

### Multi-Tenancy
- [ ] `tenant.FromContext()` extracted and nil-checked
- [ ] `tenant_id` passed to every DB query
- [ ] No cross-tenant data access possible

### Security
- [ ] Auth middleware on protected routes
- [ ] Role middleware matches required permission
- [ ] Input validated before processing
- [ ] No hardcoded secrets or credentials

## Frontend Review Checklist

### Components
- [ ] `"use client"` directive on client components
- [ ] Props typed explicitly (no `any`)
- [ ] `cn()` used for conditional Tailwind classes
- [ ] `@/` path alias for imports

### Data Fetching
- [ ] TanStack Query for server state (not `useEffect` + `useState`)
- [ ] `apiClient` from `@/lib/api-client` (not raw fetch/axios)
- [ ] Loading and error states handled

### Forms
- [ ] Zod schema for validation
- [ ] `zodResolver` with React Hook Form
- [ ] Error messages displayed to user

### Security
- [ ] No sensitive data in localStorage
- [ ] No API keys in client code
- [ ] Zod validation on all user inputs
