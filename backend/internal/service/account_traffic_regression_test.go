package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountTrafficTerminalFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		payload string
		status  int
	}{
		{`{"type":"response.done","response":{"status":"failed","error":{"code":"rate_limit_exceeded"}}}`, 429},
		{`{"type":"response.done","response":{"status":"failed","status_details":{"error":{"type":"server_error"}}}}`, 503},
		{`{"type":"response.failed","response":{"error":{"status_code":502}}}`, 502},
		{`{"type":"response.done","response":{"status":"cancelled"}}`, 499},
		{`{"type":"response.done","response":{"status":"completed"}}`, 200},
		{`{"type":"response.failed","response":{"error":{"code":"invalid_request"}}}`, 499},
	} {
		status, terminal := accountTrafficEventStatus([]byte(tc.payload))
		require.True(t, terminal)
		require.Equal(t, tc.status, status, tc.payload)
		body := &accountTrafficBody{}
		body.inspect([]byte("data: " + tc.payload + "\n\ndata: [DONE]\n\n"))
		require.EqualValues(t, tc.status, body.terminalStatus.Load(), "SSE must preserve the same terminal classification")
	}
}

func TestAccountTrafficDisabledDoesNotCreateLease(t *testing.T) {
	p := DefaultAccountTrafficPolicy()
	p.AdaptiveEnabled = true
	ctx, permit, err := NewAccountTrafficService(nil).Begin(context.Background(), AccountTrafficPlan{AccountID: 42, Policy: p})
	require.NoError(t, err)
	require.Nil(t, permit)
	require.NoError(t, ctx.Err())
}

func TestAccountTrafficOtherProvidersUnchanged(t *testing.T) {
	for _, platform := range []string{PlatformGrok, PlatformGemini, PlatformAnthropic, PlatformOpenAI} {
		a := &Account{ID: 42, Platform: platform, Type: AccountTypeAPIKey, Extra: map[string]any{AccountTrafficPolicyKey: map[string]any{"enabled": true}}}
		plan, err := AccountTrafficPlanFor(a)
		require.NoError(t, err)
		require.False(t, plan.Policy.Enabled())
		require.Equal(t, "off", a.RequestIntegrityMode())
	}
}
