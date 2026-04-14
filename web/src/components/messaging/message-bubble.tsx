"use client";

import { cn } from "@/lib/utils";
import { formatRelativeTime } from "@/lib/format";
import type { Message } from "@/types";

interface MessageBubbleProps {
  message: Message;
  isOwn: boolean;
}

export function MessageBubble({ message, isOwn }: MessageBubbleProps) {
  return (
    <div
      className={cn("flex w-full", isOwn ? "justify-end" : "justify-start")}
    >
      <div
        className={cn(
          "max-w-[75%] rounded-2xl px-4 py-2",
          isOwn
            ? "rounded-br-md bg-blue-600 text-white"
            : "rounded-bl-md bg-zinc-800 text-zinc-100",
        )}
      >
        <p className="text-sm whitespace-pre-wrap break-words">
          {message.content}
        </p>
        <p
          className={cn(
            "mt-1 text-[10px]",
            isOwn ? "text-blue-200" : "text-zinc-500",
          )}
        >
          {formatRelativeTime(message.createdAt)}
        </p>
      </div>
    </div>
  );
}
