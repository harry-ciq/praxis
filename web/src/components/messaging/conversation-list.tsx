"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { useAuth } from "@/hooks/use-auth";
import { formatRelativeTime } from "@/lib/format";
import { cn } from "@/lib/utils";
import {
  Avatar,
  AvatarImage,
  AvatarFallback,
} from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { MessageSquarePlus } from "lucide-react";
import Link from "next/link";
import type { Conversation } from "@/types";

interface ConversationListProps {
  activeId?: string;
}

export function ConversationList({ activeId }: ConversationListProps) {
  const { user } = useAuth();

  const { data: conversations = [] } = useQuery({
    queryKey: ["conversations"],
    queryFn: () => api.get<Conversation[]>("/conversations"),
  });

  function getOtherParticipant(conversation: Conversation) {
    return conversation.participants.find((p) => p.id !== user?.id) ?? conversation.participants[0];
  }

  function getInitials(name: string) {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-3">
        <h2 className="text-lg font-semibold text-zinc-50">Messages</h2>
        <Button variant="ghost" size="icon" className="text-zinc-400 hover:text-zinc-200">
          <MessageSquarePlus className="size-5" />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {conversations.length === 0 ? (
          <div className="flex flex-col items-center justify-center px-4 py-16 text-center">
            <MessageSquarePlus className="mb-3 size-10 text-zinc-600" />
            <p className="text-sm text-zinc-500">No conversations yet</p>
          </div>
        ) : (
          <div className="flex flex-col">
            {conversations.map((conversation) => {
              const other = getOtherParticipant(conversation);
              const isActive = conversation.id === activeId;
              const preview = conversation.lastMessage?.content ?? "";
              const truncated =
                preview.length > 50 ? preview.slice(0, 50) + "..." : preview;

              return (
                <Link
                  key={conversation.id}
                  href={`/messages/${conversation.id}`}
                  className={cn(
                    "flex items-center gap-3 px-4 py-3 transition-colors hover:bg-zinc-800/50",
                    isActive && "bg-zinc-800",
                  )}
                >
                  <Avatar>
                    {other.avatarUrl && (
                      <AvatarImage src={other.avatarUrl} alt={other.name} />
                    )}
                    <AvatarFallback className="bg-zinc-700 text-xs text-zinc-300">
                      {getInitials(other.name)}
                    </AvatarFallback>
                  </Avatar>

                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between">
                      <p className="truncate text-sm font-medium text-zinc-100">
                        {other.name}
                      </p>
                      {conversation.lastMessage && (
                        <span className="ml-2 shrink-0 text-[11px] text-zinc-500">
                          {formatRelativeTime(conversation.updatedAt)}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <p className="truncate text-xs text-zinc-500">
                        {truncated || "No messages yet"}
                      </p>
                      {conversation.unreadCount > 0 && (
                        <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-blue-600 text-[10px] font-bold text-white">
                          {conversation.unreadCount}
                        </span>
                      )}
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
