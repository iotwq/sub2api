package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAPIBalanceAccessTokenMigrationUsesHashedUniqueStorage(t *testing.T) {
	content, err := FS.ReadFile("231_newapi_balance_access_token.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS newapi_access_token_hash CHAR(64)")
	require.Contains(t, sql, "CREATE UNIQUE INDEX IF NOT EXISTS idx_users_newapi_access_token_hash")
	require.Contains(t, sql, "WHERE newapi_access_token_hash IS NOT NULL AND deleted_at IS NULL")
	require.NotContains(t, sql, "newapi_access_token VARCHAR")
}
