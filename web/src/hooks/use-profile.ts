"use client";

import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
} from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type {
  UserProfile,
  Achievement,
  UserSummary,
  Experience,
  Skill,
} from "@/types";

// --- Profile ---

export function useProfile(username: string) {
  return useQuery({
    queryKey: ["profile", username],
    queryFn: () => api.get<UserProfile>(`/users/${username}`),
    enabled: !!username,
  });
}

// --- Achievements ---

interface AchievementsResponse {
  achievements: Achievement[];
  nextCursor: string | null;
}

export function useProfileAchievements(username: string) {
  return useInfiniteQuery({
    queryKey: ["achievements", username],
    queryFn: ({ pageParam = "0" }) =>
      api.get<AchievementsResponse>(`/users/${username}/achievements`, {
        limit: "20",
        offset: pageParam,
      }),
    initialPageParam: "0",
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    enabled: !!username,
  });
}

// --- Follow / Unfollow ---

export function useFollow() {
  const queryClient = useQueryClient();

  const follow = useMutation({
    mutationFn: (username: string) => api.post(`/users/${username}/follow`),
    onSuccess: (_data, username) => {
      queryClient.invalidateQueries({ queryKey: ["profile", username] });
      queryClient.invalidateQueries({ queryKey: ["followers"] });
      queryClient.invalidateQueries({ queryKey: ["following"] });
    },
  });

  const unfollow = useMutation({
    mutationFn: (username: string) =>
      api.delete(`/users/${username}/follow`),
    onSuccess: (_data, username) => {
      queryClient.invalidateQueries({ queryKey: ["profile", username] });
      queryClient.invalidateQueries({ queryKey: ["followers"] });
      queryClient.invalidateQueries({ queryKey: ["following"] });
    },
  });

  return { follow, unfollow };
}

// --- Followers / Following Lists ---

interface FollowListResponse {
  users: UserSummary[];
  nextCursor: string | null;
}

export function useFollowers(username: string, enabled = true) {
  return useInfiniteQuery({
    queryKey: ["followers", username],
    queryFn: ({ pageParam = "0" }) =>
      api.get<FollowListResponse>(`/users/${username}/followers`, {
        limit: "20",
        offset: pageParam,
      }),
    initialPageParam: "0",
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    enabled: !!username && enabled,
  });
}

export function useFollowing(username: string, enabled = true) {
  return useInfiniteQuery({
    queryKey: ["following", username],
    queryFn: ({ pageParam = "0" }) =>
      api.get<FollowListResponse>(`/users/${username}/following`, {
        limit: "20",
        offset: pageParam,
      }),
    initialPageParam: "0",
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    enabled: !!username && enabled,
  });
}

// --- Update Profile ---

interface UpdateProfileData {
  name?: string;
  bio?: string;
  headline?: string;
  location?: string;
  websiteUrl?: string;
  socialLinks?: Record<string, string>;
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateProfileData) =>
      api.patch<UserProfile>("/users/me", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
  });
}

// --- Experience CRUD ---

interface ExperienceInput {
  companyName: string;
  role: string;
  startDate: string;
  endDate?: string;
  description?: string;
  isCurrent: boolean;
}

export function useCreateExperience() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: ExperienceInput) =>
      api.post<Experience>("/users/me/experiences", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
  });
}

export function useUpdateExperience() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: ExperienceInput & { id: string }) =>
      api.patch<Experience>(`/users/me/experiences/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
  });
}

export function useDeleteExperience() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.delete(`/users/me/experiences/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
  });
}

// --- Skills ---

export function useAddSkill() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (name: string) =>
      api.post<Skill>("/users/me/skills", { name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
  });
}

export function useDeleteSkill() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(`/users/me/skills/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
  });
}
