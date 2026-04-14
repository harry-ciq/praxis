"use client";

import { use } from "react";
import { ConversationList } from "@/components/messaging/conversation-list";
import { ChatThread } from "@/components/messaging/chat-thread";

export default function ConversationPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);

  return (
    <div className="flex h-[calc(100vh-3.5rem)]">
      {/* Left panel: conversation list (hidden on mobile) */}
      <div className="hidden w-80 border-r border-zinc-800 lg:block">
        <ConversationList activeId={id} />
      </div>

      {/* Right panel: chat thread */}
      <div className="flex-1">
        <ChatThread conversationId={id} />
      </div>
    </div>
  );
}
