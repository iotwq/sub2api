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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiForwardAsChatCompletions_PreservesActualImageCount(t *testing.T) {
	for _, tc := range []struct {
		name  string
		parts string
		want  int
	}{
		{"no output image", `{"text":"No image generated"}`, 0},
		{"multiple output images", `{"inlineData":{"mimeType":"image/png","data":"aGVsbG8="}},{"inlineData":{"mimeType":"image/png","data":"d29ybGQ="}}`, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			reqBody := []byte(`{"model":"gemini-3-pro-image","messages":[{"role":"user","content":"draw two cats"}]}`)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(reqBody))
			svc, upstream := newGeminiVertexAPIKeyTestService()
			upstream.response.Body = io.NopCloser(strings.NewReader(`{"candidates":[{"content":{"parts":[` + tc.parts + `]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":5}}`))

			result, err := svc.ForwardAsChatCompletions(context.Background(), c, geminiVertexAPIKeyTestAccount(""), reqBody)
			require.NoError(t, err)
			require.Equal(t, tc.want, result.ImageCount)
		})
	}
}

func TestGeminiForwardNative_ImageOnlyDisconnectRetainsUniqueImages(t *testing.T) {
	// Two distinct image chunks followed by a repeated chunk, without usageMetadata.
	first := `data: {"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"aGVsbG8="}}]}}]}` + "\n\n"
	second := strings.ReplaceAll(first, "aGVsbG8=", "d29ybGQ=")
	reqBody := []byte(`{"contents":[{"role":"user","parts":[{"text":"draw two cats"}]}]}`)
	svc, c, _ := newGeminiAbortingStreamService(t, "/v1beta/models/gemini-3-pro-image:streamGenerateContent", reqBody, first+second+first, true)

	result, err := svc.ForwardNative(context.Background(), c, geminiVertexAPIKeyTestAccount(""), "gemini-3-pro-image", "streamGenerateContent", true, reqBody)
	require.NoError(t, err)
	require.True(t, result.ClientDisconnect)
	require.Equal(t, 2, result.ImageCount)
	require.Zero(t, result.Usage.InputTokens)
	require.True(t, ForwardResultHasBillableUsage(result))
}

func TestGeminiCompatForward_CompletedImageDisconnectRetainsMultipleImages(t *testing.T) {
	first := `data: {"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"aGVsbG8="}}]}}]}` + "\n\n"
	second := strings.ReplaceAll(first, "aGVsbG8=", "d29ybGQ=")
	final := `data: {"candidates":[{"finishReason":"STOP"}]}` + "\n\n"
	reqBody := []byte(`{"model":"gemini-3-pro-image","max_tokens":16,"stream":true,"messages":[{"role":"user","content":"draw two cats"}]}`)
	svc, c, _ := newGeminiAbortingStreamService(t, "/v1/messages", reqBody, first+second+first+final, true)

	result, err := svc.Forward(context.Background(), c, geminiVertexAPIKeyTestAccount(""), reqBody)
	require.NoError(t, err)
	require.True(t, result.ClientDisconnect)
	require.Equal(t, 2, result.ImageCount)
}

func TestGeminiForwardAsChatCompletions_ClientAbortIgnoresUnterminatedUsage(t *testing.T) {
	trailing := `data: {"candidates":[{"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1000,"candidatesTokenCount":300}}`
	reqBody := []byte(`{"model":"gemini-3-flash-preview","stream":true,"messages":[{"role":"user","content":"hello"}]}`)
	svc, c, _ := newGeminiAbortingStreamService(t, "/v1/chat/completions", reqBody, geminiAbortUsageChunk+trailing, true)

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, geminiVertexAPIKeyTestAccount(""), reqBody)
	require.NoError(t, err)
	require.True(t, result.ClientDisconnect)
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, 2, result.Usage.OutputTokens)
}
