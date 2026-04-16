"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { UserSummary } from "@/types";

interface SearchUsersResponse {
  users: UserSummary[];
}

export function useSearchUsers(query: string) {
  return useQuery({
    queryKey: ["search-users", query],
    queryFn: () =>
      api.get<SearchUsersResponse>("/users/search", { q: query, limit: "10" }),
    enabled: query.length >= 2,
    staleTime: 30_000, // cache results for 30s
  });
}
