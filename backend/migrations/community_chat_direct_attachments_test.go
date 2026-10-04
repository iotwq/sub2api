package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommunityChatDirectAttachmentsMigration(t *testing.T) {
	content, err := FS.ReadFile("222_community_chat_direct_attachments.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS community_chat_direct_messages_type_check")
	require.Contains(t, sql, "CHECK (message_type IN ('text', 'image', 'file'))")
}
