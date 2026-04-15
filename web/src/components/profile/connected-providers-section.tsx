"use client";

import { useState } from "react";
import { GitFork, Loader2, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api-client";
import { formatRelativeTime } from "@/lib/format";
import type { ConnectedProvider } from "@/types";

interface ConnectedProvidersSectionProps {
  providers: ConnectedProvider[];
  isOwnProfile: boolean;
}

function providerIcon(provider: string) {
  switch (provider.toUpperCase()) {
    case "GITHUB":
      return <GitFork className="size-4" />;
    default:
      return <GitFork className="size-4" />;
  }
}

function syncStatusVariant(status: string) {
  switch (status.toLowerCase()) {
    case "synced":
      return "secondary" as const;
    case "syncing":
      return "default" as const;
    case "failed":
      return "destructive" as const;
    default:
      return "outline" as const;
  }
}

export function ConnectedProvidersSection({
  providers,
  isOwnProfile,
}: ConnectedProvidersSectionProps) {
  const [syncing, setSyncing] = useState<string | null>(null);

  const handleSync = async (provider: string) => {
    setSyncing(provider);
    try {
      await api.post("/achievements/sync", { provider: provider.toUpperCase() });
    } finally {
      setSyncing(null);
    }
  };

  if (providers.length === 0) {
    return (
      <div className="py-6 text-center text-sm text-zinc-500">
        No connected providers.
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {providers.map((p) => (
        <div
          key={p.id}
          className="flex items-center gap-3 rounded-xl border border-zinc-800 bg-zinc-900/60 px-4 py-3"
        >
          <div className="rounded-lg bg-zinc-800 p-2 text-zinc-400">
            {providerIcon(p.provider)}
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-zinc-200">
                {p.provider}
              </span>
              <Badge variant={syncStatusVariant(p.syncStatus)}>
                {p.syncStatus}
              </Badge>
            </div>
            <p className="truncate text-xs text-zinc-500">
              @{p.providerUsername}
              {p.lastSyncedAt
                ? ` \u00B7 Last synced ${formatRelativeTime(p.lastSyncedAt)}`
                : ""}
            </p>
          </div>

          {isOwnProfile && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleSync(p.provider)}
              disabled={syncing === p.provider}
            >
              {syncing === p.provider ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <RefreshCw className="size-4" />
              )}
              Sync
            </Button>
          )}
        </div>
      ))}
    </div>
  );
}
