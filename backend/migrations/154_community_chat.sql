CREATE TABLE IF NOT EXISTS community_chat_messages (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username TEXT NOT NULL,
    avatar_url TEXT NOT NULL DEFAULT '',
    message_type TEXT NOT NULL,
    content TEXT NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    image_mime_type TEXT NOT NULL DEFAULT '',
    image_size_bytes BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT community_chat_messages_type_check CHECK (message_type IN ('text', 'image')),
    CONSTRAINT community_chat_messages_content_check CHECK (length(content) > 0 AND length(content) <= 2000)
);

CREATE INDEX IF NOT EXISTS idx_community_chat_messages_visible_created
    ON community_chat_messages (created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_community_chat_messages_user
    ON community_chat_messages (user_id, created_at DESC);
