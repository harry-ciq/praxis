"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Loader2, Rss, RefreshCw, Check } from "lucide-react";
import { useFeed } from "@/hooks/use-feed";
import { AchievementCard } from "@/components/achievements/achievement-card";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api-client";

function FeedSkeleton() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 3 }).map((_, i) => (
        <div
          key={i}
          className="animate-pulse rounded-xl border border-zinc-800 bg-zinc-900/80 p-5"
        >
          <div className="mb-3 flex items-center gap-2.5">
            <div className="size-8 rounded-full bg-zinc-800" />
            <div className="space-y-1">
              <div className="h-3.5 w-24 rounded bg-zinc-800" />
              <div className="h-3 w-16 rounded bg-zinc-800" />
            </div>
          </div>
          <div className="mb-2 h-4 w-20 rounded-full bg-zinc-800" />
          <div className="mb-1 h-5 w-3/4 rounded bg-zinc-800" />
          <div className="mb-3 h-4 w-1/2 rounded bg-zinc-800" />
          <div className="flex gap-1">
            <div className="h-7 w-14 rounded-full bg-zinc-800" />
            <div className="h-7 w-14 rounded-full bg-zinc-800" />
            <div className="h-7 w-14 rounded-full bg-zinc-800" />
          </div>
        </div>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <div className="mb-4 rounded-full bg-zinc-800 p-4">
        <Rss className="size-8 text-zinc-500" />
      </div>
      <h3 className="mb-2 text-lg font-semibold text-zinc-200">
        Your feed is empty
      </h3>
      <p className="max-w-sm text-sm text-zinc-500">
        Follow other builders to see their achievements!
      </p>
    </div>
  );
}

export default function FeedPage() {
  const {
    data,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
    refetch,
  } = useFeed();

  const [syncing, setSyncing] = useState(false);
  const [syncResult, setSyncResult] = useState<string | null>(null);

  const handleSync = async () => {
    setSyncing(true);
    setSyncResult(null);
    try {
      const result = await api.post<{ newAchievements: number }>("/achievements/sync", {
        provider: "GITHUB",
      });
      const count = result.newAchievements;
      setSyncResult(
        count > 0
          ? `Found ${count} new achievement${count > 1 ? "s" : ""}!`
          : "Already up to date"
      );
      if (count > 0) {
        refetch();
      }
    } catch {
      setSyncResult("Sync failed — try again");
    } finally {
      setSyncing(false);
      setTimeout(() => setSyncResult(null), 4000);
    }
  };

  const observerRef = useRef<IntersectionObserver | null>(null);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const sentinelCallback = useCallback(
    (node: HTMLDivElement | null) => {
      if (observerRef.current) observerRef.current.disconnect();
      if (!node) return;

      sentinelRef.current = node;
      observerRef.current = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
            fetchNextPage();
          }
        },
        { threshold: 0.1 }
      );
      observerRef.current.observe(node);
    },
    [hasNextPage, isFetchingNextPage, fetchNextPage]
  );

  useEffect(() => {
    return () => {
      if (observerRef.current) observerRef.current.disconnect();
    };
  }, []);

  const achievements =
    data?.pages.flatMap((page) => page.achievements) ?? [];

  return (
    <div className="mx-auto max-w-2xl px-4 py-6">
      {/* Header */}
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">Feed</h1>
          <p className="text-sm text-zinc-500">What builders are shipping</p>
        </div>
        <div className="flex flex-col items-end gap-1">
          <Button
            size="sm"
            variant="outline"
            className="gap-1.5 border-zinc-700 text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100"
            onClick={handleSync}
            disabled={syncing}
          >
            {syncing ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <RefreshCw className="size-3.5" />
            )}
            {syncing ? "Syncing..." : "Sync GitHub"}
          </Button>
          {syncResult && (
            <span className="flex items-center gap-1 text-xs text-zinc-400">
              <Check className="size-3 text-green-400" />
              {syncResult}
            </span>
          )}
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <FeedSkeleton />
      ) : achievements.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="space-y-4">
          {achievements.map((achievement) => (
            <AchievementCard
              key={achievement.id}
              achievement={achievement}
            />
          ))}

          {/* Infinite scroll sentinel */}
          <div ref={sentinelCallback} className="h-1" />

          {isFetchingNextPage && (
            <div className="flex justify-center py-4">
              <Loader2 className="size-5 animate-spin text-zinc-500" />
            </div>
          )}

          {hasNextPage && !isFetchingNextPage && (
            <div className="flex justify-center py-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => fetchNextPage()}
              >
                Load more
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
