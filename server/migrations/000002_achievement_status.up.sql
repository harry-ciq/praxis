-- Add status column to achievements for tracking source availability
ALTER TABLE achievements ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

CREATE INDEX idx_achievements_status ON achievements(status);
