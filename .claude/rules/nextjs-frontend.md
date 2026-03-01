# Next.js Frontend Patterns

## App Router Structure

All portals use Next.js App Router. Pages go in `src/app/`.

```
src/app/
├── layout.tsx              # Root layout with providers
├── (protected)/            # Auth-required route group
│   ├── layout.tsx          # Sidebar + header
│   ├── dashboard/page.tsx
│   ├── orders/page.tsx
│   └── restaurants/[id]/page.tsx
├── auth/
│   ├── login/page.tsx
│   └── forgot-password/page.tsx
└── api/                    # API route handlers (if needed)
```

## Component Pattern

```tsx
"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";

type Props = {
  items: Item[];
  onSelect: (id: string) => void;
};

export function ItemList({ items, onSelect }: Props) {
  const [active, setActive] = useState<string | null>(null);

  return (
    <div className={cn("flex flex-col gap-2")}>
      {items.map((item) => (
        <button
          key={item.id}
          className={cn(
            "rounded-md p-3 text-sm",
            active === item.id ? "bg-slate-900 text-white" : "bg-slate-100"
          )}
          onClick={() => { setActive(item.id); onSelect(item.id); }}
        >
          {item.name}
        </button>
      ))}
    </div>
  );
}
```

**Rules:**
- Mark client components with `"use client"` at top
- Use `cn()` (clsx + tailwind-merge) for conditional classes
- Type props explicitly — no `any`
- Use `@/` path alias for imports

## API Client (Axios)

All portals use an Axios instance in `lib/api-client.ts`:
- `withCredentials: true` for cookie-based auth
- Auto-adds `X-Request-ID` header
- Auto-refreshes on 401 (single-flight pattern)
- Timeout: 15 seconds

```tsx
import { apiClient } from "@/lib/api-client";

const { data } = await apiClient.get(`/restaurants/${id}`);
await apiClient.post("/orders", orderPayload);
```

## State Management

**Server state** — TanStack React Query:
```tsx
const { data, isLoading } = useQuery({
  queryKey: ["restaurants", tenantId],
  queryFn: () => apiClient.get("/restaurants").then(r => r.data),
});
```

**Client state** — Zustand:
```tsx
export const useAuthStore = create<AuthState>((set) => ({
  email: null,
  activeRestaurantId: null,
  setSession: (email, expiresAt) => set({ email, expiresAt }),
  clearSession: () => set({ email: null, expiresAt: null }),
}));
```

## Form Handling (React Hook Form + Zod)

```tsx
const schema = z.object({
  name: z.string().min(2),
  email: z.string().email().optional().or(z.literal("")),
});

type FormValues = z.infer<typeof schema>;

const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({
  resolver: zodResolver(schema),
});
```

**Rules:**
- Every form has a Zod schema
- Use `zodResolver` — no manual validation
- Type form values with `z.infer<typeof schema>`

## Styling

- Tailwind CSS utility classes only — no CSS-in-JS, no CSS modules
- Partner/admin use shadcn/ui components in `components/ui/`
- Website has custom components
- Dark mode via Tailwind config
- Currency format: `৳` (BDT) — use `formatCurrency()` from `lib/utils.ts`

## Middleware (Auth Guard)

```tsx
// src/middleware.ts
export function middleware(request: NextRequest) {
  if (!request.cookies.has("partner_access_token")) {
    return NextResponse.redirect(new URL("/auth/login", request.url));
  }
  return NextResponse.next();
}
```

## Utility Functions (lib/utils.ts)

Available helpers: `cn()`, `formatCurrency()`, `formatDate()`, `formatDateTime()`, `timeAgo()`

## Formatting

- Prettier: semicolons, double quotes, trailing commas, 120 print width
- ESLint: `next/core-web-vitals` + `next/typescript`

---
Path scope: website/, partner/, admin/
