"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import Link from "next/link";
import { Search, Loader2, UserPlus, UserCheck } from "lucide-react";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Input } from "@/components/ui/input";
import { useSearchUsers } from "@/hooks/use-search";
import { useFollow } from "@/hooks/use-profile";
import { useAuth } from "@/hooks/use-auth";

export function SearchCommand() {
  const { user } = useAuth();
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const { follow, unfollow } = useFollow();
  const { data, isLoading } = useSearchUsers(debouncedQuery);
  const users = data?.users ?? [];

  // Debounce the query by 300ms
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query.trim());
    }, 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Open dropdown when there's a query
  useEffect(() => {
    if (debouncedQuery.length >= 2) {
      setOpen(true);
    } else {
      setOpen(false);
    }
  }, [debouncedQuery]);

  // Close on click outside
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Close on Escape
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") {
        setOpen(false);
        inputRef.current?.blur();
      }
    },
    []
  );

  const handleFollowToggle = (
    e: React.MouseEvent,
    username: string,
    isFollowing: boolean
  ) => {
    e.preventDefault();
    e.stopPropagation();
    if (isFollowing) {
      unfollow.mutate(username);
    } else {
      follow.mutate(username);
    }
  };

  return (
    <div ref={containerRef} className="relative hidden flex-1 max-w-md sm:block">
      <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
      <Input
        ref={inputRef}
        placeholder="Search users..."
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onFocus={() => {
          if (debouncedQuery.length >= 2) setOpen(true);
        }}
        onKeyDown={handleKeyDown}
        className="h-8 border-zinc-800 bg-zinc-900 pl-8 text-sm text-zinc-200 placeholder:text-zinc-500 focus:border-zinc-700"
      />

      {/* Results dropdown */}
      {open && (
        <div className="absolute inset-x-0 top-full z-50 mt-1 overflow-hidden rounded-lg border border-zinc-800 bg-zinc-900 shadow-xl shadow-black/40">
          {isLoading ? (
            <div className="flex items-center justify-center py-6">
              <Loader2 className="size-4 animate-spin text-zinc-500" />
            </div>
          ) : users.length === 0 ? (
            <div className="px-4 py-6 text-center text-sm text-zinc-500">
              No users found for &ldquo;{debouncedQuery}&rdquo;
            </div>
          ) : (
            <ul className="py-1">
              {users.map((u) => {
                const isOwnProfile = u.username === user?.username;
                return (
                  <li key={u.id}>
                    <Link
                      href={`/profile/${u.username}`}
                      onClick={() => {
                        setOpen(false);
                        setQuery("");
                      }}
                      className="flex items-center gap-3 px-3 py-2.5 transition-colors hover:bg-zinc-800/70"
                    >
                      <Avatar size="sm">
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
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium text-zinc-100">
                          {u.name}
                        </p>
                        <p className="truncate text-xs text-zinc-500">
                          @{u.username}
                          {u.headline ? ` · ${u.headline}` : ""}
                        </p>
                      </div>
                      {!isOwnProfile && (
                        <button
                          onClick={(e) =>
                            handleFollowToggle(e, u.username, u.isFollowing)
                          }
                          className={`flex shrink-0 items-center gap-1 rounded-md border px-2 py-1 text-xs font-medium transition-colors ${
                            u.isFollowing
                              ? "border-zinc-700 bg-zinc-800 text-zinc-300 hover:border-red-800 hover:bg-red-950/50 hover:text-red-400"
                              : "border-emerald-700/50 bg-emerald-950/30 text-emerald-400 hover:bg-emerald-950/50"
                          }`}
                        >
                          {u.isFollowing ? (
                            <>
                              <UserCheck className="size-3" />
                              Following
                            </>
                          ) : (
                            <>
                              <UserPlus className="size-3" />
                              Follow
                            </>
                          )}
                        </button>
                      )}
                    </Link>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
