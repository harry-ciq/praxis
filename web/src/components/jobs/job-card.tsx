"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatRelativeTime } from "@/lib/format";
import {
  MapPin,
  Building2,
  DollarSign,
  Trophy,
  CheckCircle2,
} from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";
import type { Job } from "@/types";

interface JobCardProps {
  job: Job;
}

const JOB_TYPE_LABELS: Record<string, string> = {
  FULL_TIME: "Full Time",
  PART_TIME: "Part Time",
  CONTRACT: "Contract",
  REMOTE: "Remote",
};

export function JobCard({ job }: JobCardProps) {
  const totalRequired = job.totalRequired ?? job.requiredAchievements.length;
  const matched = job.matchedAchievements ?? 0;
  const hasMatchData = job.matchedAchievements !== undefined && totalRequired > 0;
  const matchedSet = new Set(job.matchedRequirements ?? []);
  const matchPct = totalRequired > 0 ? Math.round((matched / totalRequired) * 100) : 0;
  const fullyQualified = hasMatchData && matched === totalRequired;

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 transition-colors hover:border-zinc-700">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-3">
          {job.company.logoUrl ? (
            <img
              src={job.company.logoUrl}
              alt={job.company.name}
              className="size-10 rounded-lg object-cover"
            />
          ) : (
            <div className="flex size-10 items-center justify-center rounded-lg bg-zinc-800">
              <Building2 className="size-5 text-zinc-500" />
            </div>
          )}
          <div>
            <p className="text-sm text-zinc-400">{job.company.name}</p>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          {hasMatchData && (
            <span
              className={cn(
                "inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium tabular-nums",
                fullyQualified
                  ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-400"
                  : matched > 0
                    ? "border-amber-500/30 bg-amber-500/10 text-amber-400"
                    : "border-zinc-700 bg-zinc-800/50 text-zinc-400"
              )}
            >
              {fullyQualified && <CheckCircle2 className="size-3" />}
              {matched}/{totalRequired} matched · {matchPct}%
            </span>
          )}
          <span className="text-xs text-zinc-500">
            {formatRelativeTime(job.createdAt)}
          </span>
        </div>
      </div>

      <div>
        <Link
          href={`/jobs/${job.id}`}
          className="text-lg font-semibold text-zinc-50 hover:text-blue-400 transition-colors"
        >
          {job.title}
        </Link>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <div className="flex items-center gap-1 text-xs text-zinc-400">
          <MapPin className="size-3.5" />
          <span>{job.location}</span>
        </div>
        <Badge variant="secondary" className="text-xs">
          {JOB_TYPE_LABELS[job.jobType] ?? job.jobType}
        </Badge>
        {job.salaryRange && (
          <div className="flex items-center gap-1 text-xs text-zinc-400">
            <DollarSign className="size-3.5" />
            <span>{job.salaryRange}</span>
          </div>
        )}
      </div>

      {job.requiredAchievements.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {job.requiredAchievements.map((achievement) => {
            const isMatched = matchedSet.has(achievement);
            return (
              <Badge
                key={achievement}
                variant="outline"
                className={cn(
                  "gap-1 text-xs",
                  isMatched
                    ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-400"
                    : "border-zinc-700 text-zinc-300"
                )}
              >
                {isMatched ? (
                  <CheckCircle2 className="size-3" />
                ) : (
                  <Trophy className="size-3" />
                )}
                {achievement}
              </Badge>
            );
          })}
        </div>
      )}

      <div className="mt-1">
        <Link
          href={`/jobs/${job.id}`}
          className="inline-flex h-7 items-center justify-center rounded-lg bg-blue-600 px-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          View &amp; Apply
        </Link>
      </div>
    </div>
  );
}
