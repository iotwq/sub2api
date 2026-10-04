package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCodexProtectionWSContinuationIgnoresRequestTraceID(t *testing.T) {
	for _, mode := range []string{"off", "device", "session", "full"} {
		t.Run(mode, func(t *testing.T) {
			account := protectionAccount()
			account.Extra[codexFingerprintModeExtraKey] = mode
			cfg := &config.Config{}
			cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 4
			cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 4
			pool := newOpenAIWSConnPool(cfg)
			defer pool.Close()
			headers := http.Header{}
			headers.Set("session-id", "same-session")
			headers.Set("thread-id", "same-thread")
			headers.Set("x-codex-installation-id", "same-device")
			headers.Set("x-client-request-id", "request-1")
			req := openAIWSAcquireRequest{Account: account, WSURL: "wss://example.test/responses", Headers: headers,
				PreferredConnID: "existing", ForcePreferredConn: true}
			conn := newOpenAIWSConn("existing", account.ID, &openAIWSFakeConn{}, nil)
			conn.handshakeCompatibility = openAIWSAcquireCompatibility(req)
			ap := pool.getOrCreateAccountPool(account.ID)
			ap.mu.Lock()
			ap.conns[conn.id] = conn
			ap.mu.Unlock()
			req.Headers = headers.Clone()
			req.Headers.Set("x-client-request-id", "request-2")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			lease, err := pool.Acquire(ctx, req)
			require.NoError(t, err, "changing only the trace ID must preserve the continuation connection")
			require.Equal(t, conn.id, lease.ConnID())
			lease.Release()
		})
	}
}

func TestCodexProtectionWSCompatibilityKeepsIdentityIsolation(t *testing.T) {
	account := protectionAccount()
	headers := http.Header{}
	for _, name := range []string{"session-id", "session_id", "thread-id", "conversation_id", "x-codex-installation-id", "x-codex-window-id", "x-codex-beta-features"} {
		headers.Set(name, "original")
	}
	original := normalizeOpenAIWSHandshakeCompatibility(account, headers)
	for name := range headers {
		t.Run(name, func(t *testing.T) {
			next := headers.Clone()
			next.Set(name, "different")
			require.NotEqual(t, original, normalizeOpenAIWSHandshakeCompatibility(account, next))
		})
	}
}

func TestCodexProtectionWSCompatibilityDisabledKeepsLegacyBehavior(t *testing.T) {
	account := protectionAccount()
	account.Extra[AccountTrafficPolicyKey].(map[string]any)["enabled"] = false
	headers := http.Header{}
	headers.Set("x-client-request-id", "request-1")
	next := headers.Clone()
	next.Set("x-client-request-id", "request-2")
	for _, mode := range []string{"off", "device", "session", "full"} {
		t.Run(mode, func(t *testing.T) {
			account.Extra[codexFingerprintModeExtraKey] = mode
			before := normalizeOpenAIWSHandshakeCompatibility(account, headers)
			after := normalizeOpenAIWSHandshakeCompatibility(account, next)
			if mode == "off" || mode == "device" {
				require.Equal(t, before, after)
			} else {
				require.NotEqual(t, before, after)
			}
		})
	}
}
