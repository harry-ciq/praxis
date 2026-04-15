"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  MessageSquare,
  Briefcase,
  Bell,
  User,
  Settings,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/use-auth";

const staticNavItems = [
  { href: "/feed", label: "Feed", icon: LayoutDashboard },
  { href: "/messages", label: "Messages", icon: MessageSquare },
  { href: "/jobs", label: "Jobs", icon: Briefcase },
  { href: "/notifications", label: "Notifications", icon: Bell },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuth();

  const navItems = [
    ...staticNavItems,
    { href: `/profile/${user?.username ?? ""}`, label: "Profile", icon: User },
    { href: "/settings", label: "Settings", icon: Settings },
  ];

  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden w-60 shrink-0 border-r border-zinc-800 bg-zinc-950 lg:block">
        <div className="flex h-full flex-col gap-1 px-3 py-4">
          <nav className="flex flex-col gap-0.5">
            {navItems.map((item) => {
              const isActive =
                pathname === item.href || pathname.startsWith(item.href + "/");
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-zinc-800 text-zinc-50"
                      : "text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200",
                  )}
                >
                  <item.icon className="size-4" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>
      </aside>

      {/* Mobile bottom nav */}
      <nav className="fixed inset-x-0 bottom-0 z-50 flex border-t border-zinc-800 bg-zinc-950/95 backdrop-blur lg:hidden">
        {navItems.slice(0, 5).map((item) => {
          const isActive =
            pathname === item.href || pathname.startsWith(item.href + "/");
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-1 flex-col items-center gap-0.5 py-2 text-[10px] font-medium transition-colors",
                isActive ? "text-zinc-50" : "text-zinc-500",
              )}
            >
              <item.icon className="size-5" />
              {item.label}
            </Link>
          );
        })}
      </nav>
    </>
  );
}
