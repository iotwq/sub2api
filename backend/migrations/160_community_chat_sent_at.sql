ALTER TABLE community_chat_messages
    ADD COLUMN IF NOT EXISTS sent_at TIMESTAMPTZ NULL;
