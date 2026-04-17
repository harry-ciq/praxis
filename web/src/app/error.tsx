"use client";

import { useEffect } from "react";
import Link from "next/link";
import { AlertTriangle, RotateCw } from "lucide-react";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // In production this could go to Sentry, etc.
    console.error("[GlobalError]", error);
  }, [error]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950 px-4">
      <div className="flex max-w-md flex-col items-center gap-4 text-center">
        <div className="flex size-14 items-center justify-center rounded-full bg-red-500/10">
          <AlertTriangle className="size-7 text-red-400" />
        </div>
        <div className="flex flex-col gap-1.5">
          <h1 className="text-2xl font-bold text-zinc-50">
            Something went wrong
          </h1>
          <p className="text-sm text-zinc-400">
            An unexpected error occurred. You can try again or head back to the feed.
          </p>
          {error.digest && (
            <p className="mt-1 text-[11px] font-mono text-zinc-600">
              ref: {error.digest}
            </p>
          )}
        </div>
        <div className="mt-2 flex items-center gap-2">
          <button
            onClick={reset}
            className="inline-flex h-9 items-center justify-center gap-1.5 rounded-lg bg-blue-600 px-4 text-sm font-medium text-white transition-colors hover:bg-blue-700"
          >
            <RotateCw className="size-3.5" />
            Try again
          </button>
          <Link
            href="/feed"
            className="inline-flex h-9 items-center justify-center rounded-lg border border-zinc-700 px-4 text-sm font-medium text-zinc-300 transition-colors hover:bg-zinc-800"
          >
            Go to feed
          </Link>
        </div>
      </div>
    </div>
  );
}
