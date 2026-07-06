BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY,
    nickname      TEXT        NOT NULL,
    phone         TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_nickname_unique UNIQUE (nickname),
    CONSTRAINT users_phone_unique    UNIQUE (phone)
);

CREATE TABLE IF NOT EXISTS sessions (
    token      UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id    ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

CREATE TABLE IF NOT EXISTS messages (
    id           UUID        PRIMARY KEY,
    sender_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    recipient_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    body         TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_pair      ON messages (sender_id, recipient_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages (recipient_id, created_at);

COMMIT;
