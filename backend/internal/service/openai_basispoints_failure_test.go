package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func bpsFailureWire(t *testing.T, kind, code, errorType string, status int) string {
	t.Helper()
	detail := map[string]any{"code": code, "type": errorType, "message": "PRIVATE_UPSTREAM_BODY", "param": "PRIVATE_PARAMETER"}
	event := map[string]any{"type": kind}
	if status != 0 {
		event["status"] = status
	}
	if kind == "error" {
		event["error"] = detail
	} else {
		event["response"] = map[string]any{
			"id": "resp_failure", "status": strings.TrimPrefix(kind, "response."), "error": detail,
			"output": []any{}, "usage": map[string]any{"input_tokens": 7, "output_tokens": 2},
		}
	}
	raw, err := json.Marshal(event)
	require.NoError(t, err)
	return "event: " + kind + "\ndata: " + string(raw) + "\n\n"
}

func TestBasispointsFailureAfterCompactKeepalive(t *testing.T) {
	upstream := &httpUpstreamRecorder{resp: bpsCompletionResponse(http.StatusOK,
		bpsFailureWire(t, "response.failed", "rate_limit_exceeded", "", 0))}
	svc := openAIClientToolsTestService(upstream)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", nil)
	MarkOpenAICompactClientStream(c)
	stop := StartOpenAICompactSSEKeepalive(c, time.Hour)
	defer stop()
	value, ok := c.Get(openAICompactSSEKeepaliveKey)
	require.True(t, ok)
	keepalive, ok := value.(*openAICompactSSEKeepalive)
	require.True(t, ok)
	require.True(t, keepalive.beat())
	body, err := json.Marshal(map[string]any{"model": "gpt-6-astra", "input": "continue"})
	require.NoError(t, err)
	_, err = svc.Forward(context.Background(), c, basispointsAccountForTest(), body)
	require.Error(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, strings.Count(rec.Body.String(), "event: response.failed\n"))
	require.NotContains(t, rec.Body.String(), "PRIVATE_")
	require.False(t, json.Valid(rec.Body.Bytes()), "already committed keepalive requires an SSE terminal")
	require.Contains(t, rec.Body.String(), "rate_limit_exceeded")
	observed, ok := GetOpsStreamError(c)
	require.True(t, ok)
	require.Equal(t, http.StatusTooManyRequests, observed.IntendedStatus)
	require.Len(t, upstream.requests, 1)
}

func TestBasispointsFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		name, event, code, errorType string
		explicit, status             int
		wantCode, wantType           string
	}{
		{"invalid input", "response.failed", "invalid_value", "", 0, 400, "invalid_value", "invalid_request_error"},
		{"authentication", "response.failed", "invalid_api_key", "", 0, 401, "invalid_api_key", "authentication_error"},
		{"permission", "response.failed", "permission_denied", "", 0, 403, "permission_denied", "permission_error"},
		{"rate limit", "error", "rate_limit_exceeded", "", 0, 429, "rate_limit_exceeded", "rate_limit_error"},
		{"type fallback", "response.failed", "", "overloaded_error", 0, 503, "basispoints_upstream_error", "server_error"},
		{"explicit status", "response.failed", "invalid_value", "invalid_request_error", 429, 429, "invalid_value", "rate_limit_error"},
		{"cancelled", "response.cancelled", "", "", 0, 502, "basispoints_upstream_cancelled", "server_error"},
		{"cancelled rate limit", "response.cancelled", "rate_limit_exceeded", "", 0, 429, "rate_limit_exceeded", "rate_limit_error"},
		{"unknown", "error", "PRIVATE_ERROR_CODE", "PRIVATE_ERROR_TYPE", 0, 502, "basispoints_upstream_error", "server_error"},
	} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", tc.name, stream), func(t *testing.T) {
				wire := bpsFailureWire(t, tc.event, tc.code, tc.errorType, tc.explicit)
				if stream {
					prefix, err := json.Marshal(map[string]any{"type": "response.output_text.delta", "delta": "already delivered"})
					require.NoError(t, err)
					wire = "data: " + string(prefix) + "\n\n" + wire
				}
				upstream := &httpUpstreamRecorder{resp: bpsCompletionResponse(http.StatusOK, wire)}
				svc := openAIClientToolsTestService(upstream)
				account := basispointsAccountForTest()
				account.Extra["openai_basispoints_auto_disable_on_403"] = true
				body, err := json.Marshal(map[string]any{"model": "gpt-6-astra", "stream": stream, "input": "hello"})
				require.NoError(t, err)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				result, err := svc.Forward(context.Background(), c, account, body)
				require.Error(t, err)
				if tc.event == "error" {
					require.Nil(t, result, "no billable usage means no usage row")
				} else {
					require.NotNil(t, result)
					require.Equal(t, tc.event, result.UpstreamTerminalEvent)
				}
				require.Len(t, upstream.requests, 1, "never replay a request after an upstream terminal")
				require.True(t, IsResponseCommitted(c))
				var failover *UpstreamFailoverError
				require.NotErrorAs(t, err, &failover)
				require.NotContains(t, rec.Body.String(), "PRIVATE_")
				require.NotContains(t, err.Error(), "PRIVATE_")
				require.NotContains(t, rec.Body.String(), "basispoints_stream_incomplete")
				require.NotContains(t, rec.Body.String(), "response.completed")
				if tc.event != "error" {
					require.EqualValues(t, 7, result.Usage.InputTokens)
					require.EqualValues(t, 2, result.Usage.OutputTokens)
				}
				if stream {
					require.Equal(t, http.StatusOK, rec.Code)
					require.Contains(t, rec.Body.String(), "already delivered")
					found := 0
					for _, line := range strings.Split(rec.Body.String(), "\n") {
						if !strings.HasPrefix(line, "data: ") {
							continue
						}
						raw := strings.TrimPrefix(line, "data: ")
						if gjson.Get(raw, "type").String() != tc.event {
							continue
						}
						found++
						path := "error"
						if tc.event != "error" {
							path = "response.error"
						}
						require.Equal(t, tc.wantCode, gjson.Get(raw, path+".code").String())
						require.Equal(t, tc.wantType, gjson.Get(raw, path+".type").String())
						require.EqualValues(t, tc.status, gjson.Get(raw, path+".status").Int())
					}
					require.Equal(t, 1, found)
				} else {
					require.Equal(t, tc.status, rec.Code)
					require.Equal(t, tc.wantCode, gjson.GetBytes(rec.Body.Bytes(), "error.code").String())
					require.Equal(t, tc.wantType, gjson.GetBytes(rec.Body.Bytes(), "error.type").String())
				}
				observed, ok := GetOpsStreamError(c)
				require.True(t, ok)
				require.Equal(t, tc.status, observed.IntendedStatus)
				require.Equal(t, tc.wantCode, observed.Code)
				require.Equal(t, http.StatusOK, c.GetInt(OpsUpstreamStatusCodeKey), "mapped status is not an upstream HTTP status")
				require.Equal(t, tc.status == 429, svc.isExcelBPSCoolingDown(account, "gpt-6-astra"))
				require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
				require.True(t, account.IsSchedulable())
				require.Equal(t, openAIOAuthResponsesEndpointBasis, openAIOAuthResponsesEndpointMode(account), "an in-band 403 must not disable the account")
			})
		}
	}
}

func TestBasispointsChatFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		event, code string
		status      int
	}{
		{"response.failed", "invalid_value", 400},
		{"response.failed", "invalid_api_key", 401},
		{"response.failed", "permission_denied", 403},
		{"error", "rate_limit_exceeded", 429},
		{"response.failed", "overloaded_error", 503},
		{"response.cancelled", "", 502},
		{"response.cancelled", "rate_limit_exceeded", 429},
	} {
		for _, stream := range []bool{false, true} {
			for _, visible := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/stream=%t/visible=%t", tc.event, tc.code, stream, visible), func(t *testing.T) {
					wire := bpsFailureWire(t, tc.event, tc.code, "", 0)
					if visible {
						wire = "data: {\"type\":\"response.output_text.delta\",\"delta\":\"already delivered\"}\n\n" + wire
					}
					upstream := &httpUpstreamRecorder{resp: bpsCompletionResponse(http.StatusOK, wire)}
					svc := openAIClientToolsTestService(upstream)
					account := basispointsAccountForTest()
					account.Extra[openAIBasispointsAutoDisableOn403Key] = true
					body, err := json.Marshal(gin.H{"model": "gpt-6-astra", "stream": stream, "messages": []any{gin.H{"role": "user", "content": "hello"}}})
					require.NoError(t, err)
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
					result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
					require.Error(t, err)
					var failover *UpstreamFailoverError
					require.NotErrorAs(t, err, &failover)
					if tc.event == "error" {
						require.Nil(t, result, "no billable usage means no usage row")
					} else {
						require.NotNil(t, result)
						require.Equal(t, tc.event, result.UpstreamTerminalEvent)
					}
					require.Len(t, upstream.requests, 1)
					require.NotContains(t, rec.Body.String(), "PRIVATE_")
					require.NotContains(t, rec.Body.String(), "response.completed")
					if tc.event != "error" {
						require.Equal(t, "resp_failure", result.ResponseID)
						require.Equal(t, 7, result.Usage.InputTokens)
						require.Equal(t, 2, result.Usage.OutputTokens)
					}
					if stream && visible {
						require.Equal(t, http.StatusOK, rec.Code)
						require.Contains(t, rec.Body.String(), "already delivered")
						require.Contains(t, rec.Body.String(), fmt.Sprintf("\"status\":%d", tc.status))
						require.Equal(t, 1, strings.Count(rec.Body.String(), "data: [DONE]"))
					} else {
						require.Equal(t, tc.status, rec.Code)
						require.True(t, json.Valid(rec.Body.Bytes()))
					}
					observed, ok := GetOpsStreamError(c)
					require.True(t, ok)
					require.Equal(t, tc.status, observed.IntendedStatus)
					require.Equal(t, http.StatusOK, c.GetInt(OpsUpstreamStatusCodeKey))
					require.Equal(t, tc.status == 429, svc.isExcelBPSCoolingDown(account, "gpt-6-astra"))
					require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
					require.Equal(t, openAIOAuthResponsesEndpointBasis, openAIOAuthResponsesEndpointMode(account))
				})
			}
		}
	}
}

func TestBasispointsFailurePreservesRawCacheUsage(t *testing.T) {
	for _, chat := range []bool{false, true} {
		for _, stream := range []bool{false, true} {
			for _, kind := range []string{"response.cancelled", "error"} {
				t.Run(fmt.Sprintf("chat=%t/stream=%t/%s", chat, stream, kind), func(t *testing.T) {
					event := gin.H{"type": kind, "error": gin.H{"code": "rate_limit_exceeded"}, "usage": basispointsCacheUsageFixture()}
					if kind != "error" {
						// The event type is authoritative even if the nested status is absent.
						event = gin.H{"type": kind, "response": gin.H{"id": "resp_cache_failed", "error": event["error"], "usage": event["usage"]}}
					}
					upstream := &httpUpstreamRecorder{resp: basispointsChatSSEForTest(t, event)}
					svc := openAIClientToolsTestService(upstream)
					account := basispointsAccountForTest()
					account.Extra[openAIBasispointsCacheCreationAsInputExtraKey] = true
					request := gin.H{"model": "gpt-6-sol", "input": "hi", "stream": stream}
					path := "/v1/responses"
					if chat {
						path = "/v1/chat/completions"
						delete(request, "input")
						request["messages"] = []any{gin.H{"role": "user", "content": "hi"}}
					}
					body, err := json.Marshal(request)
					require.NoError(t, err)
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, path, nil)
					var result *OpenAIForwardResult
					if chat {
						result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
					} else {
						result, err = svc.Forward(context.Background(), c, account, body)
					}
					require.Error(t, err)
					require.NotNil(t, result)
					require.Equal(t, 1000, result.Usage.InputTokens)
					require.Contains(t, rec.Body.String(), "rate_limit_exceeded")
					require.NotContains(t, rec.Body.String(), "chat.completion")
					require.Equal(t, 50, result.Usage.OutputTokens)
					require.Equal(t, 200, result.Usage.CacheCreationInputTokens, "billing receives original usage, not the downstream normalized copy")
					require.Equal(t, 100, result.Usage.CacheReadInputTokens)
					require.Len(t, upstream.requests, 1)
				})
			}
		}
	}
}
