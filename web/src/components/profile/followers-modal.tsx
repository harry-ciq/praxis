"use client";

import { useEffect, useRef, useCallback } from "react";
import { Loader2, X } from "lucide-react";
import Link from "next/link";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { useFollowers, useFollowing, useFollow } from "@/hooks/use-profile";

interface FollowersModalProps {
  open: boolean;
  onClose: () => void;
  username: string;
  type: "followers" | "following";
}

export function FollowersModal({
  open,
  onClose,
  username,
  type,
}: FollowersModalProps) {
  const followersQuery = useFollowers(username, open && type === "followers");
  const followingQuery = useFollowing(username, open && type === "following");
  const { follow, unfollow } = useFollow();

  const query = type === "followers" ? followersQuery : followingQuery;
  const users = query.data?.pages.flatMap((p) => p.users) ?? [];

  const sentinelRef = useRef<HTMLDivElement>(null);

  const handleObserver = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;
      if (entry.isIntersecting && query.hasNextPage && !query.isFetchingNextPage) {
        query.fetchNextPage();
      }
    },
    [query]
  );

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !open) return;
    const observer = new IntersectionObserver(handleObserver, {
      threshold: 0.1,
    });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [open, handleObserver]);

  if (!open) return null;

  const handleFollowToggle = (userUsername: string, isFollowing: boolean) => {
    if (isFollowing) {
      unfollow.mutate(userUsername);
    } else {
      follow.mutate(userUsername);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60"
        onClick={onClose}
        aria-hidden
      />

      <div className="relative z-10 flex max-h-[70vh] w-full max-w-md flex-col rounded-2xl border border-zinc-800 bg-zinc-900 shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-zinc-800 px-5 py-4">
          <h2 className="text-lg font-semibold capitalize text-zinc-100">
            {type}
          </h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          >
            <X className="size-5" />
          </button>
        </div>

        {/* List */}
        <div className="flex-1 overflow-y-auto px-2 py-2">
          {query.isLoading ? (
            <div className="flex justify-center py-10">
              <Loader2 className="size-5 animate-spin text-zinc-500" />
            </div>
          ) : users.length === 0 ? (
            <div className="py-10 text-center text-sm text-zinc-500">
              No {type} yet.
            </div>
          ) : (
            <>
              {users.map((u) => (
                <div
                  key={u.id}
                  className="flex items-center gap-3 rounded-xl px-3 py-2.5 hover:bg-zinc-800/60"
                >
                  <Link href={`/profile/${u.username}`} onClick={onClose}>
                    <Avatar className="size-10">
                      {u.avatarUrl ? (
                        <AvatarImage src={u.avatarUrl} alt={u.name} />
                      ) : null}
                      <AvatarFallback className="text-xs">
                        {u.name
                          .split(" ")
                          .map((w) => w[0])
                          .join("")
                          .slice(0, 2)
                          .toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                  </Link>

                  <div className="min-w-0 flex-1">
                    <Link
                      href={`/profile/${u.username}`}
                      onClick={onClose}
                      className="block truncate text-sm font-medium text-zinc-100 hover:underline"
                    >
                      {u.name}
                    </Link>
                    <p className="truncate text-xs text-zinc-500">
                      @{u.username}
                      {u.headline ? ` \u00B7 ${u.headline}` : ""}
                    </p>
                  </div>

                  <Button
                    variant={u.isFollowing ? "outline" : "default"}
                    size="sm"
                    onClick={() =>
                      handleFollowToggle(u.username, u.isFollowing)
                    }
                    disabled={follow.isPending || unfollow.isPending}
                  >
                    {u.isFollowing ? "Unfollow" : "Follow"}
                  </Button>
                </div>
              ))}

              {/* Sentinel for infinite scroll */}
              <div ref={sentinelRef} className="h-4" />
              {query.isFetchingNextPage && (
                <div className="flex justify-center py-3">
                  <Loader2 className="size-4 animate-spin text-zinc-500" />
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
