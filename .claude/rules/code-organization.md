# Code Organization

## Project Layout

```
backend/internal/
├── modules/{name}/    # Feature modules (handler, service, repository)
├── db/queries/        # SQLC query definitions (.sql)
├── db/sqlc/           # Generated code (DO NOT EDIT)
├── db/migrations/     # PostgreSQL migrations (sequential 000001-000018+)
├── middleware/         # HTTP middleware (auth, CORS, rate limiting, tenant)
├── pkg/               # Shared utilities (apperror, pagination, respond, validator)
├── adapters/          # External service integrations (bkash, fcm, sms)
└── config/            # Configuration loading

website/src/           # Customer storefront
├── app/               # Next.js App Router pages
├── components/        # Reusable components
├── lib/               # API client, utilities
└── stores/            # Zustand stores

partner/src/           # Vendor portal (same structure as admin)
admin/src/             # Super-admin panel (same structure as partner)
├── app/(protected)/   # Auth-required pages
├── components/ui/     # shadcn/ui components
├── hooks/             # Custom React hooks
├── lib/               # API client (api-client.ts), utils
├── stores/            # Zustand stores
└── providers/         # Context providers
```

## Module Pattern (Backend)

Every backend module in `internal/modules/{name}/` follows:

```
{name}/
├── handler.go      # HTTP handlers (parse request → call service → respond)
├── service.go      # Business logic (receives parsed data, returns errors)
├── repository.go   # Data access wrapper (optional, for complex modules)
├── context.go      # Context helpers (WithUser, FromContext)
├── middleware.go    # Module-specific middleware (optional)
└── *_test.go       # Tests
```

**Layer discipline is mandatory:**
- Handlers NEVER call DB directly
- Services NEVER write HTTP responses
- Repositories NEVER contain business logic

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Go packages | lowercase, no underscores | `auth`, `order`, `catalog` |
| Go types | PascalCase | `Handler`, `Service`, `Repository` |
| Go constructors | `New` prefix | `NewHandler(svc *Service)` |
| Go files | lowercase, underscores ok | `handler.go`, `middleware_test.go` |
| TS/TSX files | kebab-case | `area-selector.tsx`, `api-client.ts` |
| React components | PascalCase | `AreaSelector`, `OrderKanban` |
| Hooks | `use` prefix, camelCase | `useSSE`, `useAuthStore` |
| Zustand stores | `use{Name}Store` | `useAuthStore` |
| JSON keys | snake_case | `tenant_id`, `created_at` |
| API routes | lowercase, slashes | `/api/v1/auth/otp/send` |
| DB migrations | `000XXX_description.up.sql` / `.down.sql` |

## Cross-System Changes

When changes span backend + frontend:
1. Check if an openspec proposal exists in `openspec/changes/`
2. If breaking: create a new proposal via `openspec/AGENTS.md`
3. Backend API changes come first, then frontend integration
4. Update SQLC queries → regenerate → update service → update handler → update frontend

---
Path scope: All subsystems
