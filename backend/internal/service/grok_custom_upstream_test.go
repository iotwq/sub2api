package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func grokCustomUpstreamTestConfig() *config.Config {
	return &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
		Enabled: false,
	}}}
}

func grokCustomUpstreamAPIKeyAccount(baseURL string) *Account {
	return &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "grok-test-key",
			"base_url": baseURL,
		},
	}
}

func TestGrokAPIKeyCustomUpstreamBuildsAllEndpointURLs(t *testing.T) {
	account := grokCustomUpstreamAPIKeyAccount("https://1.cnmz.de/v1")
	cfg := grokCustomUpstreamTestConfig()
	tests := []struct {
		build func() (string, error)
		want  string
	}{
		{build: func() (string, error) { return buildGrokModelsURL(account, cfg) }, want: "https://1.cnmz.de/v1/models"},
		{build: func() (string, error) { return buildGrokResponsesURL(account, cfg) }, want: "https://1.cnmz.de/v1/responses"},
		{build: func() (string, error) { return buildGrokChatCompletionsURL(account, cfg) }, want: "https://1.cnmz.de/v1/chat/completions"},
		{build: func() (string, error) { return buildGrokMediaURL(account, cfg, GrokMediaEndpointImagesGenerations, "") }, want: "https://1.cnmz.de/v1/images/generations"},
		{build: func() (string, error) { return buildGrokMediaURL(account, cfg, GrokMediaEndpointImagesEdits, "") }, want: "https://1.cnmz.de/v1/images/edits"},
		{build: func() (string, error) { return buildGrokMediaURL(account, cfg, GrokMediaEndpointVideosGenerations, "") }, want: "https://1.cnmz.de/v1/videos/generations"},
		{build: func() (string, error) {
			return buildGrokMediaURL(account, cfg, GrokMediaEndpointVideoStatus, "req 123")
		}, want: "https://1.cnmz.de/v1/videos/req%20123"},
		{build: func() (string, error) {
			return buildGrokMediaURL(account, cfg, GrokMediaEndpointVideoContent, "req 123")
		}, want: "https://1.cnmz.de/v1/videos/req%20123/content"},
	}

	for _, tt := range tests {
		got, err := tt.build()
		require.NoError(t, err)
		require.Equal(t, tt.want, got)
	}
}

func TestGrokAPIKeyCustomUpstreamHonorsGlobalAllowlistPolicy(t *testing.T) {
	account := grokCustomUpstreamAPIKeyAccount("https://1.cnmz.de/v1")
	allowedCfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
		Enabled:       true,
		UpstreamHosts: []string{"1.cnmz.de"},
	}}}

	got, err := buildGrokResponsesURL(account, allowedCfg)
	require.NoError(t, err)
	require.Equal(t, "https://1.cnmz.de/v1/responses", got)

	blockedCfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
		Enabled:       true,
		UpstreamHosts: []string{"api.x.ai"},
	}}}
	_, err = buildGrokResponsesURL(account, blockedCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "base URL rejected by URL security policy")

	privateAccount := grokCustomUpstreamAPIKeyAccount("https://127.0.0.1/v1")
	privateCfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
		Enabled:       true,
		UpstreamHosts: []string{"127.0.0.1"},
	}}}
	_, err = buildGrokResponsesURL(privateAccount, privateCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "base URL rejected by URL security policy")
}

func TestGrokOAuthCustomUpstreamFollowsGlobalPolicy(t *testing.T) {
	t.Setenv(xai.EnvAllowUnsafeURLOverrides, "false")
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"base_url": "https://1.cnmz.de/v1",
		},
	}

	got, err := buildGrokResponsesURL(account, grokCustomUpstreamTestConfig())
	require.NoError(t, err)
	require.Equal(t, "https://1.cnmz.de/v1/responses", got)
}

func TestGrokCustomUpstreamCallSitesUseAPIKeyPolicy(t *testing.T) {
	account := grokCustomUpstreamAPIKeyAccount("https://1.cnmz.de/v1")
	svc := &OpenAIGatewayService{cfg: grokCustomUpstreamTestConfig()}

	req, err := buildGrokResponsesRequest(context.Background(), nil, account, []byte(`{"model":"grok-4.5"}`), "grok-test-key", "", svc.cfg)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "https://1.cnmz.de/v1/responses", req.URL.String())

	chatURL, err := svc.rawChatCompletionsURL(account)
	require.NoError(t, err)
	require.Equal(t, "https://1.cnmz.de/v1/chat/completions", chatURL)

	mediaURL, err := buildGrokMediaURL(account, svc.cfg, GrokMediaEndpointImagesGenerations, "")
	require.NoError(t, err)
	require.Equal(t, "https://1.cnmz.de/v1/images/generations", mediaURL)

	videoURL, err := buildGrokMediaURL(account, svc.cfg, GrokMediaEndpointVideoContent, "req 123")
	require.NoError(t, err)
	require.Equal(t, "https://1.cnmz.de/v1/videos/req%20123/content", videoURL)
}

func TestAccountTestServiceGrokAPIKeyUsesCustomUpstream(t *testing.T) {
	account := grokCustomUpstreamAPIKeyAccount("https://1.cnmz.de/v1")
	account.ID = 71
	account.Concurrency = 1
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"data: {\"type\":\"response.output_text.delta\",\"delta\":\"ok\"}\n\n" +
				"data: {\"type\":\"response.completed\"}\n\n",
		)),
	}}
	svc := &AccountTestService{
		httpUpstream: upstream,
		cfg:          grokCustomUpstreamTestConfig(),
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/71/test", nil)

	err := svc.testGrokAccountConnection(c, account, "grok-4.5", "", AccountTestModeDefault, AccountTestOptions{})
	require.NoError(t, err)
	require.Equal(t, "https://1.cnmz.de/v1/responses", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer grok-test-key", upstream.lastReq.Header.Get("Authorization"))
	require.Contains(t, recorder.Body.String(), `"type":"test_complete"`)
}
