CREATE TABLE IF NOT EXISTS sessions (
    id  TEXT PRIMARY KEY,
    avatar_url TEXT NOT NULL,
    avatar_id INT NOT NULL,
    user_name TEXT NOT NULL, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL 
)

CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_url TEXT,
    user_name TEXT NOT NULL,
    avatar_url TEXT NOT NULL,
    session_id TEXT REFERENCES sessions (id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'active'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_comment_at TIMESTAMPTZ
)

CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES posts (id) on DELETE CASCADE,
    reply_to_id BIGINT,
    content TEXT NOT NULL,
    image_url TEXT,
    user_name TEXT NOT NULL,
    avatar_url TEXT NOT NULL, 
    session_id TEXT REFERENCES sessions (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)