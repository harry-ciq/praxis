"use client";

import { use } from "react";
import { Loader2, User } from "lucide-react";
import { ProfileHeader } from "@/components/profile/profile-header";
import { AchievementCard } from "@/components/achievements/achievement-card";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { useProfile, useProfileAchievements } from "@/hooks/use-profile";

interface ProfilePageProps {
  params: Promise<{ username: string }>;
}

export default function ProfilePage({ params }: ProfilePageProps) {
  const { username } = use(params);
  const { data: profile, isLoading: profileLoading } = useProfile(username);
  const { data: achievementsData, isLoading: achievementsLoading } =
    useProfileAchievements(username);

  if (profileLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="size-6 animate-spin text-zinc-500" />
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="mb-4 rounded-full bg-zinc-800 p-4">
          <User className="size-8 text-zinc-500" />
        </div>
        <h3 className="mb-2 text-lg font-semibold text-zinc-200">
          User not found
        </h3>
        <p className="text-sm text-zinc-500">
          This profile doesn&apos;t exist or has been removed.
        </p>
      </div>
    );
  }

  const achievements = achievementsData?.data ?? [];

  return (
    <div className="mx-auto max-w-2xl px-4 py-6">
      <ProfileHeader profile={profile} />

      <div className="mt-6">
        <Tabs defaultValue="achievements">
          <TabsList variant="line">
            <TabsTrigger value="achievements">Achievements</TabsTrigger>
            <TabsTrigger value="about">About</TabsTrigger>
          </TabsList>

          <TabsContent value="achievements">
            <div className="mt-4 space-y-4">
              {achievementsLoading ? (
                <div className="flex justify-center py-10">
                  <Loader2 className="size-5 animate-spin text-zinc-500" />
                </div>
              ) : achievements.length === 0 ? (
                <div className="py-10 text-center text-sm text-zinc-500">
                  No achievements yet.
                </div>
              ) : (
                achievements.map((achievement) => (
                  <AchievementCard
                    key={achievement.id}
                    achievement={achievement}
                  />
                ))
              )}
            </div>
          </TabsContent>

          <TabsContent value="about">
            <div className="mt-4 space-y-4">
              {profile.bio ? (
                <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-5">
                  <h3 className="mb-2 text-sm font-semibold text-zinc-300">
                    Bio
                  </h3>
                  <p className="text-sm leading-relaxed text-zinc-400">
                    {profile.bio}
                  </p>
                </div>
              ) : (
                <div className="py-10 text-center text-sm text-zinc-500">
                  No bio yet.
                </div>
              )}

              {profile.headline && (
                <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-5">
                  <h3 className="mb-2 text-sm font-semibold text-zinc-300">
                    Headline
                  </h3>
                  <p className="text-sm text-zinc-400">{profile.headline}</p>
                </div>
              )}

              <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-5">
                <h3 className="mb-3 text-sm font-semibold text-zinc-300">
                  Connected Providers
                </h3>
                <div className="flex items-center gap-2">
                  <span className="inline-flex items-center gap-1.5 rounded-full bg-zinc-800 px-3 py-1 text-xs text-zinc-400 ring-1 ring-inset ring-zinc-700">
                    GitHub
                  </span>
                </div>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
