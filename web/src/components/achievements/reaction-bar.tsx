"use client";

import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import type { ReactionSummary, ReactionType } from "@/types";

interface ReactionBarProps {
  achievementId: string;
  reactions: ReactionSummary;
  userReaction: ReactionType | null;
}

const REACTION_CONFIG: {
  type: ReactionType;
  emoji: string;
  label: string;
  key: keyof ReactionSummary;
  activeClass: string;
}[] = [
  {
    type: "CLAP",
    emoji: "\uD83D\uDC4F",
    label: "Clap",
    key: "clap",
    activeClass: "bg-amber-500/10 text-amber-300 border-amber-500/30",
  },
  {
    type: "FIRE",
    emoji: "\uD83D\uDD25",
    label: "Fire",
    key: "fire",
    activeClass: "bg-orange-500/10 text-orange-300 border-orange-500/30",
  },
  {
    type: "ROCKET",
    emoji: "\uD83D\uDE80",
    label: "Rocket",
    key: "rocket",
    activeClass: "bg-blue-500/10 text-blue-300 border-blue-500/30",
  },
];

export function ReactionBar({
  achievementId,
  reactions,
  userReaction,
}: ReactionBarProps) {
  const queryClient = useQueryClient();
  const [optimisticReactions, setOptimisticReactions] =
    useState<ReactionSummary>(reactions);
  const [optimisticUserReaction, setOptimisticUserReaction] =
    useState<ReactionType | null>(userReaction);

  const addReaction = useMutation({
    mutationFn: (type: ReactionType) =>
      api.post(`/achievements/${achievementId}/reactions`, { type }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["achievements"] });
    },
  });

  const removeReaction = useMutation({
    mutationFn: () =>
      api.delete(`/achievements/${achievementId}/reactions`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["achievements"] });
    },
  });

  const handleReaction = useCallback(
    (type: ReactionType) => {
      const key = type.toLowerCase() as keyof ReactionSummary;

      if (optimisticUserReaction === type) {
        setOptimisticReactions((prev) => ({
          ...prev,
          [key]: Math.max(0, (prev[key] as number) - 1),
          total: Math.max(0, prev.total - 1),
        }));
        setOptimisticUserReaction(null);

        removeReaction.mutate(undefined, {
          onError: () => {
            setOptimisticReactions(reactions);
            setOptimisticUserReaction(userReaction);
          },
        });
      } else {
        const prevKey = optimisticUserReaction
          ? (optimisticUserReaction.toLowerCase() as keyof ReactionSummary)
          : null;

        setOptimisticReactions((prev) => {
          const next = { ...prev };
          if (prevKey && prevKey !== "total") {
            (next[prevKey] as number) = Math.max(
              0,
              (next[prevKey] as number) - 1
            );
            next.total = Math.max(0, next.total - 1);
          }
          (next[key] as number) = (next[key] as number) + 1;
          next.total = next.total + 1;
          return next;
        });
        setOptimisticUserReaction(type);

        addReaction.mutate(type, {
          onError: () => {
            setOptimisticReactions(reactions);
            setOptimisticUserReaction(userReaction);
          },
        });
      }
    },
    [
      optimisticUserReaction,
      reactions,
      userReaction,
      addReaction,
      removeReaction,
    ]
  );

  return (
    <div className="flex items-center gap-1.5">
      {REACTION_CONFIG.map(({ type, emoji, label, key, activeClass }) => {
        const count = optimisticReactions[key] as number;
        const isActive = optimisticUserReaction === type;

        return (
          <button
            key={type}
            type="button"
            onClick={() => handleReaction(type)}
            aria-label={`${label} (${count})`}
            className={cn(
              "inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs font-medium transition-all duration-200",
              isActive
                ? activeClass
                : "border-transparent bg-zinc-800/40 text-zinc-500 hover:bg-zinc-800/70 hover:text-zinc-400"
            )}
          >
            <span className={cn("text-sm", isActive && "scale-110")}>
              {emoji}
            </span>
            {count > 0 && (
              <span className="tabular-nums">{count}</span>
            )}
          </button>
        );
      })}
    </div>
  );
}
