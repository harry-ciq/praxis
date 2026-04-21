"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
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
  ArrowRight,
  ChevronDown,
  ChevronUp,
  Eye,
  ExternalLink,
  Target,
} from "lucide-react";
import { formatRelativeTime } from "@/lib/format";
import { cn } from "@/lib/utils";
import Link from "next/link";
import type { ReactNode } from "react";

// --- Types ---------------------------------------------------------------

type GroupTheme =
  | "SHIPPING"
  | "CONTENT"
  | "MILESTONE"
  | "COMMUNITY"
  | "LEARNING"
  | "OTHER";

type Vibe = "MOMENTUM" | "STEADY" | "EXPLORING" | "QUIET" | "MIXED";

interface AchievementStub {
  id: string;
  title: string;
  type: string;
  proofUrl: string;
  person: string;
  username: string;
}

interface SmartFeedGroup {
  emoji: string;
  label: string;
  detail: string;
  theme: GroupTheme;
  achievements: AchievementStub[];
}

interface Milestone {
  emoji: string;
  label: string;
  current: number;
  target: number;
}

interface WatchItem {
  emoji: string;
  prediction: string;
  targetUsername: string;
}

interface SmartFeedAction {
  label: string;
  cta: string;
  href: string;
}

interface FeaturedPerson {
  username: string;
  name: string;
  avatarUrl: string;
  isMe: boolean;
  count: number;
}

interface HeroStat {
  value: string;
  label: string;
}

interface SmartFeedDigest {
  vibe: Vibe;
  heroStat: HeroStat | null;
  timeframe: string;
  headline: string;
  summary: string;
  highlight: string;
  highlightAchievement: AchievementStub | null;
  groups: SmartFeedGroup[];
  milestones: Milestone[];
  watching: WatchItem[];
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

// --- Helpers -------------------------------------------------------------

function getInitials(name: string) {
  return name
    .split(" ")
    .map((s) => s[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase();
}

// ISO week-of-year number. Used as the "issue number" of the digest.
function getISOWeek(d: Date) {
  const date = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()));
  const dayNum = date.getUTCDay() || 7;
  date.setUTCDate(date.getUTCDate() + 4 - dayNum);
  const yearStart = new Date(Date.UTC(date.getUTCFullYear(), 0, 1));
  return Math.ceil(((+date - +yearStart) / 86400000 + 1) / 7);
}

// Count up an integer from 0 to `target` over `duration` ms (ease-out cubic).
// Uses setInterval (not RAF) so it keeps ticking even if the tab is
// backgrounded or the embedding iframe throttles rAF. 60 ticks / sec is
// plenty for a 1.1s animation.
function useCountUp(target: number, duration = 1100) {
  const [value, setValue] = useState(0);
  useEffect(() => {
    if (target <= 0) {
      setValue(0);
      return;
    }
    const stepMs = 16;
    const start = Date.now();
    const id = setInterval(() => {
      const elapsed = Date.now() - start;
      const progress = Math.min(1, elapsed / duration);
      const eased = 1 - Math.pow(1 - progress, 3);
      setValue(Math.round(eased * target));
      if (progress >= 1) clearInterval(id);
    }, stepMs);
    return () => clearInterval(id);
  }, [target, duration]);
  return value;
}

// Render a hero stat value like "17" with count-up, "1.2K" verbatim, "50+" too.
function HeroStatValue({ raw }: { raw: string }) {
  // Extract a leading integer if present (e.g. "17", "17 commits", "17+")
  const m = raw.match(/^(\d+)([^\d]*)$/);
  const target = m ? parseInt(m[1], 10) : 0;
  const suffix = m ? m[2] : "";
  const animated = useCountUp(target);
  if (m && target > 0) {
    return (
      <>
        <span className="tabular-nums">{animated}</span>
        {suffix && <span>{suffix}</span>}
      </>
    );
  }
  // Fallback for non-integer values like "1.2K"
  return <span>{raw}</span>;
}

// Split text on occurrences of any `name`-like strings from featuredPeople,
// yielding a mix of plain text and avatar chips. Matches whole words only and
// supports possessive forms like "Harry's".
function renderProseWithAvatars(
  text: string,
  people: FeaturedPerson[]
): ReactNode[] {
  if (!text || people.length === 0) return [text];

  // Build unique lowercase name tokens, longest first so "harry-ciq" is
  // matched before "harry".
  const tokens = new Map<string, FeaturedPerson>();
  for (const p of people) {
    if (p.username) tokens.set(p.username.toLowerCase(), p);
    if (p.name) {
      tokens.set(p.name.toLowerCase(), p);
      const first = p.name.split(" ")[0];
      if (first) tokens.set(first.toLowerCase(), p);
    }
  }
  const sortedKeys = Array.from(tokens.keys()).sort(
    (a, b) => b.length - a.length
  );
  if (sortedKeys.length === 0) return [text];

  // Escape regex metachars and allow an optional possessive ('s)
  const esc = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const pattern = new RegExp(
    `\\b(@?(?:${sortedKeys.map(esc).join("|")}))(['’]s)?\\b`,
    "gi"
  );

  const out: ReactNode[] = [];
  let lastIdx = 0;
  let key = 0;
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(text)) !== null) {
    const start = match.index;
    const raw = match[0];
    const nameToken = match[1].replace(/^@/, "").toLowerCase();
    const possessive = match[2] ?? "";
    const person = tokens.get(nameToken);
    if (!person) continue;

    if (start > lastIdx) out.push(text.slice(lastIdx, start));
    out.push(
      <InlineAvatarChip
        key={`chip-${key++}`}
        person={person}
        label={raw.replace(/['’]s$/, "")}
      />
    );
    if (possessive) out.push(possessive);
    lastIdx = start + raw.length;
  }
  if (lastIdx < text.length) out.push(text.slice(lastIdx));
  return out;
}

function InlineAvatarChip({
  person,
  label,
}: {
  person: FeaturedPerson;
  label: string;
}) {
  return (
    <Link
      href={`/profile/${person.username}`}
      className="inline-flex items-center gap-1 rounded-full bg-zinc-800/70 px-1.5 py-0.5 align-middle text-[12.5px] font-medium text-zinc-100 ring-1 ring-inset ring-zinc-700/50 transition-colors hover:bg-zinc-700/70 hover:ring-zinc-600"
    >
      <Avatar size="sm" className="size-4">
        {person.avatarUrl ? (
          <AvatarImage src={person.avatarUrl} alt={person.name} />
        ) : null}
        <AvatarFallback className="bg-zinc-700 text-[8px]">
          {getInitials(person.name)}
        </AvatarFallback>
      </Avatar>
      {label}
    </Link>
  );
}

// SVG sparkline for the activityByDay array. Amber stroke + faint area fill.
// Annotates today and the peak day.
function Sparkline({ data }: { data: number[] }) {
  const width = 120;
  const height = 40;
  const padding = 3;
  const max = Math.max(1, ...data);
  const stepX =
    data.length > 1 ? (width - 2 * padding) / (data.length - 1) : 0;
  const yFor = (v: number) =>
    height - padding - (v / max) * (height - 2 * padding);

  const points = data.map((v, i) => [padding + i * stepX, yFor(v)] as const);
  const linePath = points
    .map(([x, y], i) => (i === 0 ? `M${x},${y}` : `L${x},${y}`))
    .join(" ");
  const areaPath = `${linePath} L${points[points.length - 1][0]},${
    height - padding
  } L${points[0][0]},${height - padding} Z`;

  const peakIdx = data.reduce(
    (acc, v, i) => (v > data[acc] ? i : acc),
    0
  );
  const todayIdx = data.length - 1;

  return (
    <svg
      viewBox={`0 0 ${width} ${height}`}
      width={width}
      height={height}
      className="overflow-visible"
      aria-label="7-day activity"
    >
      <defs>
        <linearGradient id="sf-spark" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="rgb(251 191 36)" stopOpacity="0.35" />
          <stop offset="100%" stopColor="rgb(251 191 36)" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={areaPath} fill="url(#sf-spark)" />
      <path
        d={linePath}
        fill="none"
        stroke="rgb(251 191 36)"
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {/* Peak dot */}
      {data[peakIdx] > 0 && (
        <circle
          cx={points[peakIdx][0]}
          cy={points[peakIdx][1]}
          r="2.5"
          fill="rgb(251 191 36)"
        />
      )}
      {/* Today dot */}
      <circle
        cx={points[todayIdx][0]}
        cy={points[todayIdx][1]}
        r="2.2"
        fill="rgb(24 24 27)"
        stroke="rgb(251 191 36)"
        strokeWidth="1.3"
      />
    </svg>
  );
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
  const isRegenerating = isFetching && !isPending;

  // --- Loading -----------------------------------------------------------

  if (isInitialLoading) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-3 rounded-2xl border border-zinc-800 bg-zinc-900/50 p-5">
          <div className="flex items-center gap-2 text-sm text-zinc-400">
            <Sparkles className="size-4 animate-pulse text-amber-400" />
            Generating your digest…
          </div>
          <Skeleton className="h-16 w-32" />
          <Skeleton className="h-7 w-4/5" />
          <Skeleton className="h-3 w-full" />
          <Skeleton className="h-3 w-5/6" />
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

  return <FullDigest digest={digest} onRegenerate={regenerate} isRegenerating={isRegenerating} />;
}

function FullDigest({
  digest,
  onRegenerate,
  isRegenerating,
}: {
  digest: SmartFeedDigest;
  onRegenerate: () => void;
  isRegenerating: boolean;
}) {
  const router = useRouter();
  const vibe = VIBE_STYLE[digest.vibe] ?? VIBE_STYLE.STEADY;
  const action = digest.suggestedAction;

  // Jump to the raw-feed tab and scroll the matching card into view.
  const jumpToAchievement = (id: string) => {
    const el = document.getElementById(`achievement-${id}`);
    if (el) {
      el.scrollIntoView({ behavior: "smooth", block: "center" });
      el.classList.add("smart-feed-flash");
      setTimeout(() => el.classList.remove("smart-feed-flash"), 1400);
      return;
    }
    // Card isn't in the DOM — user is on Smart tab. Flip to Raw and deep link.
    router.push(`/feed?tab=raw&focus=${encodeURIComponent(id)}`);
  };

  const generatedDate = useMemo(() => new Date(digest.generatedAt), [
    digest.generatedAt,
  ]);
  const issueNumber = getISOWeek(generatedDate);
  const issueDateLabel = generatedDate.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  // First letter drop cap — split off the first character, then render the
  // rest of the summary with inline avatar chips for any featured names.
  const firstChar = digest.summary[0] ?? "";
  const summaryRest = digest.summary.slice(1);
  const restNodes = useMemo(
    () => renderProseWithAvatars(summaryRest, digest.featuredPeople),
    [summaryRest, digest.featuredPeople]
  );

  return (
    <div className="flex flex-col gap-4">
      {/* Hero card */}
      <div
        className={cn(
          "relative overflow-hidden rounded-2xl border bg-zinc-900/60 p-6",
          vibe.ring
        )}
      >
        {isRegenerating && <div className="smart-feed-shimmer" />}

        <div
          className={cn(
            "pointer-events-none absolute -right-24 -top-24 size-64 rounded-full blur-3xl",
            vibe.glow
          )}
        />
        <div
          className={cn(
            "pointer-events-none absolute inset-x-0 top-0 h-28 bg-gradient-to-b to-transparent",
            vibe.accent
          )}
        />

        {/* Masthead — issue number + timeframe + regenerate */}
        <div
          className="smart-feed-fade relative mb-5 flex items-center justify-between"
          style={{ ["--stagger" as string]: 0 }}
        >
          <div className="flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.18em] text-zinc-500">
            <span>Issue {String(issueNumber).padStart(2, "0")}</span>
            <span className="text-zinc-700">·</span>
            <span>{issueDateLabel}</span>
            <span className="text-zinc-700">·</span>
            <span className="text-zinc-400">{digest.timeframe}</span>
          </div>
          <div className="flex items-center gap-2 text-[11px] text-zinc-500">
            <span
              className={cn(
                "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider",
                vibe.badge
              )}
            >
              <Sparkles className="size-3" />
              {digest.vibe}
            </span>
            <button
              onClick={onRegenerate}
              disabled={isRegenerating}
              className="ml-1 inline-flex size-7 items-center justify-center rounded-md text-zinc-500 transition-colors hover:bg-zinc-800 hover:text-zinc-200 disabled:opacity-50"
              title="Regenerate digest"
              aria-label="Regenerate digest"
            >
              <RotateCw
                className={cn("size-3.5", isRegenerating && "animate-spin")}
              />
            </button>
          </div>
        </div>

        {/* Hero stat + headline — two-column on desktop */}
        <div
          className="smart-feed-fade relative mb-4 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between"
          style={{ ["--stagger" as string]: 1 }}
        >
          {digest.heroStat && (
            <div className="flex flex-col">
              <span className="font-[family-name:var(--font-playfair)] text-[56px] font-semibold leading-none text-zinc-50 sm:text-[64px]">
                <HeroStatValue raw={digest.heroStat.value} />
              </span>
              <span className="mt-1 text-[11px] font-semibold uppercase tracking-widest text-zinc-500">
                {digest.heroStat.label}
              </span>
            </div>
          )}
          {digest.headline && (
            <h2
              className={cn(
                "font-[family-name:var(--font-playfair)] text-[20px] font-semibold leading-snug tracking-tight text-zinc-50 sm:text-[22px]",
                digest.heroStat ? "sm:max-w-[60%] sm:text-right" : ""
              )}
            >
              {digest.headline}
            </h2>
          )}
        </div>

        {/* People row */}
        {digest.featuredPeople.length > 0 && (
          <div
            className="smart-feed-fade relative mb-4 flex items-center gap-3"
            style={{ ["--stagger" as string]: 2 }}
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

        {/* Summary with drop cap + inline avatar chips */}
        <p
          className="smart-feed-fade relative text-[14px] leading-relaxed text-zinc-300"
          style={{ ["--stagger" as string]: 3 }}
        >
          <span className="float-left mr-2 mt-1 font-[family-name:var(--font-playfair)] text-[42px] font-semibold leading-[0.85] text-zinc-100">
            {firstChar}
          </span>
          {restNodes}
        </p>

        {/* Highlight as a pull-quote */}
        {digest.highlight && (
          <div
            className="smart-feed-fade relative mt-6 flex flex-col items-center px-4 text-center"
            style={{ ["--stagger" as string]: 4 }}
          >
            <span
              className="mb-1 font-[family-name:var(--font-playfair)] text-[48px] leading-none text-amber-400/60"
              aria-hidden
            >
              &ldquo;
            </span>
            <p className="font-[family-name:var(--font-playfair)] text-[17px] italic leading-snug text-zinc-100 sm:text-[18px]">
              {digest.highlight}
            </p>
            <div className="mt-3 h-px w-12 bg-amber-500/30" />
            <span className="mt-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-amber-400/80">
              Highlight of the week
            </span>
            {digest.highlightAchievement && (
              <button
                onClick={() =>
                  jumpToAchievement(digest.highlightAchievement!.id)
                }
                className="mt-3 inline-flex items-center gap-1 text-[12px] font-medium text-amber-300/80 transition-colors hover:text-amber-200"
              >
                See the achievement
                <ArrowRight className="size-3" />
              </button>
            )}
          </div>
        )}

        {/* Metrics row — split bar + sparkline */}
        <div
          className="smart-feed-fade relative mt-6 flex items-center gap-5 border-t border-zinc-800/50 pt-4"
          style={{ ["--stagger" as string]: 5 }}
        >
          <div className="flex-1">
            <div className="mb-1 flex items-center justify-between text-[10px]">
              <span className="font-semibold uppercase tracking-wider text-zinc-400">
                You {digest.myShare}%
              </span>
              <span className="uppercase tracking-wider text-zinc-500">
                Following {digest.followedShare}%
              </span>
            </div>
            <div className="flex h-1.5 overflow-hidden rounded-full bg-zinc-800">
              <div
                className="bg-blue-500/80 transition-all"
                style={{ width: `${digest.myShare}%` }}
              />
              <div
                className="bg-zinc-600/60 transition-all"
                style={{ width: `${digest.followedShare}%` }}
              />
            </div>
          </div>
          <div className="flex flex-col items-end">
            <span className="mb-0.5 text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
              7-day shape
            </span>
            <Sparkline data={digest.activityByDay} />
          </div>
        </div>
      </div>

      {/* Milestones — near-miss progress bars */}
      {digest.milestones.length > 0 && (
        <section className="smart-feed-fade flex flex-col gap-3 rounded-2xl border border-zinc-800 bg-zinc-900/50 p-4"
          style={{ ["--stagger" as string]: 5.5 }}
        >
          <div className="flex items-center gap-2 text-[10px] font-semibold uppercase tracking-wider text-zinc-400">
            <Target className="size-3 text-amber-400" />
            Close to done
          </div>
          <div className="flex flex-col gap-3">
            {digest.milestones.map((m, i) => {
              const pct = Math.max(
                0,
                Math.min(100, Math.round((m.current / m.target) * 100))
              );
              return (
                <div key={i} className="flex flex-col gap-1.5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span className="text-base leading-none">{m.emoji}</span>
                      <span className="text-[13px] text-zinc-200">
                        {m.label}
                      </span>
                    </div>
                    <span className="shrink-0 text-[11px] font-medium tabular-nums text-zinc-400">
                      {m.current} / {m.target}
                    </span>
                  </div>
                  <div className="relative h-1.5 overflow-hidden rounded-full bg-zinc-800">
                    <div
                      className="h-full bg-gradient-to-r from-amber-500/80 to-amber-400 transition-[width]"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </section>
      )}

      {/* Watching — forward-looking */}
      {digest.watching.length > 0 && (
        <section
          className="smart-feed-fade flex flex-col gap-2"
          style={{ ["--stagger" as string]: 5.75 }}
        >
          <div className="flex items-center gap-2 text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
            <Eye className="size-3 text-zinc-400" />
            Watching this week
          </div>
          <div className="grid gap-2 sm:grid-cols-2">
            {digest.watching.map((w, i) => {
              const person = digest.featuredPeople.find(
                (p) => p.username === w.targetUsername
              );
              return (
                <div
                  key={i}
                  className="flex items-start gap-2.5 rounded-xl border border-zinc-800 bg-gradient-to-br from-zinc-900/80 to-zinc-900/30 px-3 py-2.5"
                >
                  {person ? (
                    <Avatar size="sm" className="size-7 shrink-0">
                      {person.avatarUrl ? (
                        <AvatarImage src={person.avatarUrl} alt={person.name} />
                      ) : null}
                      <AvatarFallback className="bg-zinc-800 text-[10px]">
                        {getInitials(person.name)}
                      </AvatarFallback>
                    </Avatar>
                  ) : (
                    <span className="text-xl leading-none">{w.emoji}</span>
                  )}
                  <p className="text-[13px] leading-snug text-zinc-200">
                    {w.prediction}
                  </p>
                </div>
              );
            })}
          </div>
        </section>
      )}

      {/* Theme chips — expandable to reveal linked achievements */}
      {digest.groups.length > 0 && (
        <div className="grid gap-2 sm:grid-cols-2">
          {digest.groups.map((g, i) => (
            <ExpandableChip
              key={i}
              group={g}
              stagger={6 + i}
              onJump={jumpToAchievement}
            />
          ))}
        </div>
      )}

      {/* Suggested action */}
      {action && (
        <ActionRow
          action={action}
          stagger={6 + digest.groups.length}
          onRegenerate={onRegenerate}
        />
      )}
    </div>
  );
}

function ExpandableChip({
  group,
  stagger,
  onJump,
}: {
  group: SmartFeedGroup;
  stagger: number;
  onJump: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const hasStubs = group.achievements.length > 0;

  return (
    <div
      className={cn(
        "smart-feed-fade flex flex-col rounded-xl border transition-colors",
        THEME_STYLE[group.theme] ?? THEME_STYLE.OTHER
      )}
      style={{ ["--stagger" as string]: stagger }}
    >
      <button
        type="button"
        onClick={() => hasStubs && setOpen((v) => !v)}
        disabled={!hasStubs}
        className={cn(
          "flex items-start gap-3 px-3 py-2.5 text-left",
          hasStubs && "cursor-pointer"
        )}
        aria-expanded={open}
      >
        <span className="text-xl leading-none">{group.emoji}</span>
        <div className="flex min-w-0 flex-1 flex-col">
          <span
            className={cn(
              "truncate text-[10px] font-semibold uppercase tracking-wider",
              THEME_LABEL_STYLE[group.theme] ?? THEME_LABEL_STYLE.OTHER
            )}
          >
            {group.label}
          </span>
          <span className="text-[13px] leading-snug text-zinc-200">
            {group.detail}
          </span>
        </div>
        {hasStubs && (
          <span className="mt-0.5 shrink-0 text-zinc-500">
            {open ? (
              <ChevronUp className="size-3.5" />
            ) : (
              <ChevronDown className="size-3.5" />
            )}
          </span>
        )}
      </button>

      {/* Expanded reveal */}
      {open && (
        <div className="flex flex-col gap-1 border-t border-zinc-800/60 px-3 py-2">
          {group.achievements.map((a) => (
            <button
              key={a.id}
              onClick={() => onJump(a.id)}
              className="group/row flex items-center gap-2 rounded-md px-1.5 py-1 text-left transition-colors hover:bg-zinc-800/60"
            >
              <span className="size-1.5 shrink-0 rounded-full bg-zinc-600 group-hover/row:bg-zinc-300" />
              <span className="flex-1 truncate text-[12px] text-zinc-300 group-hover/row:text-zinc-100">
                {a.title}
              </span>
              {a.proofUrl && (
                <a
                  href={a.proofUrl}
                  target="_blank"
                  rel="noreferrer"
                  onClick={(e) => e.stopPropagation()}
                  className="shrink-0 text-zinc-500 transition-colors hover:text-zinc-200"
                  aria-label="Open proof link"
                >
                  <ExternalLink className="size-3" />
                </a>
              )}
            </button>
          ))}
        </div>
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
