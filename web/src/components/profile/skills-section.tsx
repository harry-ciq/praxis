"use client";

import { useState } from "react";
import { CheckCircle2, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAddSkill, useDeleteSkill } from "@/hooks/use-profile";
import type { Skill } from "@/types";

interface SkillsSectionProps {
  skills: Skill[];
  isOwnProfile: boolean;
}

export function SkillsSection({ skills, isOwnProfile }: SkillsSectionProps) {
  const [showInput, setShowInput] = useState(false);
  const [newSkill, setNewSkill] = useState("");

  const addSkill = useAddSkill();
  const deleteSkill = useDeleteSkill();

  const handleAdd = () => {
    const trimmed = newSkill.trim();
    if (!trimmed) return;
    addSkill.mutate(trimmed, {
      onSuccess: () => {
        setNewSkill("");
        setShowInput(false);
      },
    });
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAdd();
    }
    if (e.key === "Escape") {
      setShowInput(false);
      setNewSkill("");
    }
  };

  return (
    <div className="space-y-4">
      {skills.length === 0 && !showInput ? (
        <div className="py-10 text-center text-sm text-zinc-500">
          No skills added yet.
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          {skills.map((skill) => (
            <span
              key={skill.id}
              className="group inline-flex items-center gap-1.5 rounded-full border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-sm text-zinc-300"
            >
              {skill.verified && (
                <span title="Verified">
                  <CheckCircle2 className="size-3.5 text-emerald-400" />
                </span>
              )}
              {skill.name}
              {isOwnProfile && !skill.verified && (
                <button
                  onClick={() => deleteSkill.mutate(skill.id)}
                  className="ml-0.5 rounded-full p-0.5 text-zinc-500 opacity-0 transition-opacity hover:text-zinc-300 group-hover:opacity-100"
                >
                  <X className="size-3" />
                </button>
              )}
            </span>
          ))}
        </div>
      )}

      {isOwnProfile && (
        <div>
          {showInput ? (
            <div className="flex items-center gap-2">
              <input
                autoFocus
                type="text"
                value={newSkill}
                onChange={(e) => setNewSkill(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="e.g. TypeScript"
                className="w-48 rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-sm text-zinc-100 placeholder:text-zinc-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              />
              <Button
                size="sm"
                onClick={handleAdd}
                disabled={addSkill.isPending || !newSkill.trim()}
              >
                {addSkill.isPending ? "Adding..." : "Add"}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setShowInput(false);
                  setNewSkill("");
                }}
              >
                Cancel
              </Button>
            </div>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowInput(true)}
            >
              <Plus className="size-4" />
              Add Skill
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
