"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { useAuth } from "@/hooks/use-auth";
import type { FeedResponse } from "@/types";

export function useFeed() {
  const { user } = useAuth();

  return useInfiniteQuery({
    queryKey: ["feed", user?.id],
    queryFn: ({ pageParam = 0 }) =>
      api.get<FeedResponse>("/feed", {
        limit: "20",
        offset: String(pageParam),
      }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, pages) =>
      lastPage.nextCursor ? pages.length * 20 : undefined,
    enabled: !!user,
  });
}
