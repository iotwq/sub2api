//go:build integration

package repository

import (
	"context"
	"testing"

	dbmigrations "github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestMigration246GeminiImageGenerationPreservesOtherGroups(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	migrationSQL, err := dbmigrations.FS.ReadFile("246_enable_gemini_image_generation_groups.sql")
	require.NoError(t, err)

	for _, platform := range []string{"gemini", "antigravity", "openai", "composite"} {
		var id int64
		require.NoError(t, tx.QueryRowContext(ctx, `
INSERT INTO groups (name, platform, allow_image_generation)
VALUES ($1, $2, false) RETURNING id`, "migration-246-"+platform, platform).Scan(&id))
		_, err = tx.ExecContext(ctx, string(migrationSQL))
		require.NoError(t, err)
		var allowed bool
		require.NoError(t, tx.QueryRowContext(ctx,
			"SELECT allow_image_generation FROM groups WHERE id = $1", id).Scan(&allowed))
		require.Equal(t, platform == "gemini" || platform == "antigravity", allowed, platform)
	}

	// The runner identifies migrations by full filename, so both 246 files apply.
	for _, name := range []string{
		"246_enable_gemini_image_generation_groups.sql",
		"246_gpt61_sol_subscription_optimization.sql",
	} {
		var count int
		require.NoError(t, tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", name).Scan(&count))
		require.Equal(t, 1, count, name)
	}
	var optimized bool
	require.NoError(t, tx.QueryRowContext(ctx,
		"SELECT gpt61_sol_subscription_optimization_enabled FROM groups WHERE name = 'migration-246-openai'").Scan(&optimized))
	require.False(t, optimized)
}
