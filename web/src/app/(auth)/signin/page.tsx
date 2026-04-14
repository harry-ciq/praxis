"use client";

import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { useState } from "react";

function GitHubIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
    >
      <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
    </svg>
  );
}

export default function SignInPage() {
  const { login } = useAuth();
  const [isRedirecting, setIsRedirecting] = useState(false);

  const handleLogin = async () => {
    setIsRedirecting(true);
    try {
      await login();
    } catch {
      setIsRedirecting(false);
    }
  };

  return (
    <div className="flex w-full max-w-md flex-col items-center gap-8 px-4">
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-4xl font-bold tracking-tight text-zinc-50">
          Praxis
        </h1>
        <p className="text-sm text-zinc-400">
          Where execution speaks louder than words
        </p>
      </div>

      <Card className="w-full border-zinc-800 bg-zinc-900/80 backdrop-blur">
        <CardHeader className="text-center">
          <CardTitle className="text-xl text-zinc-100">Welcome back</CardTitle>
          <CardDescription className="text-zinc-400">
            Your achievements. Verified. Immutable.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <Button
            size="lg"
            className="h-12 w-full gap-2 bg-zinc-50 text-zinc-900 hover:bg-zinc-200"
            onClick={handleLogin}
            disabled={isRedirecting}
          >
            {isRedirecting ? (
              <Loader2 className="size-5 animate-spin" />
            ) : (
              <GitHubIcon className="size-5" />
            )}
            Continue with GitHub
          </Button>
          <p className="text-center text-xs text-zinc-500">
            By continuing, you agree to our Terms of Service and Privacy Policy.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
