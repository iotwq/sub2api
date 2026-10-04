package repository

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSchedulerMetadataRetainsBasispointsSettings(t *testing.T) {
	account := service.Account{
		Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth,
		Credentials: map[string]any{"access_token": "private-token"},
		Extra: map[string]any{
			"openai_oauth_responses_endpoint":            "basispoints",
			"openai_basispoints_auto_disable_on_403":     true,
			"openai_basispoints_cache_creation_as_input": true,
			"unrelated_private_value":                    "private-value",
		},
	}
	metadata := buildSchedulerMetadataAccount(account)
	require.True(t, metadata.UsesBasispointsResponses())
	require.True(t, metadata.IsBasispointsAutoDisableOn403Enabled())
	require.Equal(t, true, metadata.Extra["openai_basispoints_cache_creation_as_input"])
	require.NotContains(t, metadata.Extra, "unrelated_private_value")
	require.NotContains(t, metadata.Credentials, "access_token")
	account.Extra["openai_oauth_responses_endpoint"] = "chatgpt_codex"
	updated := buildSchedulerMetadataAccount(account)
	require.False(t, updated.UsesBasispointsResponses())
	require.False(t, updated.IsBasispointsAutoDisableOn403Enabled())
	require.True(t, metadata.UsesBasispointsResponses(), "previous projection must not be mutated")
}
