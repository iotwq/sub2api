package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestEstimateSD20VideoPerSecondBillingByResolution(t *testing.T) {
	const groupID = int64(711)
	svc := newSD20VideoBillingService(groupID)

	tests := []struct {
		name       string
		model      string
		resolution string
		wantTotal  float64
	}{
		{name: "fast 480p", model: "firefly-video-v2-fast", resolution: "480p", wantTotal: 2.25},
		{name: "fast 720p", model: "firefly-video-v2-fast", resolution: "720p", wantTotal: 2.55},
		{name: "standard 480p", model: "firefly-video-v2", resolution: "480p", wantTotal: 2.55},
		{name: "standard 720p", model: "firefly-video-v2", resolution: "720p", wantTotal: 3.60},
		{name: "standard 1080p", model: "firefly-video-v2", resolution: "1080p", wantTotal: 8.10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, inputSeconds, err := svc.EstimateOpenAIVideoCreateBilling(
				context.Background(),
				miniMaxH3BillingAPIKey(groupID, 0.5),
				&User{ID: 42},
				tt.model,
				[]byte(`{"model":"`+tt.model+`","duration":15,"resolution":"`+tt.resolution+`"}`),
			)

			require.NoError(t, err)
			require.Zero(t, inputSeconds)
			require.Equal(t, string(BillingModeVideo), cost.BillingMode)
			require.InDelta(t, tt.wantTotal, cost.TotalCost, 1e-12)
			require.InDelta(t, tt.wantTotal*0.5, cost.ActualCost, 1e-12)
		})
	}
}

func TestEstimateSD20FastVideoRejectsUnsupported1080PPrice(t *testing.T) {
	const groupID = int64(714)
	svc := newSD20VideoBillingService(groupID)

	_, _, err := svc.EstimateOpenAIVideoCreateBilling(
		context.Background(),
		miniMaxH3BillingAPIKey(groupID, 1),
		&User{ID: 42},
		"firefly-video-v2-fast",
		[]byte(`{"model":"firefly-video-v2-fast","duration":5,"resolution":"1080p"}`),
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrModelPricingUnavailable)
	require.Contains(t, err.Error(), "has no 1080p per-second price")
}

func TestSD20VideoPerSecondBillingIgnoresReferenceMediaDuration(t *testing.T) {
	const (
		groupID = int64(712)
	)
	svc := newSD20VideoBillingService(groupID)

	cost, err := svc.calculateOpenAIVideoCost(context.Background(), "firefly-video-v2", miniMaxH3BillingAPIKey(groupID, 0.5), &OpenAIForwardResult{
		Model:                     "firefly-video-v2",
		MediaType:                 "video",
		VideoCount:                2,
		VideoResolution:           "720p",
		VideoDurationSeconds:      15,
		VideoInputDurationSeconds: 23,
	}, 0.5)

	require.NoError(t, err)
	require.InDelta(t, 7.20, cost.TotalCost, 1e-12)
	require.InDelta(t, 3.60, cost.ActualCost, 1e-12)
}

func TestRecordUsageSD20VideoPerSecondBilling(t *testing.T) {
	const (
		groupID = int64(713)
	)
	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	svc := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
	svc.resolver = newSD20VideoPricingResolver(groupID)

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID:                 "generic-video-per-second-final",
			Model:                     "firefly-video-v2-fast",
			UpstreamModel:             "firefly-video-v2-fast",
			MediaType:                 "video",
			VideoCount:                1,
			VideoResolution:           "720p",
			VideoDurationSeconds:      15,
			VideoInputDurationSeconds: 23,
			Duration:                  time.Second,
		},
		APIKey: miniMaxH3BillingAPIKey(groupID, 0.5),
		User:   &User{ID: 42},
		Account: &Account{
			ID:       74,
			Platform: PlatformOpenAI,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.InDelta(t, 2.55, usageRepo.lastLog.TotalCost, 1e-12)
	require.InDelta(t, 1.275, usageRepo.lastLog.ActualCost, 1e-12)
	require.Equal(t, string(BillingModeVideo), *usageRepo.lastLog.BillingMode)
	require.Equal(t, 15, *usageRepo.lastLog.VideoDurationSeconds)
	require.Equal(t, "720p", *usageRepo.lastLog.VideoResolution)
}

func newSD20VideoBillingService(groupID int64) *OpenAIGatewayService {
	resolver := newSD20VideoPricingResolver(groupID)
	billingService := NewBillingService(&config.Config{}, nil)
	return &OpenAIGatewayService{
		billingService: billingService,
		resolver:       resolver,
	}
}

func newSD20VideoPricingResolver(groupID int64) *ModelPricingResolver {
	cache := newEmptyChannelCache()
	prices := map[string]map[string]float64{
		"firefly-video-v2":      {"480p": 0.17, "720p": 0.24, "1080p": 0.54},
		"firefly-video-v2-fast": {"480p": 0.15, "720p": 0.17},
	}
	for model, modelPrices := range prices {
		intervals := make([]PricingInterval, 0, len(modelPrices))
		for _, resolution := range []string{"480p", "720p", "1080p"} {
			price, ok := modelPrices[resolution]
			if !ok {
				continue
			}
			priceCopy := price
			intervals = append(intervals, PricingInterval{TierLabel: resolution, PerRequestPrice: &priceCopy})
		}
		cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: model}] = &ChannelModelPricing{
			Platform:    PlatformOpenAI,
			Models:      []string{model},
			BillingMode: BillingModeVideo,
			Intervals:   intervals,
		}
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	billingService := NewBillingService(&config.Config{}, nil)
	return NewModelPricingResolver(channelService, billingService)
}

func newGenericVideoPricingResolver(groupID int64, secondPrice float64) *ModelPricingResolver {
	cache := newEmptyChannelCache()
	for _, model := range []string{"firefly-video-v2", "firefly-video-v2-fast"} {
		intervals := make([]PricingInterval, 0, 3)
		for _, resolution := range sd20VideoBillingResolutions(model) {
			priceCopy := secondPrice
			intervals = append(intervals, PricingInterval{TierLabel: resolution, PerRequestPrice: &priceCopy})
		}
		cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: model}] = &ChannelModelPricing{
			Platform:    PlatformOpenAI,
			Models:      []string{model},
			BillingMode: BillingModeVideo,
			Intervals:   intervals,
		}
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	return NewModelPricingResolver(channelService, NewBillingService(&config.Config{}, nil))
}
