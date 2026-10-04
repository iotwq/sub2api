ALTER TABLE community_chat_messages
    ADD COLUMN IF NOT EXISTS reply_to_message_id BIGINT NULL REFERENCES community_chat_messages(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reply_to_username TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reply_to_content TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reply_to_message_type TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reply_to_image_url TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_community_chat_messages_reply_to
    ON community_chat_messages (reply_to_message_id)
    WHERE reply_to_message_id IS NOT NULL;
