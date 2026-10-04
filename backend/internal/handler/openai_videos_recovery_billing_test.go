package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type miniMaxRecoveryHandlerRepoStub struct {
	service.OpenAIVideoTaskBindingRepository
	binding *service.OpenAIVideoTaskBinding
}

func (s *miniMaxRecoveryHandlerRepoStub) PrepareOpenAIVideoRecovery(_ context.Context, binding service.OpenAIVideoTaskBinding) error {
	copy := binding
	copy.ID = 1
	s.binding = &copy
	return nil
}

func (s *miniMaxRecoveryHandlerRepoStub) GetOpenAIVideoRecovery(_ context.Context, groupID, userID int64, taskID string) (*service.OpenAIVideoTaskBinding, error) {
	if s.binding == nil || s.binding.GroupID != groupID || s.binding.UserID != userID || s.binding.TaskID != taskID {
		return nil, nil
	}
	copy := *s.binding
	return &copy, nil
}

func (s *miniMaxRecoveryHandlerRepoStub) UpsertOpenAIVideoTaskRoute(_ context.Context, binding service.OpenAIVideoTaskBinding) error {
	copy := binding
	s.binding = &copy
	return nil
}

func (s *miniMaxRecoveryHandlerRepoStub) UpdateOpenAIVideoRecovery(_ context.Context, _ int64, status, upstreamTaskID string, nextCheckAt *time.Time, lastError string) error {
	if s.binding == nil {
		return nil
	}
	s.binding.RecoveryStatus = status
	if strings.TrimSpace(upstreamTaskID) != "" {
		s.binding.UpstreamTaskID = strings.TrimSpace(upstreamTaskID)
	}
	s.binding.RecoveryNextCheckAt = nextCheckAt
	if strings.TrimSpace(lastError) == "" {
		s.binding.RecoveryLastError = nil
	} else {
		message := strings.TrimSpace(lastError)
		s.binding.RecoveryLastError = &message
	}
	return nil
}

func (s *miniMaxRecoveryHandlerRepoStub) ClaimOpenAIVideoRecoveries(context.Context, time.Time, time.Time, int) ([]service.OpenAIVideoTaskBinding, error) {
	return nil, nil
}

func (s *miniMaxRecoveryHandlerRepoStub) UpsertOpenAIVideoTaskBinding(_ context.Context, binding service.OpenAIVideoTaskBinding) error {
	if s.binding == nil {
		copy := binding
		s.binding = &copy
		return nil
	}
	s.binding.AccountID = binding.AccountID
	s.binding.ExpiresAt = binding.ExpiresAt
	return nil
}

func (s *miniMaxRecoveryHandlerRepoStub) GetOpenAIVideoTaskBinding(_ context.Context, groupID, userID int64, taskID string) (*service.OpenAIVideoTaskBinding, error) {
	return s.GetOpenAIVideoRecovery(context.Background(), groupID, userID, taskID)
}

func (s *miniMaxRecoveryHandlerRepoStub) UpdateOpenAIVideoTaskCompensation(_ context.Context, _ int64, status string, nextCheckAt *time.Time, lastError string) error {
	if s.binding == nil {
		return nil
	}
	s.binding.CompensationStatus = status
	s.binding.NextCheckAt = nextCheckAt
	return nil
}

type miniMaxRecoveryUpstreamStub struct {
	service.HTTPUpstream
	calls        int
	statusByID   bool
	lastReqPath  string
	requestPaths []string
}

func (s *miniMaxRecoveryUpstreamStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.calls++
	s.lastReqPath = req.URL.Path
	s.requestPaths = append(s.requestPaths, req.URL.Path)
	if s.statusByID && req.URL.Host == "cdn.example.test" && req.Method == http.MethodGet {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"video/mp4"}},
			Body: io.NopCloser(bytes.NewReader([]byte{
				0x00, 0x00, 0x00, 0x18,
				0x66, 0x74, 0x79, 0x70,
				0x69, 0x73, 0x6f, 0x6d,
			}))}, nil
	}
	if s.statusByID && req.Method == http.MethodGet {
		return miniMaxRecoveryHTTPResponse(http.StatusOK, `{"task":{"id":"upstream-task-1","status":"succeeded","content":{"url":"https://cdn.example.test/video.mp4"}}}`), nil
	}
	if req.Method == http.MethodGet {
		return miniMaxRecoveryHTTPResponse(http.StatusOK, `{"items":[]}`), nil
	}
	return miniMaxRecoveryHTTPResponse(http.StatusGatewayTimeout, `{"error":{"message":"upstream timed out"}}`), nil
}

func miniMaxRecoveryHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type fireflyBoundTaskUpstreamStub struct {
	service.HTTPUpstream
	requestPaths []string
}

func (s *fireflyBoundTaskUpstreamStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.requestPaths = append(s.requestPaths, req.URL.Path)
	switch req.URL.Path {
	case "/v1/videos/firefly-task-1":
		return miniMaxRecoveryHTTPResponse(http.StatusOK, `{"id":"firefly-task-1","status":"completed"}`), nil
	case "/v1/videos/firefly-task-1/content":
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"video/mp4"}},
			Body:       io.NopCloser(bytes.NewReader([]byte("firefly-mp4"))),
		}, nil
	default:
		return miniMaxRecoveryHTTPResponse(http.StatusNotFound, `{"error":{"message":"unexpected path"}}`), nil
	}
}

func TestVideosMiniMax504KeepsHoldWithoutImmediateCharge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(7002)
	price := 1.0
	user := &service.User{ID: 8002}
	group := &service.Group{
		ID:                   groupID,
		Platform:             service.PlatformOpenAI,
		Status:               service.StatusActive,
		Hydrated:             true,
		RateMultiplier:       1,
		AllowImageGeneration: true,
		VideoRateIndependent: true,
		VideoRateMultiplier:  1,
		VideoPrice480P:       &price,
	}
	apiKey := &service.APIKey{ID: 9002, GroupID: &groupID, Group: group, User: user}
	accountRepo := &nanoBananaAccountRepoStub{account: service.Account{
		ID:          9102,
		Name:        "minimax",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 1,
		GroupIDs:    []int64{groupID},
		Credentials: map[string]any{
			"api_key":             "upstream-key",
			"base_url":            "https://metaso.example/api/minimax",
			"openai_capabilities": []any{"minimax_video"},
		},
	}}
	billingCache := &nanoBananaBillingCacheStub{balance: 100}
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1
	billingCacheService := service.NewBillingCacheService(billingCache, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billingCacheService.Stop)
	usageRepo := &nanoBananaUsageLogRepoStub{}
	billingRepo := &nanoBananaBillingRepoStub{usageRepo: usageRepo}
	upstream := &miniMaxRecoveryUpstreamStub{}
	recoveryRepo := &miniMaxRecoveryHandlerRepoStub{}
	gatewayService := service.NewOpenAIGatewayService(
		accountRepo,
		usageRepo,
		billingRepo,
		nil,
		nil,
		nil,
		nil,
		cfg,
		nil,
		nil,
		service.NewBillingService(cfg, nil),
		nil,
		billingCacheService,
		upstream,
		&service.DeferredService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		recoveryRepo,
	)
	handler := NewOpenAIGatewayHandler(
		gatewayService,
		service.NewConcurrencyService(nil),
		billingCacheService,
		&service.APIKeyService{},
		nil,
		nil,
		nil,
		nil,
		cfg,
	)
	body := `{"model":"MiniMax-H3","content":[{"type":"text","text":"cinematic waves"}],"resolution":"2K","duration":5,"ratio":"16:9"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.Videos(c)

	require.Equalf(t, http.StatusOK, rec.Code, "response: %s", rec.Body.String())
	require.True(t, strings.HasPrefix(gjson.GetBytes(rec.Body.Bytes(), "task_id").String(), "task_recovery_"))
	require.Equal(t, "processing", gjson.GetBytes(rec.Body.Bytes(), "status").String())
	require.Equal(t, 2, upstream.calls, "baseline lookup and one create request are expected")
	require.NotNil(t, billingRepo.reserveCmd)
	require.Nil(t, billingRepo.captureCmd, "an unconfirmed 504 submission must not capture the hold")
	require.Nil(t, billingRepo.applyCmd, "an unconfirmed 504 submission must not charge usage")
	require.Nil(t, billingRepo.releaseCmd, "the persisted recovery must retain the hold")
	require.Nil(t, usageRepo.lastLog, "an unconfirmed 504 submission must not create a usage record")
	require.NotNil(t, recoveryRepo.binding)
	require.Equal(t, service.OpenAIVideoRecoveryPending, recoveryRepo.binding.RecoveryStatus)
}

func TestVideoStatusUsesIdentifiedUpstreamTaskBeforeMatchedUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(7003)
	user := &service.User{ID: 8003}
	group := &service.Group{
		ID:                   groupID,
		Platform:             service.PlatformOpenAI,
		Status:               service.StatusActive,
		Hydrated:             true,
		RateMultiplier:       1,
		AllowImageGeneration: true,
	}
	apiKey := &service.APIKey{ID: 9003, GroupID: &groupID, Group: group, User: user}
	accountRepo := &nanoBananaAccountRepoStub{account: service.Account{
		ID:          9103,
		Name:        "minimax",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 1,
		GroupIDs:    []int64{groupID},
		Credentials: map[string]any{
			"api_key":             "upstream-key",
			"base_url":            "https://metaso.example/api/minimax",
			"openai_capabilities": []any{"minimax_video"},
		},
	}}
	nextCheckAt := time.Now().Add(time.Minute)
	recoveryRepo := &miniMaxRecoveryHandlerRepoStub{binding: &service.OpenAIVideoTaskBinding{
		ID:                  1,
		GroupID:             groupID,
		UserID:              user.ID,
		TaskID:              "task_recovery_local",
		AccountID:           9103,
		UpstreamTaskID:      "upstream-task-1",
		BillingTaskID:       "task_recovery_local",
		RecoveryStatus:      service.OpenAIVideoRecoveryIdentified,
		RecoveryNextCheckAt: &nextCheckAt,
		CompensationStatus:  "",
		ExpiresAt:           time.Now().Add(time.Hour),
	}}
	upstream := &miniMaxRecoveryUpstreamStub{statusByID: true}
	cfg := &config.Config{}
	billingCacheService := service.NewBillingCacheService(&nanoBananaBillingCacheStub{balance: 100}, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billingCacheService.Stop)
	gatewayService := service.NewOpenAIGatewayService(
		accountRepo,
		&nanoBananaUsageLogRepoStub{},
		&nanoBananaBillingRepoStub{},
		nil,
		nil,
		nil,
		nil,
		cfg,
		nil,
		nil,
		service.NewBillingService(cfg, nil),
		nil,
		billingCacheService,
		upstream,
		&service.DeferredService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		recoveryRepo,
	)
	handler := NewOpenAIGatewayHandler(
		gatewayService,
		service.NewConcurrencyService(nil),
		billingCacheService,
		&service.APIKeyService{},
		nil,
		nil,
		nil,
		nil,
		cfg,
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/videos/task_recovery_local", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "task_id", Value: "task_recovery_local"}}
	c.Request = req
	c.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.VideoStatus(c)

	require.Equalf(t, http.StatusOK, rec.Code, "response: %s", rec.Body.String())
	require.Equal(t, "succeeded", gjson.GetBytes(rec.Body.Bytes(), "task.status").String())
	require.Equal(t, "task_recovery_local", gjson.GetBytes(rec.Body.Bytes(), "task.id").String())
	require.Contains(t, upstream.lastReqPath, "/v2/query/video_generation/upstream-task-1")
}

func TestFireflyVideoStatusAndContentDoNotUseInactiveMiniMaxRecoveryRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(7006)
	user := &service.User{ID: 8006}
	group := &service.Group{
		ID:                   groupID,
		Platform:             service.PlatformOpenAI,
		Status:               service.StatusActive,
		Hydrated:             true,
		RateMultiplier:       1,
		AllowImageGeneration: true,
	}
	apiKey := &service.APIKey{ID: 9006, GroupID: &groupID, Group: group, User: user}
	accountRepo := &nanoBananaAccountRepoStub{account: service.Account{
		ID:          9106,
		Name:        "firefly",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 1,
		GroupIDs:    []int64{groupID},
		Credentials: map[string]any{
			"api_key":  "upstream-key",
			"base_url": "https://adobe.example",
		},
	}}
	require.False(t, accountRepo.account.SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityMiniMaxVideo))
	recoveryRepo := &miniMaxRecoveryHandlerRepoStub{binding: &service.OpenAIVideoTaskBinding{
		ID:                 1,
		GroupID:            groupID,
		UserID:             user.ID,
		TaskID:             "firefly-task-1",
		AccountID:          9106,
		UpstreamTaskID:     "firefly-task-1",
		BillingTaskID:      "firefly-task-1",
		RecoveryStatus:     "inactive",
		CompensationStatus: "pending",
		ExpiresAt:          time.Now().Add(time.Hour),
	}}
	upstream := &fireflyBoundTaskUpstreamStub{}
	cfg := &config.Config{}
	billingCacheService := service.NewBillingCacheService(&nanoBananaBillingCacheStub{balance: 100}, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billingCacheService.Stop)
	gatewayService := service.NewOpenAIGatewayService(
		accountRepo,
		&nanoBananaUsageLogRepoStub{},
		&nanoBananaBillingRepoStub{},
		nil,
		nil,
		nil,
		nil,
		cfg,
		nil,
		nil,
		service.NewBillingService(cfg, nil),
		nil,
		billingCacheService,
		upstream,
		&service.DeferredService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		recoveryRepo,
	)
	handler := NewOpenAIGatewayHandler(
		gatewayService,
		service.NewConcurrencyService(nil),
		billingCacheService,
		&service.APIKeyService{},
		nil,
		nil,
		nil,
		nil,
		cfg,
	)

	statusReq := httptest.NewRequest(http.MethodGet, "/v1/videos/firefly-task-1", nil)
	statusRec := httptest.NewRecorder()
	statusContext, _ := gin.CreateTestContext(statusRec)
	statusContext.Params = gin.Params{{Key: "task_id", Value: "firefly-task-1"}}
	statusContext.Request = statusReq
	statusContext.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	statusContext.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.VideoStatus(statusContext)

	require.Equalf(t, http.StatusOK, statusRec.Code, "response: %s", statusRec.Body.String())
	require.Equal(t, "completed", gjson.GetBytes(statusRec.Body.Bytes(), "status").String())

	contentReq := httptest.NewRequest(http.MethodGet, "/v1/videos/firefly-task-1/content", nil)
	contentRec := httptest.NewRecorder()
	contentContext, _ := gin.CreateTestContext(contentRec)
	contentContext.Params = gin.Params{{Key: "task_id", Value: "firefly-task-1"}}
	contentContext.Request = contentReq
	contentContext.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	contentContext.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.VideoContent(contentContext)

	require.Equal(t, http.StatusOK, contentRec.Code)
	require.Equal(t, "video/mp4", contentRec.Header().Get("Content-Type"))
	require.Equal(t, "firefly-mp4", contentRec.Body.String())
	require.Equal(t, []string{
		"/v1/videos/firefly-task-1",
		"/v1/videos/firefly-task-1/content",
	}, upstream.requestPaths)
}

func TestRecoveredVideoStatusAndContentUseUpstreamTaskAndKeepLocalID(t *testing.T) {
	for _, recoveryStatus := range []string{service.OpenAIVideoRecoveryMatched, service.OpenAIVideoRecoveryBillingReview} {
		t.Run(recoveryStatus, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			groupID := int64(7005)
			user := &service.User{ID: 8005}
			group := &service.Group{
				ID:                   groupID,
				Platform:             service.PlatformOpenAI,
				Status:               service.StatusActive,
				Hydrated:             true,
				RateMultiplier:       1,
				AllowImageGeneration: true,
			}
			apiKey := &service.APIKey{ID: 9005, GroupID: &groupID, Group: group, User: user}
			accountRepo := &nanoBananaAccountRepoStub{account: service.Account{
				ID:          9105,
				Name:        "minimax",
				Platform:    service.PlatformOpenAI,
				Type:        service.AccountTypeAPIKey,
				Status:      service.StatusActive,
				Schedulable: true,
				Concurrency: 1,
				GroupIDs:    []int64{groupID},
				Credentials: map[string]any{
					"api_key":             "upstream-key",
					"base_url":            "https://metaso.example/api/minimax",
					"openai_capabilities": []any{"minimax_video"},
				},
			}}
			recoveryRepo := &miniMaxRecoveryHandlerRepoStub{binding: &service.OpenAIVideoTaskBinding{
				ID:                 1,
				GroupID:            groupID,
				UserID:             user.ID,
				TaskID:             "task_recovery_local_delivery",
				AccountID:          9105,
				UpstreamTaskID:     "upstream-task-1",
				BillingTaskID:      "task_recovery_local_delivery",
				RecoveryStatus:     recoveryStatus,
				CompensationStatus: "pending",
				ExpiresAt:          time.Now().Add(time.Hour),
			}}
			upstream := &miniMaxRecoveryUpstreamStub{statusByID: true}
			cfg := &config.Config{}
			billingCacheService := service.NewBillingCacheService(&nanoBananaBillingCacheStub{balance: 100}, nil, nil, nil, nil, nil, cfg, nil)
			t.Cleanup(billingCacheService.Stop)
			gatewayService := service.NewOpenAIGatewayService(
				accountRepo,
				&nanoBananaUsageLogRepoStub{},
				&nanoBananaBillingRepoStub{},
				nil,
				nil,
				nil,
				nil,
				cfg,
				nil,
				nil,
				service.NewBillingService(cfg, nil),
				nil,
				billingCacheService,
				upstream,
				&service.DeferredService{},
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				recoveryRepo,
			)
			handler := NewOpenAIGatewayHandler(
				gatewayService,
				service.NewConcurrencyService(nil),
				billingCacheService,
				&service.APIKeyService{},
				nil,
				nil,
				nil,
				nil,
				cfg,
			)

			statusReq := httptest.NewRequest(http.MethodGet, "/v1/videos/task_recovery_local_delivery", nil)
			statusRec := httptest.NewRecorder()
			statusContext, _ := gin.CreateTestContext(statusRec)
			statusContext.Params = gin.Params{{Key: "task_id", Value: "task_recovery_local_delivery"}}
			statusContext.Request = statusReq
			statusContext.Set(string(middleware2.ContextKeyAPIKey), apiKey)
			statusContext.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

			handler.VideoStatus(statusContext)

			require.Equalf(t, http.StatusOK, statusRec.Code, "response: %s", statusRec.Body.String())
			require.Equal(t, "succeeded", gjson.GetBytes(statusRec.Body.Bytes(), "task.status").String())
			require.Equal(t, "task_recovery_local_delivery", gjson.GetBytes(statusRec.Body.Bytes(), "task.id").String())
			require.Equal(t, "/v1/videos/task_recovery_local_delivery/content", gjson.GetBytes(statusRec.Body.Bytes(), "task.content.url").String())

			contentReq := httptest.NewRequest(http.MethodGet, "/v1/videos/task_recovery_local_delivery/content", nil)
			contentRec := httptest.NewRecorder()
			contentContext, _ := gin.CreateTestContext(contentRec)
			contentContext.Params = gin.Params{{Key: "task_id", Value: "task_recovery_local_delivery"}}
			contentContext.Request = contentReq
			contentContext.Set(string(middleware2.ContextKeyAPIKey), apiKey)
			contentContext.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

			handler.VideoContent(contentContext)

			require.Equal(t, http.StatusOK, contentRec.Code)
			require.Equal(t, "video/mp4", contentRec.Header().Get("Content-Type"))
			require.Equal(t, []byte{
				0x00, 0x00, 0x00, 0x18,
				0x66, 0x74, 0x79, 0x70,
				0x69, 0x73, 0x6f, 0x6d,
			}, contentRec.Body.Bytes())
			require.Equal(t, []string{
				"/api/minimax/v2/query/video_generation/upstream-task-1",
				"/api/minimax/v2/query/video_generation/upstream-task-1",
				"/video.mp4",
			}, upstream.requestPaths)
		})
	}
}

func TestVideoStatusKeepsUnidentifiedTaskProcessing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(7004)
	user := &service.User{ID: 8004}
	group := &service.Group{
		ID:                   groupID,
		Platform:             service.PlatformOpenAI,
		Status:               service.StatusActive,
		Hydrated:             true,
		RateMultiplier:       1,
		AllowImageGeneration: true,
	}
	apiKey := &service.APIKey{ID: 9004, GroupID: &groupID, Group: group, User: user}
	nextCheckAt := time.Now().Add(time.Minute)
	recoveryRepo := &miniMaxRecoveryHandlerRepoStub{binding: &service.OpenAIVideoTaskBinding{
		ID:                  1,
		GroupID:             groupID,
		UserID:              user.ID,
		TaskID:              "task_recovery_unbilled",
		AccountID:           9104,
		UpstreamTaskID:      "",
		BillingTaskID:       "task_recovery_unbilled",
		RecoveryStatus:      service.OpenAIVideoRecoveryIdentified,
		CompensationStatus:  "pending",
		RecoveryNextCheckAt: &nextCheckAt,
		ExpiresAt:           time.Now().Add(time.Hour),
	}}
	upstream := &miniMaxRecoveryUpstreamStub{statusByID: true}
	cfg := &config.Config{}
	billingCacheService := service.NewBillingCacheService(&nanoBananaBillingCacheStub{balance: 100}, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billingCacheService.Stop)
	gatewayService := service.NewOpenAIGatewayService(
		&nanoBananaAccountRepoStub{},
		&nanoBananaUsageLogRepoStub{},
		&nanoBananaBillingRepoStub{},
		nil,
		nil,
		nil,
		nil,
		cfg,
		nil,
		nil,
		service.NewBillingService(cfg, nil),
		nil,
		billingCacheService,
		upstream,
		&service.DeferredService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		recoveryRepo,
	)
	handler := NewOpenAIGatewayHandler(
		gatewayService,
		service.NewConcurrencyService(nil),
		billingCacheService,
		&service.APIKeyService{},
		nil,
		nil,
		nil,
		nil,
		cfg,
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/videos/task_recovery_unbilled", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "task_id", Value: "task_recovery_unbilled"}}
	c.Request = req
	c.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.VideoStatus(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "processing", gjson.GetBytes(rec.Body.Bytes(), "task.status").String())
	require.Zero(t, upstream.calls, "a task without a verified upstream ID must remain processing")
}
