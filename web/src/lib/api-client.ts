import type { ApiError } from "@/types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private getAccessToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("praxis_access_token");
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const parsed = await response.json().catch(() => null);
      const error: ApiError = {
        error:
          (parsed && typeof parsed === "object" && "error" in parsed
            ? (parsed as { error?: string }).error
            : undefined) ?? "unknown_error",
        message:
          (parsed && typeof parsed === "object" && "message" in parsed
            ? (parsed as { message?: string }).message
            : undefined) ?? response.statusText,
        statusCode: response.status,
      };
      throw error;
    }
    return response.json();
  }

  async get<T>(path: string, params?: Record<string, string>): Promise<T> {
    const url = new URL(`${this.baseUrl}${path}`);
    if (params) {
      Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));
    }

    const token = this.getAccessToken();
    const response = await fetch(url.toString(), {
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });

    return this.handleResponse<T>(response);
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    const token = this.getAccessToken();
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    return this.handleResponse<T>(response);
  }

  async patch<T>(path: string, body?: unknown): Promise<T> {
    const token = this.getAccessToken();
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    return this.handleResponse<T>(response);
  }

  async delete<T>(path: string): Promise<T> {
    const token = this.getAccessToken();
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "DELETE",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });

    return this.handleResponse<T>(response);
  }

  setTokens(accessToken: string, refreshToken: string) {
    localStorage.setItem("praxis_access_token", accessToken);
    localStorage.setItem("praxis_refresh_token", refreshToken);
  }

  clearTokens() {
    localStorage.removeItem("praxis_access_token");
    localStorage.removeItem("praxis_refresh_token");
  }
}

export const api = new ApiClient(API_BASE_URL);
