CREATE TABLE IF NOT EXISTS community_chat_direct_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username TEXT NOT NULL,
    avatar_url TEXT NOT NULL DEFAULT '',
    message_type TEXT NOT NULL,
    content TEXT NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    image_mime_type TEXT NOT NULL DEFAULT '',
    image_size_bytes BIGINT NOT NULL DEFAULT 0,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT community_chat_direct_messages_type_check CHECK (message_type IN ('text')),
    CONSTRAINT community_chat_direct_messages_content_check CHECK (length(content) > 0 AND length(content) <= 2000)
);

CREATE INDEX IF NOT EXISTS idx_community_chat_direct_messages_conversation_created
    ON community_chat_direct_messages (conversation_user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_community_chat_direct_messages_sender
    ON community_chat_direct_messages (user_id, created_at DESC);
