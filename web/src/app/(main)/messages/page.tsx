"use client";

import { useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { ConversationList } from "@/components/messaging/conversation-list";
import { MessageSquare, Loader2 } from "lucide-react";
import { api } from "@/lib/api-client";

export default function MessagesPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const toUsername = searchParams.get("to");
  const [redirecting, setRedirecting] = useState(false);

  // If ?to=username is present, create/get a conversation and redirect to it
  useEffect(() => {
    if (!toUsername) return;

    setRedirecting(true);
    api
      .post<{ id: string }>("/conversations", { username: toUsername })
      .then((res) => {
        router.replace(`/messages/${res.id}`);
      })
      .catch(() => {
        setRedirecting(false);
      });
  }, [toUsername, router]);

  return (
    <div className="flex h-[calc(100vh-3.5rem)]">
      {/* Left panel: conversation list */}
      <div className="w-full border-r border-zinc-800 lg:w-80">
        <ConversationList />
      </div>

      {/* Right panel: placeholder (desktop only) */}
      <div className="hidden flex-1 items-center justify-center lg:flex">
        {redirecting ? (
          <div className="flex flex-col items-center gap-3 text-center">
            <Loader2 className="size-8 animate-spin text-zinc-500" />
            <p className="text-sm text-zinc-500">Opening conversation...</p>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-3 text-center">
            <MessageSquare className="size-12 text-zinc-700" />
            <p className="text-sm text-zinc-500">
              Select a conversation to start messaging
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
