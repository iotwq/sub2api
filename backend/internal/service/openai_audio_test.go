package service

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestForwardOpenAIAudioSpeechStreamsBinaryResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/audio/speech", strings.NewReader(`{"model":"tts-1","input":"hello","voice":"alloy"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"audio/mpeg"}, "X-Request-Id": []string{"audio-1"}}, Body: io.NopCloser(strings.NewReader("mp3-bytes"))}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	account := &Account{ID: 7, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-audio", "base_url": "https://audio.example"}}

	result, err := svc.ForwardOpenAIAudio(context.Background(), c, account, OpenAIAudioEndpointSpeech, []byte(`{"model":"tts-1","input":"hello","voice":"alloy"}`), "application/json", "tts-1")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "mp3-bytes", recorder.Body.String())
	require.Equal(t, "https://audio.example/v1/audio/speech", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer sk-audio", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "audio-1", result.RequestID)
	require.Equal(t, "/v1/audio/speech", result.UpstreamEndpoint)
}

func TestForwardOpenAIAudioTranscriptionPreservesMultipartAndUsage(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "whisper-1"))
	file, err := writer.CreateFormFile("file", "sample.mp3")
	require.NoError(t, err)
	_, err = file.Write([]byte("audio"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", bytes.NewReader(body.Bytes()))
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"text":"hello","usage":{"input_tokens":4,"output_tokens":2}}`))}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	account := &Account{ID: 8, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-audio", "base_url": "https://audio.example"}}

	result, err := svc.ForwardOpenAIAudio(context.Background(), c, account, OpenAIAudioEndpointTranscriptions, body.Bytes(), writer.FormDataContentType(), "whisper-1")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(4), gjson.GetBytes(result.ResponseBody, "usage.input_tokens").Int())
	require.Equal(t, 4, result.Usage.InputTokens)
	require.Contains(t, string(upstream.lastBody), "name=\"model\"")
	require.Contains(t, string(upstream.lastBody), "whisper-1")
	require.Contains(t, upstream.lastReq.Header.Get("Content-Type"), "multipart/form-data")
}
