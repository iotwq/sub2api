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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestBasispointsChatCompletionsWireAndResponse(t *testing.T) {
	for _, kind := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", kind, stream), func(t *testing.T) {
				account := basispointsAccountForTest()
				account.Type = kind
				account.Credentials["model_mapping"] = map[string]any{"public-chat": "gpt-6-sol", "gpt-6-sol": "must-not-map-twice"}
				upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
					require.Equal(t, basispointsResponsesURL, req.URL.String())
					require.Equal(t, "chatgpt", req.Header.Get("x-basispoints-auth-mode"))
					require.Empty(t, req.Header.Get("x-codex-turn-state"))
					wire := mustReadRequestBody(t, req)
					require.Equal(t, "gpt-6-sol", gjson.GetBytes(wire, "model").String())
					require.Equal(t, "xhigh", gjson.GetBytes(wire, "reasoning_effort").String())
					require.True(t, gjson.GetBytes(wire, "stream").Bool())
					require.False(t, gjson.GetBytes(wire, "messages").Exists())
					return basispointsTestCompletedResponse(t, []any{gin.H{"type": "message", "id": "msg_bps_chat", "role": "assistant", "content": []any{gin.H{"type": "output_text", "text": "BPS chat bridge is working."}}}}), nil
				}}
				body, err := json.Marshal(gin.H{"model": "public-chat", "messages": []any{gin.H{"role": "user", "content": "hi"}}, "reasoning_effort": "max", "stream": stream})
				require.NoError(t, err)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(string(body)))
				c.Request.Header.Set("x-codex-turn-state", "codex-only-state")
				svc := openAIClientToolsTestService(nil)
				svc.httpUpstream = upstream
				result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, rec.Code)
				require.NotNil(t, result)
				require.Equal(t, "/basispoints/api/responses", result.UpstreamEndpoint)
				require.Equal(t, "public-chat", result.Model)
				require.Equal(t, "gpt-6-sol", result.UpstreamModel)
				require.Equal(t, "resp_bps", result.ResponseID)
				require.Equal(t, "xhigh", optionalStringValue(result.ReasoningEffort))
				require.Equal(t, "max", optionalStringValue(result.RequestedReasoningEffort))
				require.Positive(t, result.Usage.InputTokens)
				require.NotContains(t, rec.Body.String(), "response.completed")
				if stream {
					require.Contains(t, rec.Header().Get("Content-Type"), "text/event-stream")
					require.Contains(t, rec.Body.String(), "chat.completion.chunk")
					require.Contains(t, rec.Body.String(), "\"role\":\"assistant\"")
					require.Contains(t, rec.Body.String(), "data: [DONE]")
					require.Contains(t, rec.Body.String(), "BPS chat bridge is working.")
				} else {
					require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
					require.Equal(t, "chat.completion", gjson.Get(rec.Body.String(), "object").String())
					require.Equal(t, "public-chat", gjson.Get(rec.Body.String(), "model").String())
					require.Equal(t, "BPS chat bridge is working.", gjson.Get(rec.Body.String(), "choices.0.message.content").String())
				}
			})
		}
	}
}

func forwardBasispointsChatForTest(t *testing.T, svc *OpenAIGatewayService, account *Account, request gin.H) (*OpenAIForwardResult, *httptest.ResponseRecorder, error) {
	t.Helper()
	body, err := json.Marshal(request)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(string(body)))
	c.Request.Header.Set("thread-id", t.Name())
	result, err := svc.ForwardAsChatCompletions(c.Request.Context(), c, account, body, "", "")
	return result, rec, err
}

func basispointsChatSSEForTest(t *testing.T, events ...gin.H) *http.Response {
	t.Helper()
	var wire strings.Builder
	for _, event := range events {
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		fmt.Fprintf(&wire, "data: %s\n\n", payload)
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(wire.String()))}
}

func TestBasispointsChatCompletionsCacheAccounting(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, enabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("stream=%t/cache-as-input=%t", stream, enabled), func(t *testing.T) {
				account := basispointsAccountForTest()
				account.Extra[openAIBasispointsCacheCreationAsInputExtraKey] = enabled
				upstream := &httpUpstreamRecorder{resp: basispointsChatSSEForTest(t, gin.H{"type": "response.completed", "response": gin.H{
					"id": "resp_chat_cache", "status": "completed", "output": []any{}, "usage": basispointsCacheUsageFixture(),
				}})}
				result, rec, err := forwardBasispointsChatForTest(t, openAIClientToolsTestService(upstream), account, gin.H{
					"model": "gpt-6-sol", "stream": stream, "messages": []any{gin.H{"role": "user", "content": "hi"}},
				})
				require.NoError(t, err)
				require.Equal(t, 1000, result.Usage.InputTokens)
				require.Equal(t, 200, result.Usage.CacheCreationInputTokens)
				require.Equal(t, 100, result.Usage.CacheReadInputTokens)
				usage := gjson.Get(rec.Body.String(), "usage")
				if stream {
					for _, line := range strings.Split(rec.Body.String(), "\n") {
						if candidate := gjson.Get(strings.TrimPrefix(line, "data: "), "usage"); candidate.Exists() {
							usage = candidate
						}
					}
				}
				require.Equal(t, int64(1000), usage.Get("prompt_tokens").Int())
				require.Equal(t, int64(50), usage.Get("completion_tokens").Int())
				require.Equal(t, int64(100), usage.Get("prompt_tokens_details.cached_tokens").Int())
				wantCreation := 200
				if enabled {
					wantCreation = 0
				}
				require.Equal(t, int64(wantCreation), usage.Get("prompt_tokens_details.cache_write_tokens").Int())
				require.Equal(t, int64(wantCreation), usage.Get("prompt_tokens_details.cache_creation_tokens").Int())
				usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
				usageService := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
				require.NoError(t, usageService.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
					Result: result, APIKey: &APIKey{ID: 10}, User: &User{ID: 20}, Account: account,
				}))
				require.Equal(t, wantCreation, usageRepo.lastLog.CacheCreationTokens)
				require.Equal(t, 900-wantCreation, usageRepo.lastLog.InputTokens)
				require.Equal(t, 1050, usageRepo.lastLog.TotalTokens())
			})
		}
	}
}

func TestBasispointsChatCompletionsToolRoundTrip(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			account := basispointsAccountForTest()
			code, err := json.Marshal(gin.H{"name": "lookup", "arguments": gin.H{"id": json.Number("9007199254740993")}})
			require.NoError(t, err)
			args, err := json.Marshal(gin.H{"code": string(code)})
			require.NoError(t, err)
			calls := 0
			upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				calls++
				wire := mustReadRequestBody(t, req)
				require.Equal(t, basispointsResponsesURL, req.URL.String())
				require.False(t, gjson.GetBytes(wire, "tools").Exists())
				if calls == 1 {
					return basispointsTestCompletedResponse(t, []any{gin.H{"type": "function_call", "id": "fc_chat", "call_id": "call_chat", "name": "run_officejs", "arguments": string(args)}}), nil
				}
				input := gjson.GetBytes(wire, "input").Array()
				require.Equal(t, "run_officejs", input[len(input)-2].Get("name").String())
				require.Equal(t, "call_chat", input[len(input)-1].Get("call_id").String())
				require.Equal(t, "found", input[len(input)-1].Get("output").String())
				return basispointsTestCompletedResponse(t, []any{}), nil
			}}
			svc := openAIClientToolsTestService(nil)
			svc.httpUpstream = upstream
			messages := []any{gin.H{"role": "user", "content": "look up the record"}}
			for turn := 0; turn < 2; turn++ {
				_, rec, err := forwardBasispointsChatForTest(t, svc, account, gin.H{
					"model": "gpt-6-sol", "stream": stream, "messages": messages,
					"tools": []any{gin.H{"type": "function", "function": gin.H{"name": "lookup", "parameters": gin.H{"type": "object"}}}},
				})
				require.NoError(t, err)
				require.NotContains(t, rec.Body.String(), "run_officejs")
				if turn == 0 {
					id, name, arguments := "", "", ""
					if stream {
						for _, line := range strings.Split(rec.Body.String(), "\n") {
							call := gjson.Get(strings.TrimPrefix(line, "data: "), "choices.0.delta.tool_calls.0")
							id += call.Get("id").String()
							name += call.Get("function.name").String()
							arguments += call.Get("function.arguments").String()
						}
					} else {
						call := gjson.Get(rec.Body.String(), "choices.0.message.tool_calls.0")
						id, name, arguments = call.Get("id").String(), call.Get("function.name").String(), call.Get("function.arguments").String()
					}
					require.Equal(t, "call_chat", id)
					require.Equal(t, "lookup", name)
					require.Equal(t, "9007199254740993", gjson.Get(arguments, "id").Raw)
					messages = append(messages,
						gin.H{"role": "assistant", "tool_calls": []any{gin.H{"id": id, "type": "function", "function": gin.H{"name": name, "arguments": arguments}}}},
						gin.H{"role": "tool", "tool_call_id": id, "content": "found"})
				}
			}
			require.Equal(t, 2, calls)
		})
	}
}

func TestBasispointsChatCompletionsFailuresKeepUsageWithoutFailover(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, delta := range []bool{false, true} {
			t.Run(fmt.Sprintf("stream=%t/delta=%t", stream, delta), func(t *testing.T) {
				events := []gin.H{}
				if delta {
					events = append(events, gin.H{"type": "response.output_text.delta", "delta": "partial answer"})
				}
				events = append(events, gin.H{"type": "response.failed", "response": gin.H{
					"id": "resp_failed_chat", "status": "failed", "usage": basispointsCacheUsageFixture(),
					"error": gin.H{"code": "server_error", "message": "Generation failed"},
				}})
				upstream := &httpUpstreamRecorder{resp: basispointsChatSSEForTest(t, events...)}
				result, rec, err := forwardBasispointsChatForTest(t, openAIClientToolsTestService(upstream), basispointsAccountForTest(), gin.H{
					"model": "gpt-6-sol", "stream": stream, "messages": []any{gin.H{"role": "user", "content": "hi"}},
				})
				require.Error(t, err)
				var failover *UpstreamFailoverError
				require.NotErrorAs(t, err, &failover)
				require.NotNil(t, result)
				require.Equal(t, 1000, result.Usage.InputTokens)
				require.Equal(t, "resp_failed_chat", result.ResponseID)
				require.Contains(t, rec.Body.String(), "server_error")
				require.NotContains(t, rec.Body.String(), "Generation failed", "upstream free-form errors are not exposed")
				require.NotContains(t, rec.Body.String(), "response.failed")
				require.Len(t, upstream.requests, 1)
			})
		}
	}
}

func TestBasispointsChatCompletionsHTTPErrorPolicy(t *testing.T) {
	for _, status := range []int{400, 401, 403, 422, 429, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("private upstream content"))}}
			result, rec, err := forwardBasispointsChatForTest(t, openAIClientToolsTestService(upstream), basispointsAccountForTest(), gin.H{
				"model": "gpt-6-sol", "messages": []any{gin.H{"role": "user", "content": "hi"}},
			})
			require.Error(t, err)
			require.Nil(t, result)
			require.Len(t, upstream.requests, 1)
			require.NotContains(t, rec.Body.String(), "private upstream content")
			var failover *UpstreamFailoverError
			if status == 503 {
				require.ErrorAs(t, err, &failover)
				require.Equal(t, status, failover.StatusCode)
			} else if status == http.StatusTooManyRequests {
				require.ErrorAs(t, err, &failover)
				require.Equal(t, status, failover.StatusCode)
			} else {
				require.NotErrorAs(t, err, &failover)
				require.Equal(t, status, rec.Code)
			}
		})
	}
}

func TestBasispointsChatCompletionsDefaultRoutesUnchanged(t *testing.T) {
	for _, kind := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
		t.Run(kind, func(t *testing.T) {
			account := basispointsAccountForTest()
			account.Type = kind
			wantURL := chatgptCodexURL
			if kind == AccountTypeAPIKey {
				account.Credentials["api_key"] = "test-key"
				account.Credentials["base_url"] = "https://example.com"
				account.Extra["openai_responses_supported"] = true
				wantURL = "https://example.com/v1/responses"
			} else {
				account.Extra[openAIOAuthResponsesEndpointExtraKey] = "chatgpt_codex"
			}
			upstream := &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				require.Equal(t, wantURL, req.URL.String())
				require.Empty(t, req.Header.Get("x-basispoints-auth-mode"))
				require.Equal(t, "max", gjson.GetBytes(mustReadRequestBody(t, req), "reasoning.effort").String())
				return basispointsTestCompletedResponse(t, []any{}), nil
			}}
			svc := openAIClientToolsTestService(nil)
			svc.httpUpstream = upstream
			_, _, err := forwardBasispointsChatForTest(t, svc, account, gin.H{
				"model": "gpt-6-sol", "reasoning_effort": "max", "messages": []any{gin.H{"role": "user", "content": "hi"}},
			})
			require.NoError(t, err)
		})
	}
}

func TestBasispointsChatCompletionsCancellationAndDisconnect(t *testing.T) {
	for _, cancelBeforeSend := range []bool{false, true} {
		t.Run(fmt.Sprint(cancelBeforeSend), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Writer = &openAIChatFailingWriter{ResponseWriter: c.Writer, failAfter: 0}
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)
			body, err := json.Marshal(gin.H{"model": "gpt-6-sol", "stream": true, "messages": []any{gin.H{"role": "user", "content": "hi"}}})
			require.NoError(t, err)
			upstream := &httpUpstreamRecorder{resp: basispointsChatSSEForTest(t,
				gin.H{"type": "response.output_text.delta", "delta": "hello"},
				gin.H{"type": "response.completed", "response": gin.H{"id": "resp_disconnect", "status": "completed", "output": []any{}, "usage": basispointsCacheUsageFixture()}},
			)}
			if cancelBeforeSend {
				cancel()
			}
			result, err := openAIClientToolsTestService(upstream).ForwardAsChatCompletions(ctx, c, basispointsAccountForTest(), body, "", "")
			if cancelBeforeSend {
				require.ErrorIs(t, err, context.Canceled)
				require.Nil(t, result)
				require.Empty(t, upstream.requests)
			} else {
				require.NoError(t, err)
				require.True(t, result.ClientDisconnect)
				require.Equal(t, 1000, result.Usage.InputTokens)
				require.Equal(t, 50, result.Usage.OutputTokens)
			}
		})
	}
}

func TestBasispointsChatCompletionsIncrementalTextAndResponsesShape(t *testing.T) {
	for _, responsesShape := range []bool{false, true} {
		t.Run(fmt.Sprint(responsesShape), func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: basispointsChatSSEForTest(t,
				gin.H{"type": "response.created", "response": gin.H{"id": "resp_text", "status": "in_progress"}},
				gin.H{"type": "response.output_text.delta", "delta": "hello"},
				gin.H{"type": "response.output_text.delta", "delta": " world"},
				gin.H{"type": "response.completed", "response": gin.H{"id": "resp_text", "status": "completed", "usage": basispointsCacheUsageFixture(),
					"output": []any{gin.H{"type": "message", "content": []any{gin.H{"type": "output_text", "text": "hello world"}}}},
				}},
			)}
			body := gin.H{"model": "gpt-6-sol", "stream": true, "reasoning": gin.H{"effort": "max"}}
			if responsesShape {
				body["input"] = "hi"
			} else {
				body["messages"] = []any{gin.H{"role": "user", "content": "hi"}}
			}
			result, rec, err := forwardBasispointsChatForTest(t, openAIClientToolsTestService(upstream), basispointsAccountForTest(), body)
			require.NoError(t, err)
			require.Equal(t, "xhigh", optionalStringValue(result.ReasoningEffort))
			require.Equal(t, "max", optionalStringValue(result.RequestedReasoningEffort))
			var text strings.Builder
			for _, line := range strings.Split(rec.Body.String(), "\n") {
				text.WriteString(gjson.Get(strings.TrimPrefix(line, "data: "), "choices.0.delta.content").String())
			}
			require.Equal(t, "hello world", text.String(), "terminal text must not duplicate streamed deltas")
		})
	}
}

func TestBasispointsChatCompletionsImages(t *testing.T) {
	for _, mode := range []string{"disabled", "ignore", "relay", "native"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("DATA_DIR", t.TempDir())
			account := basispointsAccountForTest()
			account.Extra[BasispointsIgnoreImagesKey] = mode == "ignore"
			svc := openAIClientToolsTestService(nil)
			t.Cleanup(func() { require.NoError(t, svc.CloseExcelBPSImages()) })
			if mode == "native" {
				enableNativeAttachments(svc)
			} else if mode == "relay" {
				svc.settingService = NewSettingService(&excelBPSImageSettingsRepo{values: map[string]string{
					SettingKeyExcelBPSImageRelayEnabled: "true", SettingKeyExcelBPSImageBaseURL: "https://relay.example",
				}}, svc.cfg)
			}
			calls := 0
			svc.httpUpstream = &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				calls++
				if mode == "native" && calls == 1 {
					require.Equal(t, "/basispoints/api/attachments", req.URL.Path)
					return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{\"openai_file_id\":\"file-chat-image\"}"))}, nil
				}
				require.Equal(t, basispointsResponsesURL, req.URL.String())
				wire := string(mustReadRequestBody(t, req))
				require.NotContains(t, wire, "data:image")
				switch mode {
				case "native":
					require.Contains(t, wire, "file-chat-image")
				case "relay":
					require.Contains(t, wire, "https://relay.example/api/bps-images/")
				case "ignore":
					require.NotContains(t, wire, "input_image")
				}
				return basispointsTestCompletedResponse(t, []any{}), nil
			}}
			_, rec, err := forwardBasispointsChatForTest(t, svc, account, gin.H{"model": "gpt-6-sol",
				"messages": []any{gin.H{"role": "user", "content": []any{
					gin.H{"type": "text", "text": "describe this"},
					gin.H{"type": "image_url", "image_url": gin.H{"url": "data:image/png;base64," + basispointsImageTestPNG}},
				}}},
			})
			if mode == "disabled" {
				require.Error(t, err)
				require.Equal(t, 400, rec.Code)
				require.Zero(t, calls)
			} else {
				require.NoError(t, err)
				wantCalls := 1
				if mode == "native" {
					wantCalls = 2
				}
				require.Equal(t, wantCalls, calls)
			}
		})
	}
}

func TestBasispointsChatCompletionsCorrectionFailure(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			first := &basispointsRepairBody{Reader: strings.NewReader(basispointsRepairWire(t, "chat_correction", "Run", "text(42);"))}
			rejected := &basispointsRepairBody{Reader: strings.NewReader("PRIVATE_UPSTREAM")}
			upstream := &httpUpstreamRecorder{responses: []*http.Response{
				{StatusCode: 200, Header: http.Header{}, Body: first},
				{StatusCode: 429, Header: http.Header{}, Body: rejected},
			}}
			result, rec, err := forwardBasispointsChatForTest(t, openAIClientToolsTestService(upstream), basispointsAccountForTest(), gin.H{
				"model": "gpt-6-sol", "stream": stream, "messages": []any{gin.H{"role": "user", "content": "run lookup"}},
				"tools": []any{gin.H{"type": "function", "function": gin.H{"name": "lookup", "parameters": gin.H{"type": "object"}}}},
			})
			require.Error(t, err)
			var failover *UpstreamFailoverError
			require.NotErrorAs(t, err, &failover)
			require.NotNil(t, result)
			require.Equal(t, 10, result.Usage.InputTokens)
			require.Len(t, upstream.requests, 2)
			require.True(t, first.closed.Load())
			require.True(t, rejected.closed.Load())
			require.NotContains(t, rec.Body.String(), "PRIVATE_UPSTREAM")
			require.NotContains(t, rec.Body.String(), "run_officejs")
		})
	}
}
