"use client";

import { useNotifications } from "@/hooks/use-notifications";
import { formatRelativeTime } from "@/lib/format";
import { cn } from "@/lib/utils";
import {
  UserPlus,
  Flame,
  MessageSquare,
  Briefcase,
  Trophy,
  Bell,
  Loader2,
} from "lucide-react";
import type { NotificationType } from "@/types";

const NOTIFICATION_ICONS: Record<NotificationType, typeof Bell> = {
  NEW_FOLLOWER: UserPlus,
  REACTION: Flame,
  MESSAGE: MessageSquare,
  JOB_MATCH: Briefcase,
  ACHIEVEMENT: Trophy,
};

const NOTIFICATION_COLORS: Record<NotificationType, string> = {
  NEW_FOLLOWER: "text-blue-400",
  REACTION: "text-orange-400",
  MESSAGE: "text-green-400",
  JOB_MATCH: "text-purple-400",
  ACHIEVEMENT: "text-amber-400",
};

export default function NotificationsPage() {
  const { notifications, markRead } = useNotifications();

  return (
    <div className="mx-auto max-w-2xl px-4 py-6">
      <h1 className="mb-6 text-2xl font-bold text-zinc-50">Notifications</h1>

      {notifications.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Bell className="mb-3 size-12 text-zinc-700" />
          <p className="text-sm text-zinc-500">No notifications yet</p>
        </div>
      ) : (
        <div className="flex flex-col gap-1">
          {notifications.map((notification) => {
            const Icon =
              NOTIFICATION_ICONS[notification.type] ?? Bell;
            const iconColor =
              NOTIFICATION_COLORS[notification.type] ?? "text-zinc-400";

            return (
              <button
                key={notification.id}
                onClick={() => {
                  if (!notification.read) {
                    markRead(notification.id);
                  }
                }}
                className={cn(
                  "flex w-full items-start gap-3 rounded-lg px-4 py-3 text-left transition-colors hover:bg-zinc-800/50",
                  !notification.read && "border-l-2 border-blue-500 bg-zinc-900/50",
                )}
              >
                <div
                  className={cn(
                    "mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-full bg-zinc-800",
                    iconColor,
                  )}
                >
                  <Icon className="size-4" />
                </div>

                <div className="min-w-0 flex-1">
                  <p
                    className={cn(
                      "text-sm",
                      notification.read
                        ? "text-zinc-400"
                        : "font-medium text-zinc-100",
                    )}
                  >
                    {notification.title}
                  </p>
                  <p className="mt-0.5 text-xs text-zinc-500">
                    {notification.body}
                  </p>
                  <p className="mt-1 text-[11px] text-zinc-600">
                    {formatRelativeTime(notification.createdAt)}
                  </p>
                </div>

                {!notification.read && (
                  <span className="mt-2 size-2 shrink-0 rounded-full bg-blue-500" />
                )}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
