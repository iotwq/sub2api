package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestParseOpenAINanoBananaRequest(t *testing.T) {
	parsed, err := ParseOpenAINanoBananaRequest([]byte(`{
		"model": "nano-banana-pro",
		"prompt": "draw a poster",
		"aspectRatio": "16:9",
		"imageSize": "4K"
	}`))

	require.NoError(t, err)
	require.Equal(t, "nano-banana-pro", parsed.Model)
	require.Equal(t, "4K", parsed.ImageSize)
	require.Equal(t, 1, parsed.N)
}

func TestParseOpenAINanoBananaRequestRetainsRequestedImageCount(t *testing.T) {
	parsed, err := ParseOpenAINanoBananaRequest([]byte(`{"model":"nano-banana-pro","prompt":"draw","n":3}`))

	require.NoError(t, err)
	require.Equal(t, 3, parsed.N)
}

func TestParseOpenAINanoBananaRequestRejectsInvalidImageCount(t *testing.T) {
	for _, body := range []string{
		`{"model":"nano-banana-pro","prompt":"draw","n":0}`,
		`{"model":"nano-banana-pro","prompt":"draw","n":1.5}`,
		`{"model":"nano-banana-pro","prompt":"draw","n":"2"}`,
	} {
		_, err := ParseOpenAINanoBananaRequest([]byte(body))
		require.ErrorContains(t, err, "positive integer")
	}
}

func TestParseOpenAINanoBananaRequestRejectsWrongModel(t *testing.T) {
	_, err := ParseOpenAINanoBananaRequest([]byte(`{"model":"gpt-image-2","prompt":"draw a cat"}`))

	require.Error(t, err)
	require.Contains(t, err.Error(), "nano-banana-pro")
}

func TestBuildNanoBananaEndpointURL(t *testing.T) {
	require.Equal(t, "https://visionary.beer/v1/api/nano-banana", buildNanoBananaEndpointURL("https://visionary.beer"))
	require.Equal(t, "https://visionary.beer/v1/api/nano-banana", buildNanoBananaEndpointURL("https://visionary.beer/v1"))
	require.Equal(t, "https://visionary.beer/v1/api/nano-banana", buildNanoBananaEndpointURL("https://visionary.beer/v1/api/nano-banana"))
}

func TestPrepareNanoBananaUpstreamBodyNormalizesVisionaryFields(t *testing.T) {
	body, err := prepareNanoBananaUpstreamBody([]byte(`{
		"model": "Nano_Banana_Pro",
		"prompt": "draw a poster",
		"ratio": "9:16",
		"imageSize": "4K",
		"replyType": "json"
	}`), "Nano_Banana_Pro")

	require.NoError(t, err)
	require.Equal(t, "nano-banana-pro", gjson.GetBytes(body, "model").String())
	require.Equal(t, "9:16", gjson.GetBytes(body, "ratio").String())
	require.Equal(t, "9:16", gjson.GetBytes(body, "aspectRatio").String())
	require.Equal(t, "4K", gjson.GetBytes(body, "imageSize").String())
}

func TestPrepareNanoBananaUpstreamBodyPreservesAspectRatioAndAddsRatio(t *testing.T) {
	body, err := prepareNanoBananaUpstreamBody([]byte(`{
		"model": "nano-banana-pro",
		"prompt": "draw a poster",
		"aspectRatio": "1:1",
		"image_size": "4K"
	}`), "nano-banana-pro")

	require.NoError(t, err)
	require.Equal(t, "nano-banana-pro", gjson.GetBytes(body, "model").String())
	require.Equal(t, "1:1", gjson.GetBytes(body, "aspectRatio").String())
	require.Equal(t, "1:1", gjson.GetBytes(body, "ratio").String())
	require.Equal(t, "4K", gjson.GetBytes(body, "imageSize").String())
}

func TestExtractNanoBananaImageCountSupportsArrayObjectAndURLResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{name: "results array", body: `{"results":[{"url":"https://example.com/1.png"},{"url":"https://example.com/2.png"}]}`, want: 2},
		{name: "nested image object", body: `{"data":{"imageUrl":"https://example.com/image.png"}}`, want: 1},
		{name: "direct url", body: `{"result":"https://example.com/image.png"}`, want: 1},
		{name: "status only", body: `{"data":"success"}`, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, extractNanoBananaImageCountFromJSONBytes([]byte(tt.body)))
		})
	}
}

func TestAccountSupportsNanoBananaEndpoint(t *testing.T) {
	visionaryAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"base_url": "https://visionary.beer",
		},
	}
	require.True(t, AccountSupportsNanoBananaEndpoint(visionaryAccount, "nano-banana-pro"))

	mappedAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"base_url": "https://relay.example.com/v1",
			"model_mapping": map[string]any{
				"nano-banana-pro": "nano-banana-pro",
			},
		},
	}
	require.False(t, AccountSupportsNanoBananaEndpoint(mappedAccount, "nano-banana-pro"))

	nativePathAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"base_url": "https://relay.example.com/v1/api/nano-banana",
			"model_mapping": map[string]any{
				"nano-banana-pro": "nano-banana-pro",
			},
		},
	}
	require.True(t, AccountSupportsNanoBananaEndpoint(nativePathAccount, "nano-banana-pro"))

	genericAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"base_url": "https://api.example.com/v1",
		},
	}
	require.False(t, AccountSupportsNanoBananaEndpoint(genericAccount, "nano-banana-pro"))
}
