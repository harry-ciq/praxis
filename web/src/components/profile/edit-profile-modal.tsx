"use client";

import { useState } from "react";
import { Globe, MapPin, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useUpdateProfile } from "@/hooks/use-profile";
import type { UserProfile } from "@/types";

interface EditProfileModalProps {
  open: boolean;
  onClose: () => void;
  profile: UserProfile;
}

export function EditProfileModal({
  open,
  onClose,
  profile,
}: EditProfileModalProps) {
  const [name, setName] = useState(profile.name);
  const [headline, setHeadline] = useState(profile.headline ?? "");
  const [bio, setBio] = useState(profile.bio ?? "");
  const [location, setLocation] = useState(profile.location ?? "");
  const [websiteUrl, setWebsiteUrl] = useState(profile.websiteUrl ?? "");
  const [twitter, setTwitter] = useState(profile.socialLinks?.twitter ?? "");
  const [linkedin, setLinkedin] = useState(profile.socialLinks?.linkedin ?? "");

  const updateProfile = useUpdateProfile();

  if (!open) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateProfile.mutate(
      {
        name,
        headline,
        bio,
        location,
        websiteUrl,
        socialLinks: { twitter, linkedin },
      },
      { onSuccess: onClose }
    );
  };

  const inputClass =
    "w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Overlay */}
      <div
        className="absolute inset-0 bg-black/60"
        onClick={onClose}
        aria-hidden
      />

      {/* Panel */}
      <div className="relative z-10 w-full max-w-lg rounded-2xl border border-zinc-800 bg-zinc-900 p-6 shadow-xl">
        {/* Header */}
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-zinc-100">Edit Profile</h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          >
            <X className="size-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Name */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-zinc-300">
              Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your name"
              className={inputClass}
            />
          </div>

          {/* Headline */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-zinc-300">
              Headline
            </label>
            <input
              type="text"
              value={headline}
              onChange={(e) => setHeadline(e.target.value)}
              placeholder="e.g. Full-stack developer"
              className={inputClass}
            />
          </div>

          {/* Bio */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-zinc-300">
              Bio
            </label>
            <textarea
              rows={4}
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="Tell us about yourself..."
              className={inputClass + " resize-none"}
            />
          </div>

          {/* Location */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-zinc-300">
              Location
            </label>
            <div className="relative">
              <MapPin className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
              <input
                type="text"
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                placeholder="San Francisco, CA"
                className={inputClass + " pl-9"}
              />
            </div>
          </div>

          {/* Website URL */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-zinc-300">
              Website
            </label>
            <div className="relative">
              <Globe className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
              <input
                type="text"
                value={websiteUrl}
                onChange={(e) => setWebsiteUrl(e.target.value)}
                placeholder="https://yoursite.com"
                className={inputClass + " pl-9"}
              />
            </div>
          </div>

          {/* Social Links */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-zinc-300">
                Twitter / X
              </label>
              <input
                type="text"
                value={twitter}
                onChange={(e) => setTwitter(e.target.value)}
                placeholder="https://x.com/username"
                className={inputClass}
              />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-zinc-300">
                LinkedIn
              </label>
              <input
                type="text"
                value={linkedin}
                onChange={(e) => setLinkedin(e.target.value)}
                placeholder="https://linkedin.com/in/username"
                className={inputClass}
              />
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateProfile.isPending}>
              {updateProfile.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
