package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIImagesAPIKeyConversionEnabledConvertsEveryURLResponse(t *testing.T) {
	imageBytes := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0x01}, 32)...)
	dataURLBytes := append([]byte("\xff\xd8\xff"), bytes.Repeat([]byte{0x02}, 16)...)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"created":1788000603,
				"data":[
					{"url":"https://cdn.example.test/generated.png","mime_type":"image/png"},
					{"url":"data:image/jpeg;base64,` + base64.StdEncoding.EncodeToString(dataURLBytes) + `"},
					{"b64_json":"YWxyZWFkeQ==","url":"https://public.example.test/image.png"}
				],
				"model":"gpt-image-2",
				"quality":"medium",
				"size":"1280x1280",
				"usage":{"input_tokens":5,"output_tokens":2094,"total_tokens":2099}
			}`)),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"image/png"}},
			Body:       io.NopCloser(bytes.NewReader(imageBytes)),
		},
	}}
	svc := &OpenAIGatewayService{
		cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
			Enabled: false,
		}}},
		httpUpstream: upstream,
	}
	body := []byte(`{"model":"gpt-image-2","prompt":"draw","response_format":"url"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	account := &Account{
		ID: 81, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://api.example.test/v1"},
		Extra:       map[string]any{OpenAIImageURLToB64JSONExtraKey: true},
	}

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 3, result.ImageCount)
	require.Equal(t, 5, result.Usage.InputTokens)
	require.Equal(t, 2094, result.Usage.OutputTokens)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, base64.StdEncoding.EncodeToString(imageBytes), gjson.Get(rec.Body.String(), "data.0.b64_json").String())
	require.Equal(t, base64.StdEncoding.EncodeToString(dataURLBytes), gjson.Get(rec.Body.String(), "data.1.b64_json").String())
	require.Equal(t, "YWxyZWFkeQ==", gjson.Get(rec.Body.String(), "data.2.b64_json").String())
	require.False(t, gjson.Get(rec.Body.String(), "data.0.url").Exists())
	require.False(t, gjson.Get(rec.Body.String(), "data.1.url").Exists())
	require.False(t, gjson.Get(rec.Body.String(), "data.2.url").Exists())
	require.NotContains(t, rec.Body.String(), "cdn.example.test")
	require.Len(t, upstream.requests, 2)
	require.Equal(t, http.MethodGet, upstream.requests[1].Method)
	require.Equal(t, "https://cdn.example.test/generated.png", upstream.requests[1].URL.String())
	require.Empty(t, upstream.requests[1].Header.Get("Authorization"))
}

func TestOpenAIImagesAPIKeyConversionDisabledPreservesBase64AndURL(t *testing.T) {
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"created":1787998906,"data":[{"b64_json":"aW1hZ2U=","url":"https://should-not-leak.example/image.png"}]}`,
		)),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	body := []byte(`{"model":"gpt-image-2","prompt":"draw"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	account := &Account{ID: 82, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test"}}

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "aW1hZ2U=", gjson.Get(rec.Body.String(), "data.0.b64_json").String())
	require.Equal(t, "https://should-not-leak.example/image.png", gjson.Get(rec.Body.String(), "data.0.url").String())
	require.Len(t, upstream.requests, 1)
}

func TestOpenAIImagesAPIKeyConversionDownloadFailureDoesNotExposeURL(t *testing.T) {
	const privateURL = "https://private-upstream.example/signed/image.png?secret=hidden"
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"url":"` + privateURL + `"}]}`)),
		},
		{
			StatusCode: http.StatusForbidden,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"error":"expired"}`)),
		},
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	body := []byte(`{"model":"gpt-image-2","prompt":"draw"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	account := &Account{
		ID: 83, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "sk-test"},
		Extra:       map[string]any{OpenAIImageURLToB64JSONExtraKey: true},
	}

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")

	require.Nil(t, result)
	require.Error(t, err)
	require.NotContains(t, err.Error(), privateURL)
	require.NotContains(t, rec.Body.String(), privateURL)
	require.False(t, c.Writer.Written())
}

func TestOpenAIImageResultDownloadValidatesEveryRedirectHost(t *testing.T) {
	imageBytes := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0x03}, 16)...)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{"https://cdn-two.example.test/final.png"}},
			Body:       io.NopCloser(strings.NewReader("")),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"image/png"}},
			Body:       io.NopCloser(bytes.NewReader(imageBytes)),
		},
	}}
	svc := &OpenAIGatewayService{
		cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
			Enabled:       true,
			UpstreamHosts: []string{"cdn-one.example.test", "cdn-two.example.test"},
		}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 85, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

	got, err := svc.downloadOpenAIImageResult(context.Background(), account, "", "https://cdn-one.example.test/start.png")

	require.NoError(t, err)
	require.Equal(t, imageBytes, got)
	require.Len(t, upstream.requests, 2)
	require.True(t, HTTPUpstreamRedirectsDisabled(upstream.requests[0].Context()))
	require.Equal(t, "https://cdn-two.example.test/final.png", upstream.requests[1].URL.String())
}

func TestOpenAIImageResultDownloadRejectsRedirectOutsideAllowlist(t *testing.T) {
	upstream := &httpUpstreamRecorder{responses: []*http.Response{{
		StatusCode: http.StatusFound,
		Header:     http.Header{"Location": []string{"https://blocked.example.test/final.png"}},
		Body:       io.NopCloser(strings.NewReader("")),
	}}}
	svc := &OpenAIGatewayService{
		cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
			Enabled:       true,
			UpstreamHosts: []string{"cdn-one.example.test"},
		}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 86, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

	_, err := svc.downloadOpenAIImageResult(context.Background(), account, "", "https://cdn-one.example.test/start.png")

	require.Error(t, err)
	require.Len(t, upstream.requests, 1)
}

func TestOpenAIImagesAPIKeyConversionEnabledAppliesToOtherImageModels(t *testing.T) {
	imageBytes := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0x04}, 16)...)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"url":"https://cdn.example.test/image.png"}]}`)),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"image/png"}},
			Body:       io.NopCloser(bytes.NewReader(imageBytes)),
		},
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	body := []byte(`{"model":"gpt-image-1","prompt":"draw","response_format":"url"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	account := &Account{
		ID: 84, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "sk-test"},
		Extra:       map[string]any{OpenAIImageURLToB64JSONExtraKey: true},
	}

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, base64.StdEncoding.EncodeToString(imageBytes), gjson.Get(rec.Body.String(), "data.0.b64_json").String())
	require.False(t, gjson.Get(rec.Body.String(), "data.0.url").Exists())
	require.Len(t, upstream.requests, 2)
}

func TestAccountShouldConvertOpenAIImageURLsToBase64(t *testing.T) {
	require.False(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}).ShouldConvertOpenAIImageURLsToBase64())
	require.False(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: map[string]any{OpenAIImageURLToB64JSONExtraKey: "true"}}).ShouldConvertOpenAIImageURLsToBase64())
	require.False(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{OpenAIImageURLToB64JSONExtraKey: true}}).ShouldConvertOpenAIImageURLsToBase64())
	require.False(t, (&Account{Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Extra: map[string]any{OpenAIImageURLToB64JSONExtraKey: true}}).ShouldConvertOpenAIImageURLsToBase64())
	require.True(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: map[string]any{OpenAIImageURLToB64JSONExtraKey: true}}).ShouldConvertOpenAIImageURLsToBase64())
}
