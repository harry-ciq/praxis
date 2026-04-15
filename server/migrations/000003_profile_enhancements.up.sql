-- Add profile fields to users
ALTER TABLE users ADD COLUMN location TEXT DEFAULT '';
ALTER TABLE users ADD COLUMN website_url TEXT DEFAULT '';
ALTER TABLE users ADD COLUMN social_links JSONB DEFAULT '{}';

-- Work experience
CREATE TABLE experiences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name TEXT NOT NULL,
    role TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    description TEXT DEFAULT '',
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_experiences_user_id ON experiences(user_id);

-- Skills (verified from achievements or manually added)
CREATE TABLE skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT false,
    source TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_skills_user_id ON skills(user_id);
