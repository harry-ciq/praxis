"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { UserProfile, Achievement, PaginatedResponse } from "@/types";

export function useProfile(username: string) {
  return useQuery({
    queryKey: ["profile", username],
    queryFn: () => api.get<UserProfile>(`/users/${username}`),
    enabled: !!username,
  });
}

export function useProfileAchievements(username: string, page = 0) {
  return useQuery({
    queryKey: ["achievements", username, page],
    queryFn: () =>
      api.get<PaginatedResponse<Achievement>>(
        `/users/${username}/achievements`,
        { limit: "20", offset: String(page * 20) }
      ),
    enabled: !!username,
  });
}

export function useFollow() {
  const queryClient = useQueryClient();

  const follow = useMutation({
    mutationFn: (username: string) => api.post(`/users/${username}/follow`),
    onSuccess: (_data, username) => {
      queryClient.invalidateQueries({ queryKey: ["profile", username] });
    },
  });

  const unfollow = useMutation({
    mutationFn: (username: string) => api.delete(`/users/${username}/follow`),
    onSuccess: (_data, username) => {
      queryClient.invalidateQueries({ queryKey: ["profile", username] });
    },
  });

  return { follow, unfollow };
}
