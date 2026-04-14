"use client";

import {
  createContext,
  useContext,
  useCallback,
  useEffect,
  useState,
  useRef,
  type ReactNode,
} from "react";
import { createElement } from "react";
import { api } from "@/lib/api-client";
import type { AuthUser, AuthTokens } from "@/types";

interface AuthContextValue {
  user: AuthUser | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: () => Promise<void>;
  handleCallback: (code: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const TOKEN_REFRESH_INTERVAL = 4 * 60 * 1000; // 4 minutes

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const refreshIntervalRef = useRef<ReturnType<typeof setInterval> | null>(
    null,
  );

  const clearRefreshInterval = useCallback(() => {
    if (refreshIntervalRef.current) {
      clearInterval(refreshIntervalRef.current);
      refreshIntervalRef.current = null;
    }
  }, []);

  const refreshToken = useCallback(async () => {
    const storedRefreshToken = localStorage.getItem("praxis_refresh_token");
    if (!storedRefreshToken) return false;

    try {
      const tokens = await api.post<AuthTokens>("/auth/refresh", {
        refreshToken: storedRefreshToken,
      });
      api.setTokens(tokens.accessToken, tokens.refreshToken);
      return true;
    } catch {
      api.clearTokens();
      setUser(null);
      clearRefreshInterval();
      return false;
    }
  }, [clearRefreshInterval]);

  const startRefreshInterval = useCallback(() => {
    clearRefreshInterval();
    refreshIntervalRef.current = setInterval(() => {
      refreshToken();
    }, TOKEN_REFRESH_INTERVAL);
  }, [clearRefreshInterval, refreshToken]);

  useEffect(() => {
    async function init() {
      const accessToken = localStorage.getItem("praxis_access_token");
      if (!accessToken) {
        setIsLoading(false);
        return;
      }

      try {
        const currentUser = await api.get<AuthUser>("/auth/me");
        setUser(currentUser);
        startRefreshInterval();
      } catch {
        const refreshed = await refreshToken();
        if (refreshed) {
          try {
            const currentUser = await api.get<AuthUser>("/auth/me");
            setUser(currentUser);
            startRefreshInterval();
          } catch {
            api.clearTokens();
          }
        }
      } finally {
        setIsLoading(false);
      }
    }

    init();

    return () => {
      clearRefreshInterval();
    };
  }, [clearRefreshInterval, refreshToken, startRefreshInterval]);

  const login = useCallback(async () => {
    const data = await api.get<{ url: string }>("/auth/github");
    window.location.href = data.url;
  }, []);

  const handleCallback = useCallback(
    async (code: string) => {
      const data = await api.post<{
        tokens: AuthTokens;
        user: AuthUser;
      }>("/auth/github/callback", { code });
      api.setTokens(data.tokens.accessToken, data.tokens.refreshToken);
      setUser(data.user);
      startRefreshInterval();

      // Auto-sync GitHub achievements after login
      try {
        await api.post("/achievements/sync", { provider: "GITHUB" });
      } catch {
        // Non-blocking — sync failure shouldn't prevent login
      }
    },
    [startRefreshInterval],
  );

  const logout = useCallback(async () => {
    const storedRefreshToken = localStorage.getItem("praxis_refresh_token");
    try {
      if (storedRefreshToken) {
        await api.post("/auth/logout", { refreshToken: storedRefreshToken });
      }
    } catch {
      // Proceed with local cleanup even if API call fails
    } finally {
      api.clearTokens();
      setUser(null);
      clearRefreshInterval();
      window.location.href = "/";
    }
  }, [clearRefreshInterval]);

  const value: AuthContextValue = {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    handleCallback,
    logout,
  };

  return createElement(AuthContext.Provider, { value }, children);
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
