package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountCodexSignalProjectionDoesNotExposeState(t *testing.T) {
	now := time.Now().UTC()
	account := &service.Account{ID: 41, Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth,
		Extra: map[string]any{
			service.CodexSignalExtraKey:     service.CodexSignalStatus{Length: 312, ObservedAt: now, Last312At: &now},
			"codex_turn_ticket:gpt-6-astra": map[string]any{"state": "old-private-blob"},
			"codex_harvest_proxy_url":       "socks5://user:secret@example.com:1080",
		}}
	full := AccountFromService(account)
	list := AccountListItemFromAccount(full)
	require.NotNil(t, full.CodexSignal)
	require.Equal(t, 312, full.CodexSignal.Length)
	require.Equal(t, full.CodexSignal, list.CodexSignal)
	for _, item := range []any{full, list} {
		data, err := json.Marshal(item)
		require.NoError(t, err)
		require.NotContains(t, string(data), "old-private-blob")
		require.NotContains(t, string(data), "secret")
		require.Contains(t, string(data), `"codex_signal":{"length":312`)
	}
	require.NotContains(t, full.Extra, service.CodexSignalExtraKey)
	account.Type = service.AccountTypeAPIKey
	require.Nil(t, AccountFromService(account).CodexSignal)
	account.Type = service.AccountTypeOAuth
	account.Extra = nil
	require.Nil(t, AccountFromService(account).CodexSignal)
}
