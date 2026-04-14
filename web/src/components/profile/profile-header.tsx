"use client";

import { GitFork, Pencil } from "lucide-react";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/use-auth";
import { useFollow } from "@/hooks/use-profile";
import { formatNumber } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { UserProfile } from "@/types";

interface ProfileHeaderProps {
  profile: UserProfile;
}

export function ProfileHeader({ profile }: ProfileHeaderProps) {
  const { user } = useAuth();
  const { follow, unfollow } = useFollow();

  const isOwnProfile = user?.username === profile.username;

  const handleFollowToggle = () => {
    if (profile.isFollowing) {
      unfollow.mutate(profile.username);
    } else {
      follow.mutate(profile.username);
    }
  };

  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-6">
      <div className="flex flex-col items-start gap-5 sm:flex-row sm:items-center">
        {/* Avatar */}
        <Avatar
          className={cn("size-20")}
        >
          {profile.avatarUrl ? (
            <AvatarImage src={profile.avatarUrl} alt={profile.name} />
          ) : null}
          <AvatarFallback className="text-xl">
            {profile.name
              .split(" ")
              .map((w) => w[0])
              .join("")
              .slice(0, 2)
              .toUpperCase()}
          </AvatarFallback>
        </Avatar>

        {/* Info */}
        <div className="flex-1">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h1 className="text-xl font-bold text-zinc-100">
                {profile.name}
              </h1>
              <p className="text-sm text-zinc-500">@{profile.username}</p>
              {profile.headline && (
                <p className="mt-1 text-sm text-zinc-400">
                  {profile.headline}
                </p>
              )}
            </div>

            {/* Action button */}
            {isOwnProfile ? (
              <Button variant="outline" size="sm">
                <Pencil className="size-3.5" />
                Edit Profile
              </Button>
            ) : (
              <Button
                variant={profile.isFollowing ? "outline" : "default"}
                size="sm"
                onClick={handleFollowToggle}
                disabled={follow.isPending || unfollow.isPending}
              >
                {profile.isFollowing ? "Unfollow" : "Follow"}
              </Button>
            )}
          </div>

          {/* Stats row */}
          <div className="mt-4 flex items-center gap-5 text-sm">
            <div>
              <span className="font-semibold text-zinc-100">
                {formatNumber(profile.achievementCount)}
              </span>{" "}
              <span className="text-zinc-500">achievements</span>
            </div>
            <div>
              <span className="font-semibold text-zinc-100">
                {formatNumber(profile.followerCount)}
              </span>{" "}
              <span className="text-zinc-500">followers</span>
            </div>
            <div>
              <span className="font-semibold text-zinc-100">
                {formatNumber(profile.followingCount)}
              </span>{" "}
              <span className="text-zinc-500">following</span>
            </div>
          </div>

          {/* Connected providers */}
          <div className="mt-3 flex items-center gap-2">
            <span className="inline-flex items-center gap-1 rounded-full bg-zinc-800 px-2.5 py-0.5 text-xs text-zinc-400 ring-1 ring-inset ring-zinc-700">
              <GitFork className="size-3" />
              GitHub
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
