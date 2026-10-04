//go:build unit

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

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestMonitorCandyFinalAnswer(t *testing.T) {
	for _, tc := range []struct{ name, text, raw, status, reason string }{
		{"correct", "分析后\n<answer>21</answer>", `{}`, "passed", ""},
		{"null error is normal", "21", `{"status":"completed","error":null}`, "passed", ""},
		{"bare", "**21**", `{}`, "passed", ""},
		{"bare with unit", "21颗。", `{}`, "passed", ""},
		{"wrong final despite incidental 21", "21 不够\n<answer>23</answer>", `{}`, "failed", "answer_mismatch"},
		{"substring", "121", `{}`, "failed", "answer_mismatch"},
		{"incidental answer", "考虑 21 或 23，还需要分析", `{}`, "inconclusive", "missing_final_answer"},
		{"ambiguous", "<answer>23</answer><answer>21</answer>", `{}`, "inconclusive", "missing_final_answer"},
		{"empty", "", `{}`, "inconclusive", "empty_response"},
		{"chat truncated", "21", `{"choices":[{"finish_reason":"length"}]}`, "inconclusive", "response_incomplete"},
		{"anthropic truncated", "21", `{"stop_reason":"max_tokens"}`, "inconclusive", "response_incomplete"},
		{"gemini truncated", "21", `{"candidates":[{"finishReason":"MAX_TOKENS"}]}`, "inconclusive", "response_incomplete"},
		{"responses incomplete", "21", `{"status":"incomplete"}`, "inconclusive", "response_incomplete"},
		{"truncated json", "21", `{"choices":`, "inconclusive", "response_incomplete"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := judgeMonitorCandy(tc.text, tc.raw)
			require.Equal(t, tc.status, result.Status)
			require.Equal(t, tc.reason, result.Reason)
		})
	}
}

func TestMonitorCandyRecordedAnswer(t *testing.T) {
	for _, tc := range []struct{ name, text, raw, want string }{
		{"correct", "private explanation\n<answer>21</answer>", "{}", "21"},
		{"incorrect", "private explanation mentions 21\n<answer>29</answer>", "{}", "29"},
		{"bare number", "29颗。", "{}", "29"},
		{"zero", "0", "{}", "0"},
		{"leading zeros", "021", "{}", "021"},
		{"ambiguous", "<answer>21</answer><answer>29</answer>", "{}", ""},
		{"missing", "private explanation without a final answer", "{}", ""},
		{"incomplete", "21", "{\"status\":\"incomplete\"}", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(judgeMonitorCandy(tc.text, tc.raw))
			require.NoError(t, err)
			require.Equal(t, tc.want, gjson.GetBytes(encoded, "answer").String())
			require.Equal(t, tc.want != "", gjson.GetBytes(encoded, "answer").Exists())
			require.NotContains(t, string(encoded), "private explanation")
		})
	}
}

func TestMonitorIntelligenceProviderRequests(t *testing.T) {
	swapMonitorHTTPClient(t)
	for _, tc := range []struct{ provider, apiMode, promptPath, tokenPath, reply, healthReply string }{
		{MonitorProviderOpenAI, MonitorAPIModeChatCompletions, "messages.0.content", "max_tokens", `{"choices":[{"message":{"content":"<answer>21</answer>"},"finish_reason":"stop"}]}`, `{"choices":[{"message":{"content":"%s"}}]}`},
		{MonitorProviderOpenAI, MonitorAPIModeResponses, "input", "max_output_tokens", `{"status":"completed","error":null,"output":[{"type":"reasoning"},{"type":"message","content":[{"type":"output_text","text":"<answer>21</answer>"}]}]}`, `{"output":[{"type":"message","content":[{"type":"output_text","text":"%s"}]}]}`},
		{MonitorProviderAnthropic, "", "messages.0.content", "max_tokens", `{"stop_reason":"end_turn","content":[{"type":"thinking","thinking":"23"},{"type":"text","text":"<answer>21</answer>"}]}`, `{"content":[{"type":"text","text":"%s"}]}`},
		{MonitorProviderGemini, "", "contents.0.parts.0.text", "generationConfig.maxOutputTokens", `{"candidates":[{"finishReason":"STOP","content":{"parts":[{"thought":true,"text":"23"},{"text":"<answer>2"},{"text":"1</answer>"}]}}]}`, `{"candidates":[{"content":{"parts":[{"text":"%s"}]}}]}`},
	} {
		t.Run(tc.provider+tc.apiMode, func(t *testing.T) {
			var payload, healthPayload []byte
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				payload, _ = io.ReadAll(r.Body)
				prompt := gjson.GetBytes(payload, tc.promptPath).String()
				if prompt != monitorCandyPrompt {
					healthPayload = payload
					w.Header().Set(ChannelMonitorProbeLatencyHeader, "123")
					_, _ = fmt.Fprintf(w, tc.healthReply, answerFromChallengePrompt(prompt))
					return
				}
				w.Header().Set(ChannelMonitorProbeLatencyHeader, "9000")
				time.Sleep(5 * time.Millisecond)
				_, _ = io.WriteString(w, tc.reply)
			}))
			defer srv.Close()
			opts := &CheckOptions{IntelligenceEnabled: true, APIMode: tc.apiMode, BodyOverrideMode: "merge", BodyOverride: map[string]any{"max_tokens": 50, "temperature": 0.2}}
			if tc.provider == MonitorProviderOpenAI {
				opts.BodyOverride["reasoning_effort"] = "xhigh"
				opts.BodyOverride["reasoning"] = map[string]any{"effort": "high", "summary": "auto"}
			}
			result := runCheckForModel(context.Background(), tc.provider, srv.URL, "test-key", "test-model", opts)
			require.Equal(t, "passed", result.Intelligence.Status)
			require.Equal(t, "21", result.Intelligence.Answer)
			require.Equal(t, MonitorStatusOperational, result.Status)
			require.Equal(t, 123, *result.LatencyMs, "only the ordinary probe determines latency")
			require.Equal(t, 2, calls)
			require.EqualValues(t, monitorChallengeMaxTokens, gjson.GetBytes(healthPayload, tc.tokenPath).Int())
			require.Equal(t, monitorCandyPrompt, gjson.GetBytes(payload, tc.promptPath).String())
			require.EqualValues(t, monitorIntelligenceMaxTokens, gjson.GetBytes(payload, tc.tokenPath).Int())
			require.Equal(t, 50, opts.BodyOverride["max_tokens"], "do not mutate saved templates")
			if tc.provider == MonitorProviderOpenAI {
				require.Equal(t, "xhigh", gjson.GetBytes(healthPayload, "reasoning_effort").String())
				require.Equal(t, "high", gjson.GetBytes(healthPayload, "reasoning.effort").String())
				require.Equal(t, "xhigh", opts.BodyOverride["reasoning_effort"])
				require.Equal(t, "high", opts.BodyOverride["reasoning"].(map[string]any)["effort"])
				if tc.apiMode == MonitorAPIModeResponses {
					require.Equal(t, "low", gjson.GetBytes(payload, "reasoning.effort").String())
					require.Equal(t, "auto", gjson.GetBytes(payload, "reasoning.summary").String())
					require.False(t, gjson.GetBytes(payload, "reasoning_effort").Exists())
				} else {
					require.Equal(t, "low", gjson.GetBytes(payload, "reasoning_effort").String())
					require.False(t, gjson.GetBytes(payload, "reasoning").Exists())
				}
			} else {
				require.False(t, gjson.GetBytes(payload, "reasoning_effort").Exists())
				require.False(t, gjson.GetBytes(payload, "reasoning").Exists())
			}
			if tc.apiMode == MonitorAPIModeResponses {
				require.NotContains(t, gjson.GetBytes(payload, "instructions").String(), "arithmetic")
			}
		})
	}
}

func TestMonitorIntelligenceWrongAnswerDoesNotRetryOrFailHealth(t *testing.T) {
	swapMonitorHTTPClient(t)
	calls := 0
	var candyPayload []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		payload, _ := io.ReadAll(r.Body)
		prompt := gjson.GetBytes(payload, "messages.0.content").String()
		if prompt != monitorCandyPrompt {
			_, _ = fmt.Fprintf(w, `{"choices":[{"message":{"content":"%s"}}]}`, answerFromChallengePrompt(prompt))
			return
		}
		candyPayload = payload
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"按完全随机抓取计算，最终答案为 <answer>29</answer>"},"finish_reason":"stop"}]}`)
	}))
	defer srv.Close()
	result := runCheckForModel(context.Background(), MonitorProviderOpenAI, srv.URL, "key", "test", &CheckOptions{IntelligenceEnabled: true})
	require.Equal(t, "failed", result.Intelligence.Status)
	require.Equal(t, "29", result.Intelligence.Answer)
	require.True(t, isMonitorSuccessStatus(result.Status))
	require.Equal(t, 2, calls, "one health probe and one candy check, no retry for a wrong answer")
	require.Equal(t, "low", gjson.GetBytes(candyPayload, "reasoning_effort").String(), "low is also sent without a request template")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result = runCheckForModel(ctx, MonitorProviderOpenAI, srv.URL, "key", "test", &CheckOptions{IntelligenceEnabled: true})
	require.Equal(t, "inconclusive", result.Intelligence.Status)
	require.Equal(t, "request_failed", result.Intelligence.Reason)
	require.Equal(t, MonitorStatusError, result.Status)
}

func TestMonitorIntelligenceSlowAnswerDoesNotDegradeHealth(t *testing.T) {
	swapMonitorHTTPClient(t)
	monitorHTTPClient.Timeout = 2 * monitorDegradedThreshold
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		payload, _ := io.ReadAll(r.Body)
		prompt := gjson.GetBytes(payload, "messages.0.content").String()
		if prompt == monitorCandyPrompt {
			time.Sleep(monitorDegradedThreshold + 50*time.Millisecond)
			_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"<answer>21</answer>"}}]}`)
			return
		}
		w.Header().Set(ChannelMonitorProbeLatencyHeader, "123")
		_, _ = fmt.Fprintf(w, `{"choices":[{"message":{"content":"%s"}}]}`, answerFromChallengePrompt(prompt))
	}))
	defer srv.Close()
	start := time.Now()
	result := runCheckForModel(context.Background(), MonitorProviderOpenAI, srv.URL, "key", "test", &CheckOptions{IntelligenceEnabled: true})
	require.GreaterOrEqual(t, time.Since(start), monitorDegradedThreshold)
	require.Equal(t, MonitorStatusOperational, result.Status)
	require.Equal(t, 123, *result.LatencyMs)
	require.Empty(t, result.Message)
	require.Equal(t, "passed", result.Intelligence.Status)
	require.Equal(t, 2, calls)
}

func TestMonitorIntelligenceHealthIsolation(t *testing.T) {
	swapMonitorHTTPClient(t)
	for _, tc := range []struct {
		name, healthLatency, candyReply, status, intelligence, reason string
		healthError, candyError, cancelCandy                          bool
	}{
		{name: "slow health still degrades", healthLatency: "7000", candyReply: `{"choices":[{"message":{"content":"<answer>21</answer>"}}]}`, status: MonitorStatusDegraded, intelligence: "passed"},
		{name: "candy HTTP error", healthLatency: "123", candyError: true, status: MonitorStatusOperational, intelligence: "inconclusive", reason: "request_failed"},
		{name: "candy canceled", healthLatency: "123", cancelCandy: true, status: MonitorStatusOperational, intelligence: "inconclusive", reason: "request_failed"},
		{name: "candy truncated", healthLatency: "123", candyReply: `{"choices":[{"message":{"content":"21"},"finish_reason":"length"}]}`, status: MonitorStatusOperational, intelligence: "inconclusive", reason: "response_incomplete"},
		{name: "candy empty", healthLatency: "123", candyReply: `{"choices":[{"message":{"content":""}}]}`, status: MonitorStatusOperational, intelligence: "inconclusive", reason: "empty_response"},
		{name: "health failed skips candy", healthError: true, status: MonitorStatusError, intelligence: "inconclusive", reason: "request_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				payload, _ := io.ReadAll(r.Body)
				prompt := gjson.GetBytes(payload, "messages.0.content").String()
				if prompt == monitorCandyPrompt {
					if tc.cancelCandy {
						cancel()
						// Wait for the client to disconnect before the handler can send an implicit 200.
						<-r.Context().Done()
						return
					}
					if tc.candyError {
						w.WriteHeader(http.StatusServiceUnavailable)
					}
					_, _ = io.WriteString(w, tc.candyReply)
					return
				}
				if tc.healthError {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.Header().Set(ChannelMonitorProbeLatencyHeader, tc.healthLatency)
				_, _ = fmt.Fprintf(w, `{"choices":[{"message":{"content":"%s"}}]}`, answerFromChallengePrompt(prompt))
			}))
			defer srv.Close()
			result := runCheckForModel(ctx, MonitorProviderOpenAI, srv.URL, "key", "test", &CheckOptions{IntelligenceEnabled: true})
			require.Equal(t, tc.status, result.Status)
			require.Equal(t, tc.intelligence, result.Intelligence.Status)
			require.Equal(t, tc.reason, result.Intelligence.Reason)
			if tc.healthError {
				require.Equal(t, 1, calls)
				require.Contains(t, result.Message, "upstream HTTP 503")
			} else {
				require.Equal(t, 2, calls)
				require.Equal(t, tc.healthLatency, fmt.Sprint(*result.LatencyMs))
				if tc.status == MonitorStatusDegraded {
					require.Equal(t, "slow response: 7000ms", result.Message)
				} else {
					require.Empty(t, result.Message)
				}
			}
		})
	}
}

func TestMonitorIntelligenceResponsesFallbackPreservesHealth(t *testing.T) {
	swapMonitorHTTPClient(t)
	health := &openAICaptureHandler{rejectChatModelNotFound: true, probeLatencyHeader: "7000"}
	calls := 0
	var payloads [][]byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		payload, _ := io.ReadAll(r.Body)
		payloads = append(payloads, payload)
		if r.URL.Path == providerOpenAIResponsesPath && gjson.GetBytes(payload, "input").String() == monitorCandyPrompt {
			_, _ = io.WriteString(w, `{"output":[{"type":"message","content":[{"type":"output_text","text":"<answer>21</answer>"}]}]}`)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(payload)))
		health.ServeHTTP(w, r)
	}))
	defer srv.Close()
	result := runCheckForModel(context.Background(), MonitorProviderOpenAI, srv.URL, "key", "gpt-5.5", &CheckOptions{IntelligenceEnabled: true})
	require.Equal(t, 4, calls, "each probe falls back once without starting another combined check")
	require.Equal(t, MonitorStatusDegraded, result.Status)
	require.Equal(t, 7000, *result.LatencyMs)
	require.Equal(t, "chat_completions model_not_found; retried via responses; slow response: 7000ms", result.Message)
	require.Equal(t, "passed", result.Intelligence.Status)
	require.False(t, gjson.GetBytes(payloads[0], "reasoning_effort").Exists(), "ordinary probe keeps upstream defaults")
	require.False(t, gjson.GetBytes(payloads[1], "reasoning").Exists())
	require.Equal(t, "low", gjson.GetBytes(payloads[2], "reasoning_effort").String())
	require.Equal(t, "low", gjson.GetBytes(payloads[3], "reasoning.effort").String())
	require.False(t, gjson.GetBytes(payloads[3], "reasoning_effort").Exists())
}

func TestMonitorIntelligenceDefaultOffAndConfigValidation(t *testing.T) {
	swapMonitorHTTPClient(t)
	handler := &captureHandler{format: "openai_chat"}
	srv := httptest.NewServer(handler)
	defer srv.Close()
	result := runCheckForModel(context.Background(), MonitorProviderOpenAI, srv.URL, "key", "test", nil)
	require.Nil(t, result.Intelligence)
	require.True(t, isMonitorSuccessStatus(result.Status))
	require.EqualValues(t, monitorChallengeMaxTokens, handler.lastBody["max_tokens"])
	for _, mode := range []string{"probe", "quota_probe"} {
		require.NoError(t, validateMonitorIntelligence(true, mode, "merge"))
	}
	require.Error(t, validateMonitorIntelligence(true, "quota", "off"))
	require.Error(t, validateMonitorIntelligence(true, "probe", "replace"))
	require.NoError(t, validateMonitorIntelligence(false, "quota", "replace"))
	handler.lastBody = nil
	result = runCheckForModel(context.Background(), MonitorProviderOpenAI, srv.URL, "key", "test", &CheckOptions{IntelligenceEnabled: true, BodyOverrideMode: "replace"})
	require.Equal(t, "incompatible_configuration", result.Intelligence.Reason)
	require.Nil(t, handler.lastBody, "invalid configuration must not send either probe")
}

func TestMonitorIntelligenceAggregation(t *testing.T) {
	pass := &domain.MonitorIntelligenceResult{Status: "passed", Answer: "21"}
	fail := &domain.MonitorIntelligenceResult{Status: "failed", Reason: "answer_mismatch", Answer: "29"}
	m := &ChannelMonitor{PrimaryModel: "main", ExtraModels: []string{"extra"}, IntelligenceEnabled: true}
	latest := []*ChannelMonitorLatest{{Model: "main", Status: MonitorStatusOperational, Intelligence: pass}, {Model: "extra", Status: MonitorStatusOperational, Intelligence: fail}}
	summary := buildStatusSummary(indexLatestByModel(latest), nil, "main", m.ExtraModels)
	view := buildUserViewFromSummary(m, summary, latest[0], []*ChannelMonitorHistoryEntry{{Intelligence: pass}, {Intelligence: fail}, {}})
	require.True(t, view.IntelligenceEnabled)
	require.Equal(t, pass, view.Intelligence)
	require.Equal(t, fail, view.ExtraModels[0].Intelligence)
	require.Equal(t, pass, view.Timeline[0].Intelligence)
	require.Equal(t, fail, view.Timeline[1].Intelligence)
	require.Nil(t, view.Timeline[2].Intelligence)
	encoded, err := json.Marshal(view.Timeline)
	require.NoError(t, err)
	require.Equal(t, "passed", gjson.GetBytes(encoded, "0.intelligence.status").String())
	require.Equal(t, "21", gjson.GetBytes(encoded, "0.intelligence.answer").String())
	require.Equal(t, "29", gjson.GetBytes(encoded, "1.intelligence.answer").String())
}

type intelligenceLifecycleRepo struct {
	*duplicateChannelMonitorRepoStub
	rows       []*ChannelMonitorHistoryRow
	persistErr error
}

func (r *intelligenceLifecycleRepo) GetByID(ctx context.Context, id int64) (*ChannelMonitor, error) {
	m, err := r.duplicateChannelMonitorRepoStub.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	copy := *m
	return &copy, nil
}

func (r *intelligenceLifecycleRepo) Update(_ context.Context, m *ChannelMonitor) error {
	copy := *m
	r.source = &copy
	return nil
}

func (r *intelligenceLifecycleRepo) InsertHistoryBatch(ctx context.Context, rows []*ChannelMonitorHistoryRow) error {
	r.persistErr = ctx.Err()
	r.rows = rows
	return r.persistErr
}

func (r *intelligenceLifecycleRepo) MarkChecked(ctx context.Context, _ int64, _ time.Time) error {
	return ctx.Err()
}

func TestChannelMonitorIntelligenceLifecycleAndTimedOutHistory(t *testing.T) {
	repo := &intelligenceLifecycleRepo{duplicateChannelMonitorRepoStub: &duplicateChannelMonitorRepoStub{}}
	svc := NewChannelMonitorService(repo, &duplicateChannelMonitorEncryptor{})
	svc.SetRuntimeReader(channelMonitorRuntimeStub{rt: ChannelMonitorRuntime{Enabled: true, Mode: ChannelMonitorModeV1}})
	m, err := svc.Create(context.Background(), ChannelMonitorCreateParams{
		Name: "candy", Provider: MonitorProviderOpenAI, Endpoint: "https://8.8.8.8", APIKey: "test",
		PrimaryModel: "main", ExtraModels: []string{"extra"}, Enabled: true, IntervalSeconds: 300, IntelligenceEnabled: true,
	})
	require.NoError(t, err)
	require.True(t, repo.created[0].IntelligenceEnabled)
	repo.source = repo.created[0]
	enabled := false
	_, err = svc.Update(context.Background(), m.ID, ChannelMonitorUpdateParams{IntelligenceEnabled: &enabled})
	require.NoError(t, err)
	require.False(t, repo.source.IntelligenceEnabled)
	enabled = true
	_, err = svc.Update(context.Background(), m.ID, ChannelMonitorUpdateParams{IntelligenceEnabled: &enabled})
	require.NoError(t, err)
	copy, err := svc.Duplicate(context.Background(), m.ID, 1, "admin:1", "copy-candy")
	require.NoError(t, err)
	require.True(t, copy.IntelligenceEnabled)
	require.False(t, copy.Enabled)
	require.Nil(t, copy.LastCheckedAt)

	swapMonitorHTTPClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			cancel()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	oldPingClient := monitorPingHTTPClient
	monitorPingHTTPClient = srv.Client()
	t.Cleanup(func() { monitorPingHTTPClient = oldPingClient })
	repo.source.Endpoint = srv.URL
	results, err := svc.RunCheck(ctx, m.ID)
	require.NoError(t, err)
	require.Error(t, ctx.Err())
	require.Len(t, results, 2)
	require.Len(t, repo.rows, 2)
	require.NoError(t, repo.persistErr, "timed out probe must still persist a gray sample")
	for _, row := range repo.rows {
		require.Equal(t, "inconclusive", row.Intelligence.Status)
	}

	t.Run("successful health survives candy deadline in history", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)
				return
			}
			payload, _ := io.ReadAll(r.Body)
			prompt := gjson.GetBytes(payload, "messages.0.content").String()
			if prompt == monitorCandyPrompt {
				<-r.Context().Done()
				return
			}
			w.Header().Set(ChannelMonitorProbeLatencyHeader, "123")
			_, _ = fmt.Fprintf(w, `{"choices":[{"message":{"content":"%s"}}]}`, answerFromChallengePrompt(prompt))
		}))
		defer srv.Close()
		repo.source.Endpoint = srv.URL
		repo.source.ExtraModels = nil
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		results, err := svc.RunCheck(ctx, m.ID)
		require.NoError(t, err)
		require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
		require.NoError(t, repo.persistErr)
		require.Len(t, results, 1)
		require.Len(t, repo.rows, 1)
		row := repo.rows[0]
		require.Equal(t, MonitorStatusOperational, row.Status)
		require.Equal(t, 123, *row.LatencyMs)
		require.Empty(t, row.Message)
		require.Equal(t, "inconclusive", row.Intelligence.Status)
		require.Equal(t, "request_failed", row.Intelligence.Reason)
	})

	svc.SetRuntimeReader(channelMonitorRuntimeStub{rt: ChannelMonitorRuntime{Enabled: true, Mode: ChannelMonitorModeV2}})
	_, err = svc.RunCheck(context.Background(), m.ID)
	require.ErrorIs(t, err, ErrChannelMonitorActiveProbesRetired)
}

// Minimax enumeration independently verifies the published answer. After each
// draw the player picks a shape; an adversary chooses any remaining flavor.
func TestMonitorCandyReferenceAnswerByEnumeration(t *testing.T) {
	limits := [6]int{7, 9, 8, 7, 6, 4}
	memo := map[[6]int]int{}
	var solve func([6]int) int
	solve = func(drawn [6]int) int {
		if drawn[0]*drawn[4] > 0 || drawn[1]*drawn[3] > 0 {
			return 0
		}
		if n, ok := memo[drawn]; ok {
			return n
		}
		best := 42
		for shape := 0; shape < 2; shape++ {
			worst := -1
			for k := shape * 3; k < shape*3+3; k++ {
				if drawn[k] == limits[k] {
					continue
				}
				next := drawn
				next[k]++
				worst = max(worst, solve(next))
			}
			if worst >= 0 {
				best = min(best, 1+worst)
			}
		}
		memo[drawn] = best
		return best
	}
	require.Equal(t, 21, solve([6]int{}))
}
