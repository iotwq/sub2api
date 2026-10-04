package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestEstimateMiniMaxH3VideoBillingChargesOutputAndVerifiedInputVideo(t *testing.T) {
	dataDir := t.TempDir()
	writeMiniMaxH3TestVideo(t, dataDir, "asset-1", "reference.mp4", 7.5)
	groupID := int64(701)
	price768P, price2K := 0.08, 0.20
	svc := newMiniMaxH3VideoBillingService(t, dataDir, groupID, price768P, price2K)
	body := []byte(`{
		"model":"MiniMax-H3",
		"content":[
			{"type":"text","text":"cinematic scene"},
			{"type":"image_url","image_url":{"url":"https://api.example.test/pg/assets/image/frame.png"},"role":"reference_image"},
			{"type":"video_url","video_url":{"url":"https://api.example.test/pg/assets/asset-1/reference.mp4"},"role":"reference_video"},
			{"type":"audio_url","audio_url":{"url":"https://api.example.test/pg/assets/audio/voice.mp3"},"role":"reference_audio"}
		],
		"resolution":"2K",
		"duration":5,
		"ratio":"adaptive"
	}`)

	cost, inputSeconds, err := svc.EstimateOpenAIVideoCreateBilling(context.Background(), miniMaxH3BillingAPIKey(groupID, 0.5), &User{ID: 42}, "MiniMax-H3", body)

	require.NoError(t, err)
	require.InDelta(t, 7.5, inputSeconds, 1e-9)
	require.Equal(t, string(BillingModeVideo), cost.BillingMode)
	require.InDelta(t, 12.5*price2K, cost.TotalCost, 1e-12)
	require.InDelta(t, 12.5*price2K*0.5, cost.ActualCost, 1e-12)
}

func TestEstimateMiniMaxH3VideoBillingUses768PPriceAndKeepsImagesAudioFree(t *testing.T) {
	groupID := int64(702)
	price768P, price2K := 0.08, 0.20
	svc := newMiniMaxH3VideoBillingService(t, t.TempDir(), groupID, price768P, price2K)
	content := []string{`{"type":"text","text":"cinematic scene"}`}
	for i := 0; i < miniMaxH3MaxReferenceImages; i++ {
		content = append(content, fmt.Sprintf(`{"type":"image_url","image_url":{"url":"https://cdn.example.test/%d.png"},"role":"reference_image"}`, i))
	}
	for i := 0; i < miniMaxH3MaxReferenceAudios; i++ {
		content = append(content, fmt.Sprintf(`{"type":"audio_url","audio_url":{"url":"https://cdn.example.test/%d.mp3"},"role":"reference_audio"}`, i))
	}
	body := []byte(fmt.Sprintf(`{"model":"MiniMax-H3","content":[%s],"resolution":"768P","duration":5}`, strings.Join(content, ",")))

	cost, inputSeconds, err := svc.EstimateOpenAIVideoCreateBilling(context.Background(), miniMaxH3BillingAPIKey(groupID, 1), &User{ID: 42}, "MiniMax-H3", body)

	require.NoError(t, err)
	require.Zero(t, inputSeconds)
	require.InDelta(t, 5*price768P, cost.TotalCost, 1e-12)
}

func TestEstimateMiniMaxH3VideoBillingRejectsUnverifiableOrExcessMedia(t *testing.T) {
	groupID := int64(703)
	svc := newMiniMaxH3VideoBillingService(t, t.TempDir(), groupID, 0.08, 0.20)
	apiKey := miniMaxH3BillingAPIKey(groupID, 1)

	_, _, err := svc.EstimateOpenAIVideoCreateBilling(context.Background(), apiKey, &User{ID: 42}, "MiniMax-H3", []byte(`{
		"model":"MiniMax-H3",
		"content":[{"type":"text","text":"scene"},{"type":"video_url","video_url":{"url":"https://cdn.example.test/reference.mp4"},"role":"reference_video"}],
		"resolution":"2K","duration":5
	}`))
	require.ErrorContains(t, err, "uploaded through this service")

	content := []string{`{"type":"text","text":"scene"}`}
	for i := 0; i < miniMaxH3MaxReferenceVideos+1; i++ {
		content = append(content, fmt.Sprintf(`{"type":"video_url","video_url":{"url":"https://api.example.test/pg/assets/a%d/reference.mp4"},"role":"reference_video"}`, i))
	}
	_, _, err = svc.EstimateOpenAIVideoCreateBilling(context.Background(), apiKey, &User{ID: 42}, "MiniMax-H3", []byte(fmt.Sprintf(
		`{"model":"MiniMax-H3","content":[%s],"resolution":"2K","duration":5}`,
		strings.Join(content, ","),
	)))
	require.ErrorContains(t, err, "at most 3 reference videos")

	content = []string{`{"type":"text","text":"scene"}`}
	for i := 0; i < miniMaxH3MaxReferenceImages+1; i++ {
		content = append(content, fmt.Sprintf(`{"type":"image_url","image_url":{"url":"https://cdn.example.test/%d.png"},"role":"reference_image"}`, i))
	}
	_, _, err = svc.EstimateOpenAIVideoCreateBilling(context.Background(), apiKey, &User{ID: 42}, "MiniMax-H3", []byte(fmt.Sprintf(
		`{"model":"MiniMax-H3","content":[%s],"resolution":"2K","duration":5}`,
		strings.Join(content, ","),
	)))
	require.ErrorContains(t, err, "at most 9 reference images")

	content = []string{`{"type":"text","text":"scene"}`}
	for i := 0; i < miniMaxH3MaxReferenceAudios+1; i++ {
		content = append(content, fmt.Sprintf(`{"type":"audio_url","audio_url":{"url":"https://cdn.example.test/%d.mp3"},"role":"reference_audio"}`, i))
	}
	_, _, err = svc.EstimateOpenAIVideoCreateBilling(context.Background(), apiKey, &User{ID: 42}, "MiniMax-H3", []byte(fmt.Sprintf(
		`{"model":"MiniMax-H3","content":[%s],"resolution":"2K","duration":5}`,
		strings.Join(content, ","),
	)))
	require.ErrorContains(t, err, "at most 3 reference audios")
}

func TestRecordUsageMiniMaxH3VideoBillingKeepsVerifiedInputCharge(t *testing.T) {
	groupID := int64(704)
	price768P, price2K := 0.08, 0.20
	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	svc := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
	svc.resolver = newMiniMaxH3VideoPricingResolverForTest(groupID, price768P, price2K)

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID:                 "minimax-video-billing-final",
			Model:                     "MiniMax-H3",
			UpstreamModel:             "MiniMax-H3",
			MediaType:                 "video",
			VideoCount:                1,
			VideoResolution:           "2K",
			VideoDurationSeconds:      5,
			VideoInputDurationSeconds: 7.5,
			Duration:                  time.Second,
		},
		APIKey: miniMaxH3BillingAPIKey(groupID, 0.5),
		User:   &User{ID: 42},
		Account: &Account{
			ID:       73,
			Platform: PlatformOpenAI,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.InDelta(t, 12.5*price2K, usageRepo.lastLog.TotalCost, 1e-12)
	require.InDelta(t, 12.5*price2K*0.5, usageRepo.lastLog.ActualCost, 1e-12)
	require.Equal(t, string(BillingModeVideo), *usageRepo.lastLog.BillingMode)
	require.Equal(t, "2K", *usageRepo.lastLog.VideoResolution)
	require.Equal(t, 5, *usageRepo.lastLog.VideoDurationSeconds)
}

func newMiniMaxH3VideoBillingService(t *testing.T, dataDir string, groupID int64, price768P, price2K float64) *OpenAIGatewayService {
	t.Helper()
	resolver := newMiniMaxH3VideoPricingResolverForTest(groupID, price768P, price2K)
	cfg := &config.Config{}
	cfg.Pricing.DataDir = dataDir
	cfg.Server.FrontendURL = "https://api.example.test"
	billingService := NewBillingService(cfg, nil)
	resolver.billingService = billingService
	return &OpenAIGatewayService{
		cfg:            cfg,
		billingService: billingService,
		resolver:       resolver,
	}
}

func newMiniMaxH3VideoPricingResolverForTest(groupID int64, price768P, price2K float64) *ModelPricingResolver {
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: openAIVideoModelMiniMaxH3}] = &ChannelModelPricing{
		Platform:    PlatformOpenAI,
		Models:      []string{openAIVideoModelMiniMaxH3Canonical},
		BillingMode: BillingModeVideo,
		Intervals: []PricingInterval{
			{TierLabel: miniMaxH3VideoResolution768P, PerRequestPrice: &price768P},
			{TierLabel: miniMaxH3VideoResolution2K, PerRequestPrice: &price2K},
		},
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	return NewModelPricingResolver(channelService, NewBillingService(&config.Config{}, nil))
}

func miniMaxH3BillingAPIKey(groupID int64, multiplier float64) *APIKey {
	return &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:             groupID,
			RateMultiplier: multiplier,
		},
	}
}

func writeMiniMaxH3TestVideo(t *testing.T, dataDir, assetID, filename string, durationSeconds float64) {
	t.Helper()
	dir := filepath.Join(dataDir, "pg", "assets", assetID)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	file, err := os.Create(filepath.Join(dir, filename))
	require.NoError(t, err)

	ftyp := mp4.NewFtyp("isom", 0, []string{"isom", "mp42"})
	moov := mp4.NewMoovBox()
	mvhd := mp4.CreateMvhd()
	mvhd.Timescale = 1000
	mvhd.Duration = uint64(durationSeconds * 1000)
	moov.AddChild(mvhd)
	require.NoError(t, ftyp.Encode(file))
	require.NoError(t, moov.Encode(file))
	require.NoError(t, file.Close())
}
