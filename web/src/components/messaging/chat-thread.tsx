"use client";

import { useEffect, useRef, useState, useCallback, useMemo } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api-client";
import { useAuth } from "@/hooks/use-auth";
import { socket } from "@/lib/socket";
import { MessageBubble } from "@/components/messaging/message-bubble";
import {
  Avatar,
  AvatarImage,
  AvatarFallback,
} from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ArrowLeft, Send, Loader2, Trash2 } from "lucide-react";
import Link from "next/link";
import type { Message, Conversation } from "@/types";

interface ChatThreadProps {
  conversationId: string;
}

export function ChatThread({ conversationId }: ChatThreadProps) {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const router = useRouter();
  const [newMessage, setNewMessage] = useState("");
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Track locally-added messages (from send + WebSocket) separately
  // so we don't fight with React Query's reference identity
  const [localMessages, setLocalMessages] = useState<Message[]>([]);

  // Typing indicator state — set when other user is typing, auto-clears
  const [otherTyping, setOtherTyping] = useState(false);
  const typingClearRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastTypingSentRef = useRef<number>(0);

  // Fetch conversations list to get participant info
  const { data: conversations } = useQuery({
    queryKey: ["conversations"],
    queryFn: () => api.get<Conversation[]>("/conversations"),
  });

  const currentConversation = useMemo(
    () => (conversations ?? []).find((c) => c.id === conversationId),
    [conversations, conversationId],
  );

  const otherParticipant = useMemo(() => {
    if (!currentConversation) return null;
    return (
      currentConversation.participants.find((p) => p.id !== user?.id) ??
      currentConversation.participants[0] ??
      null
    );
  }, [currentConversation, user?.id]);

  // Fetch messages — backend returns DESC, reverse for chronological display
  const { data: fetchedMessages, isLoading } = useQuery({
    queryKey: ["messages", conversationId],
    queryFn: async () => {
      const msgs = await api.get<Message[]>(
        `/conversations/${conversationId}/messages`,
      );
      return [...msgs].reverse();
    },
  });

  // Reset local messages when fetched data changes (e.g. refetch after delete)
  const fetchedIds = useMemo(
    () => (fetchedMessages ?? []).map((m) => m.id).join(","),
    [fetchedMessages],
  );

  useEffect(() => {
    setLocalMessages([]);
  }, [fetchedIds]);

  // Merge fetched + local messages, deduplicating by id
  const messages = useMemo(() => {
    const fetched = fetchedMessages ?? [];
    if (localMessages.length === 0) return fetched;
    const ids = new Set(fetched.map((m) => m.id));
    const unique = localMessages.filter((m) => !ids.has(m.id));
    return [...fetched, ...unique];
  }, [fetchedMessages, localMessages]);

  // Mark as read on mount
  useEffect(() => {
    api.patch(`/conversations/${conversationId}/read`).catch(() => {});
    queryClient.invalidateQueries({ queryKey: ["conversations"] });
  }, [conversationId, queryClient]);

  // Listen for new messages via WebSocket
  useEffect(() => {
    const unsubMessage = socket.on("message", (data) => {
      const msg = data as Message;
      if (msg.conversationId === conversationId) {
        setLocalMessages((prev) => {
          if (prev.some((m) => m.id === msg.id)) return prev;
          return [...prev, msg];
        });
        api.patch(`/conversations/${conversationId}/read`).catch(() => {});
      }
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
    });

    const unsubDeleted = socket.on("message_deleted", (data) => {
      const { messageId, conversationId: convId } = data as {
        messageId: string;
        conversationId: string;
      };
      if (convId === conversationId) {
        setLocalMessages((prev) => prev.filter((m) => m.id !== messageId));
        // Also remove from query cache
        queryClient.setQueryData<Message[]>(
          ["messages", conversationId],
          (old) => old?.filter((m) => m.id !== messageId),
        );
      }
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
    });

    const unsubTyping = socket.on("typing", (data) => {
      const t = data as { conversationId: string; userId: string };
      if (t.conversationId !== conversationId) return;
      if (t.userId === user?.id) return; // ignore our own
      setOtherTyping(true);
      if (typingClearRef.current) clearTimeout(typingClearRef.current);
      typingClearRef.current = setTimeout(() => setOtherTyping(false), 3000);
    });

    return () => {
      unsubMessage();
      unsubDeleted();
      unsubTyping();
      if (typingClearRef.current) clearTimeout(typingClearRef.current);
    };
  }, [conversationId, queryClient, user?.id]);

  // Send a typing event when the user types (throttled to once per 2s)
  const sendTyping = useCallback(() => {
    const now = Date.now();
    if (now - lastTypingSentRef.current < 2000) return;
    lastTypingSentRef.current = now;
    socket.sendRaw({ type: "typing", conversationId });
  }, [conversationId]);

  // Auto-scroll to bottom when message count changes
  const messageCount = messages.length;
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messageCount]);

  const handleSend = useCallback(async () => {
    const content = newMessage.trim();
    if (!content) return;

    setSending(true);
    try {
      const msg = await api.post<Message>(
        `/conversations/${conversationId}/messages`,
        { content },
      );
      setLocalMessages((prev) => {
        if (prev.some((m) => m.id === msg.id)) return prev;
        return [...prev, msg];
      });
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
      setNewMessage("");
    } finally {
      setSending(false);
    }
  }, [newMessage, conversationId, queryClient]);

  const handleDeleteMessage = useCallback(
    async (messageId: string) => {
      try {
        await api.delete(
          `/conversations/${conversationId}/messages/${messageId}`,
        );
        // Remove from local state
        setLocalMessages((prev) => prev.filter((m) => m.id !== messageId));
        // Remove from query cache
        queryClient.setQueryData<Message[]>(
          ["messages", conversationId],
          (old) => old?.filter((m) => m.id !== messageId),
        );
        queryClient.invalidateQueries({ queryKey: ["conversations"] });
      } catch {
        // ignore
      }
    },
    [conversationId, queryClient],
  );

  const handleDeleteConversation = useCallback(async () => {
    try {
      await api.delete(`/conversations/${conversationId}`);
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
      router.push("/messages");
    } catch {
      // ignore
    }
  }, [conversationId, queryClient, router]);

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
        {otherParticipant ? (
          <Link
            href={`/profile/${otherParticipant.username}`}
            className="flex flex-1 items-center gap-3 transition-opacity hover:opacity-80"
          >
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
          </Link>
        ) : (
          <div className="flex-1" />
        )}

        {/* Delete conversation */}
        <button
          onClick={handleDeleteConversation}
          className="flex size-8 items-center justify-center rounded-lg text-zinc-500 transition-colors hover:bg-red-950/50 hover:text-red-400"
          title="Delete conversation"
        >
          <Trash2 className="size-4" />
        </button>
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
              onDelete={handleDeleteMessage}
            />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Typing indicator */}
      {otherTyping && otherParticipant && (
        <div className="px-4 pb-1 text-xs text-zinc-500">
          {otherParticipant.name} is typing
          <span className="ml-1 inline-block animate-pulse">…</span>
        </div>
      )}

      {/* Input */}
      <div className="border-t border-zinc-800 px-4 py-3">
        <div className="flex items-center gap-2">
          <Input
            value={newMessage}
            onChange={(e) => {
              setNewMessage(e.target.value);
              if (e.target.value) sendTyping();
            }}
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
