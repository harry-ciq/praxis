"use client";

import { useEffect, useRef } from "react";
import { socket } from "@/lib/socket";
import { useAuth } from "@/hooks/use-auth";

/**
 * Connects the global WebSocket when authenticated.
 * Call this once at a high level (e.g., MainLayout).
 * Components that need to listen for events can import `socket` directly
 * from "@/lib/socket" and use `socket.on(...)`.
 */
export function useSocket() {
  const { isAuthenticated, isLoading } = useAuth();
  const wasAuthenticated = useRef(false);

  useEffect(() => {
    // Don't act during the initial auth loading phase
    if (isLoading) return;

    if (isAuthenticated) {
      const token = localStorage.getItem("praxis_access_token");
      if (token) {
        socket.connect(token);
        wasAuthenticated.current = true;
      }
    } else if (wasAuthenticated.current) {
      // Only disconnect if the user was previously authenticated
      // (i.e., they logged out), not during initial load
      socket.disconnect();
      wasAuthenticated.current = false;
    }
  }, [isAuthenticated, isLoading]);

  return { socket };
}
