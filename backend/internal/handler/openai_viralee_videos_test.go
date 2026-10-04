package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type viraleeTaskRepoStub struct {
	miniMaxRecoveryHandlerRepoStub
}

func (r *viraleeTaskRepoStub) UpsertOpenAIVideoTaskRoute(ctx context.Context, binding service.OpenAIVideoTaskBinding) error {
	// The persistent repository writes this default for ordinary video tasks.
	binding.RecoveryStatus = "inactive"
	return r.miniMaxRecoveryHandlerRepoStub.UpsertOpenAIVideoTaskRoute(ctx, binding)
}

func (r *viraleeTaskRepoStub) ArmOpenAIVideoTaskCompensation(_ context.Context, _, _ int64, _ string, key *service.APIKey, next time.Time) error {
	r.binding.APIKeyID = key.ID
	r.binding.CompensationStatus = "pending"
	r.binding.NextCheckAt = &next
	return nil
}

type viraleeVideoUpstreamStub struct {
	service.HTTPUpstream
	t         *testing.T
	model     string
	accountID int64
	calls     int
}

func (s *viraleeVideoUpstreamStub) Do(req *http.Request, _ string, accountID int64, _ int) (*http.Response, error) {
	s.calls++
	require.Equal(s.t, s.accountID, accountID, "poll on the account that created the task")
	require.Equal(s.t, "Bearer upstream-key", req.Header.Get("Authorization"))
	if req.Method == http.MethodPost {
		require.Equal(s.t, "/v1/videos", req.URL.Path)
		return miniMaxRecoveryHTTPResponse(http.StatusOK, fmt.Sprintf(`{"task_id":"task-viralee","model":%q,"status":"queued"}`, s.model)), nil
	}
	require.Equal(s.t, "/v1/videos/task-viralee", req.URL.Path)
	return miniMaxRecoveryHTTPResponse(http.StatusOK, `{"id":"task-viralee","status":"completed","url":"https://cdn.test/video.mp4"}`), nil
}

func TestViraleeVideosCreateBindChargeAndPoll(t *testing.T) {
	for _, model := range []string{"wan3.0x", "wan3.0x-480p", "wan3.0x-1080p", "viraldance933", "viraldance933-fast", "viraldance2.5-30", "viraldance2.5-15", "viraldance2.5-480p-15", "dola-viraldance2.0", "dola-viraldance2.5"} {
		t.Run(model, func(t *testing.T) {
			groupID := int64(7088)
			price := .5
			user := &service.User{ID: 8088}
			key := &service.APIKey{ID: 9088, GroupID: &groupID, User: user, Group: &service.Group{
				ID: groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, Hydrated: true,
				RateMultiplier: 1, VideoRateMultiplier: 1, AllowImageGeneration: true,
				VideoPrice480P: &price, VideoPrice720P: &price, VideoPrice1080P: &price,
			}}
			accounts := &nanoBananaAccountRepoStub{account: service.Account{
				ID: 9188, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
				Status: service.StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []int64{groupID},
				Credentials: map[string]any{"api_key": "upstream-key", "base_url": "https://video.test/v1", "model_mapping": map[string]any{model: model}},
			}}
			cfg := &config.Config{}
			cache := service.NewBillingCacheService(&nanoBananaBillingCacheStub{balance: 100}, nil, nil, nil, nil, nil, cfg, nil)
			t.Cleanup(cache.Stop)
			usage := &nanoBananaUsageLogRepoStub{}
			billing := &nanoBananaBillingRepoStub{usageRepo: usage}
			routes := &viraleeTaskRepoStub{}
			upstream := &viraleeVideoUpstreamStub{t: t, model: model, accountID: accounts.account.ID}
			gateway := service.NewOpenAIGatewayService(accounts, usage, billing, nil, nil, nil, nil, cfg, nil, nil,
				service.NewBillingService(cfg, nil), nil, cache, upstream, &service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil, routes)
			h := NewOpenAIGatewayHandler(gateway, service.NewConcurrencyService(nil), cache, &service.APIKeyService{}, nil, nil, nil, nil, cfg)
			seconds := 10
			if model == "dola-viraldance2.5" {
				seconds = 30
			}
			body := fmt.Sprintf(`{"model":%q,"prompt":"waves","duration":%d,"ratio":"16:9"}`, model, seconds)
			perform := func(method, path, body string) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Set(string(middleware2.ContextKeyAPIKey), key)
				c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})
				if method == http.MethodPost {
					h.Videos(c)
				} else {
					c.Params = gin.Params{{Key: "request_id", Value: "task-viralee"}}
					h.VideoStatus(c)
				}
				return rec
			}
			rec := perform(http.MethodPost, "/v1/videos", body)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.Equal(t, "task-viralee", gjson.GetBytes(rec.Body.Bytes(), "task_id").String())
			require.NotNil(t, billing.captureCmd)
			require.InDelta(t, float64(seconds)*price, billing.reserveCmd.HoldAmount, 1e-8)
			require.InDelta(t, float64(seconds)*price, billing.captureCmd.ActualAmount, 1e-8)
			require.Equal(t, seconds, *usage.lastLog.VideoDurationSeconds)
			require.Nil(t, billing.releaseCmd)
			require.Equal(t, service.OpenAIVideoUsageRequestID("task-viralee"), usage.lastLog.RequestID)
			require.Equal(t, "pending", routes.binding.CompensationStatus)
			require.Equal(t, accounts.account.ID, routes.binding.AccountID)
			require.Equal(t, "inactive", routes.binding.RecoveryStatus)
			originalUsage := usage.lastLog
			rec = perform(http.MethodGet, "/v1/videos/task-viralee", "")
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.Equal(t, "https://cdn.test/video.mp4", gjson.GetBytes(rec.Body.Bytes(), "url").String())
			require.Equal(t, "completed", routes.binding.CompensationStatus)
			require.Same(t, originalUsage, usage.lastLog, "polling must not add usage charges")
			require.Equal(t, 2, upstream.calls)
		})
	}
}
