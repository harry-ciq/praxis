"use client";

import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import { Code2, Loader2, CheckCircle2, Link2, Video } from "lucide-react";
import type { UserProfile } from "@/types";

export default function SettingsPage() {
  const { user } = useAuth();

  const [name, setName] = useState(user?.name ?? "");
  const [headline, setHeadline] = useState(user?.headline ?? "");
  const [bio, setBio] = useState("");

  const profileMutation = useMutation({
    mutationFn: () =>
      api.patch("/auth/me", {
        name: name.trim(),
        headline: headline.trim() || null,
        bio: bio.trim() || null,
      }),
  });

  const { data: githubStatus } = useQuery({
    queryKey: ["github-connection"],
    queryFn: () =>
      api
        .get<{ connected: boolean; username?: string }>("/auth/github/status")
        .catch(() => ({ connected: true, username: user?.username })),
  });

  // Fetch full profile (including connected providers) for the current user
  const { data: profile } = useQuery({
    queryKey: ["my-profile", user?.username],
    queryFn: () =>
      api.get<UserProfile>(`/users/${user!.username}`),
    enabled: !!user?.username,
  });

  const youtubeProvider = profile?.providers?.find(
    (p) => p.provider === "YOUTUBE"
  );
  const youtubeConnected = !!youtubeProvider;

  const [youtubeConnecting, setYoutubeConnecting] = useState(false);

  const connectYouTube = async () => {
    try {
      setYoutubeConnecting(true);
      const data = await api.get<{ url: string }>("/auth/youtube");
      window.location.href = data.url;
    } catch {
      setYoutubeConnecting(false);
    }
  };

  return (
    <div className="mx-auto max-w-2xl px-4 py-6">
      <h1 className="mb-6 text-2xl font-bold text-zinc-50">Settings</h1>

      {/* Profile section */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6">
        <h2 className="mb-4 text-lg font-semibold text-zinc-100">Profile</h2>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm text-zinc-400">Name</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="border-zinc-700 bg-zinc-900 text-zinc-100"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm text-zinc-400">Headline</label>
            <Input
              value={headline}
              onChange={(e) => setHeadline(e.target.value)}
              placeholder="e.g. Full-stack developer, open source contributor"
              className="border-zinc-700 bg-zinc-900 text-zinc-100 placeholder:text-zinc-600"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm text-zinc-400">Bio</label>
            <Textarea
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="Tell people about yourself..."
              className="min-h-24 border-zinc-700 bg-zinc-900 text-zinc-100 placeholder:text-zinc-600"
              rows={4}
            />
          </div>

          <div className="flex items-center gap-3">
            <Button
              onClick={() => profileMutation.mutate()}
              disabled={profileMutation.isPending}
              className="bg-blue-600 text-white hover:bg-blue-700"
            >
              {profileMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 size-4 animate-spin" />
                  Saving...
                </>
              ) : (
                "Save Changes"
              )}
            </Button>
            {profileMutation.isSuccess && (
              <span className="flex items-center gap-1 text-sm text-green-400">
                <CheckCircle2 className="size-4" />
                Saved
              </span>
            )}
          </div>
        </div>
      </section>

      <Separator className="my-6 bg-zinc-800" />

      {/* Connected accounts */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6">
        <h2 className="mb-4 text-lg font-semibold text-zinc-100">
          Connected Accounts
        </h2>

        <div className="flex flex-col gap-3">
          {/* GitHub */}
          <div className="flex items-center justify-between rounded-lg border border-zinc-800 bg-zinc-900 px-4 py-3">
            <div className="flex items-center gap-3">
              <Code2 className="size-5 text-zinc-300" />
              <div>
                <p className="text-sm font-medium text-zinc-100">GitHub</p>
                {githubStatus?.connected ? (
                  <p className="text-xs text-zinc-500">
                    Connected as @{githubStatus.username ?? user?.username}
                  </p>
                ) : (
                  <p className="text-xs text-zinc-500">Not connected</p>
                )}
              </div>
            </div>
            <Button
              variant={githubStatus?.connected ? "outline" : "default"}
              size="sm"
              className={
                githubStatus?.connected
                  ? "border-zinc-700 text-zinc-300"
                  : "bg-zinc-800 text-zinc-100"
              }
            >
              {githubStatus?.connected ? (
                <>
                  <Link2 className="mr-1.5 size-3.5" />
                  Disconnect
                </>
              ) : (
                <>
                  <Code2 className="mr-1.5 size-3.5" />
                  Connect
                </>
              )}
            </Button>
          </div>

          {/* YouTube */}
          <div className="flex items-center justify-between rounded-lg border border-zinc-800 bg-zinc-900 px-4 py-3">
            <div className="flex items-center gap-3">
              <Video className="size-5 text-red-500" />
              <div>
                <p className="text-sm font-medium text-zinc-100">YouTube</p>
                {youtubeConnected ? (
                  <p className="text-xs text-zinc-500">
                    Connected
                    {youtubeProvider?.providerUsername
                      ? ` as ${youtubeProvider.providerUsername}`
                      : ""}
                    {youtubeProvider?.lastSyncedAt
                      ? ` · synced ${new Date(youtubeProvider.lastSyncedAt).toLocaleString()}`
                      : ""}
                  </p>
                ) : (
                  <p className="text-xs text-zinc-500">
                    Connect to track video milestones, subscribers & views
                  </p>
                )}
              </div>
            </div>
            {youtubeConnected ? (
              <Button
                variant="outline"
                size="sm"
                className="border-zinc-700 text-zinc-300"
              >
                <Link2 className="mr-1.5 size-3.5" />
                Disconnect
              </Button>
            ) : (
              <Button
                onClick={connectYouTube}
                disabled={youtubeConnecting}
                size="sm"
                className="bg-zinc-800 text-zinc-100 hover:bg-zinc-700"
              >
                {youtubeConnecting ? (
                  <>
                    <Loader2 className="mr-1.5 size-3.5 animate-spin" />
                    Connecting...
                  </>
                ) : (
                  <>
                    <Video className="mr-1.5 size-3.5" />
                    Connect
                  </>
                )}
              </Button>
            )}
          </div>
        </div>
      </section>

      <Separator className="my-6 bg-zinc-800" />

      {/* Placeholder sections */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6">
        <h2 className="mb-2 text-lg font-semibold text-zinc-100">
          Email Preferences
        </h2>
        <p className="text-sm text-zinc-500">
          Email notification settings coming soon.
        </p>
      </section>
    </div>
  );
}
