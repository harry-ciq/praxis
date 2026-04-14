"use client";

import { useEffect, useRef, useState } from "react";
import { socket } from "@/lib/socket";
import { useAuth } from "@/hooks/use-auth";

export function useSocket() {
  const { isAuthenticated } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const connectedRef = useRef(false);

  useEffect(() => {
    if (!isAuthenticated) return;

    const token = localStorage.getItem("praxis_access_token");
    if (!token) return;

    socket.connect(token);
    setIsConnected(true);
    connectedRef.current = true;

    return () => {
      socket.disconnect();
      setIsConnected(false);
      connectedRef.current = false;
    };
  }, [isAuthenticated]);

  return { socket, isConnected };
}
