---
title: Frontend Page Implementation
description: Step-by-step guide for adding new pages to partner/admin/website portals
tags: [frontend, nextjs, react]
---

# Implementing a New Frontend Page

## Step-by-step Workflow

### 1. Create Page File

```
src/app/(protected)/{feature}/page.tsx    # Protected page
src/app/auth/{feature}/page.tsx           # Public auth page
```

### 2. Page Skeleton

```tsx
"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";

export default function FeaturePage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["feature-items"],
    queryFn: () => apiClient.get("/feature-items").then((r) => r.data),
  });

  if (isLoading) return <PageSkeleton />;
  if (error) return <ErrorState message={error.message} />;

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Feature Title</h1>
        <Button onClick={handleCreate}>Add New</Button>
      </div>
      {/* Content */}
    </div>
  );
}
```

### 3. Form Page (Create/Edit)

```tsx
"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { apiClient } from "@/lib/api-client";

const schema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  description: z.string().min(10),
  isActive: z.boolean().default(true),
});

type FormValues = z.infer<typeof schema>;

export default function CreateFeaturePage() {
  const [saving, setSaving] = useState(false);
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (data: FormValues) => {
    setSaving(true);
    try {
      await apiClient.post("/feature-items", data);
      // Redirect or show success
    } catch (err) {
      // Show error toast
    } finally {
      setSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="max-w-2xl space-y-6 p-6">
      <div>
        <label className="text-sm font-medium">Name</label>
        <input {...register("name")} className="mt-1 w-full rounded-md border px-3 py-2" />
        {errors.name && <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>}
      </div>
      <button type="submit" disabled={saving}>
        {saving ? "Saving..." : "Create"}
      </button>
    </form>
  );
}
```

### 4. Add to Navigation

Partner/admin: update sidebar in `components/sidebar.tsx` or layout.
Website: update header/nav component.

### 5. Protect Route

Ensure the page is inside `(protected)/` route group, which is guarded by Next.js middleware.

## Key Patterns

- **Data fetching:** TanStack Query (`useQuery` for reads, `useMutation` for writes)
- **Forms:** React Hook Form + Zod (always)
- **Styling:** Tailwind utility classes + `cn()` helper
- **API calls:** `apiClient` from `@/lib/api-client` (never raw fetch)
- **State:** Zustand for client state, React Query for server state
- **Currency:** Use `formatCurrency()` from `@/lib/utils`
- **Dates:** Use `formatDate()` / `formatDateTime()` / `timeAgo()`
