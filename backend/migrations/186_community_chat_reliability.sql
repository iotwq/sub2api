ALTER TABLE community_chat_direct_messages
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS deleted_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_community_chat_direct_messages_visible_conversation
    ON community_chat_direct_messages (conversation_user_id, id DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS community_chat_direct_read_states (
    reader_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_message_id BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (reader_user_id, conversation_user_id)
);

CREATE INDEX IF NOT EXISTS idx_community_chat_direct_read_states_conversation
    ON community_chat_direct_read_states (conversation_user_id, reader_user_id);
