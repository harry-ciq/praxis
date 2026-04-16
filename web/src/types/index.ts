// ============================================================
// Praxis — Shared frontend types
// ============================================================

// --- Auth ---
export interface AuthUser {
  id: string;
  username: string;
  name: string;
  email: string;
  avatarUrl: string | null;
  headline: string | null;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

// --- Users ---
export interface UserProfile {
  id: string;
  username: string;
  name: string;
  email: string;
  bio: string | null;
  avatarUrl: string | null;
  headline: string | null;
  location: string | null;
  websiteUrl: string | null;
  socialLinks: Record<string, string>;
  createdAt: string;
  achievementCount: number;
  followerCount: number;
  followingCount: number;
  isFollowing: boolean;
  experiences: Experience[];
  skills: Skill[];
  providers: ConnectedProvider[];
}

export interface Experience {
  id: string;
  userId: string;
  companyName: string;
  role: string;
  startDate: string;
  endDate: string | null;
  description: string | null;
  isCurrent: boolean;
  createdAt: string;
}

export interface Skill {
  id: string;
  userId: string;
  name: string;
  verified: boolean;
  source: string;
  createdAt: string;
}

export interface ConnectedProvider {
  id: string;
  userId: string;
  provider: string;
  providerUsername: string;
  lastSyncedAt: string | null;
  syncStatus: string;
  createdAt: string;
}

export interface UserSummary {
  id: string;
  username: string;
  name: string;
  avatarUrl: string | null;
  headline: string | null;
  isFollowing: boolean;
}

// --- Achievements ---
export type AchievementType =
  | "REPO_CREATED"
  | "STARS_MILESTONE"
  | "COMMIT_STREAK"
  | "PR_MERGED"
  | "FIRST_CONTRIBUTION"
  | "WEEKLY_COMMITS"
  | "VIDEO_PUBLISHED"
  | "SUBSCRIBERS_MILESTONE"
  | "VIEWS_MILESTONE";

export type Provider = "GITHUB" | "YOUTUBE";

export type AchievementStatus = "active" | "archived";

export type ReactionType = "CLAP" | "FIRE" | "ROCKET";

export interface Achievement {
  id: string;
  userId: string;
  user: {
    username: string;
    name: string;
    avatarUrl: string | null;
  };
  type: AchievementType;
  title: string;
  description: string | null;
  metadata: Record<string, unknown>;
  proofUrl: string | null;
  source: Provider;
  sourceId: string;
  verificationHash: string | null;
  status: AchievementStatus;
  createdAt: string;
  reactions: ReactionSummary;
  userReaction: ReactionType | null;
}

export interface ReactionSummary {
  clap: number;
  fire: number;
  rocket: number;
  total: number;
}

// --- Feed ---
export interface FeedResponse {
  achievements: Achievement[];
  nextCursor: string | null;
}

// --- Messaging ---
export interface ConversationParticipant {
  id: string;
  username: string;
  name: string;
  avatarUrl: string | null;
}

export interface Conversation {
  id: string;
  participants: ConversationParticipant[];
  lastMessageContent: string | null;
  lastMessageSenderId: string | null;
  lastMessageAt: string | null;
  unreadCount: number;
  updatedAt: string;
}

export interface Message {
  id: string;
  conversationId: string;
  senderId: string;
  content: string;
  createdAt: string;
}

// --- Jobs ---
export type JobType = "FULL_TIME" | "PART_TIME" | "CONTRACT" | "REMOTE";
export type JobStatus = "ACTIVE" | "CLOSED" | "DRAFT";
export type ApplicationStatus = "PENDING" | "REVIEWED" | "ACCEPTED" | "REJECTED";

export interface Job {
  id: string;
  company: Company;
  title: string;
  description: string;
  location: string;
  jobType: JobType;
  salaryRange: string | null;
  requiredAchievements: string[];
  status: JobStatus;
  createdAt: string;
}

export interface Company {
  id: string;
  name: string;
  logoUrl: string | null;
  description: string | null;
  website: string | null;
}

export interface JobApplication {
  id: string;
  jobId: string;
  userId: string;
  coverNote: string | null;
  status: ApplicationStatus;
  createdAt: string;
}

// --- Notifications ---
export type NotificationType =
  | "NEW_FOLLOWER"
  | "REACTION"
  | "MESSAGE"
  | "JOB_MATCH"
  | "ACHIEVEMENT";

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  body: string;
  data: Record<string, unknown>;
  read: boolean;
  createdAt: string;
}

// --- API ---
export interface ApiError {
  error: string;
  message: string;
  statusCode: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  nextCursor: string | null;
  totalCount?: number;
}
