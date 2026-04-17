import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { createElement } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AchievementCard } from "@/components/achievements/achievement-card";
import type { Achievement } from "@/types";

// Mock next/navigation
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => "/feed",
}));

// Mock next/link using createElement to avoid JSX issues
vi.mock("next/link", () => ({
  default: (props: Record<string, unknown>) =>
    createElement("a", { href: props.href as string }, props.children as string),
}));

// Mock the Badge component (uses base-ui useRender)
vi.mock("@/components/ui/badge", () => ({
  Badge: (props: Record<string, unknown>) =>
    createElement("span", { className: props.className as string }, props.children as string),
  badgeVariants: () => "",
}));

// Mock the Avatar components (base-ui primitives don't work well in jsdom)
vi.mock("@/components/ui/avatar", () => ({
  Avatar: (props: Record<string, unknown>) =>
    createElement("div", { "data-testid": "avatar" }, props.children as string),
  AvatarImage: (props: Record<string, unknown>) =>
    createElement("img", { src: props.src as string, alt: props.alt as string }),
  AvatarFallback: (props: Record<string, unknown>) =>
    createElement("span", null, props.children as string),
  AvatarBadge: (props: Record<string, unknown>) =>
    createElement("span", null, props.children as string),
  AvatarGroup: (props: Record<string, unknown>) =>
    createElement("div", null, props.children as string),
  AvatarGroupCount: (props: Record<string, unknown>) =>
    createElement("div", null, props.children as string),
}));

// Mock useAuth
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { id: "me", username: "testuser", name: "Test User" },
    isAuthenticated: true,
  }),
}));

const mockAchievement: Achievement = {
  id: "ach-1",
  userId: "user-1",
  user: {
    username: "janedoe",
    name: "Jane Doe",
    avatarUrl: "https://example.com/avatar.jpg",
  },
  type: "REPO_CREATED",
  title: "Created awesome-project",
  description: "A blazing fast CLI tool for developers",
  metadata: {},
  proofUrl: "https://github.com/janedoe/awesome-project",
  source: "GITHUB",
  sourceId: "repo-123",
  verificationHash: "abc123hash",
  status: "active",
  createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2h ago
  reactions: { clap: 5, fire: 3, rocket: 1, total: 9 },
  userReaction: null,
};

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("AchievementCard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the achievement title", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("Created awesome-project")).toBeInTheDocument();
  });

  it("renders user info", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
    expect(screen.getByText("@janedoe")).toBeInTheDocument();
  });

  it("renders the type badge", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("New Repository")).toBeInTheDocument();
  });

  it("renders description when present", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(
      screen.getByText("A blazing fast CLI tool for developers")
    ).toBeInTheDocument();
  });

  it("shows proof link when proofUrl is present", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    const proofLink = screen.getByText("View on GitHub");
    expect(proofLink).toBeInTheDocument();
    expect(proofLink.closest("a")).toHaveAttribute(
      "href",
      "https://github.com/janedoe/awesome-project"
    );
  });

  it("does not show proof link when proofUrl is null", () => {
    const achievement = { ...mockAchievement, proofUrl: null };
    render(createElement(AchievementCard, { achievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.queryByText("View on GitHub")).not.toBeInTheDocument();
  });

  it("shows correct reaction counts", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("5")).toBeInTheDocument(); // clap count
    expect(screen.getByText("3")).toBeInTheDocument(); // fire count
    expect(screen.getByText("1")).toBeInTheDocument(); // rocket count
  });

  it("renders relative timestamp", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("2h ago")).toBeInTheDocument();
  });

  it("shows verification indicator when verificationHash is present", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("Verified")).toBeInTheDocument();
  });

  it("shows GitHub source badge", () => {
    render(createElement(AchievementCard, { achievement: mockAchievement }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("GitHub")).toBeInTheDocument();
  });
});
