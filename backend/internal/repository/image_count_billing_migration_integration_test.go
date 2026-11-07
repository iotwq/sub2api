//go:build integration

package repository

import (
	"context"
	"testing"

	dbmigrations "github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestMigration248ImageCountBillingPreservesExistingGroupSettings(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	migrationSQL, err := dbmigrations.FS.ReadFile("248_group_image_count_billing.sql")
	require.NoError(t, err)

	var id int64
	require.NoError(t, tx.QueryRowContext(ctx, `
INSERT INTO groups (name, platform, allow_image_generation, image_rate_independent,
 image_rate_multiplier, gpt61_sol_subscription_optimization_enabled)
VALUES ('migration-248', 'openai', true, true, 1, true) RETURNING id`).Scan(&id))
	var countBilling, independent, optimized bool
	var multiplier float64
	readGroup := func() {
		t.Helper()
		require.NoError(t, tx.QueryRowContext(ctx, `
SELECT image_count_billing, image_rate_independent, image_rate_multiplier,
 gpt61_sol_subscription_optimization_enabled FROM groups WHERE id = $1`, id).
			Scan(&countBilling, &independent, &multiplier, &optimized))
	}
	readGroup()
	require.False(t, countBilling, "existing billing mode must remain the default")
	_, err = tx.ExecContext(ctx, "UPDATE groups SET image_count_billing = true WHERE id = $1", id)
	require.NoError(t, err)
	for i := 0; i < 2; i++ {
		_, err = tx.ExecContext(ctx, string(migrationSQL))
		require.NoError(t, err)
	}
	readGroup()
	require.True(t, countBilling, "reapplying migration must preserve explicit opt-in")
	require.True(t, independent)
	require.Equal(t, float64(1), multiplier)
	require.True(t, optimized, "the local subscription optimization setting must be preserved")
	for _, name := range []string{"247_scoped_admin_api_keys.sql", "248_group_image_count_billing.sql"} {
		var count int
		require.NoError(t, tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", name).Scan(&count))
		require.Equal(t, 1, count, name)
	}
}
