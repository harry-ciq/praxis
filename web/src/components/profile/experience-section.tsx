"use client";

import { useState } from "react";
import { Briefcase, Pencil, Plus, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  useCreateExperience,
  useUpdateExperience,
  useDeleteExperience,
} from "@/hooks/use-profile";
import type { Experience } from "@/types";

interface ExperienceSectionProps {
  experiences: Experience[];
  isOwnProfile: boolean;
}

interface ExperienceFormData {
  companyName: string;
  role: string;
  startDate: string;
  endDate: string;
  isCurrent: boolean;
  description: string;
}

const emptyForm: ExperienceFormData = {
  companyName: "",
  role: "",
  startDate: "",
  endDate: "",
  isCurrent: false,
  description: "",
};

function formatMonthYear(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
}

function ExperienceForm({
  initial,
  onSubmit,
  onCancel,
  isPending,
}: {
  initial: ExperienceFormData;
  onSubmit: (data: ExperienceFormData) => void;
  onCancel: () => void;
  isPending: boolean;
}) {
  const [form, setForm] = useState<ExperienceFormData>(initial);

  const inputClass =
    "w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500";

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form);
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-3 rounded-xl border border-zinc-700 bg-zinc-800/50 p-4"
    >
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="mb-1 block text-xs font-medium text-zinc-400">
            Company Name
          </label>
          <input
            type="text"
            required
            value={form.companyName}
            onChange={(e) =>
              setForm((f) => ({ ...f, companyName: e.target.value }))
            }
            placeholder="Acme Inc."
            className={inputClass}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-zinc-400">
            Role
          </label>
          <input
            type="text"
            required
            value={form.role}
            onChange={(e) => setForm((f) => ({ ...f, role: e.target.value }))}
            placeholder="Software Engineer"
            className={inputClass}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="mb-1 block text-xs font-medium text-zinc-400">
            Start Date
          </label>
          <input
            type="month"
            required
            value={form.startDate}
            onChange={(e) =>
              setForm((f) => ({ ...f, startDate: e.target.value }))
            }
            className={inputClass}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-zinc-400">
            End Date
          </label>
          <input
            type="month"
            disabled={form.isCurrent}
            value={form.isCurrent ? "" : form.endDate}
            onChange={(e) =>
              setForm((f) => ({ ...f, endDate: e.target.value }))
            }
            className={inputClass + " disabled:opacity-50"}
          />
        </div>
      </div>

      <label className="flex items-center gap-2 text-sm text-zinc-300">
        <input
          type="checkbox"
          checked={form.isCurrent}
          onChange={(e) =>
            setForm((f) => ({ ...f, isCurrent: e.target.checked, endDate: "" }))
          }
          className="size-4 rounded border-zinc-600 bg-zinc-800 text-emerald-500 focus:ring-emerald-500"
        />
        I currently work here
      </label>

      <div>
        <label className="mb-1 block text-xs font-medium text-zinc-400">
          Description
        </label>
        <textarea
          rows={3}
          value={form.description}
          onChange={(e) =>
            setForm((f) => ({ ...f, description: e.target.value }))
          }
          placeholder="What did you work on?"
          className={inputClass + " resize-none"}
        />
      </div>

      <div className="flex items-center justify-end gap-2 pt-1">
        <Button type="button" variant="ghost" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" size="sm" disabled={isPending}>
          {isPending ? "Saving..." : "Save"}
        </Button>
      </div>
    </form>
  );
}

export function ExperienceSection({
  experiences,
  isOwnProfile,
}: ExperienceSectionProps) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  const createExperience = useCreateExperience();
  const updateExperience = useUpdateExperience();
  const deleteExperience = useDeleteExperience();

  const handleCreate = (data: ExperienceFormData) => {
    createExperience.mutate(
      {
        companyName: data.companyName,
        role: data.role,
        startDate: data.startDate,
        endDate: data.isCurrent ? undefined : data.endDate || undefined,
        description: data.description || undefined,
        isCurrent: data.isCurrent,
      },
      { onSuccess: () => setShowAddForm(false) }
    );
  };

  const handleUpdate = (id: string, data: ExperienceFormData) => {
    updateExperience.mutate(
      {
        id,
        companyName: data.companyName,
        role: data.role,
        startDate: data.startDate,
        endDate: data.isCurrent ? undefined : data.endDate || undefined,
        description: data.description || undefined,
        isCurrent: data.isCurrent,
      },
      { onSuccess: () => setEditingId(null) }
    );
  };

  const handleDelete = (id: string) => {
    deleteExperience.mutate(id);
  };

  return (
    <div className="space-y-4">
      {isOwnProfile && (
        <div className="flex justify-end">
          {!showAddForm && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowAddForm(true)}
            >
              <Plus className="size-4" />
              Add Experience
            </Button>
          )}
        </div>
      )}

      {showAddForm && (
        <ExperienceForm
          initial={emptyForm}
          onSubmit={handleCreate}
          onCancel={() => setShowAddForm(false)}
          isPending={createExperience.isPending}
        />
      )}

      {experiences.length === 0 && !showAddForm ? (
        <div className="py-10 text-center text-sm text-zinc-500">
          No experience added yet.
        </div>
      ) : (
        <div className="space-y-3">
          {experiences.map((exp) =>
            editingId === exp.id ? (
              <ExperienceForm
                key={exp.id}
                initial={{
                  companyName: exp.companyName,
                  role: exp.role,
                  startDate: exp.startDate.slice(0, 7),
                  endDate: exp.endDate ? exp.endDate.slice(0, 7) : "",
                  isCurrent: exp.isCurrent,
                  description: exp.description ?? "",
                }}
                onSubmit={(data) => handleUpdate(exp.id, data)}
                onCancel={() => setEditingId(null)}
                isPending={updateExperience.isPending}
              />
            ) : (
              <div
                key={exp.id}
                className="group flex items-start gap-3 rounded-xl border border-zinc-800 bg-zinc-900/60 p-4"
              >
                <div className="mt-0.5 rounded-lg bg-zinc-800 p-2">
                  <Briefcase className="size-4 text-zinc-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <h4 className="text-sm font-semibold text-zinc-100">
                        {exp.companyName}
                      </h4>
                      <p className="text-sm text-zinc-400">{exp.role}</p>
                    </div>
                    {isOwnProfile && (
                      <div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                        <button
                          onClick={() => setEditingId(exp.id)}
                          className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-800 hover:text-zinc-300"
                        >
                          <Pencil className="size-3.5" />
                        </button>
                        <button
                          onClick={() => handleDelete(exp.id)}
                          className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-800 hover:text-red-400"
                        >
                          <Trash2 className="size-3.5" />
                        </button>
                      </div>
                    )}
                  </div>
                  <p className="mt-1 text-xs text-zinc-500">
                    {formatMonthYear(exp.startDate)} &mdash;{" "}
                    {exp.isCurrent
                      ? "Present"
                      : exp.endDate
                        ? formatMonthYear(exp.endDate)
                        : ""}
                  </p>
                  {exp.description && (
                    <p className="mt-2 text-sm leading-relaxed text-zinc-400">
                      {exp.description}
                    </p>
                  )}
                </div>
              </div>
            )
          )}
        </div>
      )}
    </div>
  );
}
