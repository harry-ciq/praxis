"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api-client";
import { useAuth } from "@/hooks/use-auth";
import { useSearchUsers } from "@/hooks/use-search";
import { formatRelativeTime } from "@/lib/format";
import { cn } from "@/lib/utils";
import {
  Avatar,
  AvatarImage,
  AvatarFallback,
} from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MessageSquarePlus, Search, X, Loader2 } from "lucide-react";
import Link from "next/link";
import type { Conversation } from "@/types";

interface ConversationListProps {
  activeId?: string;
}

export function ConversationList({ activeId }: ConversationListProps) {
  const { user } = useAuth();
  const router = useRouter();
  const [showNewChat, setShowNewChat] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [creating, setCreating] = useState(false);

  const { data: conversations = [] } = useQuery({
    queryKey: ["conversations"],
    queryFn: () => api.get<Conversation[]>("/conversations"),
  });

  const { data: searchResults, isLoading: searching } =
    useSearchUsers(searchQuery);

  function getOtherParticipant(conversation: Conversation) {
    return (
      conversation.participants.find((p) => p.id !== user?.id) ??
      conversation.participants[0]
    );
  }

  function getInitials(name: string) {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  }

  async function startConversation(username: string) {
    setCreating(true);
    try {
      const res = await api.post<{ id: string }>("/conversations", {
        username,
      });
      setShowNewChat(false);
      setSearchQuery("");
      router.push(`/messages/${res.id}`);
    } catch {
      // handle error
    } finally {
      setCreating(false);
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-3">
        <h2 className="text-lg font-semibold text-zinc-50">Messages</h2>
        <Button
          variant="ghost"
          size="icon"
          className="text-zinc-400 hover:text-zinc-200"
          onClick={() => {
            setShowNewChat(!showNewChat);
            setSearchQuery("");
          }}
        >
          {showNewChat ? (
            <X className="size-5" />
          ) : (
            <MessageSquarePlus className="size-5" />
          )}
        </Button>
      </div>

      {/* New conversation search */}
      {showNewChat && (
        <div className="border-b border-zinc-800 p-3">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-zinc-500" />
            <Input
              placeholder="Search users to message..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-8 border-zinc-700 bg-zinc-900 pl-8 text-sm text-zinc-200 placeholder:text-zinc-500"
              autoFocus
            />
          </div>
          {searchQuery.length >= 2 && (
            <div className="mt-2 max-h-48 overflow-y-auto">
              {searching ? (
                <div className="flex justify-center py-3">
                  <Loader2 className="size-4 animate-spin text-zinc-500" />
                </div>
              ) : !searchResults?.users.length ? (
                <p className="py-3 text-center text-xs text-zinc-500">
                  No users found
                </p>
              ) : (
                searchResults.users
                  .filter((u) => u.username !== user?.username)
                  .map((u) => (
                    <button
                      key={u.id}
                      onClick={() => startConversation(u.username)}
                      disabled={creating}
                      className="flex w-full items-center gap-2.5 rounded-lg px-2 py-2 text-left transition-colors hover:bg-zinc-800/70 disabled:opacity-50"
                    >
                      <Avatar size="sm">
                        {u.avatarUrl ? (
                          <AvatarImage src={u.avatarUrl} alt={u.name} />
                        ) : null}
                        <AvatarFallback className="text-xs">
                          {getInitials(u.name)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-zinc-100">
                          {u.name}
                        </p>
                        <p className="truncate text-xs text-zinc-500">
                          @{u.username}
                        </p>
                      </div>
                    </button>
                  ))
              )}
            </div>
          )}
        </div>
      )}

      <div className="flex-1 overflow-y-auto">
        {conversations.length === 0 && !showNewChat ? (
          <div className="flex flex-col items-center justify-center px-4 py-16 text-center">
            <MessageSquarePlus className="mb-3 size-10 text-zinc-600" />
            <p className="text-sm text-zinc-500">No conversations yet</p>
            <p className="mt-1 text-xs text-zinc-600">
              Click the + button to start a new chat
            </p>
          </div>
        ) : (
          <div className="flex flex-col">
            {conversations.map((conversation) => {
              const other = getOtherParticipant(conversation);
              const isActive = conversation.id === activeId;
              const preview = conversation.lastMessageContent ?? "";
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
                      {conversation.lastMessageAt && (
                        <span className="ml-2 shrink-0 text-[11px] text-zinc-500">
                          {formatRelativeTime(conversation.lastMessageAt)}
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
