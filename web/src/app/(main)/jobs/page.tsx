"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { JobCard } from "@/components/jobs/job-card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Search, Briefcase, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Job, JobType } from "@/types";

const JOB_TYPES: { value: JobType | "ALL"; label: string }[] = [
  { value: "ALL", label: "All" },
  { value: "FULL_TIME", label: "Full Time" },
  { value: "PART_TIME", label: "Part Time" },
  { value: "CONTRACT", label: "Contract" },
  { value: "REMOTE", label: "Remote" },
];

export default function JobsPage() {
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<JobType | "ALL">("ALL");

  const { data: jobs = [], isLoading } = useQuery({
    queryKey: ["jobs", search, typeFilter],
    queryFn: () => {
      const params: Record<string, string> = {};
      if (search) params.search = search;
      if (typeFilter !== "ALL") params.jobType = typeFilter;
      return api.get<Job[]>("/jobs", params);
    },
  });

  return (
    <div className="mx-auto max-w-4xl px-4 py-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-zinc-50">Job Board</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Find jobs that match your verified achievements
        </p>
      </div>

      {/* Search */}
      <div className="mb-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search jobs..."
            className="h-10 border-zinc-800 bg-zinc-900 pl-9 text-zinc-100 placeholder:text-zinc-500"
          />
        </div>
      </div>

      {/* Filters */}
      <div className="mb-6 flex flex-wrap gap-2">
        {JOB_TYPES.map((type) => (
          <Button
            key={type.value}
            variant={typeFilter === type.value ? "default" : "outline"}
            size="sm"
            onClick={() => setTypeFilter(type.value)}
            className={cn(
              typeFilter === type.value
                ? "bg-blue-600 text-white hover:bg-blue-700"
                : "border-zinc-700 text-zinc-400 hover:text-zinc-200",
            )}
          >
            {type.label}
          </Button>
        ))}
      </div>

      {/* Job listing */}
      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="size-8 animate-spin text-zinc-500" />
        </div>
      ) : jobs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Briefcase className="mb-3 size-12 text-zinc-700" />
          <p className="text-sm text-zinc-500">No jobs found</p>
          <p className="mt-1 text-xs text-zinc-600">
            Try adjusting your search or filters
          </p>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {jobs.map((job) => (
            <JobCard key={job.id} job={job} />
          ))}
        </div>
      )}
    </div>
  );
}
