"use client";

import { useEffect, useRef, useState, use } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api-client";
import { Loader2, AlertCircle, CheckCircle2 } from "lucide-react";
import Link from "next/link";

export default function YouTubeCallbackPage({
  searchParams,
}: {
  searchParams: Promise<{ code?: string; error?: string }>;
}) {
  const params = use(searchParams);
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const processedRef = useRef(false);

  useEffect(() => {
    if (processedRef.current) return;
    processedRef.current = true;

    // Handle Google OAuth error (e.g., user denied access)
    if (params.error) {
      setError(
        params.error === "access_denied"
          ? "YouTube access was denied. Please try again and grant the required permissions."
          : `Google returned an error: ${params.error}`
      );
      return;
    }

    const code = params.code;
    if (!code) {
      setError("No authorization code received from Google.");
      return;
    }

    // Check if user is authenticated
    const token = localStorage.getItem("praxis_access_token");
    if (!token) {
      setError(
        "You must be signed in to connect a YouTube account. Please sign in first."
      );
      return;
    }

    api
      .post("/auth/youtube/callback", { code })
      .then(() =>
        // Trigger initial sync to fetch achievements and create connected_providers entry
        api.post("/achievements/sync", { provider: "YOUTUBE" }).catch((err) => {
          console.warn("[YouTube callback] initial sync failed:", err);
          // Don't block success — user can manually sync later
        })
      )
      .then(() => {
        setSuccess(true);
        setTimeout(() => {
          router.replace("/settings");
        }, 1500);
      })
      .catch((err: unknown) => {
        const message =
          err && typeof err === "object" && "message" in err
            ? (err as { message: string }).message
            : "Failed to connect YouTube account.";
        console.error("[YouTube callback] error:", err);
        setError(message + " Please try again.");
      });
  }, [params.code, params.error, router]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-950">
        <div className="flex max-w-md flex-col items-center gap-4 px-4 text-center">
          <div className="flex size-12 items-center justify-center rounded-full bg-red-500/10">
            <AlertCircle className="size-6 text-red-400" />
          </div>
          <h2 className="text-lg font-semibold text-zinc-100">
            Connection Failed
          </h2>
          <p className="text-sm text-zinc-400">{error}</p>
          <Link
            href="/settings"
            className="mt-2 inline-flex h-8 items-center justify-center rounded-lg border border-zinc-700 px-4 text-sm font-medium text-zinc-300 transition-colors hover:bg-zinc-800"
          >
            Back to Settings
          </Link>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-950">
        <div className="flex flex-col items-center gap-4">
          <CheckCircle2 className="size-8 text-green-400" />
          <p className="text-sm text-zinc-400">
            YouTube connected! Redirecting...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950">
      <div className="flex flex-col items-center gap-4">
        <Loader2 className="size-8 animate-spin text-zinc-400" />
        <p className="text-sm text-zinc-400">Connecting YouTube account...</p>
      </div>
    </div>
  );
}
