import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { createElement } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ConversationList } from "@/components/messaging/conversation-list";

// Mock next/navigation
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => "/messages",
}));

// Mock next/link
vi.mock("next/link", () => ({
  default: (props: Record<string, unknown>) =>
    createElement("a", { href: props.href as string }, props.children as string),
}));

// Mock Badge
vi.mock("@/components/ui/badge", () => ({
  Badge: (props: Record<string, unknown>) =>
    createElement("span", { className: props.className as string }, props.children as string),
  badgeVariants: () => "",
}));

// Mock Avatar
vi.mock("@/components/ui/avatar", () => ({
  Avatar: (props: Record<string, unknown>) =>
    createElement("div", { "data-testid": "avatar" }, props.children as string),
  AvatarImage: (props: Record<string, unknown>) =>
    createElement("img", { src: props.src as string, alt: props.alt as string }),
  AvatarFallback: (props: Record<string, unknown>) =>
    createElement("span", null, props.children as string),
}));

// Mock Button
vi.mock("@/components/ui/button", () => ({
  Button: (props: Record<string, unknown>) =>
    createElement("button", {}, props.children as string),
  buttonVariants: () => "",
}));

// Mock useAuth
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { id: "me", username: "testuser", name: "Test User" },
    isAuthenticated: true,
  }),
}));

// Mock the API call - data must be inline since vi.mock is hoisted
vi.mock("@/lib/api-client", () => ({
  api: {
    get: vi.fn().mockResolvedValue([
      {
        id: "conv-1",
        participants: [
          { id: "me", username: "testuser", name: "Test User", avatarUrl: null },
          {
            id: "user-2",
            username: "janedoe",
            name: "Jane Doe",
            avatarUrl: "https://example.com/jane.jpg",
          },
        ],
        lastMessageContent: "Hey, that looks amazing!",
        lastMessageSenderId: "user-2",
        lastMessageAt: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
        unreadCount: 2,
        updatedAt: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      },
      {
        id: "conv-2",
        participants: [
          { id: "me", username: "testuser", name: "Test User", avatarUrl: null },
          {
            id: "user-3",
            username: "bobsmith",
            name: "Bob Smith",
            avatarUrl: null,
          },
        ],
        lastMessageContent: null,
        lastMessageSenderId: null,
        lastMessageAt: null,
        unreadCount: 0,
        updatedAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
      },
    ]),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("ConversationList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders conversations", async () => {
    render(createElement(ConversationList), {
      wrapper: createWrapper(),
    });
    expect(await screen.findByText("Jane Doe")).toBeInTheDocument();
    expect(await screen.findByText("Bob Smith")).toBeInTheDocument();
  });

  it("shows unread badge", async () => {
    render(createElement(ConversationList), {
      wrapper: createWrapper(),
    });
    // Wait for data to load and check for the unread count
    expect(await screen.findByText("2")).toBeInTheDocument();
  });

  it("shows last message preview", async () => {
    render(createElement(ConversationList), {
      wrapper: createWrapper(),
    });
    expect(
      await screen.findByText("Hey, that looks amazing!"),
    ).toBeInTheDocument();
  });

  it("shows placeholder for conversations without messages", async () => {
    render(createElement(ConversationList), {
      wrapper: createWrapper(),
    });
    // Wait for Bob Smith's row, which has no messages, then assert placeholder
    await screen.findByText("Bob Smith");
    expect(screen.getAllByText("No messages yet").length).toBeGreaterThan(0);
  });

  it("renders the Messages heading", () => {
    render(createElement(ConversationList), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("Messages")).toBeInTheDocument();
  });
});
