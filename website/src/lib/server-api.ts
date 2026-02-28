import { headers } from "next/headers";

const INTERNAL_API_BASE = process.env.INTERNAL_API_BASE_URL || "http://localhost:8080/api/v1";

/**
 * Helper to fetch data from the backend APIs during SSR.
 * It automatically forwards the Host header so the backend can resolve the tenant.
 */
export async function fetchServerApi<T>(path: string, options: RequestInit = {}): Promise<T> {
  const requestHeaders = await headers();
  const host = requestHeaders.get("x-forwarded-host") || requestHeaders.get("host") || "localhost";

  const res = await fetch(`${INTERNAL_API_BASE}${path}`, {
    ...options,
    headers: {
      ...options.headers,
      "Content-Type": "application/json",
      Host: host,
    },
    // Prevent Next.js aggressive caching for tenant-specific data unless specified
    cache: options.cache || "no-store", 
  });

  if (!res.ok) {
    let errMessage = "Failed to fetch from API";
    try {
      const errBody = await res.json();
      errMessage = errBody?.error?.message || errBody?.message || errMessage;
    } catch {
      // Ignore parse error
    }
    throw new Error(`API Error (${res.status}): ${errMessage}`);
  }

  return res.json() as Promise<T>;
}
