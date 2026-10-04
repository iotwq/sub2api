package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func basispointsCacheUsageFixture() map[string]any {
	return map[string]any{
		"input_tokens": 1000, "output_tokens": 50, "total_tokens": 1050,
		"input_tokens_details":  map[string]any{"cached_tokens": 100, "cache_write_tokens": 200, "cache_creation_tokens": 200},
		"prompt_tokens_details": map[string]any{"cached_tokens": 100, "cache_write_tokens": 200, "cache_creation_tokens": 200},
		"cache_write_tokens":    200, "cache_creation_input_tokens": 200, "cache_write_input_tokens": 200, "cache_creation_tokens": 200,
		"cache_creation":          map[string]any{"ephemeral_5m_input_tokens": 150, "ephemeral_1h_input_tokens": 50},
		"cache_read_input_tokens": 100,
	}
}

func requireBasispointsCacheUsage(t *testing.T, payload []byte, prefix string, wantCreation int) {
	t.Helper()
	for _, field := range []string{
		"input_tokens_details.cache_write_tokens", "prompt_tokens_details.cache_write_tokens",
		"input_tokens_details.cache_creation_tokens", "prompt_tokens_details.cache_creation_tokens",
		"cache_write_tokens", "cache_creation_input_tokens", "cache_write_input_tokens", "cache_creation_tokens",
	} {
		require.Equal(t, int64(wantCreation), gjson.GetBytes(payload, prefix+field).Int(), prefix+field)
		require.True(t, gjson.GetBytes(payload, prefix+field).Exists(), prefix+field)
	}
	for field, want := range map[string]int{
		"input_tokens": 1000, "output_tokens": 50, "total_tokens": 1050,
		"input_tokens_details.cached_tokens": 100, "prompt_tokens_details.cached_tokens": 100, "cache_read_input_tokens": 100,
	} {
		require.Equal(t, int64(want), gjson.GetBytes(payload, prefix+field).Int(), prefix+field)
	}
	for field, original := range map[string]int{"ephemeral_5m_input_tokens": 150, "ephemeral_1h_input_tokens": 50} {
		want := original
		if wantCreation == 0 {
			want = 0
		}
		require.Equal(t, int64(want), gjson.GetBytes(payload, prefix+"cache_creation."+field).Int())
	}
}

func TestNormalizeOpenAIBasispointsUsage(t *testing.T) {
	payload, err := json.Marshal(map[string]any{
		"usage": basispointsCacheUsageFixture(), "response": map[string]any{"usage": basispointsCacheUsageFixture()},
		"large_id": json.Number("9007199254740993"), "text": "cache_write_tokens: 200",
	})
	require.NoError(t, err)
	got, err := normalizeOpenAIBasispointsUsage(payload)
	require.NoError(t, err)
	requireBasispointsCacheUsage(t, got, "usage.", 0)
	requireBasispointsCacheUsage(t, got, "response.usage.", 0)
	require.Equal(t, "9007199254740993", gjson.GetBytes(got, "large_id").Raw)
	require.Equal(t, "cache_write_tokens: 200", gjson.GetBytes(got, "text").String())
	for _, original := range [][]byte{[]byte("[DONE]"), []byte("{}"), got} {
		normalized, err := normalizeOpenAIBasispointsUsage(original)
		require.NoError(t, err)
		require.Equal(t, original, normalized)
	}
}

func TestBasispointsCacheCreationAsInputForward(t *testing.T) {
	for _, kind := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
		for _, stream := range []bool{false, true} {
			for _, enabled := range []bool{false, true} {
				for _, path := range []string{"/v1/responses", "/v1/responses/compact"} {
					t.Run(fmt.Sprintf("%s/stream=%t/enabled=%t/%s", kind, stream, enabled, path), func(t *testing.T) {
						rec := httptest.NewRecorder()
						c, _ := gin.CreateTestContext(rec)
						c.Request = httptest.NewRequest(http.MethodPost, path, nil)
						SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)
						account := &Account{ID: 41, Platform: PlatformOpenAI, Type: kind,
							Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "acct-test", "model_mapping": map[string]any{"public-cache-model": "gpt-5.6-sol"}},
							Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints", openAIBasispointsCacheCreationAsInputExtraKey: enabled}}
						payload, err := json.Marshal(map[string]any{"type": "response.completed", "response": map[string]any{
							"id": "resp_cache", "status": "completed", "model": "gpt-5.6-sol", "output": []any{}, "usage": basispointsCacheUsageFixture(),
						}})
						require.NoError(t, err)
						upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
							require.Equal(t, basispointsResponsesURL, req.URL.String())
							return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader("data: " + string(payload) + "\n\n"))}, nil
						}}
						svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
						body, err := json.Marshal(map[string]any{"model": "public-cache-model", "stream": stream, "input": "test"})
						require.NoError(t, err)
						result, err := svc.Forward(context.Background(), c, account, body)
						require.NoError(t, err)
						require.Equal(t, 200, rec.Code)
						require.Equal(t, "/basispoints/api/responses", result.UpstreamEndpoint)
						require.Equal(t, 200, result.Usage.CacheCreationInputTokens, "retain original upstream usage")
						require.Equal(t, 100, result.Usage.CacheReadInputTokens)
						clientBody := rec.Body.Bytes()
						if stream {
							var ok bool
							clientBody, ok = extractCodexFinalResponse(string(clientBody))
							require.True(t, ok)
						}
						wantCreation := 200
						if enabled {
							wantCreation = 0
						}
						requireBasispointsCacheUsage(t, clientBody, "usage.", wantCreation)
						require.Equal(t, "public-cache-model", gjson.GetBytes(clientBody, "model").String(), "normalization must not undo model mapping")
						usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
						usageService := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
						require.NoError(t, usageService.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
							Result: result, APIKey: &APIKey{ID: 10}, User: &User{ID: 20}, Account: account,
						}))
						require.Equal(t, wantCreation, usageRepo.lastLog.CacheCreationTokens)
						require.Equal(t, 900-wantCreation, usageRepo.lastLog.InputTokens)
						require.Equal(t, 100, usageRepo.lastLog.CacheReadTokens)
						require.Equal(t, 1050, usageRepo.lastLog.TotalTokens())
					})
				}
			}
		}
	}
}

func TestBasispointsCacheCreationAsInputResponseIsolation(t *testing.T) {
	for _, tc := range []struct {
		name, mode, endpoint string
		wantCreation         int
	}{
		{"default codex", "chatgpt_codex", "/responses", 200},
		{"native fallback", "basispoints", "/responses", 200},
		{"stale endpoint", "chatgpt_codex", "/basispoints/api/responses", 200},
		{"basispoints enabled", "basispoints", "/basispoints/api/responses", 0},
	} {
		for _, format := range []string{"stream", "sse-to-json", "json"} {
			t.Run(tc.name+"/"+format, func(t *testing.T) {
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				SetActualOpenAIUpstreamEndpoint(c, tc.endpoint)
				account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{
					openAIOAuthResponsesEndpointExtraKey: tc.mode, openAIBasispointsCacheCreationAsInputExtraKey: true,
				}}
				response := map[string]any{"id": "resp_native", "status": "completed", "output": []any{}, "usage": basispointsCacheUsageFixture()}
				payload, err := json.Marshal(response)
				require.NoError(t, err)
				contentType := "application/json"
				if format != "json" {
					payload, err = json.Marshal(map[string]any{"type": "response.completed", "response": response})
					require.NoError(t, err)
					payload = []byte("data: " + string(payload) + "\n\n")
					contentType = "text/event-stream"
				}
				resp := &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(string(payload)))}
				svc := &OpenAIGatewayService{cfg: &config.Config{}}
				if format == "stream" {
					parsed, err := svc.handleStreamingResponse(context.Background(), resp, c, account, time.Now(), "gpt-5.6-sol", "gpt-5.6-sol")
					require.NoError(t, err)
					require.Equal(t, 200, parsed.usage.CacheCreationInputTokens)
				} else {
					parsed, err := svc.handleNonStreamingResponse(context.Background(), resp, c, account, "gpt-5.6-sol", "gpt-5.6-sol")
					require.NoError(t, err)
					require.Equal(t, 200, parsed.usage.CacheCreationInputTokens)
				}
				clientBody := rec.Body.Bytes()
				if format == "stream" {
					var ok bool
					clientBody, ok = extractCodexFinalResponse(string(clientBody))
					require.True(t, ok)
				}
				requireBasispointsCacheUsage(t, clientBody, "usage.", tc.wantCreation)
			})
		}
	}
}

func TestBasispointsCacheCreationAsInputScope(t *testing.T) {
	require.False(t, openAIBasispointsCacheCreationAsInput(nil, "/basispoints/api/responses"))
	for _, tc := range []struct {
		name, platform, kind, mode, endpoint string
		enabled                              any
		want                                 bool
	}{
		{"oauth enabled", PlatformOpenAI, AccountTypeOAuth, "basispoints", "/basispoints/api/responses", true, true},
		{"setup enabled", PlatformOpenAI, AccountTypeSetupToken, "basispoints", "/basispoints/api/responses", true, true},
		{"disabled", PlatformOpenAI, AccountTypeOAuth, "basispoints", "/basispoints/api/responses", false, false},
		{"missing", PlatformOpenAI, AccountTypeOAuth, "basispoints", "/basispoints/api/responses", nil, false},
		{"invalid switch", PlatformOpenAI, AccountTypeOAuth, "basispoints", "/basispoints/api/responses", "true", false},
		{"default codex", PlatformOpenAI, AccountTypeOAuth, "chatgpt_codex", "/responses", true, false},
		{"stale switch", PlatformOpenAI, AccountTypeOAuth, "chatgpt_codex", "/basispoints/api/responses", true, false},
		{"native fallback", PlatformOpenAI, AccountTypeOAuth, "basispoints", "/responses", true, false},
		{"unknown endpoint", PlatformOpenAI, AccountTypeOAuth, "basispoints", "", true, false},
		{"images", PlatformOpenAI, AccountTypeOAuth, "basispoints", "/v1/images/generations", true, false},
		{"api key", PlatformOpenAI, AccountTypeAPIKey, "basispoints", "/basispoints/api/responses", true, false},
		{"other platform", PlatformAnthropic, AccountTypeOAuth, "basispoints", "/basispoints/api/responses", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := &Account{Platform: tc.platform, Type: tc.kind, Extra: map[string]any{
				openAIOAuthResponsesEndpointExtraKey:          tc.mode,
				openAIBasispointsCacheCreationAsInputExtraKey: tc.enabled,
			}}
			require.Equal(t, tc.want, openAIBasispointsCacheCreationAsInput(account, tc.endpoint))
		})
	}
}

func TestBasispointsCacheCreationAsInputRecordUsage(t *testing.T) {
	for _, tc := range []struct {
		name, mode, endpoint string
		enabled              bool
		wantCreation         int
	}{
		{"basispoints on", "basispoints", "/basispoints/api/responses", true, 0},
		{"basispoints off", "basispoints", "/basispoints/api/responses", false, 200},
		{"codex unchanged", "chatgpt_codex", "/responses", true, 200},
		{"native fallback unchanged", "basispoints", "/responses", true, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
			svc := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
			svc.billingService = NewBillingService(svc.cfg, &PricingService{pricingData: map[string]*LiteLLMModelPricing{
				"gpt-5.6-sol": {InputCostPerToken: 5e-6, OutputCostPerToken: 30e-6, CacheReadInputTokenCost: 0.5e-6},
			}})
			result := &OpenAIForwardResult{RequestID: "bps-cache-usage", Model: "gpt-5.6-sol", UpstreamEndpoint: tc.endpoint, Duration: time.Second,
				Usage: OpenAIUsage{InputTokens: 1000, OutputTokens: 50, CacheCreationInputTokens: 200, CacheReadInputTokens: 100}}
			account := &Account{ID: 30, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{
				openAIOAuthResponsesEndpointExtraKey: tc.mode, openAIBasispointsCacheCreationAsInputExtraKey: tc.enabled,
			}}
			require.NoError(t, svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{Result: result, APIKey: &APIKey{ID: 10}, User: &User{ID: 20}, Account: account}))
			log := usageRepo.lastLog
			require.NotNil(t, log)
			require.Equal(t, 900-tc.wantCreation, log.InputTokens)
			require.Equal(t, tc.wantCreation, log.CacheCreationTokens)
			require.Equal(t, 100, log.CacheReadTokens)
			require.Equal(t, 1050, log.TotalTokens())
			require.InDelta(t, float64(900-tc.wantCreation)*5e-6, log.InputCost, 1e-12)
			require.InDelta(t, float64(tc.wantCreation)*6.25e-6, log.CacheCreationCost, 1e-12)
			require.InDelta(t, 100*0.5e-6, log.CacheReadCost, 1e-12)
			require.InDelta(t, 50*30e-6, log.OutputCost, 1e-12)
			require.InDelta(t, log.TotalCost*1.1, log.ActualCost, 1e-12)
			require.Equal(t, 200, result.Usage.CacheCreationInputTokens, "preserve upstream measurement")
		})
	}
}
