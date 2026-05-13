CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    email         CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user'
                    CHECK (role IN ('user', 'admin')),
    status        TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'banned')),
    created_at    TIMESTAMPTZ DEFAULT NOW()
);