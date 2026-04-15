"use client";

import { useAuth } from "@/hooks/use-auth";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
  DropdownMenuGroup,
} from "@/components/ui/dropdown-menu";
import { Bell, Search, LogOut, User, Settings } from "lucide-react";
import Link from "next/link";

export function Navbar() {
  const { user, logout } = useAuth();

  const initials = user?.name
    ? user.name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .slice(0, 2)
    : "?";

  return (
    <header className="sticky top-0 z-40 flex h-14 items-center border-b border-zinc-800 bg-zinc-950/95 px-4 backdrop-blur">
      {/* Logo */}
      <Link
        href="/feed"
        className="mr-4 text-lg font-bold tracking-tight text-zinc-50 lg:mr-6"
      >
        Praxis
      </Link>

      {/* Search */}
      <div className="relative hidden flex-1 max-w-md sm:block">
        <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
        <Input
          placeholder="Search..."
          className="h-8 border-zinc-800 bg-zinc-900 pl-8 text-sm text-zinc-200 placeholder:text-zinc-500 focus:border-zinc-700"
        />
      </div>

      <div className="ml-auto flex items-center gap-2">
        {/* Mobile search button */}
        <Button
          variant="ghost"
          size="icon"
          className="text-zinc-400 hover:text-zinc-200 sm:hidden"
        >
          <Search className="size-4" />
        </Button>

        {/* Notifications */}
        <Link
          href="/notifications"
          className="relative inline-flex size-8 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-zinc-200"
        >
          <Bell className="size-4" />
          <span className="absolute right-1.5 top-1.5 size-2 rounded-full bg-red-500" />
        </Link>

        {/* User menu */}
        <DropdownMenu>
          <DropdownMenuTrigger className="cursor-pointer rounded-full outline-none focus-visible:ring-2 focus-visible:ring-zinc-600">
            <Avatar size="sm">
              {user?.avatarUrl && <AvatarImage src={user.avatarUrl} alt={user.name} />}
              <AvatarFallback className="bg-zinc-800 text-xs text-zinc-300">
                {initials}
              </AvatarFallback>
            </Avatar>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" side="bottom" sideOffset={8}>
            <DropdownMenuGroup>
              <DropdownMenuLabel className="font-normal">
                <div className="flex flex-col gap-0.5">
                  <p className="text-sm font-medium text-zinc-100">
                    {user?.name}
                  </p>
                  <p className="text-xs text-zinc-400">@{user?.username}</p>
                </div>
              </DropdownMenuLabel>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem render={<Link href={`/profile/${user?.username ?? ""}`} />}>
              <User className="mr-1.5 size-4" />
              Profile
            </DropdownMenuItem>
            <DropdownMenuItem render={<Link href="/settings" />}>
              <Settings className="mr-1.5 size-4" />
              Settings
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => logout()}>
              <LogOut className="mr-1.5 size-4" />
              Sign Out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
