"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Loader2, CheckCircle2, XCircle } from "lucide-react";
import type { JobApplication } from "@/types";

interface JobApplicationFormProps {
  jobId: string;
  onSuccess?: () => void;
}

export function JobApplicationForm({ jobId, onSuccess }: JobApplicationFormProps) {
  const [coverNote, setCoverNote] = useState("");

  const mutation = useMutation({
    mutationFn: () =>
      api.post<JobApplication>(`/jobs/${jobId}/apply`, {
        coverNote: coverNote.trim() || null,
      }),
    onSuccess: () => {
      onSuccess?.();
    },
  });

  if (mutation.isSuccess) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 text-center">
        <CheckCircle2 className="size-10 text-green-500" />
        <p className="text-sm font-medium text-zinc-100">
          Application submitted!
        </p>
        <p className="text-xs text-zinc-500">
          You&apos;ll hear back from the company soon.
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
      <h3 className="text-base font-semibold text-zinc-100">Apply Now</h3>

      <div className="flex flex-col gap-2">
        <label className="text-sm text-zinc-400">
          Cover Note (optional)
        </label>
        <Textarea
          value={coverNote}
          onChange={(e) => setCoverNote(e.target.value)}
          placeholder="Tell them why you're a great fit..."
          className="min-h-24 border-zinc-700 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500"
          rows={4}
        />
      </div>

      {mutation.isError && (
        <div className="flex items-center gap-2 text-sm text-red-400">
          <XCircle className="size-4" />
          <span>Failed to submit application. Please try again.</span>
        </div>
      )}

      <Button
        onClick={() => mutation.mutate()}
        disabled={mutation.isPending}
        className="w-full bg-blue-600 text-white hover:bg-blue-700"
      >
        {mutation.isPending ? (
          <>
            <Loader2 className="mr-2 size-4 animate-spin" />
            Submitting...
          </>
        ) : (
          "Submit Application"
        )}
      </Button>
    </div>
  );
}
