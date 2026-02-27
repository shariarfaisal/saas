import axios, { AxiosError, AxiosHeaders } from "axios";

type ApiError = {
  message: string;
  status: number;
};

const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

let refreshPromise: Promise<void> | null = null;

export const apiClient = axios.create({
  baseURL: apiBaseUrl,
  withCredentials: true,
  timeout: 15000,
});

apiClient.interceptors.request.use((config) => {
  if (!config.headers) {
    config.headers = new AxiosHeaders();
  }
  config.headers.set("X-Request-ID", crypto.randomUUID());
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<{ message?: string }>) => {
    const status = error.response?.status ?? 500;
    const originalRequest = error.config;

    if (status === 401 && originalRequest && !originalRequest.headers?.["x-refresh-retry"]) {
      refreshPromise ??= apiClient
        .post("/auth/refresh")
        .then(() => undefined)
        .finally(() => {
          refreshPromise = null;
        });

      try {
        await refreshPromise;
        const headers = AxiosHeaders.from(originalRequest.headers);
        headers.set("x-refresh-retry", "1");
        originalRequest.headers = headers;
        return apiClient.request(originalRequest);
      } catch {
        return Promise.reject<ApiError>({ message: "Session expired. Please login again.", status: 401 });
      }
    }

    return Promise.reject<ApiError>({
      message: error.response?.data?.message ?? error.message ?? "Unexpected API error",
      status,
    });
  },
);
