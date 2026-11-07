//go:build unit

package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type channelMonitorPlatformUpstream struct {
	service.HTTPUpstream
	calls     []int64
	successID int64
	firstWait time.Duration
	chat      bool
}

func (u *channelMonitorPlatformUpstream) Do(req *http.Request, _ string, accountID int64, _ int) (*http.Response, error) {
	u.calls = append(u.calls, accountID)
	if len(u.calls) == 1 && u.firstWait > 0 {
		if !sleepWithContext(req.Context(), u.firstWait) {
			return nil, req.Context().Err()
		}
	}
	status := http.StatusServiceUnavailable
	body := `{"error":{"code":503,"message":"temporary upstream failure","status":"UNAVAILABLE"}}`
	if accountID == u.successID {
		status = http.StatusOK
		body = `{"candidates":[{"content":{"role":"model","parts":[{"text":"4"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1,"totalTokenCount":2}}`
		if u.chat {
			body = `{"id":"chatcmpl-monitor","object":"chat.completion","model":"test-model","choices":[{"index":0,"message":{"role":"assistant","content":"4"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`
		}
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestGeminiV1BetaModels_ChannelMonitorProbe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name         string
		accountCount int
		successID    int64
		monitor      bool
		maxSwitches  int
		wantCalls    []int64
		wantSuccess  bool
	}{
		{"first success stops probing", 6, 1, true, 10, []int64{1}, true},
		{"switches to successful account", 6, 2, true, 10, []int64{1, 2}, true},
		{"fifth account can succeed", 6, 5, true, 10, []int64{1, 2, 3, 4, 5}, true},
		{"five failed accounts exhaust budget", 6, 6, true, 10, []int64{1, 2, 3, 4, 5}, false},
		{"available accounts exhausted without recycling", 2, 0, true, 10, []int64{1, 2}, false},
		{"lower configured limit is respected", 6, 3, true, 1, []int64{1, 2}, false},
		{"ordinary request keeps configured switches", 6, 6, false, 10, []int64{1, 2, 3, 4, 5, 6}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			groupID := int64(9300)
			group := &service.Group{ID: groupID, Hydrated: true, Platform: service.PlatformGemini, Status: service.StatusActive}
			accounts := make([]*service.Account, tt.accountCount)
			for i := range accounts {
				accountID := int64(i + 1)
				accounts[i] = &service.Account{
					ID: accountID, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey,
					Status: service.StatusActive, Schedulable: true, Concurrency: 1, Priority: i,
					Credentials:   map[string]any{"api_key": "test-key"},
					AccountGroups: []service.AccountGroup{{AccountID: accountID, GroupID: groupID}},
				}
			}
			h, cleanup := newTestGatewayHandler(t, group, accounts)
			t.Cleanup(cleanup)
			cfg := &config.Config{RunMode: config.RunModeSimple}
			cfg.Gateway.MaxSameAccountRetriesGemini = 0
			h.cfg = cfg
			h.maxAccountSwitchesGemini = tt.maxSwitches
			upstream := &channelMonitorPlatformUpstream{successID: tt.successID}
			if tt.successID == 2 {
				upstream.firstWait = 100 * time.Millisecond
			}
			h.geminiCompatService = service.NewGeminiMessagesCompatService(nil, nil, nil, nil, nil, nil, upstream, nil, cfg)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.8-flash:generateContent", strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"2+2=?"}]}]}`))
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.Group, group))
			c.Request.Header.Set("Content-Type", "application/json")
			if tt.monitor {
				c.Request.Header.Set(service.ChannelMonitorProbeAttemptsHeader, strconv.Itoa(service.ChannelMonitorProbeAttempts))
			}
			c.Params = gin.Params{{Key: "modelAction", Value: "/gemini-3.8-flash:generateContent"}}
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
				ID: 9301, UserID: 9302, GroupID: &groupID, Group: group,
				User: &service.User{ID: 9302, Balance: 100},
			})
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 9302, Concurrency: 10})

			started := time.Now()
			h.GeminiV1BetaModels(c)
			elapsed := time.Since(started)

			require.Equal(t, tt.wantCalls, upstream.calls)
			if tt.wantSuccess {
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), `"text":"4"`)
			} else {
				require.GreaterOrEqual(t, rec.Code, 400, rec.Body.String())
			}
			if tt.monitor && tt.wantSuccess {
				latency, err := strconv.Atoi(rec.Header().Get(service.ChannelMonitorProbeLatencyHeader))
				require.NoError(t, err)
				require.GreaterOrEqual(t, latency, 0)
				if upstream.firstWait > 0 {
					require.Less(t, time.Duration(latency)*time.Millisecond, elapsed-upstream.firstWait/2,
						"the previous failed account must not contribute to successful probe latency")
				}
			} else if !tt.monitor {
				require.Empty(t, rec.Header().Get(service.ChannelMonitorProbeLatencyHeader))
			}
		})
	}
}

func TestChatCompletions_ChannelMonitorCompatiblePlatforms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, platform := range []string{service.PlatformDeepseek, service.PlatformZhipu, service.PlatformKimi} {
		model := "deepseek-v4-flash"
		switch platform {
		case service.PlatformZhipu:
			model = "glm-5"
		case service.PlatformKimi:
			model = "kimi-k2.5"
		}
		for _, tt := range []struct {
			name         string
			accountCount int
			successID    int64
			wantCalls    []int64
			wantSuccess  bool
		}{
			{"first success stops probing", 3, 1, []int64{1}, true},
			{"switches to successful account", 3, 2, []int64{1, 2}, true},
			{"fifth account can succeed", 6, 5, []int64{1, 2, 3, 4, 5}, true},
			{"all available accounts fail", 2, 0, []int64{1, 2}, false},
			{"five-account budget", 6, 6, []int64{1, 2, 3, 4, 5}, false},
		} {
			t.Run(platform+"/"+tt.name, func(t *testing.T) {
				groupID := int64(9400)
				group := &service.Group{ID: groupID, Hydrated: true, Platform: platform, Status: service.StatusActive}
				accounts := make([]service.Account, tt.accountCount)
				for i := range accounts {
					accounts[i] = service.Account{
						ID: int64(i + 1), Platform: platform, Type: service.AccountTypeAPIKey,
						Status: service.StatusActive, Schedulable: true, Concurrency: 1, Priority: i,
						Credentials: map[string]any{"api_key": "test-key", "base_url": "https://upstream.example.com"},
					}
				}
				cfg := &config.Config{RunMode: config.RunModeSimple}
				cfg.Default.RateMultiplier = 1
				cfg.Gateway.MaxAccountSwitches = 10
				billingCache := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
				t.Cleanup(billingCache.Stop)
				upstream := &channelMonitorPlatformUpstream{successID: tt.successID, chat: true, firstWait: 100 * time.Millisecond}
				gateway := service.NewOpenAIGatewayService(
					&openAIWSFailoverHandlerAccountRepoStub{accounts: accounts},
					nil, nil, nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, nil),
					nil, billingCache, upstream, &service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil, nil,
				)
				h := NewOpenAIGatewayHandler(
					gateway, service.NewConcurrencyService(&fakeConcurrencyCache{}), billingCache,
					service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg), nil, nil, nil, nil, cfg,
				)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"`+model+`","messages":[{"role":"user","content":"2+2=?"}],"stream":false}`))
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.Group, group))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Request.Header.Set(service.ChannelMonitorProbeAttemptsHeader, strconv.Itoa(service.ChannelMonitorProbeAttempts))
				c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
					ID: 9401, UserID: 9402, GroupID: &groupID, Group: group,
					User: &service.User{ID: 9402, Balance: 100},
				})
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 9402, Concurrency: 0})

				started := time.Now()
				h.ChatCompletions(c)
				elapsed := time.Since(started)
				require.Equal(t, tt.wantCalls, upstream.calls, "HTTP %d: %s", rec.Code, rec.Body.String())
				if tt.wantSuccess {
					require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
					require.Contains(t, rec.Body.String(), `"content":"4"`)
					latency, err := strconv.Atoi(rec.Header().Get(service.ChannelMonitorProbeLatencyHeader))
					require.NoError(t, err)
					require.GreaterOrEqual(t, latency, 0)
					if len(tt.wantCalls) > 1 {
						require.Less(t, time.Duration(latency)*time.Millisecond, elapsed-upstream.firstWait/2)
					}
				} else {
					require.GreaterOrEqual(t, rec.Code, 400, rec.Body.String())
				}
			})
		}
	}
}
