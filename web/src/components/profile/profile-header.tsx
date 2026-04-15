"use client";

import { useState } from "react";
import Link from "next/link";
import {
  GitFork,
  Globe,
  MapPin,
  MessageSquare,
  Pencil,
} from "lucide-react";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useAuth } from "@/hooks/use-auth";
import { useFollow } from "@/hooks/use-profile";
import { formatNumber } from "@/lib/format";
import { EditProfileModal } from "@/components/profile/edit-profile-modal";
import { FollowersModal } from "@/components/profile/followers-modal";
import type { UserProfile } from "@/types";

interface ProfileHeaderProps {
  profile: UserProfile;
}

export function ProfileHeader({ profile }: ProfileHeaderProps) {
  const { user } = useAuth();
  const { follow, unfollow } = useFollow();

  const [editOpen, setEditOpen] = useState(false);
  const [followersOpen, setFollowersOpen] = useState(false);
  const [followersType, setFollowersType] = useState<
    "followers" | "following"
  >("followers");

  const isOwnProfile = user?.username === profile.username;

  const handleFollowToggle = () => {
    if (profile.isFollowing) {
      unfollow.mutate(profile.username);
    } else {
      follow.mutate(profile.username);
    }
  };

  const openFollowersModal = (type: "followers" | "following") => {
    setFollowersType(type);
    setFollowersOpen(true);
  };

  return (
    <>
      <div className="overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/80">
        {/* Banner / Cover */}
        <div className="h-32 bg-gradient-to-r from-emerald-600/40 via-zinc-800 to-zinc-900" />

        {/* Content */}
        <div className="relative px-6 pb-6">
          {/* Avatar overlapping the banner */}
          <div className="-mt-12 mb-4 flex items-end justify-between">
            <Avatar className="size-24 border-4 border-zinc-900 shadow-lg">
              {profile.avatarUrl ? (
                <AvatarImage src={profile.avatarUrl} alt={profile.name} />
              ) : null}
              <AvatarFallback className="text-2xl">
                {profile.name
                  .split(" ")
                  .map((w) => w[0])
                  .join("")
                  .slice(0, 2)
                  .toUpperCase()}
              </AvatarFallback>
            </Avatar>

            {/* Action buttons */}
            <div className="flex items-center gap-2 pt-12">
              {isOwnProfile ? (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setEditOpen(true)}
                >
                  <Pencil className="size-3.5" />
                  Edit Profile
                </Button>
              ) : (
                <>
                  <Button
                    variant={profile.isFollowing ? "outline" : "default"}
                    size="sm"
                    onClick={handleFollowToggle}
                    disabled={follow.isPending || unfollow.isPending}
                  >
                    {profile.isFollowing ? "Unfollow" : "Follow"}
                  </Button>
                  <Link
                    href={`/messages?to=${profile.username}`}
                    className="inline-flex h-7 items-center gap-1 rounded-[min(var(--radius-md),12px)] border border-input bg-input/30 px-2.5 text-[0.8rem] font-medium hover:bg-input/50 [&_svg]:size-3.5"
                  >
                    <MessageSquare className="size-3.5" />
                    Message
                  </Link>
                </>
              )}
            </div>
          </div>

          {/* Name & info */}
          <div>
            <h1 className="text-xl font-bold text-zinc-100">{profile.name}</h1>
            <p className="text-sm text-zinc-500">@{profile.username}</p>
            {profile.headline && (
              <p className="mt-1 text-sm text-zinc-300">{profile.headline}</p>
            )}
          </div>

          {/* Location, website, social */}
          <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-zinc-400">
            {profile.location && (
              <span className="inline-flex items-center gap-1">
                <MapPin className="size-3.5" />
                {profile.location}
              </span>
            )}
            {profile.websiteUrl && (
              <a
                href={profile.websiteUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 hover:text-zinc-200"
              >
                <Globe className="size-3.5" />
                {profile.websiteUrl.replace(/^https?:\/\//, "")}
              </a>
            )}
            {profile.socialLinks?.twitter && (
              <a
                href={profile.socialLinks.twitter}
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-zinc-200"
              >
                Twitter/X
              </a>
            )}
            {profile.socialLinks?.linkedin && (
              <a
                href={profile.socialLinks.linkedin}
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-zinc-200"
              >
                LinkedIn
              </a>
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
            <button
              onClick={() => openFollowersModal("followers")}
              className="hover:underline"
            >
              <span className="font-semibold text-zinc-100">
                {formatNumber(profile.followerCount)}
              </span>{" "}
              <span className="text-zinc-500">followers</span>
            </button>
            <button
              onClick={() => openFollowersModal("following")}
              className="hover:underline"
            >
              <span className="font-semibold text-zinc-100">
                {formatNumber(profile.followingCount)}
              </span>{" "}
              <span className="text-zinc-500">following</span>
            </button>
          </div>

          {/* Connected provider badges */}
          {profile.providers.length > 0 && (
            <div className="mt-3 flex items-center gap-2">
              {profile.providers.map((p) => (
                <Badge key={p.id} variant="outline" className="gap-1">
                  <GitFork className="size-3" />
                  {p.provider}
                </Badge>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Modals */}
      <EditProfileModal
        open={editOpen}
        onClose={() => setEditOpen(false)}
        profile={profile}
      />
      <FollowersModal
        open={followersOpen}
        onClose={() => setFollowersOpen(false)}
        username={profile.username}
        type={followersType}
      />
    </>
  );
}
