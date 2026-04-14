DROP INDEX IF EXISTS idx_achievements_status;
ALTER TABLE achievements DROP COLUMN IF EXISTS status;
