import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { createElement } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactionBar } from "@/components/achievements/reaction-bar";
import type { ReactionSummary, ReactionType } from "@/types";

const mockPost = vi.fn().mockResolvedValue({});
const mockDelete = vi.fn().mockResolvedValue({});

// Mock the API client
vi.mock("@/lib/api-client", () => ({
  api: {
    get: vi.fn(),
    post: (...args: unknown[]) => mockPost(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
  },
}));

const mockReactions: ReactionSummary = {
  clap: 10,
  fire: 5,
  rocket: 2,
  total: 17,
};

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("ReactionBar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders three reaction buttons", () => {
    render(
      createElement(ReactionBar, {
        achievementId: "ach-1",
        reactions: mockReactions,
        userReaction: null,
      }),
      { wrapper: createWrapper() }
    );

    expect(screen.getByLabelText("Clap (10)")).toBeInTheDocument();
    expect(screen.getByLabelText("Fire (5)")).toBeInTheDocument();
    expect(screen.getByLabelText("Rocket (2)")).toBeInTheDocument();
  });

  it("shows reaction counts", () => {
    render(
      createElement(ReactionBar, {
        achievementId: "ach-1",
        reactions: mockReactions,
        userReaction: null,
      }),
      { wrapper: createWrapper() }
    );

    expect(screen.getByText("10")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("highlights the user's current reaction", () => {
    render(
      createElement(ReactionBar, {
        achievementId: "ach-1",
        reactions: mockReactions,
        userReaction: "FIRE" as ReactionType,
      }),
      { wrapper: createWrapper() }
    );

    const fireButton = screen.getByLabelText("Fire (5)");
    // Active reaction is highlighted in orange (its activeClass)
    expect(fireButton.className).toContain("orange");
  });

  it("calls API when clicking a reaction", async () => {
    render(
      createElement(ReactionBar, {
        achievementId: "ach-1",
        reactions: mockReactions,
        userReaction: null,
      }),
      { wrapper: createWrapper() }
    );

    const clapButton = screen.getByLabelText("Clap (10)");
    fireEvent.click(clapButton);

    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledWith("/achievements/ach-1/reactions", {
        type: "CLAP",
      });
    });
  });

  it("calls delete API when clicking active reaction to toggle off", async () => {
    render(
      createElement(ReactionBar, {
        achievementId: "ach-1",
        reactions: mockReactions,
        userReaction: "CLAP" as ReactionType,
      }),
      { wrapper: createWrapper() }
    );

    const clapButton = screen.getByLabelText("Clap (10)");
    fireEvent.click(clapButton);

    await waitFor(() => {
      expect(mockDelete).toHaveBeenCalledWith(
        "/achievements/ach-1/reactions"
      );
    });
  });

  it("does not show count when reaction count is 0", () => {
    const reactions: ReactionSummary = {
      clap: 0,
      fire: 3,
      rocket: 0,
      total: 3,
    };

    render(
      createElement(ReactionBar, {
        achievementId: "ach-1",
        reactions,
        userReaction: null,
      }),
      { wrapper: createWrapper() }
    );

    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByLabelText("Clap (0)")).toBeInTheDocument();
    expect(screen.getByLabelText("Rocket (0)")).toBeInTheDocument();
  });
});
