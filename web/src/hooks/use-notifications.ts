"use client";

import { useEffect, useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { useSocket } from "@/hooks/use-socket";
import type { Notification } from "@/types";

export function useNotifications() {
  const { socket } = useSocket();
  const queryClient = useQueryClient();

  const { data: unreadCount = 0 } = useQuery({
    queryKey: ["notifications", "unread-count"],
    queryFn: () =>
      api.get<{ count: number }>("/notifications/unread-count").then((r) => r.count),
  });

  const {
    data: notifications = [],
    refetch,
  } = useQuery({
    queryKey: ["notifications"],
    queryFn: () => api.get<Notification[]>("/notifications"),
  });

  useEffect(() => {
    const unsubscribe = socket.on("notification", () => {
      queryClient.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    });

    return () => { unsubscribe(); };
  }, [socket, queryClient]);

  const markRead = useCallback(
    async (id: string) => {
      await api.patch(`/notifications/${id}/read`);
      queryClient.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
    [queryClient],
  );

  return { unreadCount, notifications, markRead, refetch };
}
