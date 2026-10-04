package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type protectionTestCache struct {
	AccountTrafficCache
	calls      atomic.Int32
	finishes   atomic.Int32
	status     atomic.Int32
	allowAfter int32
}

func (s *protectionTestCache) Acquire(context.Context, AccountTrafficPlan, string) (AccountTrafficAdmission, error) {
	return AccountTrafficAdmission{Allowed: s.calls.Add(1) > s.allowAfter, RetryAfter: time.Millisecond}, nil
}
func (s *protectionTestCache) Finish(_ context.Context, _ AccountTrafficPlan, _ string, status int, _ int64) error {
	s.status.Store(int32(status))
	s.finishes.Add(1)
	return nil
}

type protectionTestUpstream struct {
	HTTPUpstream
	control  *AccountTrafficService
	calls    int
	tlsCalls int
}

func (s *protectionTestUpstream) AccountTrafficController() *AccountTrafficService { return s.control }
func (s *protectionTestUpstream) Do(*http.Request, string, int64, int) (*http.Response, error) {
	s.calls++
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok"))}, nil
}
func (s *protectionTestUpstream) DoWithCodexTLS(r *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	s.tlsCalls++
	return s.Do(r, proxy, id, concurrency)
}
func protectionAccount() *Account {
	return &Account{ID: 91, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 8,
		Credentials: map[string]any{"chatgpt_account_id": "test-account"},
		Extra:       map[string]any{codexFingerprintSeedExtraKey: "00000000-0000-4000-8000-000000000091", AccountTrafficPolicyKey: map[string]any{"enabled": true, "wait_seconds": 0}}}
}

func TestCodexProtectionHTTPRouteScopeAndRetries(t *testing.T) {
	for _, route := range []string{"/backend-api/codex/responses", "/backend-api/codex/responses/compact", "/v1/images/generations", "/v1/videos", "/v1/models"} {
		t.Run(route, func(t *testing.T) {
			a := protectionAccount()
			a.Extra[AccountTrafficPolicyKey].(map[string]any)["tls_profile"] = "nodejs24"
			cache := &protectionTestCache{}
			up := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
			cfg := &config.Config{}
			cfg.Gateway.TLSFingerprint.Enabled = true
			svc := &OpenAIGatewayService{httpUpstream: up, cfg: cfg}
			for i := 0; i < 2; i++ {
				req := httptest.NewRequest(http.MethodPost, "https://example.test"+route, nil)
				resp, err := svc.doOpenAIUpstream(req, "", a)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())
			}
			expected := 0
			if strings.Contains(route, "/responses") {
				expected = 2
			}
			require.EqualValues(t, expected, cache.calls.Load())
			require.Equal(t, expected, up.tlsCalls)
			require.Equal(t, 2, up.calls)
		})
	}
}

func TestCodexProtectionBoundedWaitAndCancellation(t *testing.T) {
	plan, err := AccountTrafficPlanFor(protectionAccount())
	require.NoError(t, err)
	plan.Policy.WaitSeconds = 1
	cache := &protectionTestCache{allowAfter: 1}
	_, permit, err := NewAccountTrafficService(cache).Begin(context.Background(), plan)
	require.NoError(t, err)
	require.EqualValues(t, 2, cache.calls.Load())
	permit.Finish(200)
	require.EqualValues(t, 1, cache.finishes.Load())
	for _, timeout := range []time.Duration{20 * time.Millisecond, 2 * time.Second} {
		cache := &protectionTestCache{allowAfter: 1000}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		start := time.Now()
		_, permit, err := NewAccountTrafficService(cache).Begin(ctx, plan)
		cancel()
		require.Error(t, err)
		require.Nil(t, permit)
		require.Less(t, time.Since(start), 1500*time.Millisecond)
		require.Zero(t, cache.finishes.Load())
	}
}

func TestCodexProtectionWSAdmissionRespectsWriteCancellation(t *testing.T) {
	plan, err := AccountTrafficPlanFor(protectionAccount())
	require.NoError(t, err)
	plan.Policy.WaitSeconds = 30
	cache := &protectionTestCache{allowAfter: 1000}
	inner := &trafficFrameStub{}
	frame := &accountTrafficFrameConn{inner: inner, control: NewAccountTrafficService(cache), plan: plan, ctx: context.Background()}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, frame.WriteFrame(ctx, coderws.MessageText, []byte(`{"type":"response.create"}`)), context.DeadlineExceeded)
	require.Zero(t, inner.writes)
	require.NoError(t, frame.Close())
}

func TestCodexProtectionWSTerminalFailureIsNotSuccess(t *testing.T) {
	plan, err := AccountTrafficPlanFor(protectionAccount())
	require.NoError(t, err)
	cache := &protectionTestCache{}
	_, permit, err := NewAccountTrafficService(cache).Begin(context.Background(), plan)
	require.NoError(t, err)
	permit.ObserveEvent([]byte(`{"type":"response.failed","response":{"error":{"code":"rate_limit_exceeded"}}}`))
	finishAccountTrafficTurn(permit, nil)
	require.EqualValues(t, 429, cache.status.Load())
	require.EqualValues(t, 1, cache.finishes.Load())
	_, truncated, err := NewAccountTrafficService(cache).Begin(context.Background(), plan)
	require.NoError(t, err)
	finishAccountTrafficTurn(truncated, nil)
	require.Zero(t, cache.status.Load(), "no terminal event is not recovery evidence")
}

func TestCodexProtectionNativeDrainRetainsLeaseAfterClientDisconnect(t *testing.T) {
	cache := &protectionTestCache{}
	upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
	ctx, disconnect := context.WithCancel(context.Background())
	operation, permit, err := beginAccountTrafficTurn(ctx, upstream, protectionAccount())
	require.NoError(t, err)
	disconnect()
	require.ErrorIs(t, operation.Err(), context.Canceled)
	require.NoError(t, permit.ctx.Err(), "draining upstream still owns the concurrency lease")
	require.Zero(t, cache.finishes.Load())
	permit.ObserveEvent([]byte(`{"type":"response.completed"}`))
	finishAccountTrafficTurn(permit, nil)
	require.EqualValues(t, 1, cache.finishes.Load())
	require.EqualValues(t, 200, cache.status.Load())
	require.ErrorIs(t, permit.ctx.Err(), context.Canceled)
}

func TestCodexProtectionDeviceStableButConversationsIsolated(t *testing.T) {
	a := protectionAccount()
	render := func(key int64, session string) (string, string, openAIWSHandshakeCompatibilityKey) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
		c.Request.Header.Set("session-id", session)
		c.Set("api_key", &APIKey{ID: key})
		raw := []byte(`{"type":"response.create","input":"context","reasoning":{"effort":"high"},"client_metadata":{"session_id":"` + session + `","installation_id":"client-device"}}`)
		body, _, err := applyCodexAccountAndProtectionIdentityRaw(c, a, raw)
		require.NoError(t, err)
		headers := c.Request.Header.Clone()
		applyStagedCodexFingerprintHeaders(c, a, headers)
		device := gjson.GetBytes(body, `client_metadata.x-codex-installation-id`).String()
		require.Equal(t, device, headers.Get("x-codex-installation-id"))
		require.Equal(t, device, gjson.GetBytes(body, `client_metadata.installation_id`).String())
		httpIDs := resolveCodexFingerprintIDsFromRequest(a, c.Request.Header)
		require.Equal(t, httpIDs.installationID, device)
		require.Equal(t, "high", gjson.GetBytes(body, "reasoning.effort").String())
		again, _, err := applyCodexAccountAndProtectionIdentityRaw(c, a, raw)
		require.NoError(t, err)
		require.Equal(t, gjson.GetBytes(body, "client_metadata").Raw, gjson.GetBytes(again, "client_metadata").Raw)
		return device, gjson.GetBytes(body, "client_metadata.session_id").String(), normalizeOpenAIWSHandshakeCompatibility(a, headers)
	}
	d1, s1, k1 := render(1, "session-one")
	d2, s2, k2 := render(1, "session-two")
	d3, s3, _ := render(2, "session-one")
	require.NotEmpty(t, d1)
	require.Equal(t, d1, d2)
	require.Equal(t, d1, d3)
	require.NotEqual(t, s1, s2)
	require.NotEqual(t, s1, s3)
	require.NotEqual(t, k1, k2)
	a.Extra[codexFingerprintModeExtraKey] = "off"
	off1 := normalizeOpenAIWSHandshakeCompatibility(a, http.Header{"Session-Id": {"session-one"}})
	off2 := normalizeOpenAIWSHandshakeCompatibility(a, http.Header{"Session-Id": {"session-two"}})
	require.NotEqual(t, off1, off2)
}

func TestCodexProtectionSeedLifecycleAndOptOut(t *testing.T) {
	a := protectionAccount()
	extra := prepareCodexFingerprintExtraForCreate(a.Platform, a.Type, a.Extra)
	seed, ok := codexFingerprintSeed(extra)
	require.True(t, ok)
	a.Extra = extra
	updated := prepareCodexFingerprintExtraForUpdate(a, extra)
	require.Equal(t, seed, updated[codexFingerprintSeedExtraKey])
	a.Extra[codexFingerprintModeExtraKey] = "off"
	require.Equal(t, codexFingerprintOff, a.GetCodexFingerprintMode())
	a.Extra[AccountTrafficPolicyKey].(map[string]any)["tls_profile"] = "nodejs24"
	_, err := codexProtectionTLSProfile(a, &config.Config{})
	require.Error(t, err)
	a.Type = AccountTypeAPIKey
	profile, err := codexProtectionTLSProfile(a, &config.Config{})
	require.NoError(t, err)
	require.Nil(t, profile)
}
