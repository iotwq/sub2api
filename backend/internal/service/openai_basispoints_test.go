package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIOAuthResponsesEndpointModeDefaultsToChatGPT(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	require.Equal(t, openAIOAuthResponsesEndpointChatGPT, openAIOAuthResponsesEndpointMode(account))

	account.Extra = map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis}
	require.Equal(t, openAIOAuthResponsesEndpointBasis, openAIOAuthResponsesEndpointMode(account))

	account.Extra[openAIOAuthResponsesEndpointExtraKey] = "unknown"
	require.Equal(t, openAIOAuthResponsesEndpointChatGPT, openAIOAuthResponsesEndpointMode(account))
}

func TestPrepareOpenAIBasispointsBodyDowngradesMax(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra:    map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
	}
	got, bridge, err := prepareOpenAIBasispointsBody(nil, account, []byte(`{"model":"gpt-6-sol","input":"hi","reasoning":{"effort":"max"}}`))
	require.NoError(t, err)
	require.Equal(t, "xhigh", gjson.GetBytes(got, "reasoning_effort").String())
	require.False(t, gjson.GetBytes(got, "reasoning").Exists())
	require.Equal(t, "max", bridge.RequestedEffort)
}

func TestApplyOpenAIBasispointsHeadersUsesStoredAccountID(t *testing.T) {
	account := &Account{
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "acct-test"},
		Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
	}
	headers := make(http.Header)
	require.NoError(t, applyOpenAIBasispointsHeaders(context.Background(), nil, account, "opaque-token", headers))
	require.Equal(t, "acct-test", headers.Get("chatgpt-account-id"))
	require.Equal(t, "acct-test", headers.Get("x-openai-account-id"))
	require.Equal(t, "chatgpt", headers.Get("x-basispoints-auth-mode"))
	require.Equal(t, "", headers.Get("authorization"))
}

func TestNewBasispointsRequestUsesDedicatedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	account := &Account{
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "acct-test"},
		Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
	}
	wire, _, err := prepareOpenAIBasispointsBody(c, account, []byte(`{"model":"gpt-6-sol","input":"hi","reasoning":{"effort":"max"}}`))
	require.NoError(t, err)
	req, err := newOpenAIBasispointsRequest(context.Background(), nil, account, wire, "oauth-token")
	require.NoError(t, err)
	require.Equal(t, basispointsResponsesURL, req.URL.String())
	require.Equal(t, "acct-test", req.Header.Get("chatgpt-account-id"))
	require.Equal(t, "acct-test", req.Header.Get("x-openai-account-id"))
	require.Equal(t, "chatgpt", req.Header.Get("x-basispoints-auth-mode"))
	require.Equal(t, "xhigh", gjson.GetBytes(mustReadRequestBody(t, req), "reasoning_effort").String())
	require.Equal(t, "https://bps.openai.com", req.Header.Get("Origin"))
	require.Equal(t, "excel", req.Header.Get("x-openai-internal-basispoints-client-agent-profile"))
	require.Equal(t, "Bearer oauth-token", req.Header.Get("Authorization"))
	require.True(t, HTTPUpstreamRedirectsDisabled(req.Context()))
}

func TestBuildUpstreamRequestBasispointsDoesNotAffectSelfBuiltImages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	account := &Account{
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "acct-test"},
		Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
	}
	ctx := withOpenAIImagesSelfBuiltRequest(context.Background())
	svc := &OpenAIGatewayService{}
	req, err := svc.buildUpstreamRequest(ctx, c, account, []byte(`{"model":"gpt-image-2","reasoning":{"effort":"max"}}`), "oauth-token", false, "", false)
	require.NoError(t, err)
	require.Equal(t, chatgptCodexURL, req.URL.String())
	require.Empty(t, req.Header.Get("x-openai-account-id"))
	require.Empty(t, req.Header.Get("x-basispoints-auth-mode"))
	require.JSONEq(t, `{"model":"gpt-image-2","reasoning":{"effort":"max"}}`, string(mustReadRequestBody(t, req)))
}

func TestOpenAIChatGPTAccountIDFromAccessToken(t *testing.T) {
	payload, err := json.Marshal(map[string]any{
		"https://api.openai.com/auth": map[string]string{"chatgpt_account_id": "acct-from-token"},
	})
	require.NoError(t, err)
	token := "header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
	require.Equal(t, "acct-from-token", openAIChatGPTAccountIDFromAccessToken(token))
}

func mustReadRequestBody(t *testing.T, req *http.Request) []byte {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	return body
}

type basispointsTestUpstream struct {
	HTTPUpstream
	send func(*http.Request) (*http.Response, error)
}

func (u *basispointsTestUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return u.send(req)
}

func TestBasispointsForwardWireProtocol(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(strconv.FormatBool(stream), func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			c.Request.Header.Set("x-codex-turn-state", "codex-only-state")
			SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)
			account := &Account{ID: 9, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
				Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "acct-test"},
				Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis}}
			upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				body := mustReadRequestBody(t, req)
				require.Equal(t, basispointsResponsesURL, req.URL.String())
				require.Equal(t, "xhigh", gjson.GetBytes(body, "reasoning_effort").String())
				require.False(t, gjson.GetBytes(body, "reasoning").Exists())
				require.Equal(t, "explicit", gjson.GetBytes(body, "model_selection").String())
				require.True(t, gjson.GetBytes(body, "stream").Bool())
				require.Equal(t, "basispoints-excel-plugin", req.Header.Get("x-openai-internal-basispoints-client-product"))
				require.Empty(t, req.Header.Get("x-codex-turn-state"))
				require.Empty(t, req.Header.Get("originator"))
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(
					"data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n" +
						"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_bps\",\"model\":\"gpt-6-sol\",\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"hello\"}]}],\"usage\":{\"input_tokens\":10,\"output_tokens\":3}}}\n\n"))}, nil
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			body := []byte(`{"model":"gpt-6-sol","input":"hi","reasoning":{"effort":"max"},"stream":` + strconv.FormatBool(stream) + `}`)
			result, err := svc.Forward(context.Background(), c, account, body)
			require.NoError(t, err)
			require.Equal(t, 200, rec.Code)
			require.Contains(t, rec.Body.String(), "hello")
			require.Equal(t, "xhigh", *result.ReasoningEffort)
			require.Equal(t, "max", *result.RequestedReasoningEffort)
			require.Equal(t, 10, result.Usage.InputTokens)
			require.Equal(t, 3, result.Usage.OutputTokens)
			require.Equal(t, "/basispoints/api/responses", result.UpstreamEndpoint)
		})
	}
}

func TestBasispointsUsagePreservesRequestedEffortBeforeGroupMapping(t *testing.T) {
	for _, tc := range []struct {
		name, requested, ceiling, forwarded string
	}{
		{"bps caps max", "max", "", "xhigh"},
		{"group maps max to xhigh", "max", "xhigh", "xhigh"},
		{"group maps max to high", "max", "high", "high"},
		{"unchanged high", "high", "", "high"},
	} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", tc.name, stream), func(t *testing.T) {
				upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
					require.Equal(t, tc.forwarded, gjson.GetBytes(mustReadRequestBody(t, req), "reasoning_effort").String())
					return basispointsTestCompletedResponse(t, []any{}), nil
				}}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
				body := []byte(fmt.Sprintf(`{"model":"gpt-6-sol","stream":%t,"reasoning":{"effort":%q},"input":"test"}`, stream, tc.requested))
				ctx := WithRequestedReasoningEffort(context.Background(), tc.requested)
				if tc.ceiling != "" {
					var changed bool
					var err error
					body, changed, err = ApplyOpenAIReasoningEffortPolicy(body, tc.ceiling, nil, "downgrade")
					require.NoError(t, err)
					require.True(t, changed)
				}
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(string(body))).WithContext(ctx)
				account := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
					Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "acct-test"},
					Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis}}
				result, err := svc.Forward(ctx, c, account, body)
				require.NoError(t, err)
				require.Equal(t, tc.requested, optionalStringValue(result.RequestedReasoningEffort))
				require.Equal(t, tc.forwarded, optionalStringValue(result.ReasoningEffort))

				usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
				usageService := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
				require.NoError(t, usageService.RecordUsage(ctx, &OpenAIRecordUsageInput{
					Result: result, APIKey: &APIKey{ID: 10}, User: &User{ID: 20}, Account: account,
				}))
				require.NotNil(t, usageRepo.lastLog)
				require.Equal(t, tc.requested, optionalStringValue(usageRepo.lastLog.RequestedReasoningEffort))
				require.Equal(t, tc.forwarded, optionalStringValue(usageRepo.lastLog.ReasoningEffort))
			})
		}
	}
}

func TestBasispointsAccountSelectionAndDefaultIsolation(t *testing.T) {
	for _, tc := range []struct {
		name, platform, kind, mode string
		want                       bool
	}{
		{"oauth", PlatformOpenAI, AccountTypeOAuth, "basispoints", true},
		{"setup", PlatformOpenAI, AccountTypeSetupToken, "basispoints", true},
		{"default", PlatformOpenAI, AccountTypeOAuth, "", false},
		{"unknown", PlatformOpenAI, AccountTypeOAuth, "invalid", false},
		{"apikey", PlatformOpenAI, AccountTypeAPIKey, "basispoints", false},
		{"other-platform", PlatformAnthropic, AccountTypeOAuth, "basispoints", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := &Account{Platform: tc.platform, Type: tc.kind, Extra: map[string]any{openAIOAuthResponsesEndpointExtraKey: tc.mode}}
			require.Equal(t, tc.want, shouldForwardOpenAIBasispoints(context.Background(), nil, account, []byte(`{"model":"gpt-6-sol"}`)))
		})
	}
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	body := []byte(`{"model":"gpt-6-sol","reasoning":{"effort":"max"}}`)
	SetOpenAIClientTransport(c, OpenAIClientTransportWS)
	require.False(t, shouldForwardOpenAIBasispoints(context.Background(), c, account, body))
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)
	require.False(t, shouldForwardOpenAIBasispoints(withOpenAIImagesSelfBuiltRequest(context.Background()), c, account, body))
	require.True(t, shouldForwardOpenAIBasispoints(context.Background(), c, account, []byte(`{"tools":[{"type":"image_generation"}]}`)))
	account.Extra[openAIOAuthResponsesEndpointExtraKey] = "chatgpt_codex"
	svc := &OpenAIGatewayService{}
	for _, passthrough := range []bool{false, true} {
		var req *http.Request
		var err error
		if passthrough {
			req, err = svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, body, "token")
		} else {
			req, err = svc.buildUpstreamRequest(context.Background(), c, account, body, "token", false, "", false)
		}
		require.NoError(t, err)
		require.Equal(t, chatgptCodexURL, req.URL.String())
		require.Empty(t, req.Header.Get("x-basispoints-auth-mode"))
		require.Equal(t, "max", gjson.GetBytes(mustReadRequestBody(t, req), "reasoning.effort").String())
	}
}

func TestBasispointsUsesSelectedWorkspaceBeforeTokenClaim(t *testing.T) {
	claim := base64.RawURLEncoding.EncodeToString([]byte(`{"https://api.openai.com/auth":{"chatgpt_account_id":"personal"}}`))
	token := "header." + claim + ".signature"
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"chatgpt_account_id": "team"}}
	id, err := resolveOpenAIBasispointsAccountID(context.Background(), nil, account, token)
	require.NoError(t, err)
	require.Equal(t, "team", id)
	delete(account.Credentials, "chatgpt_account_id")
	id, err = resolveOpenAIBasispointsAccountID(context.Background(), nil, account, token)
	require.NoError(t, err)
	require.Equal(t, "personal", id)
}

func basispointsTestCompletedResponse(t *testing.T, output any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(gin.H{"type": "response.completed", "response": gin.H{
		"id": "resp_bps", "status": "completed", "model": "gpt-6-sol", "output": output,
		"usage": gin.H{"input_tokens": 12, "output_tokens": 4},
	}})
	require.NoError(t, err)
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader("data: " + string(payload) + "\n\n"))}
}

func TestBasispointsForwardToolRoundTrip(t *testing.T) {
	account := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeSetupToken,
		Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "acct-test", "model_mapping": map[string]any{"public-model": "gpt-6-sol"}},
		Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints", "openai_passthrough": true}}
	code, err := json.Marshal(gin.H{"name": "shell", "arguments": gin.H{"command": "pwd"}})
	require.NoError(t, err)
	args, err := json.Marshal(gin.H{"code": string(code)})
	require.NoError(t, err)
	call := gin.H{"type": "function_call", "id": "fc_native", "call_id": "call_native", "name": "run_officejs", "arguments": string(args)}
	requests := 0
	upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
		requests++
		body := mustReadRequestBody(t, req)
		require.Equal(t, "gpt-6-sol", gjson.GetBytes(body, "model").String())
		require.False(t, gjson.GetBytes(body, "tools").Exists())
		if requests == 1 {
			require.Contains(t, string(body), "shell")
			return basispointsTestCompletedResponse(t, []any{call}), nil
		}
		input := gjson.GetBytes(body, "input").Array()
		require.Equal(t, "run_officejs", input[len(input)-2].Get("name").String())
		require.Equal(t, "call_native", input[len(input)-1].Get("call_id").String())
		require.Equal(t, "/workspace", input[len(input)-1].Get("output").String())
		return basispointsTestCompletedResponse(t, []any{}), nil
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	input := []any{gin.H{"role": "user", "content": "where am I?"}}
	for turn := 0; turn < 2; turn++ {
		body, err := json.Marshal(gin.H{"model": "public-model", "input": input, "tools": []any{gin.H{"type": "function", "name": "shell", "parameters": gin.H{"type": "object"}}}})
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
		c.Request.Header.Set("thread-id", "basispoints-tool-test")
		result, err := svc.Forward(context.Background(), c, account, body)
		require.NoError(t, err)
		require.Equal(t, 12, result.Usage.InputTokens)
		require.Equal(t, "public-model", result.Model)
		require.Equal(t, "gpt-6-sol", result.UpstreamModel)
		if turn == 0 {
			require.NotContains(t, rec.Body.String(), "run_officejs")
			translated := gjson.Get(rec.Body.String(), "output.0")
			require.Equal(t, "shell", translated.Get("name").String())
			require.JSONEq(t, `{"command":"pwd"}`, translated.Get("arguments").String())
			input = append(input, translated.Value(), gin.H{"type": "function_call_output", "call_id": "call_native", "output": "/workspace"})
		}
	}
	require.Equal(t, 2, requests)
}

func TestBasispointsAccountTestUsesSelectedEndpoint(t *testing.T) {
	for _, kind := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
		t.Run(kind, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/1/test", nil)
			upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				require.Equal(t, basispointsResponsesURL, req.URL.String())
				require.Contains(t, string(mustReadRequestBody(t, req)), "my test prompt")
				return basispointsTestCompletedResponse(t, []any{}), nil
			}}
			account := &Account{ID: 43, Platform: PlatformOpenAI, Type: kind,
				Credentials: map[string]any{"access_token": "token", "chatgpt_account_id": "acct"},
				Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
			svc := &AccountTestService{httpUpstream: upstream}
			require.NoError(t, svc.testOpenAIAccountConnection(c, account, "gpt-6-sol", "my test prompt", ""))
			require.Contains(t, recorder.Body.String(), `"success":true`)
		})
	}
}

func TestBasispointsCompactUsesResponsesEndpoint(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", nil)
	account := &Account{ID: 45, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Credentials: map[string]any{"access_token": "token", "chatgpt_account_id": "acct"},
		Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
	upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
		require.Equal(t, basispointsResponsesURL, req.URL.String())
		input := gjson.GetBytes(mustReadRequestBody(t, req), "input").Array()
		require.Equal(t, "compaction_trigger", input[len(input)-1].Get("type").String())
		return basispointsTestCompletedResponse(t, []any{gin.H{"type": "compaction", "encrypted_content": "compact-state"}}), nil
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	_, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-6-sol","input":"compress this"}`))
	require.NoError(t, err)
	require.Contains(t, rec.Body.String(), "compact-state")
}

func TestBasispointsUpstreamErrorsPreserveStatusAndFailover(t *testing.T) {
	for _, status := range []int{400, 401, 403, 422, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			account := &Account{ID: 46, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
				Credentials: map[string]any{"access_token": "secret-token", "chatgpt_account_id": "acct"},
				Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
			upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Header: http.Header{"X-Request-Id": {"bps-error-id"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"secret-token"}}`))}, nil
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			result, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-6-sol","input":"hi"}`))
			require.Error(t, err)
			require.Nil(t, result)
			require.NotContains(t, rec.Body.String(), "secret-token")
			if status == http.StatusTooManyRequests {
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Equal(t, status, failover.StatusCode)
				require.False(t, c.Writer.Written())
			} else if status == 503 {
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Equal(t, status, failover.StatusCode)
				require.False(t, c.Writer.Written())
			} else {
				require.Equal(t, status, rec.Code)
				if status == 429 {
					require.Contains(t, rec.Body.String(), "basispoints_rate_limited")
				} else {
					require.Contains(t, rec.Body.String(), "basispoints_upstream_error")
				}
			}
		})
	}
}

func TestBasispointsInvalidHistoryFailsBeforeSending(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Extra: map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}}
	_, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-6-sol","input":"hi","previous_response_id":"resp_other_route"}`))
	require.Error(t, err)
	require.Equal(t, 400, c.Writer.Status())
}

func TestBasispointsKeepsConcurrencyControlWithoutCodexTLS(t *testing.T) {
	account := protectionAccount()
	account.Extra[openAIOAuthResponsesEndpointExtraKey] = "basispoints"
	account.Extra[AccountTrafficPolicyKey].(map[string]any)["tls_profile"] = "nodejs24"
	cache := &protectionTestCache{}
	upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	req, err := newOpenAIBasispointsRequest(context.Background(), nil, account, []byte(`{}`), "token")
	require.NoError(t, err)
	resp, err := svc.doOpenAIBasispoints(req, account)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 1, upstream.calls)
	require.Zero(t, upstream.tlsCalls)
	require.EqualValues(t, 1, cache.calls.Load())
	require.EqualValues(t, 1, cache.finishes.Load())
}

func TestBasispointsCanceledAdmissionDoesNotSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache := &cancellationTrafficCache{cancel: cancel, grantOnFirst: true}
	upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
	account := protectionAccount()
	account.Extra[openAIOAuthResponsesEndpointExtraKey] = "basispoints"
	account.Credentials["access_token"] = "token"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	_, err := svc.Forward(ctx, c, account, []byte(`{"model":"gpt-6-sol","input":"hi"}`))
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, upstream.calls)
	require.Eventually(t, func() bool { return cache.finishes.Load() == 1 }, time.Second, time.Millisecond)
}

func TestBasispointsTruncatedOrInvalidStreamDoesNotSucceed(t *testing.T) {
	for _, wire := range []string{
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial\"}\n\n",
		"data: {\"type\":\"response.completed\",\"response\":{\"output\":[{\"type\":\"function_call\",\"name\":\"undeclared\",\"call_id\":\"unknown\"}]}}\n\n",
	} {
		for _, stream := range []bool{false, true} {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(wire))}, nil
			}}
			account := &Account{ID: 47, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
				Credentials: map[string]any{"access_token": "token", "chatgpt_account_id": "acct"},
				Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			result, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-6-sol","input":"hi","stream":`+strconv.FormatBool(stream)+`}`))
			require.Error(t, err, "a partial response or an undeclared tool must not be reported as successful")
			require.Nil(t, result, "failed attempts without billable usage must not create usage rows")
		}
	}
}
