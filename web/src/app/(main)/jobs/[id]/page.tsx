"use client";

import { use, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { JobApplicationForm } from "@/components/jobs/job-application-form";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  MapPin,
  Building2,
  DollarSign,
  Trophy,
  Globe,
  ArrowLeft,
  Loader2,
} from "lucide-react";
import { formatRelativeTime } from "@/lib/format";
import Link from "next/link";
import type { Job } from "@/types";

const JOB_TYPE_LABELS: Record<string, string> = {
  FULL_TIME: "Full Time",
  PART_TIME: "Part Time",
  CONTRACT: "Contract",
  REMOTE: "Remote",
};

export default function JobDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const [showApplicationForm, setShowApplicationForm] = useState(false);

  const { data: job, isLoading } = useQuery({
    queryKey: ["job", id],
    queryFn: () => api.get<Job>(`/jobs/${id}`),
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="size-8 animate-spin text-zinc-500" />
      </div>
    );
  }

  if (!job) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="text-sm text-zinc-500">Job not found</p>
        <Link
          href="/jobs"
          className="mt-4 inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-200"
        >
          <ArrowLeft className="size-4" />
          Back to jobs
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-6">
      <Link
        href="/jobs"
        className="mb-4 inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-sm text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-200"
      >
        <ArrowLeft className="size-4" />
        Back to jobs
      </Link>

      <div className="grid gap-6 lg:grid-cols-[1fr_300px]">
        {/* Main content */}
        <div className="flex flex-col gap-6">
          {/* Header */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6">
            <div className="flex items-start gap-4">
              {job.company.logoUrl ? (
                <img
                  src={job.company.logoUrl}
                  alt={job.company.name}
                  className="size-14 rounded-xl object-cover"
                />
              ) : (
                <div className="flex size-14 items-center justify-center rounded-xl bg-zinc-800">
                  <Building2 className="size-7 text-zinc-500" />
                </div>
              )}
              <div className="flex-1">
                <h1 className="text-2xl font-bold text-zinc-50">
                  {job.title}
                </h1>
                <p className="mt-1 text-sm text-zinc-400">
                  {job.company.name}
                </p>
              </div>
            </div>

            <div className="mt-4 flex flex-wrap items-center gap-3">
              <div className="flex items-center gap-1.5 text-sm text-zinc-400">
                <MapPin className="size-4" />
                <span>{job.location}</span>
              </div>
              <Badge variant="secondary">
                {JOB_TYPE_LABELS[job.jobType] ?? job.jobType}
              </Badge>
              {job.salaryRange && (
                <div className="flex items-center gap-1.5 text-sm text-zinc-400">
                  <DollarSign className="size-4" />
                  <span>{job.salaryRange}</span>
                </div>
              )}
              <span className="text-xs text-zinc-500">
                Posted {formatRelativeTime(job.createdAt)}
              </span>
            </div>
          </div>

          {/* Description */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6">
            <h2 className="mb-3 text-lg font-semibold text-zinc-100">
              Description
            </h2>
            <div className="prose prose-sm prose-invert max-w-none text-zinc-300">
              <p className="whitespace-pre-wrap">{job.description}</p>
            </div>
          </div>

          {/* Required achievements */}
          {job.requiredAchievements.length > 0 && (
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6">
              <h2 className="mb-3 text-lg font-semibold text-zinc-100">
                Required Achievements
              </h2>
              <div className="flex flex-col gap-2">
                {job.requiredAchievements.map((achievement) => (
                  <div
                    key={achievement}
                    className="flex items-center gap-3 rounded-lg border border-zinc-800 bg-zinc-900 px-4 py-3"
                  >
                    <Trophy className="size-5 text-amber-500" />
                    <span className="text-sm text-zinc-200">{achievement}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Apply section (mobile) */}
          <div className="lg:hidden">
            {showApplicationForm ? (
              <JobApplicationForm
                jobId={job.id}
                onSuccess={() => setShowApplicationForm(false)}
              />
            ) : (
              <Button
                className="w-full bg-blue-600 text-white hover:bg-blue-700"
                onClick={() => setShowApplicationForm(true)}
              >
                Apply for this position
              </Button>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="hidden flex-col gap-4 lg:flex">
          {/* Company info */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
            <h3 className="mb-3 text-base font-semibold text-zinc-100">
              About {job.company.name}
            </h3>
            {job.company.description && (
              <p className="mb-3 text-sm text-zinc-400">
                {job.company.description}
              </p>
            )}
            {job.company.website && (
              <a
                href={job.company.website}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1.5 text-sm text-blue-400 hover:text-blue-300"
              >
                <Globe className="size-3.5" />
                {job.company.website.replace(/^https?:\/\//, "")}
              </a>
            )}
          </div>

          {/* Apply */}
          {showApplicationForm ? (
            <JobApplicationForm
              jobId={job.id}
              onSuccess={() => setShowApplicationForm(false)}
            />
          ) : (
            <Button
              className="w-full bg-blue-600 text-white hover:bg-blue-700"
              onClick={() => setShowApplicationForm(true)}
            >
              Apply for this position
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
