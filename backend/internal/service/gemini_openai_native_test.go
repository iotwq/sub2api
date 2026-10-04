//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newOpenAIGeminiNativeTestService(upstream HTTPUpstream) *GeminiMessagesCompatService {
	return &GeminiMessagesCompatService{
		httpUpstream: upstream,
		cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
			Enabled:           false,
			AllowInsecureHTTP: true,
		}}},
	}
}

func openAIGeminiNativeTestAccount(baseURL string, enabled bool) *Account {
	capabilities := []any{"chat_completions", "embeddings"}
	if enabled {
		capabilities = append(capabilities, "gemini_native")
	}
	return &Account{
		ID:          301,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":             "sk-openai-native",
			"base_url":            baseURL,
			"openai_capabilities": capabilities,
			"model_mapping": map[string]any{
				"gemini-3-pro-image-preview": "gemini-3-pro-image-preview",
			},
		},
	}
}

func TestOpenAIGeminiNativeCapabilityDefaultsOff(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	require.False(t, account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityGeminiNative))

	account.Credentials = map[string]any{"openai_capabilities": []any{"gemini_native"}}
	require.True(t, account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityGeminiNative))

	account.Type = AccountTypeOAuth
	require.False(t, account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityGeminiNative))
}

func TestForwardNativeOpenAIUsesBearerAndPreservesBody(t *testing.T) {
	body := []byte(`{"contents":[{"role":"user","parts":[{"text":"hello"},{"functionCall":{"name":"lookup","args":{}}}]}],"generationConfig":{"responseModalities":["TEXT","IMAGE"]}}`)
	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"candidates":[],"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":1}}`)),
	}}
	svc := newOpenAIGeminiNativeTestService(upstream)
	account := openAIGeminiNativeTestAccount("http://upstream.example/v1", true)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3-pro-image-preview:generateContent", bytes.NewReader(body))

	result, err := svc.ForwardNative(context.Background(), c, account, "gemini-3-pro-image-preview", "generateContent", false, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "http://upstream.example/v1beta/models/gemini-3-pro-image-preview:generateContent", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer sk-openai-native", upstream.lastReq.Header.Get("Authorization"))
	require.Empty(t, upstream.lastReq.Header.Get("x-goog-api-key"))
	sentBody, err := io.ReadAll(upstream.lastReq.Body)
	require.NoError(t, err)
	require.Equal(t, body, sentBody)
}

func TestForwardNativeOpenAIStreamPreservesSSE(t *testing.T) {
	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hello\"}]}}]}\n\n",
		)),
	}}
	svc := newOpenAIGeminiNativeTestService(upstream)
	account := openAIGeminiNativeTestAccount("http://upstream.example", true)
	body := []byte(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3-pro-image-preview:streamGenerateContent?alt=sse", bytes.NewReader(body))

	result, err := svc.ForwardNative(context.Background(), c, account, "gemini-3-pro-image-preview", "streamGenerateContent", true, body)
	require.NoError(t, err)
	require.True(t, result.Stream)
	require.Equal(t, "http://upstream.example/v1beta/models/gemini-3-pro-image-preview:streamGenerateContent?alt=sse", upstream.lastReq.URL.String())
	require.Equal(t, "text/event-stream", strings.Split(recorder.Header().Get("Content-Type"), ";")[0])
	require.Contains(t, recorder.Body.String(), `data: {"candidates"`)
}

func TestForwardAIStudioGETOpenAIUsesBearer(t *testing.T) {
	upstream := &geminiCompatHTTPUpstreamStub{response: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"models":[]}`)),
	}}
	svc := newOpenAIGeminiNativeTestService(upstream)
	account := openAIGeminiNativeTestAccount("http://upstream.example/v1", true)

	result, err := svc.ForwardAIStudioGET(context.Background(), account, "/v1beta/models")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.Equal(t, "http://upstream.example/v1beta/models", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer sk-openai-native", upstream.lastReq.Header.Get("Authorization"))
	require.Empty(t, upstream.lastReq.Header.Get("x-goog-api-key"))
}

func TestForwardNativeOpenAIRejectsAccountWithoutCapability(t *testing.T) {
	svc := newOpenAIGeminiNativeTestService(&geminiCompatHTTPUpstreamStub{})
	account := openAIGeminiNativeTestAccount("http://upstream.example", false)
	body := []byte(`{"contents":[{"parts":[{"text":"hello"}]}]}`)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3-pro-image-preview:generateContent", bytes.NewReader(body))

	result, err := svc.ForwardNative(context.Background(), c, account, "gemini-3-pro-image-preview", "generateContent", false, body)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, http.StatusBadGateway, recorder.Code)
	require.Contains(t, recorder.Body.String(), "not configured for Gemini native passthrough")
}

func TestForwardAIStudioGETOpenAIRejectsAccountWithoutCapability(t *testing.T) {
	svc := newOpenAIGeminiNativeTestService(&geminiCompatHTTPUpstreamStub{})
	account := openAIGeminiNativeTestAccount("http://upstream.example", false)

	result, err := svc.ForwardAIStudioGET(context.Background(), account, "/v1beta/models")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "not configured for Gemini native passthrough")
}
