ALTER TABLE community_chat_messages
    DROP CONSTRAINT IF EXISTS community_chat_messages_type_check;

ALTER TABLE community_chat_messages
    ADD CONSTRAINT community_chat_messages_type_check
    CHECK (message_type IN ('text', 'image', 'file'));
