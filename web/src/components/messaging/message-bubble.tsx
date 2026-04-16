"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { formatRelativeTime } from "@/lib/format";
import { Trash2 } from "lucide-react";
import type { Message } from "@/types";

interface MessageBubbleProps {
  message: Message;
  isOwn: boolean;
  onDelete?: (messageId: string) => void;
}

export function MessageBubble({ message, isOwn, onDelete }: MessageBubbleProps) {
  const [hovered, setHovered] = useState(false);

  return (
    <div
      className={cn("group flex w-full items-center gap-1", isOwn ? "justify-end" : "justify-start")}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {/* Delete button for own messages (appears on hover, before the bubble) */}
      {isOwn && onDelete && (
        <button
          onClick={() => onDelete(message.id)}
          className={cn(
            "flex size-6 shrink-0 items-center justify-center rounded-full text-zinc-500 transition-all hover:bg-red-950/50 hover:text-red-400",
            hovered ? "opacity-100" : "opacity-0",
          )}
          title="Delete message"
        >
          <Trash2 className="size-3" />
        </button>
      )}

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
