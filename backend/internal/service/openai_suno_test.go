package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestIsSunoMusicModel(t *testing.T) {
	require.True(t, IsSunoMusicModel("suno"))
	require.True(t, IsSunoMusicModel(" SUNO "))
	require.True(t, IsSunoMusicModel("suno-v6-mini"))
	require.True(t, IsSunoMusicModel("suno-v6"))
	require.True(t, IsSunoMusicModel("suno-v6-wild"))
	require.True(t, IsSunoV6MusicModel(" SUNO-V6 "))
	require.False(t, IsSunoMusicModel("suno-v2"))
	require.False(t, IsSunoMusicModel("gpt-5.6-sol"))
}

func TestAccountShouldProxySunoAudioURLs(t *testing.T) {
	apiKey := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: map[string]any{
		OpenAISunoAudioURLToProxyExtraKey: true,
	}}
	oauth := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{
		OpenAISunoAudioURLToProxyExtraKey: true,
	}}
	require.True(t, apiKey.ShouldProxySunoAudioURLs())
	require.False(t, oauth.ShouldProxySunoAudioURLs())
	require.False(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}).ShouldProxySunoAudioURLs())
	require.True(t, shouldRelaySunoAudioURL(apiKey, "suno-v6-mini"))
}

func TestValidateSunoChatRequest(t *testing.T) {
	require.NoError(t, ValidateSunoChatRequest([]byte(`{"model":"suno","messages":[{"role":"user","content":"Create warm piano music"}]}`)))
	require.NoError(t, ValidateSunoChatRequest([]byte(`{"model":"suno","messages":[{"role":"user","content":[{"type":"text","text":"gentle piano"}]}],"n":1,"stream_options":{"include_usage":true}}`)))
	require.NoError(t, ValidateSunoChatRequest([]byte(`{"model":"suno","messages":[{"role":"user","content":"Create a warm song"}],"lyrics":"春风经过小小街道","style_text":"中文民谣，温暖木吉他"}`)))
	require.NoError(t, ValidateSunoChatRequest([]byte(`{"model":"suno","messages":[{"role":"user","content":"music"}],"lyrics":"","style_text":""}`)))
	require.NoError(t, ValidateSunoChatRequest([]byte(`{"model":"suno","messages":[{"role":"user","content":"music"}],"lyrics":"`+strings.Repeat("中", sunoMaxTextCharacters)+`"}`)))
	require.NoError(t, ValidateSunoChatRequest([]byte(`{"model":"suno-v6","messages":[{"role":"user","content":"Gentle piano"}],"instrumental":true,"n":1}`)))

	tests := []struct {
		name string
		body string
		want string
	}{
		{"missing messages", `{"model":"suno"}`, "1 to 32"},
		{"image content", `{"model":"suno","messages":[{"role":"user","content":[{"type":"image_url","image_url":{"url":"https://example.com/image.png"}}]}]}`, "plain text"},
		{"too long prompt", `{"model":"suno","messages":[{"role":"user","content":"` + strings.Repeat("a", sunoMaxTextCharacters+1) + `"}]}`, "520000"},
		{"v6 instrumental type", `{"model":"suno-v6","messages":[{"role":"user","content":"music"}],"instrumental":"true"}`, "instrumental"},
		{"legacy instrumental", `{"model":"suno","messages":[{"role":"user","content":"music"}],"instrumental":true}`, "instrumental"},
		{"v6 prompt too long", `{"model":"suno-v6","messages":[{"role":"user","content":"` + strings.Repeat("中", sunoV6MaxPromptUTF16Units+1) + `"}]}`, "3000"},
		{"too long lyrics", `{"model":"suno","messages":[{"role":"user","content":"music"}],"lyrics":"` + strings.Repeat("a", sunoMaxTextCharacters+1) + `"}`, "lyrics"},
		{"too long style", `{"model":"suno","messages":[{"role":"user","content":"music"}],"style_text":"` + strings.Repeat("a", sunoMaxTextCharacters+1) + `"}`, "style_text"},
		{"lyrics null", `{"model":"suno","messages":[{"role":"user","content":"music"}],"lyrics":null}`, "lyrics"},
		{"style not string", `{"model":"suno","messages":[{"role":"user","content":"music"}],"style_text":[]}`, "style_text"},
		{"v6 lyrics unsupported", `{"model":"suno-v6","messages":[{"role":"user","content":"music"}],"lyrics":"words"}`, "lyrics"},
		{"n two", `{"model":"suno","messages":[{"role":"user","content":"music"}],"n":2}`, "n=1"},
		{"unsupported field", `{"model":"suno","messages":[{"role":"user","content":"music"}],"duration":10}`, "duration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSunoChatRequest([]byte(tt.body))
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.want)
		})
	}
	require.NoError(t, ValidateSunoChatRequest([]byte(`{"model":"gpt-5.6-sol","messages":[{"role":"user","content":[{"type":"image_url"}]}]}`)))
}

func TestSunoMusicCostAndResultMarker(t *testing.T) {
	cost := sunoMusicCost(1.5)
	require.Equal(t, string(BillingModePerRequest), cost.BillingMode)
	require.InDelta(t, 0.08, cost.TotalCost, 1e-12)
	require.InDelta(t, 0.12, cost.ActualCost, 1e-12)

	result := &OpenAIForwardResult{Model: "suno"}
	markSunoForwardResult(result)
	require.Equal(t, 1, result.RequestCount)
	require.Equal(t, "audio", result.MediaType)

	v6Result := &OpenAIForwardResult{Model: "suno-v6-wild"}
	markSunoForwardResult(v6Result)
	require.Equal(t, 1, v6Result.RequestCount)
	require.Equal(t, "audio", v6Result.MediaType)
}

func TestSunoChannelPricingOverridesDefaultCost(t *testing.T) {
	price := 0.9
	group := &Group{ID: 42, Platform: PlatformOpenAI, RateMultiplier: 1, ModelPricing: []ChannelModelPricing{{
		Models: []string{"suno", "suno-v6"}, BillingMode: BillingModePerRequest, PerRequestPrice: &price,
	}}}
	cfg := &config.Config{}
	billing := NewBillingService(cfg, nil)
	svc := &OpenAIGatewayService{
		billingService: billing,
		resolver:       NewModelPricingResolver(nil, billing),
	}
	apiKey := &APIKey{User: &User{ID: 1}, GroupID: &group.ID, Group: group}
	cost, err := svc.calculateSunoMusicCost(context.Background(), "suno", apiKey, 1, 1)
	require.NoError(t, err)
	require.InDelta(t, 0.9, cost.TotalCost, 1e-12)
	require.InDelta(t, 0.9, cost.ActualCost, 1e-12)
	v6Cost, err := svc.calculateSunoMusicCost(context.Background(), "suno-v6", apiKey, 1, 1)
	require.NoError(t, err)
	require.InDelta(t, 0.9, v6Cost.TotalCost, 1e-12)

	estimate, priced := (inflightEstimateDeps{cfg: cfg, resolver: svc.resolver}).estimate(
		context.Background(), apiKey, InflightEstimateRequest{Model: "suno"},
	)
	require.True(t, priced)
	require.InDelta(t, 0.9, estimate, 1e-12)
}

func TestSunoRawHeaderAllowsIdempotencyKey(t *testing.T) {
	require.True(t, openaiCCRawAllowedHeaders["idempotency-key"])
}
