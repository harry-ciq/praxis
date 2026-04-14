"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { useAuth } from "@/hooks/use-auth";
import { useSocket } from "@/hooks/use-socket";
import { MessageBubble } from "@/components/messaging/message-bubble";
import {
  Avatar,
  AvatarImage,
  AvatarFallback,
} from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ArrowLeft, Send, Loader2 } from "lucide-react";
import Link from "next/link";
import type { Message, Conversation } from "@/types";

interface ChatThreadProps {
  conversationId: string;
}

export function ChatThread({ conversationId }: ChatThreadProps) {
  const { user } = useAuth();
  const { socket } = useSocket();
  const queryClient = useQueryClient();
  const [newMessage, setNewMessage] = useState("");
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const { data: conversation } = useQuery({
    queryKey: ["conversation", conversationId],
    queryFn: () => api.get<Conversation>(`/conversations/${conversationId}`),
    enabled: false,
  });

  const { data: messages = [], isLoading } = useQuery({
    queryKey: ["messages", conversationId],
    queryFn: () =>
      api.get<Message[]>(`/conversations/${conversationId}/messages`),
  });

  const otherParticipant = conversation?.participants.find(
    (p) => p.id !== user?.id,
  ) ?? conversation?.participants[0];

  // Mark as read on mount
  useEffect(() => {
    api.patch(`/conversations/${conversationId}/read`).catch(() => {});
    queryClient.invalidateQueries({ queryKey: ["conversations"] });
  }, [conversationId, queryClient]);

  // Listen for new messages via WebSocket
  useEffect(() => {
    const unsubscribe = socket.on("message", (data) => {
      const msg = data as Message;
      if (msg.conversationId === conversationId) {
        queryClient.setQueryData<Message[]>(
          ["messages", conversationId],
          (old) => (old ? [...old, msg] : [msg]),
        );
        // Mark as read since we're viewing this conversation
        api.patch(`/conversations/${conversationId}/read`).catch(() => {});
      }
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
    });

    return () => { unsubscribe(); };
  }, [socket, conversationId, queryClient]);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = useCallback(async () => {
    const content = newMessage.trim();
    if (!content) return;

    setSending(true);
    try {
      const msg = await api.post<Message>(
        `/conversations/${conversationId}/messages`,
        { content },
      );
      queryClient.setQueryData<Message[]>(
        ["messages", conversationId],
        (old) => (old ? [...old, msg] : [msg]),
      );
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
      setNewMessage("");
    } finally {
      setSending(false);
    }
  }, [newMessage, conversationId, queryClient]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

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
      {/* Header */}
      <div className="flex items-center gap-3 border-b border-zinc-800 px-4 py-3">
        <Link
          href="/messages"
          className="inline-flex size-8 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-200 lg:hidden"
        >
          <ArrowLeft className="size-5" />
        </Link>
        {otherParticipant && (
          <>
            <Avatar size="sm">
              {otherParticipant.avatarUrl && (
                <AvatarImage
                  src={otherParticipant.avatarUrl}
                  alt={otherParticipant.name}
                />
              )}
              <AvatarFallback className="bg-zinc-700 text-xs text-zinc-300">
                {getInitials(otherParticipant.name)}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="text-sm font-medium text-zinc-100">
                {otherParticipant.name}
              </p>
              <p className="text-xs text-zinc-500">
                @{otherParticipant.username}
              </p>
            </div>
          </>
        )}
      </div>

      {/* Messages */}
      <div
        ref={containerRef}
        className="flex flex-1 flex-col gap-2 overflow-y-auto px-4 py-4"
      >
        {isLoading ? (
          <div className="flex flex-1 items-center justify-center">
            <Loader2 className="size-6 animate-spin text-zinc-500" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-1 items-center justify-center">
            <p className="text-sm text-zinc-500">
              No messages yet. Say hello!
            </p>
          </div>
        ) : (
          messages.map((message) => (
            <MessageBubble
              key={message.id}
              message={message}
              isOwn={message.senderId === user?.id}
            />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="border-t border-zinc-800 px-4 py-3">
        <div className="flex items-center gap-2">
          <Input
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a message..."
            className="flex-1 border-zinc-700 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500"
          />
          <Button
            onClick={handleSend}
            disabled={!newMessage.trim() || sending}
            size="icon"
            className="shrink-0 bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {sending ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Send className="size-4" />
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
