import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import { AuthProvider, useAuth } from "@/hooks/use-auth";
import { api } from "@/lib/api-client";
import type { ReactNode } from "react";

vi.mock("@/lib/api-client", () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    setTokens: vi.fn(),
    clearTokens: vi.fn(),
  },
}));

const mockGet = api.get as Mock;
const mockPost = api.post as Mock;

function wrapper({ children }: { children: ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}

// Stub window.location
const originalLocation = window.location;

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();

  Object.defineProperty(window, "location", {
    writable: true,
    value: { ...originalLocation, href: "http://localhost:3000", assign: vi.fn() },
  });
});

describe("useAuth", () => {
  it("throws when used outside AuthProvider", () => {
    expect(() => {
      renderHook(() => useAuth());
    }).toThrow("useAuth must be used within an AuthProvider");
  });

  it("resolves to unauthenticated when no token is stored", async () => {
    const { result } = renderHook(() => useAuth(), { wrapper });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBeNull();
  });

  it("validates existing token on mount and sets user", async () => {
    localStorage.setItem("praxis_access_token", "valid-token");
    const mockUser = {
      id: "1",
      username: "testuser",
      name: "Test User",
      email: "test@example.com",
      avatarUrl: null,
      headline: null,
    };
    mockGet.mockResolvedValueOnce(mockUser);

    const { result } = renderHook(() => useAuth(), { wrapper });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.user).toEqual(mockUser);
  });

  it("login redirects to GitHub OAuth URL", async () => {
    mockGet.mockResolvedValueOnce({ url: "https://github.com/login/oauth/authorize?client_id=abc" });

    const { result } = renderHook(() => useAuth(), { wrapper });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    await act(async () => {
      await result.current.login();
    });

    expect(mockGet).toHaveBeenCalledWith("/auth/github");
    expect(window.location.href).toBe(
      "https://github.com/login/oauth/authorize?client_id=abc",
    );
  });

  it("handleCallback exchanges code for tokens and sets user", async () => {
    const callbackResponse = {
      accessToken: "access-123",
      refreshToken: "refresh-123",
      user: {
        id: "1",
        username: "testuser",
        name: "Test User",
        email: "test@example.com",
        avatarUrl: null,
        headline: null,
      },
    };
    mockPost.mockResolvedValueOnce(callbackResponse);

    const { result } = renderHook(() => useAuth(), { wrapper });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    await act(async () => {
      await result.current.handleCallback("test-code");
    });

    expect(mockPost).toHaveBeenCalledWith("/auth/github/callback", {
      code: "test-code",
    });
    expect(api.setTokens).toHaveBeenCalledWith("access-123", "refresh-123");
    expect(result.current.user).toEqual(callbackResponse.user);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it("logout clears state and redirects", async () => {
    localStorage.setItem("praxis_access_token", "valid-token");
    localStorage.setItem("praxis_refresh_token", "refresh-token");

    const mockUser = {
      id: "1",
      username: "testuser",
      name: "Test User",
      email: "test@example.com",
      avatarUrl: null,
      headline: null,
    };
    mockGet.mockResolvedValueOnce(mockUser);
    mockPost.mockResolvedValueOnce({});

    const { result } = renderHook(() => useAuth(), { wrapper });

    await waitFor(() => {
      expect(result.current.isAuthenticated).toBe(true);
    });

    await act(async () => {
      await result.current.logout();
    });

    expect(mockPost).toHaveBeenCalledWith("/auth/logout", {
      refreshToken: "refresh-token",
    });
    expect(api.clearTokens).toHaveBeenCalled();
    expect(window.location.href).toBe("/");
  });

  it("falls back to unauthenticated when /auth/me and refresh both fail", async () => {
    localStorage.setItem("praxis_access_token", "expired-token");
    localStorage.setItem("praxis_refresh_token", "bad-refresh");

    mockGet.mockRejectedValueOnce(new Error("Unauthorized"));
    mockPost.mockRejectedValueOnce(new Error("Refresh failed"));

    const { result } = renderHook(() => useAuth(), { wrapper });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBeNull();
  });
});
