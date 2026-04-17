"use client";

import Link from "next/link";
import {
  Hammer,
  Star,
  Flame,
  GitPullRequest,
  PartyPopper,
  ExternalLink,
  CheckCircle2,
  GitFork,
  GitCommitHorizontal,
  Play,
  ArrowUpRight,
  ArchiveIcon,
  Video,
  Users,
  Eye,
  TvMinimalPlay,
} from "lucide-react";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { ReactionBar } from "@/components/achievements/reaction-bar";
import { formatRelativeTime } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { Achievement, AchievementType, Provider } from "@/types";

const ACHIEVEMENT_TYPE_CONFIG: Record<
  AchievementType,
  {
    icon: React.ElementType;
    label: string;
    badgeClass: string;
    accentGradient: string;
    iconBg: string;
  }
> = {
  REPO_CREATED: {
    icon: Hammer,
    label: "New Repository",
    badgeClass: "bg-green-500/10 text-green-400 border-green-500/20",
    accentGradient: "from-green-500/20 via-transparent to-transparent",
    iconBg: "bg-green-500/10 text-green-400",
  },
  STARS_MILESTONE: {
    icon: Star,
    label: "Stars Milestone",
    badgeClass: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    accentGradient: "from-amber-500/20 via-transparent to-transparent",
    iconBg: "bg-amber-500/10 text-amber-400",
  },
  COMMIT_STREAK: {
    icon: Flame,
    label: "Commit Streak",
    badgeClass: "bg-orange-500/10 text-orange-400 border-orange-500/20",
    accentGradient: "from-orange-500/20 via-transparent to-transparent",
    iconBg: "bg-orange-500/10 text-orange-400",
  },
  PR_MERGED: {
    icon: GitPullRequest,
    label: "PR Merged",
    badgeClass: "bg-blue-500/10 text-blue-400 border-blue-500/20",
    accentGradient: "from-blue-500/20 via-transparent to-transparent",
    iconBg: "bg-blue-500/10 text-blue-400",
  },
  FIRST_CONTRIBUTION: {
    icon: PartyPopper,
    label: "First Contribution",
    badgeClass: "bg-purple-500/10 text-purple-400 border-purple-500/20",
    accentGradient: "from-purple-500/20 via-transparent to-transparent",
    iconBg: "bg-purple-500/10 text-purple-400",
  },
  WEEKLY_COMMITS: {
    icon: GitCommitHorizontal,
    label: "Weekly Activity",
    badgeClass: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    accentGradient: "from-emerald-500/20 via-transparent to-transparent",
    iconBg: "bg-emerald-500/10 text-emerald-400",
  },
  VIDEO_PUBLISHED: {
    icon: Play,
    label: "Video Published",
    badgeClass: "bg-red-500/10 text-red-400 border-red-500/20",
    accentGradient: "from-red-500/20 via-transparent to-transparent",
    iconBg: "bg-red-500/10 text-red-400",
  },
  SUBSCRIBERS_MILESTONE: {
    icon: Users,
    label: "Subscribers Milestone",
    badgeClass: "bg-rose-500/10 text-rose-400 border-rose-500/20",
    accentGradient: "from-rose-500/20 via-transparent to-transparent",
    iconBg: "bg-rose-500/10 text-rose-400",
  },
  VIEWS_MILESTONE: {
    icon: Eye,
    label: "Views Milestone",
    badgeClass: "bg-pink-500/10 text-pink-400 border-pink-500/20",
    accentGradient: "from-pink-500/20 via-transparent to-transparent",
    iconBg: "bg-pink-500/10 text-pink-400",
  },
  CHANNEL_CREATED: {
    icon: TvMinimalPlay,
    label: "Channel Created",
    badgeClass: "bg-red-500/10 text-red-400 border-red-500/20",
    accentGradient: "from-red-500/20 via-transparent to-transparent",
    iconBg: "bg-red-500/10 text-red-400",
  },
};

const SOURCE_LABEL: Record<Provider, string> = {
  GITHUB: "GitHub",
  YOUTUBE: "YouTube",
};

const SOURCE_ICON: Record<Provider, React.ElementType> = {
  GITHUB: GitFork,
  YOUTUBE: Video,
};

interface AchievementCardProps {
  achievement: Achievement;
}

export function AchievementCard({ achievement }: AchievementCardProps) {
  const {
    id,
    user,
    type,
    title,
    description,
    metadata,
    proofUrl,
    source,
    verificationHash,
    status,
    createdAt,
    reactions,
    userReaction,
  } = achievement;

  const typeConfig = ACHIEVEMENT_TYPE_CONFIG[type];
  const TypeIcon = typeConfig.icon;
  const SourceIcon = SOURCE_ICON[source];
  const isArchived = status === "archived";

  return (
    <div className={cn(
      "group relative overflow-hidden rounded-2xl border transition-all duration-300",
      isArchived
        ? "border-zinc-800/50 bg-zinc-900/50 opacity-75"
        : "border-zinc-800/80 bg-zinc-900/90 hover:border-zinc-700/80 hover:shadow-lg hover:shadow-black/20"
    )}>
      {/* Top gradient accent */}
      <div
        className={cn(
          "pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r",
          typeConfig.accentGradient
        )}
      />

      {/* Subtle corner glow */}
      <div
        className={cn(
          "pointer-events-none absolute -left-12 -top-12 h-32 w-32 rounded-full bg-gradient-to-br opacity-[0.07] blur-2xl transition-opacity duration-300 group-hover:opacity-[0.12]",
          typeConfig.accentGradient
        )}
      />

      <div className="relative p-5">
        {/* Header: user info + timestamp */}
        <div className="mb-4 flex items-center justify-between">
          <Link
            href={`/profile/${user.username}`}
            className="flex items-center gap-3 transition-opacity hover:opacity-80"
          >
            <Avatar size="default">
              {user.avatarUrl ? (
                <AvatarImage src={user.avatarUrl} alt={user.name} />
              ) : null}
              <AvatarFallback>
                {user.name
                  .split(" ")
                  .map((w) => w[0])
                  .join("")
                  .slice(0, 2)
                  .toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col">
              <span className="text-sm font-semibold text-zinc-100">
                {user.name}
              </span>
              <span className="text-xs text-zinc-500">@{user.username}</span>
            </div>
          </Link>
          <span className="text-xs tabular-nums text-zinc-500">
            {formatRelativeTime(createdAt)}
          </span>
        </div>

        {/* Achievement type icon + title row */}
        <div className="mb-3 flex items-start gap-3">
          <div
            className={cn(
              "flex size-10 shrink-0 items-center justify-center rounded-xl",
              typeConfig.iconBg
            )}
          >
            <TypeIcon className="size-5" />
          </div>
          <div className="min-w-0 flex-1">
            <h3 className="text-[15px] font-bold leading-snug text-zinc-100">
              {title}
            </h3>
            {/* Inline badges */}
            <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
              <span
                className={cn(
                  "inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium",
                  typeConfig.badgeClass
                )}
              >
                {typeConfig.label}
              </span>
              <span className="inline-flex items-center gap-1 rounded-md border border-zinc-800 bg-zinc-800/50 px-2 py-0.5 text-[11px] font-medium text-zinc-500">
                <SourceIcon className="size-2.5" />
                {SOURCE_LABEL[source]}
              </span>
              {verificationHash && (
                <span className="inline-flex items-center gap-0.5 text-[11px] font-medium text-emerald-500">
                  <CheckCircle2 className="size-3" />
                  Verified
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Description (non-weekly) */}
        {description && type !== "WEEKLY_COMMITS" && (
          <p className="mb-3 pl-[52px] text-[13px] leading-relaxed text-zinc-400">
            {description}
          </p>
        )}

        {/* Repo breakdown for weekly commits */}
        {type === "WEEKLY_COMMITS" && Array.isArray(metadata?.repos) && (
          <div className="mb-3 ml-[52px] space-y-1">
            {(
              metadata.repos as {
                repo: string;
                commits: number;
                owned?: boolean;
              }[]
            ).map((r) => (
              <a
                key={r.repo}
                href={`https://github.com/${r.repo}`}
                target="_blank"
                rel="noopener noreferrer"
                className="group/repo flex items-center gap-2.5 rounded-lg border border-zinc-800/60 bg-zinc-800/30 px-3 py-2 transition-all hover:border-zinc-700/60 hover:bg-zinc-800/50"
              >
                <GitCommitHorizontal className="size-3.5 shrink-0 text-emerald-500/70" />
                <span className="truncate text-[13px] font-medium text-zinc-300 group-hover/repo:text-zinc-100">
                  {r.repo}
                </span>
                {!r.owned && (
                  <span className="shrink-0 rounded border border-zinc-700/50 bg-zinc-800 px-1.5 py-0.5 text-[10px] font-medium text-zinc-500">
                    contributor
                  </span>
                )}
                <span className="ml-auto shrink-0 text-[12px] tabular-nums text-zinc-500">
                  {r.commits} commit{r.commits !== 1 ? "s" : ""}
                </span>
              </a>
            ))}
          </div>
        )}

        {/* Proof link or archived notice */}
        {isArchived ? (
          <div className="mb-4 ml-[52px] inline-flex items-center gap-2 rounded-md border border-zinc-700/50 bg-zinc-800/50 px-3 py-1.5 text-[13px] font-medium text-zinc-300">
            <ArchiveIcon className="size-4 text-zinc-400" />
            Source no longer available
          </div>
        ) : proofUrl ? (
          <a
            href={proofUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="mb-4 ml-[52px] inline-flex items-center gap-1.5 text-[13px] font-medium text-zinc-300 transition-colors hover:text-white"
          >
            <ArrowUpRight className="size-3.5" />
            View on {SOURCE_LABEL[source]}
          </a>
        ) : null}

        {/* Divider */}
        <div className="ml-[52px] border-t border-zinc-800/60 pt-3">
          <ReactionBar
            achievementId={id}
            reactions={reactions}
            userReaction={userReaction}
          />
        </div>
      </div>
    </div>
  );
}
