import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { createElement } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { JobCard } from "@/components/jobs/job-card";
import type { Job } from "@/types";

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

// Mock Button
vi.mock("@/components/ui/button", () => ({
  Button: ({
    children,
    render: renderProp,
    ...props
  }: Record<string, unknown>) => {
    if (renderProp && typeof renderProp === "object" && renderProp !== null) {
      const linkProps = renderProp as { props?: { href?: string } };
      return createElement(
        "a",
        { href: linkProps.props?.href },
        children as string,
      );
    }
    return createElement("button", props, children as string);
  },
  buttonVariants: () => "",
}));

const mockJob: Job = {
  id: "job-1",
  company: {
    id: "company-1",
    name: "Acme Corp",
    logoUrl: "https://example.com/logo.png",
    description: "A great company",
    website: "https://acme.com",
  },
  title: "Senior Frontend Engineer",
  description: "Build amazing products.",
  location: "San Francisco, CA",
  jobType: "FULL_TIME",
  salaryRange: "$150k - $200k",
  requiredAchievements: ["30-day streak", "100+ stars"],
  status: "ACTIVE",
  createdAt: new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(),
};

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("JobCard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders job title", () => {
    render(createElement(JobCard, { job: mockJob }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("Senior Frontend Engineer")).toBeInTheDocument();
  });

  it("renders company name", () => {
    render(createElement(JobCard, { job: mockJob }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("Acme Corp")).toBeInTheDocument();
  });

  it("renders location", () => {
    render(createElement(JobCard, { job: mockJob }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("San Francisco, CA")).toBeInTheDocument();
  });

  it("shows required achievements as badges", () => {
    render(createElement(JobCard, { job: mockJob }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("30-day streak")).toBeInTheDocument();
    expect(screen.getByText("100+ stars")).toBeInTheDocument();
  });

  it("shows salary range when present", () => {
    render(createElement(JobCard, { job: mockJob }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("$150k - $200k")).toBeInTheDocument();
  });

  it("does not show salary when null", () => {
    const jobWithoutSalary = { ...mockJob, salaryRange: null };
    render(createElement(JobCard, { job: jobWithoutSalary }), {
      wrapper: createWrapper(),
    });
    expect(screen.queryByText("$150k - $200k")).not.toBeInTheDocument();
  });

  it("shows job type badge", () => {
    render(createElement(JobCard, { job: mockJob }), {
      wrapper: createWrapper(),
    });
    expect(screen.getByText("Full Time")).toBeInTheDocument();
  });
});
