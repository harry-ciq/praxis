"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Sparkles, RotateCw, AlertCircle, Rss } from "lucide-react";
import { formatRelativeTime } from "@/lib/format";

interface SmartFeedGroup {
  emoji: string;
  label: string;
  detail: string;
}

interface SmartFeedDigest {
  summary: string;
  groups: SmartFeedGroup[];
  sourceCount: number;
  generatedAt: string;
  cached: boolean;
}

export function SmartFeed() {
  const query = useQuery<SmartFeedDigest, { message?: string; statusCode?: number }>({
    queryKey: ["smart-feed"],
    queryFn: () => api.get<SmartFeedDigest>("/feed/smart"),
    staleTime: 5 * 60 * 1000, // 5 min — server caches 15m so this keeps us from hammering
    retry: false, // 503s mean "disabled in this env", no point retrying
  });
  const { data: digest, isPending, isFetching, refetch, error } = query;
  const isInitialLoading = isPending && !error;

  const regenerate = () =>
    api.get<SmartFeedDigest>("/feed/smart", { refresh: "true" }).then(() => refetch());

  if (isInitialLoading) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-3 rounded-2xl border border-zinc-800 bg-zinc-900/50 p-5">
          <div className="flex items-center gap-2 text-sm text-zinc-500">
            <Sparkles className="size-4 animate-pulse text-amber-400" />
            Generating your smart digest…
          </div>
          <Skeleton className="h-3 w-full" />
          <Skeleton className="h-3 w-5/6" />
          <Skeleton className="h-3 w-3/4" />
          <div className="mt-2 flex gap-2">
            <Skeleton className="h-6 w-24 rounded-full" />
            <Skeleton className="h-6 w-28 rounded-full" />
            <Skeleton className="h-6 w-20 rounded-full" />
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    const notConfigured =
      (error as { statusCode?: number }).statusCode === 503;
    return (
      <div className="flex flex-col items-center gap-3 rounded-2xl border border-zinc-800 bg-zinc-900/50 px-6 py-12 text-center">
        <div className="flex size-11 items-center justify-center rounded-full bg-amber-500/10">
          <AlertCircle className="size-5 text-amber-400" />
        </div>
        <h3 className="text-base font-semibold text-zinc-100">
          {notConfigured
            ? "Smart feed is not configured"
            : "Couldn't generate a digest"}
        </h3>
        <p className="max-w-sm text-sm text-zinc-500">
          {notConfigured
            ? "The server is missing an Anthropic API key. Smart feed is disabled for now — use the Raw feed tab instead."
            : "Something went wrong calling the AI model. Try again in a moment."}
        </p>
        {!notConfigured && (
          <Button
            size="sm"
            variant="outline"
            className="mt-1 gap-1.5 border-zinc-700 text-zinc-300"
            onClick={() => refetch()}
          >
            <RotateCw className="size-3.5" />
            Retry
          </Button>
        )}
      </div>
    );
  }

  if (!digest) return null;

  if (digest.sourceCount === 0) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-2xl border border-zinc-800 bg-zinc-900/50 px-6 py-12 text-center">
        <div className="flex size-11 items-center justify-center rounded-full bg-zinc-800">
          <Rss className="size-5 text-zinc-500" />
        </div>
        <p className="text-sm text-zinc-400">{digest.summary}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Summary card */}
      <div className="relative overflow-hidden rounded-2xl border border-zinc-800 bg-gradient-to-br from-zinc-900 to-zinc-900/40 p-5">
        <div className="pointer-events-none absolute -right-8 -top-8 size-40 rounded-full bg-amber-500/5 blur-3xl" />
        <div className="mb-3 flex items-center justify-between">
          <div className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wide text-amber-400">
            <Sparkles className="size-3.5" />
            AI Digest
          </div>
          <div className="flex items-center gap-2 text-[11px] text-zinc-500">
            <span>
              {digest.sourceCount} event{digest.sourceCount === 1 ? "" : "s"} ·{" "}
              {formatRelativeTime(digest.generatedAt)}
            </span>
            <button
              onClick={regenerate}
              disabled={isFetching}
              className="inline-flex size-6 items-center justify-center rounded-md text-zinc-500 transition-colors hover:bg-zinc-800 hover:text-zinc-200 disabled:opacity-50"
              title="Regenerate digest"
            >
              <RotateCw className={`size-3 ${isFetching ? "animate-spin" : ""}`} />
            </button>
          </div>
        </div>
        <p className="text-[15px] leading-relaxed text-zinc-100">
          {digest.summary}
        </p>
      </div>

      {/* Theme groups */}
      {digest.groups.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {digest.groups.map((g, i) => (
            <div
              key={i}
              className="flex items-center gap-2 rounded-xl border border-zinc-800 bg-zinc-900/80 px-3 py-2"
            >
              <span className="text-lg leading-none">{g.emoji}</span>
              <div className="flex flex-col">
                <span className="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">
                  {g.label}
                </span>
                <span className="text-sm text-zinc-200">{g.detail}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
