package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdminAccountPreservesCodexObservationAndLegacySecrets(t *testing.T) {
	for _, key := range []string{CodexSignalExtraKey, "codex_turn_ticket:gpt-6-astra"} {
		persisted := map[string]any{"length": 312, "observed_at": time.Now().UTC().Format(time.RFC3339Nano)}
		account := &Account{ID: 41, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{key: persisted}}
		repo := &upstreamBillingProbeAccountRepo{accounts: map[int64]*Account{41: account}}
		svc := &adminServiceImpl{accountRepo: repo}
		for _, extra := range []map[string]any{{key: map[string]any{"length": 292}, "custom": true}, {"custom": true}} {
			updated, err := svc.UpdateAccount(context.Background(), 41, &UpdateAccountInput{Extra: extra})
			require.NoError(t, err)
			require.Equal(t, persisted, updated.Extra[key])
			require.Equal(t, true, updated.Extra["custom"])
		}
		require.NoError(t, svc.UpdateAccountExtra(context.Background(), 41, map[string]any{key: map[string]any{"length": 292}}))
		require.Equal(t, persisted, repo.accounts[41].Extra[key])
		created, err := buildAccountForCreate(&CreateAccountInput{Platform: PlatformOpenAI, Type: AccountTypeOAuth}, map[string]any{key: persisted, "custom": true})
		require.NoError(t, err)
		require.NotContains(t, created.Extra, key)
		require.Equal(t, true, created.Extra["custom"])
	}
}
