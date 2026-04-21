"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Avatar,
  AvatarImage,
  AvatarFallback,
} from "@/components/ui/avatar";
import {
  Sparkles,
  RotateCw,
  AlertCircle,
  Rss,
  Star,
  ArrowRight,
} from "lucide-react";
import { formatRelativeTime } from "@/lib/format";
import { cn } from "@/lib/utils";
import Link from "next/link";

// --- Types ---------------------------------------------------------------

type GroupTheme =
  | "SHIPPING"
  | "CONTENT"
  | "MILESTONE"
  | "COMMUNITY"
  | "LEARNING"
  | "OTHER";

type Vibe = "MOMENTUM" | "STEADY" | "EXPLORING" | "QUIET" | "MIXED";

interface SmartFeedGroup {
  emoji: string;
  label: string;
  detail: string;
  theme: GroupTheme;
}

interface SmartFeedAction {
  label: string;
  cta: string;
  href: string; // "" when no navigation target
}

interface FeaturedPerson {
  username: string;
  name: string;
  avatarUrl: string;
  isMe: boolean;
  count: number;
}

interface SmartFeedDigest {
  vibe: Vibe;
  timeframe: string;
  headline: string;
  summary: string;
  highlight: string;
  groups: SmartFeedGroup[];
  suggestedAction: SmartFeedAction | null;
  featuredPeople: FeaturedPerson[];
  myShare: number;
  followedShare: number;
  activityByDay: number[];
  sourceCount: number;
  generatedAt: string;
  cached: boolean;
}

// --- Styling vocabulary --------------------------------------------------

const VIBE_STYLE: Record<
  Vibe,
  { badge: string; ring: string; glow: string; accent: string }
> = {
  MOMENTUM: {
    badge:
      "bg-amber-500/15 text-amber-300 ring-1 ring-inset ring-amber-500/30",
    ring: "border-amber-500/20",
    glow: "bg-amber-500/5",
    accent: "from-amber-500/20",
  },
  STEADY: {
    badge: "bg-sky-500/15 text-sky-300 ring-1 ring-inset ring-sky-500/30",
    ring: "border-sky-500/20",
    glow: "bg-sky-500/5",
    accent: "from-sky-500/20",
  },
  EXPLORING: {
    badge:
      "bg-violet-500/15 text-violet-300 ring-1 ring-inset ring-violet-500/30",
    ring: "border-violet-500/20",
    glow: "bg-violet-500/5",
    accent: "from-violet-500/20",
  },
  MIXED: {
    badge:
      "bg-emerald-500/15 text-emerald-300 ring-1 ring-inset ring-emerald-500/30",
    ring: "border-emerald-500/20",
    glow: "bg-emerald-500/5",
    accent: "from-emerald-500/20",
  },
  QUIET: {
    badge: "bg-zinc-700/50 text-zinc-300 ring-1 ring-inset ring-zinc-600",
    ring: "border-zinc-800",
    glow: "bg-zinc-700/5",
    accent: "from-zinc-700/20",
  },
};

const THEME_STYLE: Record<GroupTheme, string> = {
  SHIPPING:
    "border-emerald-500/20 bg-emerald-500/[0.04] hover:border-emerald-500/40",
  CONTENT: "border-rose-500/20 bg-rose-500/[0.04] hover:border-rose-500/40",
  MILESTONE:
    "border-amber-500/20 bg-amber-500/[0.04] hover:border-amber-500/40",
  COMMUNITY: "border-sky-500/20 bg-sky-500/[0.04] hover:border-sky-500/40",
  LEARNING:
    "border-violet-500/20 bg-violet-500/[0.04] hover:border-violet-500/40",
  OTHER: "border-zinc-800 bg-zinc-900/40 hover:border-zinc-700",
};

const THEME_LABEL_STYLE: Record<GroupTheme, string> = {
  SHIPPING: "text-emerald-400",
  CONTENT: "text-rose-400",
  MILESTONE: "text-amber-400",
  COMMUNITY: "text-sky-400",
  LEARNING: "text-violet-400",
  OTHER: "text-zinc-400",
};

function getInitials(name: string) {
  return name
    .split(" ")
    .map((s) => s[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase();
}

// --- Component -----------------------------------------------------------

export function SmartFeed() {
  const query = useQuery<SmartFeedDigest, { message?: string; statusCode?: number }>({
    queryKey: ["smart-feed"],
    queryFn: () => api.get<SmartFeedDigest>("/feed/smart"),
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
  const { data: digest, isPending, isFetching, refetch, error } = query;

  const regenerate = () =>
    api
      .get<SmartFeedDigest>("/feed/smart", { refresh: "true" })
      .then(() => refetch());

  const isInitialLoading = isPending && !error;

  // --- Loading -----------------------------------------------------------

  if (isInitialLoading) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-3 rounded-2xl border border-zinc-800 bg-zinc-900/50 p-5">
          <div className="flex items-center gap-2 text-sm text-zinc-400">
            <Sparkles className="size-4 animate-pulse text-amber-400" />
            Generating your digest…
          </div>
          <Skeleton className="h-7 w-4/5" />
          <Skeleton className="h-3 w-full" />
          <Skeleton className="h-3 w-5/6" />
          <Skeleton className="h-3 w-3/4" />
          <div className="mt-2 grid grid-cols-2 gap-2">
            <Skeleton className="h-14 w-full rounded-xl" />
            <Skeleton className="h-14 w-full rounded-xl" />
          </div>
        </div>
      </div>
    );
  }

  // --- Error -------------------------------------------------------------

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

  // --- Empty state -------------------------------------------------------

  if (digest.sourceCount === 0) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-2xl border border-zinc-800 bg-zinc-900/50 px-6 py-12 text-center">
        <div className="flex size-11 items-center justify-center rounded-full bg-zinc-800">
          <Rss className="size-5 text-zinc-500" />
        </div>
        <p className="max-w-sm text-sm text-zinc-400">{digest.headline}</p>
        <p className="max-w-sm text-xs text-zinc-600">{digest.summary}</p>
      </div>
    );
  }

  // --- Full digest -------------------------------------------------------

  const vibe = VIBE_STYLE[digest.vibe] ?? VIBE_STYLE.STEADY;
  const action = digest.suggestedAction;
  const maxActivity = Math.max(1, ...digest.activityByDay);
  const dayLabels = ["Su", "M", "Tu", "W", "Th", "F", "Sa"];
  // Compute weekday labels for activityByDay (oldest first, i.e. 6 days ago → today)
  const now = new Date();
  const activityLabels = digest.activityByDay.map((_, i) => {
    const d = new Date(now);
    d.setDate(now.getDate() - (6 - i));
    return dayLabels[d.getDay()];
  });

  return (
    <div className="flex flex-col gap-4">
      {/* Hero card */}
      <div
        className={cn(
          "relative overflow-hidden rounded-2xl border bg-zinc-900/60 p-5",
          vibe.ring
        )}
      >
        <div
          className={cn(
            "pointer-events-none absolute -right-20 -top-20 size-64 rounded-full blur-3xl",
            vibe.glow
          )}
        />
        {/* Subtle gradient wash behind the content */}
        <div
          className={cn(
            "pointer-events-none absolute inset-x-0 top-0 h-24 bg-gradient-to-b to-transparent",
            vibe.accent
          )}
        />

        {/* Meta row */}
        <div
          className="smart-feed-fade relative mb-3 flex items-center justify-between"
          style={{ ["--stagger" as string]: 0 }}
        >
          <div className="flex items-center gap-2">
            <span
              className={cn(
                "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider",
                vibe.badge
              )}
            >
              <Sparkles className="size-3" />
              {digest.vibe}
            </span>
            <span className="text-[11px] text-zinc-500">·</span>
            <span className="text-[11px] font-medium text-zinc-400">
              {digest.timeframe}
            </span>
          </div>
          <div className="flex items-center gap-2 text-[11px] text-zinc-500">
            <span>
              {digest.sourceCount} event{digest.sourceCount === 1 ? "" : "s"}
            </span>
            <span>·</span>
            <span>{formatRelativeTime(digest.generatedAt)}</span>
            <button
              onClick={regenerate}
              disabled={isFetching}
              className="ml-1 inline-flex size-6 items-center justify-center rounded-md text-zinc-500 transition-colors hover:bg-zinc-800 hover:text-zinc-200 disabled:opacity-50"
              title="Regenerate digest"
              aria-label="Regenerate digest"
            >
              <RotateCw
                className={cn("size-3", isFetching && "animate-spin")}
              />
            </button>
          </div>
        </div>

        {/* People row */}
        {digest.featuredPeople.length > 0 && (
          <div
            className="smart-feed-fade relative mb-4 flex items-center gap-3"
            style={{ ["--stagger" as string]: 1 }}
          >
            <div className="flex -space-x-2">
              {digest.featuredPeople.slice(0, 5).map((p) => (
                <Link
                  key={p.username}
                  href={`/profile/${p.username}`}
                  title={`${p.name} · ${p.count} event${p.count === 1 ? "" : "s"}`}
                  className="inline-block rounded-full ring-2 ring-zinc-900 transition-transform hover:z-10 hover:scale-110"
                >
                  <Avatar size="sm" className="size-7">
                    {p.avatarUrl ? (
                      <AvatarImage src={p.avatarUrl} alt={p.name} />
                    ) : null}
                    <AvatarFallback className="bg-zinc-800 text-[10px]">
                      {getInitials(p.name)}
                    </AvatarFallback>
                  </Avatar>
                </Link>
              ))}
            </div>
            <span className="text-[11px] text-zinc-500">
              {digest.featuredPeople.length === 1
                ? digest.featuredPeople[0].name
                : `${digest.featuredPeople[0].name} and ${digest.featuredPeople.length - 1} other${digest.featuredPeople.length - 1 === 1 ? "" : "s"}`}
            </span>
          </div>
        )}

        {/* Headline (serif, editorial) */}
        {digest.headline && (
          <h2
            className="smart-feed-fade relative mb-2 font-[family-name:var(--font-playfair)] text-[22px] font-semibold leading-tight tracking-tight text-zinc-50"
            style={{ ["--stagger" as string]: 2 }}
          >
            {digest.headline}
          </h2>
        )}

        {/* Summary */}
        <p
          className="smart-feed-fade relative text-[14px] leading-relaxed text-zinc-300"
          style={{ ["--stagger" as string]: 3 }}
        >
          {digest.summary}
        </p>

        {/* Highlight */}
        {digest.highlight && (
          <div
            className="smart-feed-fade relative mt-4 flex items-start gap-2.5 rounded-xl border border-amber-500/20 bg-amber-500/[0.04] px-3 py-2.5"
            style={{ ["--stagger" as string]: 4 }}
          >
            <Star className="mt-0.5 size-4 shrink-0 fill-amber-400 text-amber-400" />
            <div className="flex flex-col">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-amber-400">
                Highlight
              </span>
              <span className="text-[13px] leading-snug text-zinc-100">
                {digest.highlight}
              </span>
            </div>
          </div>
        )}

        {/* Your vs. followed split + 7-day activity strip */}
        {digest.sourceCount > 0 && (
          <div
            className="smart-feed-fade relative mt-4 flex items-center gap-4"
            style={{ ["--stagger" as string]: 5 }}
          >
            {/* Your vs. them */}
            <div className="flex-1">
              <div className="mb-1 flex items-center justify-between text-[10px]">
                <span className="font-medium text-zinc-400">
                  You {digest.myShare}%
                </span>
                <span className="text-zinc-500">
                  Following {digest.followedShare}%
                </span>
              </div>
              <div className="flex h-1.5 overflow-hidden rounded-full bg-zinc-800">
                <div
                  className="bg-blue-500/70 transition-all"
                  style={{ width: `${digest.myShare}%` }}
                />
                <div
                  className="bg-zinc-600/60 transition-all"
                  style={{ width: `${digest.followedShare}%` }}
                />
              </div>
            </div>

            {/* 7-day shape */}
            <div className="flex flex-col gap-1">
              <span className="text-[10px] font-medium text-zinc-500">
                7-day shape
              </span>
              <div className="flex items-end gap-0.5">
                {digest.activityByDay.map((count, i) => {
                  const intensity = count / maxActivity;
                  const bg =
                    count === 0
                      ? "bg-zinc-800"
                      : intensity > 0.66
                        ? "bg-emerald-500/80"
                        : intensity > 0.33
                          ? "bg-emerald-500/50"
                          : "bg-emerald-500/25";
                  return (
                    <div
                      key={i}
                      className={cn("size-3 rounded-sm", bg)}
                      title={`${activityLabels[i]}: ${count} event${count === 1 ? "" : "s"}`}
                    />
                  );
                })}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Theme chips */}
      {digest.groups.length > 0 && (
        <div className="grid gap-2 sm:grid-cols-2">
          {digest.groups.map((g, i) => (
            <div
              key={i}
              className={cn(
                "smart-feed-fade flex items-start gap-3 rounded-xl border px-3 py-2.5 transition-colors",
                THEME_STYLE[g.theme] ?? THEME_STYLE.OTHER
              )}
              style={{ ["--stagger" as string]: 6 + i }}
            >
              <span className="text-xl leading-none">{g.emoji}</span>
              <div className="flex min-w-0 flex-col">
                <span
                  className={cn(
                    "truncate text-[10px] font-semibold uppercase tracking-wider",
                    THEME_LABEL_STYLE[g.theme] ?? THEME_LABEL_STYLE.OTHER
                  )}
                >
                  {g.label}
                </span>
                <span className="text-[13px] leading-snug text-zinc-200">
                  {g.detail}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Suggested action */}
      {action && (
        <ActionRow
          action={action}
          stagger={6 + digest.groups.length}
          onRegenerate={regenerate}
        />
      )}
    </div>
  );
}

function ActionRow({
  action,
  stagger,
  onRegenerate,
}: {
  action: SmartFeedAction;
  stagger: number;
  onRegenerate: () => void;
}) {
  const content = (
    <>
      <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-amber-500/10 ring-1 ring-inset ring-amber-500/20">
        <Sparkles className="size-4 text-amber-400" />
      </div>
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
          Next step
        </span>
        <span className="truncate text-[13px] text-zinc-100">
          {action.label}
        </span>
      </div>
      <span className="inline-flex shrink-0 items-center gap-1 rounded-lg bg-zinc-800 px-3 py-1.5 text-[12px] font-medium text-zinc-100 transition-colors group-hover:bg-zinc-700">
        {action.cta || "Open"}
        <ArrowRight className="size-3.5" />
      </span>
    </>
  );

  const className = cn(
    "smart-feed-fade group flex items-center gap-3 rounded-xl border border-zinc-800 bg-gradient-to-r from-zinc-900/80 to-zinc-900/40 px-4 py-3 text-left transition-colors",
    "hover:border-zinc-700"
  );
  const style = { ["--stagger" as string]: stagger } as React.CSSProperties;

  if (action.href) {
    return (
      <Link href={action.href} className={className} style={style}>
        {content}
      </Link>
    );
  }
  return (
    <button onClick={onRegenerate} className={className} style={style}>
      {content}
    </button>
  );
}
