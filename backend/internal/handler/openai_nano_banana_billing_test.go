package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type nanoBananaBillingCacheStub struct {
	service.BillingCache
	balance float64
}

func (s *nanoBananaBillingCacheStub) GetUserBalance(context.Context, int64) (float64, error) {
	return s.balance, nil
}

func (s *nanoBananaBillingCacheStub) InvalidateUserBalance(context.Context, int64) error {
	return nil
}

type nanoBananaBillingRepoStub struct {
	service.UsageBillingRepository
	usageRepo  *nanoBananaUsageLogRepoStub
	reserveCmd *service.BatchImageBalanceHoldCommand
	captureCmd *service.BatchImageBalanceHoldCommand
	releaseCmd *service.BatchImageBalanceHoldCommand
	applyCmd   *service.UsageBillingCommand
	applyErr   error
}

func (s *nanoBananaBillingRepoStub) ReserveBatchImageBalance(_ context.Context, cmd *service.BatchImageBalanceHoldCommand) (*service.BatchImageBalanceHoldResult, error) {
	copy := *cmd
	s.reserveCmd = &copy
	return &service.BatchImageBalanceHoldResult{Applied: true}, nil
}

func (s *nanoBananaBillingRepoStub) CaptureBatchImageBalance(_ context.Context, cmd *service.BatchImageBalanceHoldCommand) (*service.BatchImageBalanceHoldResult, error) {
	copy := *cmd
	s.captureCmd = &copy
	return &service.BatchImageBalanceHoldResult{Applied: true}, nil
}

func (s *nanoBananaBillingRepoStub) ReleaseBatchImageBalance(_ context.Context, cmd *service.BatchImageBalanceHoldCommand) (*service.BatchImageBalanceHoldResult, error) {
	copy := *cmd
	s.releaseCmd = &copy
	return &service.BatchImageBalanceHoldResult{Applied: true}, nil
}

func (s *nanoBananaBillingRepoStub) Apply(_ context.Context, cmd *service.UsageBillingCommand) (*service.UsageBillingApplyResult, error) {
	copy := *cmd
	s.applyCmd = &copy
	return &service.UsageBillingApplyResult{Applied: true}, s.applyErr
}

func (s *nanoBananaBillingRepoStub) ApplyWithUsageLog(ctx context.Context, cmd *service.UsageBillingCommand, log *service.UsageLog) (*service.UsageBillingApplyResult, error) {
	result, err := s.Apply(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if cmd.CapturedBalanceHold != nil {
		copy := *cmd.CapturedBalanceHold
		s.captureCmd = &copy
	}
	if s.usageRepo != nil {
		_, err = s.usageRepo.Create(ctx, log)
	}
	return result, err
}

type nanoBananaUsageLogRepoStub struct {
	service.UsageLogRepository
	lastLog *service.UsageLog
}

func (s *nanoBananaUsageLogRepoStub) Create(_ context.Context, log *service.UsageLog) (bool, error) {
	copy := *log
	s.lastLog = &copy
	return true, nil
}

type nanoBananaAccountRepoStub struct {
	service.AccountRepository
	account service.Account
}

func (s *nanoBananaAccountRepoStub) GetByID(_ context.Context, id int64) (*service.Account, error) {
	if id != s.account.ID {
		return nil, service.ErrNoAvailableAccounts
	}
	copy := s.account
	return &copy, nil
}

func (s *nanoBananaAccountRepoStub) ListSchedulableByGroupIDAndPlatform(_ context.Context, _ int64, platform string) ([]service.Account, error) {
	return s.accountsForPlatform(platform), nil
}

func (s *nanoBananaAccountRepoStub) ListSchedulableByPlatform(_ context.Context, platform string) ([]service.Account, error) {
	return s.accountsForPlatform(platform), nil
}

func (s *nanoBananaAccountRepoStub) ListSchedulableUngroupedByPlatform(_ context.Context, platform string) ([]service.Account, error) {
	return s.accountsForPlatform(platform), nil
}

func (s *nanoBananaAccountRepoStub) accountsForPlatform(platform string) []service.Account {
	if s.account.Platform != platform {
		return nil
	}
	return []service.Account{s.account}
}

type nanoBananaUpstreamStub struct {
	service.HTTPUpstream
	calls int
}

func (s *nanoBananaUpstreamStub) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.calls++
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Request-Id": []string{"nano-request"},
		},
		Body: io.NopCloser(bytes.NewBufferString(`{"data":{"imageUrl":"https://example.com/image.png"}}`)),
	}, nil
}

func TestNanoBananaRejectsInsufficientBalanceBeforeUpstream(t *testing.T) {
	handler, billingRepo, _, upstream, apiKey := newNanoBananaBillingTestHandler(t, 5)

	rec := performNanoBananaBillingRequest(handler, apiKey, `{"model":"nano-banana-pro","prompt":"draw","imageSize":"2K","n":2}`)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Equal(t, "billing_error", gjson.GetBytes(rec.Body.Bytes(), "error.type").String())
	require.Zero(t, upstream.calls)
	require.Nil(t, billingRepo.reserveCmd)
}

func TestNanoBananaReservesEstimatedCostAndCapturesActualImageCount(t *testing.T) {
	handler, billingRepo, usageRepo, upstream, apiKey := newNanoBananaBillingTestHandler(t, 100)

	rec := performNanoBananaBillingRequest(handler, apiKey, `{"model":"nano-banana-pro","prompt":"draw","imageSize":"2K","n":2}`)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, upstream.calls)
	require.NotNil(t, billingRepo.reserveCmd)
	require.InDelta(t, 6, billingRepo.reserveCmd.HoldAmount, 1e-12)
	require.NotNil(t, billingRepo.captureCmd)
	require.InDelta(t, 3, billingRepo.captureCmd.ActualAmount, 1e-12)
	require.Nil(t, billingRepo.releaseCmd)
	require.NotNil(t, billingRepo.applyCmd)
	require.Zero(t, billingRepo.applyCmd.BalanceCost)
	require.Equal(t, 1, billingRepo.applyCmd.ImageCount)
	require.NotNil(t, usageRepo.lastLog)
	require.Equal(t, 1, usageRepo.lastLog.ImageCount)
}

func TestNanoBananaSettlementFailureDoesNotReleaseSuccessfulGeneration(t *testing.T) {
	handler, billing, usage, upstream, key := newNanoBananaBillingTestHandler(t, 100)
	billing.applyErr = errors.New("temporary settlement failure")
	rec := performNanoBananaBillingRequest(handler, key, `{"model":"nano-banana-pro","prompt":"draw","imageSize":"2K"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, upstream.calls)
	require.NotNil(t, billing.reserveCmd)
	require.NotNil(t, billing.applyCmd)
	require.Nil(t, billing.releaseCmd, "completed work retains its reservation for reconciliation")
	require.Nil(t, usage.lastLog, "a failed transaction must not create a zero-cost success row")
}

func newNanoBananaBillingTestHandler(t *testing.T, balance float64) (*OpenAIGatewayHandler, *nanoBananaBillingRepoStub, *nanoBananaUsageLogRepoStub, *nanoBananaUpstreamStub, *service.APIKey) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	groupID := int64(7001)
	imagePrice := 3.0
	user := &service.User{ID: 8001}
	group := &service.Group{
		ID:                   groupID,
		Platform:             service.PlatformOpenAI,
		RateMultiplier:       1,
		AllowImageGeneration: true,
		ImageRateMultiplier:  1,
		ImagePrice2K:         &imagePrice,
	}
	apiKey := &service.APIKey{ID: 9001, GroupID: &groupID, Group: group, User: user}
	accountRepo := &nanoBananaAccountRepoStub{account: service.Account{
		ID:          9101,
		Name:        "nano-banana",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		GroupIDs:    []int64{groupID},
		Credentials: map[string]any{
			"api_key":  "upstream-key",
			"base_url": "https://visionary.beer",
		},
	}}
	billingCache := &nanoBananaBillingCacheStub{balance: balance}
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1
	billingCacheService := service.NewBillingCacheService(billingCache, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billingCacheService.Stop)
	usageRepo := &nanoBananaUsageLogRepoStub{}
	billingRepo := &nanoBananaBillingRepoStub{usageRepo: usageRepo}
	upstream := &nanoBananaUpstreamStub{}
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
		nil,
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
	return handler, billingRepo, usageRepo, upstream, apiKey
}

func performNanoBananaBillingRequest(handler *OpenAIGatewayHandler, apiKey *service.APIKey, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/api/nano-banana", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: apiKey.User.ID})
	handler.NanoBanana(c)
	return rec
}
