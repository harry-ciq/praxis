"use client";

import { useEffect, useRef, useState, use } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { Loader2, AlertCircle } from "lucide-react";
import Link from "next/link";

export default function CallbackPage({
  searchParams,
}: {
  searchParams: Promise<{ code?: string }>;
}) {
  const params = use(searchParams);
  const { handleCallback } = useAuth();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const processedRef = useRef(false);

  useEffect(() => {
    if (processedRef.current) return;
    processedRef.current = true;

    const code = params.code;
    if (!code) {
      setError("No authorization code received from GitHub.");
      return;
    }

    handleCallback(code)
      .then(() => {
        router.replace("/feed");
      })
      .catch(() => {
        setError("Authentication failed. Please try again.");
      });
  }, [params.code, handleCallback, router]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-950">
        <div className="flex max-w-md flex-col items-center gap-4 px-4 text-center">
          <div className="flex size-12 items-center justify-center rounded-full bg-red-500/10">
            <AlertCircle className="size-6 text-red-400" />
          </div>
          <h2 className="text-lg font-semibold text-zinc-100">
            Authentication Error
          </h2>
          <p className="text-sm text-zinc-400">{error}</p>
          <Link
            href="/signin"
            className="mt-2 inline-flex h-8 items-center justify-center rounded-lg border border-zinc-700 px-4 text-sm font-medium text-zinc-300 transition-colors hover:bg-zinc-800"
          >
            Back to Sign In
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950">
      <div className="flex flex-col items-center gap-4">
        <Loader2 className="size-8 animate-spin text-zinc-400" />
        <p className="text-sm text-zinc-400">Completing sign in...</p>
      </div>
    </div>
  );
}
