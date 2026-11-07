package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type sunoAudioRelayUpstream struct {
	lastRequest *http.Request
}

func (u *sunoAudioRelayUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.lastRequest = req
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"audio/mpeg"}},
		Body:       io.NopCloser(strings.NewReader("ID3 test audio")),
	}, nil
}

func (u *sunoAudioRelayUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, accountConcurrency)
}

func TestRewriteSunoAudioResponseURLsStoresAudioOnGateway(t *testing.T) {
	dataDir := t.TempDir()
	upstream := &sunoAudioRelayUpstream{}
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{Pricing: config.PricingConfig{DataDir: dataDir}},
		httpUpstream: upstream,
	}
	account := &Account{
		ID:          7,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 2,
		Extra:       map[string]any{OpenAISunoAudioURLToProxyExtraKey: true},
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "https://api.example.test/v1/chat/completions", nil)
	body := []byte(`{"choices":[{"message":{"content":"https://cdn.example.test/media/song.mp3?signature=secret"}}]}`)

	rewritten, err := svc.rewriteSunoAudioResponseURLs(c.Request.Context(), c, account, body)
	require.NoError(t, err)
	require.NotContains(t, string(rewritten), "cdn.example.test")
	require.Contains(t, string(rewritten), "https://api.example.test/pg/assets/")
	require.NotNil(t, upstream.lastRequest)
	require.True(t, HTTPUpstreamPublicHostsOnly(upstream.lastRequest.Context()))
	require.True(t, HTTPUpstreamRedirectsDisabled(upstream.lastRequest.Context()))

	matches, err := filepath.Glob(filepath.Join(dataDir, "pg", "assets", "*", "suno-output-*.mp3"))
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Regexp(t, `suno-output-[0-9a-f]{32}\.mp3$`, matches[0])
	saved, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	require.Equal(t, "ID3 test audio", string(saved))
}

func TestRewriteSunoAudioSSERewritesFinalAudioURL(t *testing.T) {
	dataDir := t.TempDir()
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{Pricing: config.PricingConfig{DataDir: dataDir}},
		httpUpstream: &sunoAudioRelayUpstream{},
	}
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: map[string]any{OpenAISunoAudioURLToProxyExtraKey: true}}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "https://api.example.test/v1/chat/completions", nil)
	input := []byte("data: {\"choices\":[{\"delta\":{\"content\":\"https://cdn.example.test/song.mp3\"}}]}\n\ndata: [DONE]\n")

	rewritten, err := svc.rewriteSunoAudioSSE(c.Request.Context(), c, account, input)
	require.NoError(t, err)
	require.NotContains(t, string(rewritten), "cdn.example.test")
	require.Contains(t, string(rewritten), "/pg/assets/")
	require.Contains(t, string(rewritten), "data: [DONE]")
}
