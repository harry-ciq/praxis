"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatRelativeTime } from "@/lib/format";
import {
  MapPin,
  Building2,
  DollarSign,
  Trophy,
} from "lucide-react";
import Link from "next/link";
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
        <span className="shrink-0 text-xs text-zinc-500">
          {formatRelativeTime(job.createdAt)}
        </span>
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
          {job.requiredAchievements.map((achievement) => (
            <Badge
              key={achievement}
              variant="outline"
              className="gap-1 border-zinc-700 text-xs text-zinc-300"
            >
              <Trophy className="size-3" />
              {achievement}
            </Badge>
          ))}
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
