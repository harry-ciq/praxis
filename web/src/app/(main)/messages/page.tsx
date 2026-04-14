"use client";

import { ConversationList } from "@/components/messaging/conversation-list";
import { MessageSquare } from "lucide-react";

export default function MessagesPage() {
  return (
    <div className="flex h-[calc(100vh-3.5rem)]">
      {/* Left panel: conversation list */}
      <div className="w-full border-r border-zinc-800 lg:w-80">
        <ConversationList />
      </div>

      {/* Right panel: placeholder (desktop only) */}
      <div className="hidden flex-1 items-center justify-center lg:flex">
        <div className="flex flex-col items-center gap-3 text-center">
          <MessageSquare className="size-12 text-zinc-700" />
          <p className="text-sm text-zinc-500">
            Select a conversation to start messaging
          </p>
        </div>
      </div>
    </div>
  );
}
