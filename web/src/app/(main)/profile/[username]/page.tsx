"use client";

import { use, useEffect, useRef, useCallback } from "react";
import { Calendar, Loader2, User } from "lucide-react";
import { ProfileHeader } from "@/components/profile/profile-header";
import { ExperienceSection } from "@/components/profile/experience-section";
import { SkillsSection } from "@/components/profile/skills-section";
import { ConnectedProvidersSection } from "@/components/profile/connected-providers-section";
import { AchievementCard } from "@/components/achievements/achievement-card";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  useProfile,
  useProfileAchievements,
} from "@/hooks/use-profile";
import { useAuth } from "@/hooks/use-auth";

interface ProfilePageProps {
  params: Promise<{ username: string }>;
}

export default function ProfilePage({ params }: ProfilePageProps) {
  const { username } = use(params);
  const { user } = useAuth();
  const { data: profile, isLoading: profileLoading } = useProfile(username);
  const achievements = useProfileAchievements(username);

  const isOwnProfile = user?.username === profile?.username;

  // Infinite scroll sentinel for achievements
  const sentinelRef = useRef<HTMLDivElement>(null);

  const handleObserver = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;
      if (
        entry.isIntersecting &&
        achievements.hasNextPage &&
        !achievements.isFetchingNextPage
      ) {
        achievements.fetchNextPage();
      }
    },
    [achievements]
  );

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(handleObserver, {
      threshold: 0.1,
    });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [handleObserver]);

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

  const allAchievements =
    achievements.data?.pages.flatMap((p) => p.achievements) ?? [];

  return (
    <div className="mx-auto max-w-2xl px-4 py-6">
      <ProfileHeader profile={profile} />

      <div className="mt-6">
        <Tabs defaultValue="achievements">
          <TabsList variant="line">
            <TabsTrigger value="achievements">Achievements</TabsTrigger>
            <TabsTrigger value="experience">Experience</TabsTrigger>
            <TabsTrigger value="skills">Skills</TabsTrigger>
            <TabsTrigger value="about">About</TabsTrigger>
          </TabsList>

          {/* Achievements Tab */}
          <TabsContent value="achievements">
            <div className="mt-4 space-y-4">
              {achievements.isLoading ? (
                <div className="flex justify-center py-10">
                  <Loader2 className="size-5 animate-spin text-zinc-500" />
                </div>
              ) : allAchievements.length === 0 ? (
                <div className="py-10 text-center text-sm text-zinc-500">
                  No achievements yet.
                </div>
              ) : (
                <>
                  {allAchievements.map((achievement) => (
                    <AchievementCard
                      key={achievement.id}
                      achievement={achievement}
                    />
                  ))}
                  <div ref={sentinelRef} className="h-4" />
                  {achievements.isFetchingNextPage && (
                    <div className="flex justify-center py-3">
                      <Loader2 className="size-4 animate-spin text-zinc-500" />
                    </div>
                  )}
                </>
              )}
            </div>
          </TabsContent>

          {/* Experience Tab */}
          <TabsContent value="experience">
            <div className="mt-4 rounded-2xl border border-zinc-800 bg-zinc-900/80 p-5">
              <h3 className="mb-4 text-sm font-semibold text-zinc-300">
                Experience
              </h3>
              <ExperienceSection
                experiences={profile.experiences}
                isOwnProfile={isOwnProfile}
              />
            </div>
          </TabsContent>

          {/* Skills Tab */}
          <TabsContent value="skills">
            <div className="mt-4 rounded-2xl border border-zinc-800 bg-zinc-900/80 p-5">
              <h3 className="mb-4 text-sm font-semibold text-zinc-300">
                Skills
              </h3>
              <SkillsSection
                skills={profile.skills}
                isOwnProfile={isOwnProfile}
              />
            </div>
          </TabsContent>

          {/* About Tab */}
          <TabsContent value="about">
            <div className="mt-4 space-y-4">
              {/* Bio */}
              <div className="rounded-2xl border border-zinc-800 bg-zinc-900/80 p-5">
                <h3 className="mb-2 text-sm font-semibold text-zinc-300">
                  Bio
                </h3>
                {profile.bio ? (
                  <p className="text-sm leading-relaxed text-zinc-400">
                    {profile.bio}
                  </p>
                ) : (
                  <p className="text-sm text-zinc-500">No bio yet.</p>
                )}
              </div>

              {/* Connected Providers */}
              <div className="rounded-2xl border border-zinc-800 bg-zinc-900/80 p-5">
                <h3 className="mb-3 text-sm font-semibold text-zinc-300">
                  Connected Providers
                </h3>
                <ConnectedProvidersSection
                  providers={profile.providers}
                  isOwnProfile={isOwnProfile}
                />
              </div>

              {/* Member since */}
              <div className="rounded-2xl border border-zinc-800 bg-zinc-900/80 p-5">
                <h3 className="mb-2 text-sm font-semibold text-zinc-300">
                  Member Since
                </h3>
                <p className="flex items-center gap-1.5 text-sm text-zinc-400">
                  <Calendar className="size-3.5" />
                  {new Date(profile.createdAt).toLocaleDateString("en-US", {
                    month: "long",
                    year: "numeric",
                  })}
                </p>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
