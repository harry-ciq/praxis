"use client";

import { useAuth } from "@/hooks/use-auth";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { Loader2, Shield, Briefcase, ArrowRight, Code2 } from "lucide-react";
import Link from "next/link";

export default function Home() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace("/feed");
    }
  }, [isLoading, isAuthenticated, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-950">
        <Loader2 className="size-8 animate-spin text-zinc-400" />
      </div>
    );
  }

  if (isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-zinc-950">
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-blue-600/10 via-transparent to-purple-600/10" />
        <div className="relative mx-auto flex max-w-5xl flex-col items-center gap-8 px-4 pb-20 pt-24 text-center sm:pt-32">
          <div className="inline-flex items-center gap-2 rounded-full border border-zinc-800 bg-zinc-900/80 px-4 py-1.5 text-xs text-zinc-400">
            <span className="size-1.5 rounded-full bg-green-500" />
            Now in public beta
          </div>

          <h1 className="max-w-3xl text-4xl font-bold tracking-tight text-zinc-50 sm:text-5xl lg:text-6xl">
            Where Execution Speaks{" "}
            <span className="bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
              Louder Than Words
            </span>
          </h1>

          <p className="max-w-xl text-lg text-zinc-400 sm:text-xl">
            The professional platform where your verified achievements speak for
            themselves. Connect with others who ship, not just talk.
          </p>

          <div className="flex flex-col gap-3 sm:flex-row">
            <Link
              href="/signin"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-blue-600 px-8 text-sm font-medium text-white transition-colors hover:bg-blue-700"
            >
              Get Started
              <ArrowRight className="size-4" />
            </Link>
            <Link
              href="/signin"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-lg border border-zinc-700 px-8 text-sm font-medium text-zinc-300 transition-colors hover:bg-zinc-800 hover:text-zinc-100"
            >
              <Code2 className="size-4" />
              Sign in with GitHub
            </Link>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="border-t border-zinc-800/50 bg-zinc-950 py-20">
        <div className="mx-auto max-w-5xl px-4">
          <h2 className="mb-12 text-center text-3xl font-bold text-zinc-50">
            Built for Builders
          </h2>

          <div className="grid gap-6 sm:grid-cols-3">
            <FeatureCard
              icon={<Code2 className="size-6" />}
              title="Verified Achievements"
              description="Your GitHub commits, stars, and contributions automatically tracked and verified."
            />
            <FeatureCard
              icon={<Briefcase className="size-6" />}
              title="Jobs That Match Your Skills"
              description="Get matched to jobs based on what you've actually built, not what you claim."
            />
            <FeatureCard
              icon={<Shield className="size-6" />}
              title="Immutable Proof"
              description="Every achievement backed by verifiable proof. Coming soon: blockchain verification."
            />
          </div>
        </div>
      </section>

      {/* How it works */}
      <section className="border-t border-zinc-800/50 bg-zinc-900/30 py-20">
        <div className="mx-auto max-w-5xl px-4">
          <h2 className="mb-12 text-center text-3xl font-bold text-zinc-50">
            How It Works
          </h2>

          <div className="grid gap-8 sm:grid-cols-3">
            <StepCard
              step={1}
              title="Connect GitHub"
              description="Link your GitHub account in seconds. We never post on your behalf."
            />
            <StepCard
              step={2}
              title="Achievements Auto-Detected"
              description="Your repos, contributions, and milestones are automatically discovered and verified."
            />
            <StepCard
              step={3}
              title="Get Discovered"
              description="Companies find you based on your real skills and verified track record."
            />
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-zinc-800/50 py-20">
        <div className="mx-auto flex max-w-2xl flex-col items-center gap-6 px-4 text-center">
          <h2 className="text-3xl font-bold text-zinc-50">
            Join the Builders
          </h2>
          <p className="text-zinc-400">
            Stop talking about what you can do. Start proving it.
          </p>
          <Link
            href="/signin"
            className="inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-blue-600 px-8 text-sm font-medium text-white transition-colors hover:bg-blue-700"
          >
            Get Started Free
            <ArrowRight className="size-4" />
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-zinc-800/50 py-8">
        <div className="mx-auto max-w-5xl px-4 text-center">
          <p className="text-sm text-zinc-600">Praxis</p>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <div className="flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-900/50 p-6">
      <div className="flex size-12 items-center justify-center rounded-lg bg-gradient-to-br from-blue-600/20 to-purple-600/20 text-blue-400">
        {icon}
      </div>
      <h3 className="text-lg font-semibold text-zinc-100">{title}</h3>
      <p className="text-sm leading-relaxed text-zinc-400">{description}</p>
    </div>
  );
}

function StepCard({
  step,
  title,
  description,
}: {
  step: number;
  title: string;
  description: string;
}) {
  return (
    <div className="flex flex-col items-center gap-3 text-center">
      <div className="flex size-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-purple-600 text-sm font-bold text-white">
        {step}
      </div>
      <h3 className="text-lg font-semibold text-zinc-100">{title}</h3>
      <p className="text-sm leading-relaxed text-zinc-400">{description}</p>
    </div>
  );
}
