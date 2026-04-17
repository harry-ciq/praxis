import Link from "next/link";
import { Compass } from "lucide-react";

export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950 px-4">
      <div className="flex max-w-md flex-col items-center gap-4 text-center">
        <div className="flex size-14 items-center justify-center rounded-full bg-zinc-800/80">
          <Compass className="size-7 text-zinc-400" />
        </div>
        <div className="flex flex-col gap-1.5">
          <h1 className="text-2xl font-bold text-zinc-50">Page not found</h1>
          <p className="text-sm text-zinc-400">
            The page you're looking for doesn't exist or has been moved.
          </p>
        </div>
        <Link
          href="/feed"
          className="mt-2 inline-flex h-9 items-center justify-center rounded-lg bg-blue-600 px-4 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          Back to feed
        </Link>
      </div>
    </div>
  );
}
