package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	sunoAudioRelayTimeout            = 60 * time.Second
	sunoAudioRelayMaxBytes     int64 = 15 << 20
	sunoAudioRelayMaxRedirects       = 5
)

var sunoAudioURLPattern = regexp.MustCompile(`https?://[^\s"'<>]+`)

func shouldRelaySunoAudioURL(account *Account, models ...string) bool {
	if account == nil || !account.ShouldProxySunoAudioURLs() {
		return false
	}
	for _, model := range models {
		if IsSunoMusicModel(model) {
			return true
		}
	}
	return false
}

// rewriteSunoAudioResponseURLs rewrites audio URLs in a Chat Completions JSON
// response. It is intentionally limited to content/url fields so unrelated
// metadata links are not downloaded or changed.
func (s *OpenAIGatewayService) rewriteSunoAudioResponseURLs(ctx context.Context, c *gin.Context, account *Account, body []byte) ([]byte, error) {
	var payload any
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&payload); err != nil {
		return nil, errors.New("invalid Suno upstream response")
	}
	cache := make(map[string]string)
	if err := s.rewriteSunoAudioValue(ctx, c, account, &payload, cache); err != nil {
		return nil, err
	}
	return json.Marshal(payload)
}

func (s *OpenAIGatewayService) rewriteSunoAudioValue(ctx context.Context, c *gin.Context, account *Account, value *any, cache map[string]string) error {
	switch typed := (*value).(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "content" || key == "url" || key == "audio_url" {
				if text, ok := child.(string); ok {
					rewritten, err := s.rewriteSunoAudioText(ctx, c, account, text, cache)
					if err != nil {
						return err
					}
					typed[key] = rewritten
					continue
				}
			}
			if err := s.rewriteSunoAudioValue(ctx, c, account, &child, cache); err != nil {
				return err
			}
			typed[key] = child
		}
	case []any:
		for index := range typed {
			if err := s.rewriteSunoAudioValue(ctx, c, account, &typed[index], cache); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *OpenAIGatewayService) rewriteSunoAudioText(ctx context.Context, c *gin.Context, account *Account, text string, cache map[string]string) (string, error) {
	var firstErr error
	return sunoAudioURLPattern.ReplaceAllStringFunc(text, func(raw string) string {
		if firstErr != nil {
			return raw
		}
		trimmed := strings.TrimRight(raw, ".,;:!?)]}")
		trailing := raw[len(trimmed):]
		if replacement, ok := cache[trimmed]; ok {
			return replacement + trailing
		}
		replacement, err := s.storeSunoAudio(ctx, c, account, trimmed)
		if err != nil {
			firstErr = err
			return raw
		}
		cache[trimmed] = replacement
		return replacement + trailing
	}), firstErr
}

func (s *OpenAIGatewayService) storeSunoAudio(ctx context.Context, c *gin.Context, account *Account, rawURL string) (string, error) {
	if s == nil || s.httpUpstream == nil || account == nil {
		return "", errors.New("Suno audio relay is unavailable")
	}
	currentURL, err := s.validateOpenAIImageResultURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid Suno audio URL: %w", err)
	}
	data, contentType, err := s.downloadSunoAudio(ctx, account, currentURL)
	if err != nil {
		return "", err
	}

	dataDir := ""
	if s.cfg != nil {
		dataDir = strings.TrimSpace(s.cfg.Pricing.DataDir)
	}
	if dataDir == "" {
		return "", errors.New("Suno audio relay data directory is not configured")
	}
	assetID, err := randomSunoAudioAssetID()
	if err != nil {
		return "", fmt.Errorf("allocate Suno audio asset: %w", err)
	}
	filename := sunoAudioFilename(contentType, assetID)
	assetDir := filepath.Join(dataDir, "pg", "assets", assetID)
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return "", fmt.Errorf("create Suno audio asset directory: %w", err)
	}
	path := filepath.Join(assetDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		_ = os.RemoveAll(assetDir)
		return "", fmt.Errorf("save Suno audio asset: %w", err)
	}

	return sunoAudioPublicURL(c, assetID, filename), nil
}

func (s *OpenAIGatewayService) downloadSunoAudio(ctx context.Context, account *Account, currentURL string) ([]byte, string, error) {
	downloadCtx, cancel := context.WithTimeout(WithHTTPUpstreamPublicHostsOnly(ctx), sunoAudioRelayTimeout)
	defer cancel()
	proxyURL := resolveAccountProxyURL(account)

	for redirects := 0; redirects <= sunoAudioRelayMaxRedirects; redirects++ {
		validatedURL, err := s.validateOpenAIImageResultURL(currentURL)
		if err != nil {
			return nil, "", errors.New("Suno audio URL is not allowed")
		}
		req, err := http.NewRequestWithContext(WithHTTPUpstreamRedirectsDisabled(downloadCtx), http.MethodGet, validatedURL, nil)
		if err != nil {
			return nil, "", errors.New("failed to build Suno audio download request")
		}
		req.Header.Set("Accept", "audio/*,*/*;q=0.8")
		resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
		if err != nil {
			return nil, "", fmt.Errorf("download Suno audio: %w", err)
		}
		if resp == nil || resp.Body == nil {
			return nil, "", errors.New("Suno audio download returned no response")
		}
		if isOpenAIImageDownloadRedirect(resp.StatusCode) {
			location := strings.TrimSpace(resp.Header.Get("Location"))
			_ = resp.Body.Close()
			if location == "" || redirects == sunoAudioRelayMaxRedirects {
				return nil, "", errors.New("Suno audio download redirect is invalid")
			}
			next, resolveErr := resolveOpenAIImageDownloadRedirect(validatedURL, location)
			if resolveErr != nil {
				return nil, "", errors.New("Suno audio download redirect is invalid")
			}
			currentURL = next
			continue
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			_ = resp.Body.Close()
			return nil, "", fmt.Errorf("Suno audio download returned status %d", resp.StatusCode)
		}
		if resp.ContentLength > sunoAudioRelayMaxBytes {
			_ = resp.Body.Close()
			return nil, "", errors.New("Suno audio file is too large")
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, sunoAudioRelayMaxBytes+1))
		contentType := resp.Header.Get("Content-Type")
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, "", fmt.Errorf("read Suno audio: %w", readErr)
		}
		if int64(len(data)) == 0 || int64(len(data)) > sunoAudioRelayMaxBytes {
			return nil, "", errors.New("Suno audio file size is invalid")
		}
		contentType = normalizeSunoAudioContentType(contentType, validatedURL, data)
		if contentType == "" {
			return nil, "", errors.New("Suno response is not an allowed audio format")
		}
		return data, contentType, nil
	}
	return nil, "", errors.New("too many Suno audio redirects")
}

func normalizeSunoAudioContentType(header, rawURL string, data []byte) string {
	contentType, _, err := mime.ParseMediaType(strings.TrimSpace(header))
	if err != nil {
		contentType = ""
	}
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if !isAllowedSunoAudioContentType(contentType) {
		contentType = ""
	}
	if contentType == "" {
		ext := strings.ToLower(filepath.Ext(strings.Split(rawURL, "?")[0]))
		contentType = map[string]string{".mp3": "audio/mpeg", ".m4a": "audio/mp4", ".wav": "audio/wav", ".aac": "audio/aac", ".ogg": "audio/ogg"}[ext]
	}
	if contentType == "" {
		detected := http.DetectContentType(data)
		if isAllowedSunoAudioContentType(detected) {
			contentType = detected
		}
	}
	if !isAllowedSunoAudioContentType(contentType) {
		return ""
	}
	return contentType
}

func isAllowedSunoAudioContentType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "audio/mpeg", "audio/mp4", "audio/wav", "audio/x-wav", "audio/aac", "audio/ogg", "audio/webm":
		return true
	default:
		return false
	}
}

func sunoAudioFilename(contentType, assetID string) string {
	prefix := "suno-output-" + assetID
	switch strings.ToLower(contentType) {
	case "audio/mp4":
		return prefix + ".m4a"
	case "audio/wav", "audio/x-wav":
		return prefix + ".wav"
	case "audio/aac":
		return prefix + ".aac"
	case "audio/ogg":
		return prefix + ".ogg"
	case "audio/webm":
		return prefix + ".webm"
	default:
		return prefix + ".mp3"
	}
}

func randomSunoAudioAssetID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func sunoAudioPublicURL(c *gin.Context, assetID, filename string) string {
	scheme := "http"
	if c != nil {
		if sunoRequestIsHTTPS(c) || strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https") {
			scheme = "https"
		}
		host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
		if host == "" && c.Request != nil {
			host = strings.TrimSpace(c.Request.Host)
		}
		return scheme + "://" + host + "/pg/assets/" + assetID + "/" + filename
	}
	return "/pg/assets/" + assetID + "/" + filename
}

func sunoRequestIsHTTPS(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	return c.Request.TLS != nil || strings.EqualFold(strings.TrimSpace(c.Request.URL.Scheme), "https")
}

// rewriteSunoAudioSSE rewrites complete SSE data lines before they are exposed
// to the client. Suno emits the final result only after generation completes,
// so buffering this dedicated stream does not hide intermediate progress.
func (s *OpenAIGatewayService) rewriteSunoAudioSSE(ctx context.Context, c *gin.Context, account *Account, body []byte) ([]byte, error) {
	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 4096), 4<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload != "" && payload != "[DONE]" && json.Valid([]byte(payload)) {
				rewritten, err := s.rewriteSunoAudioResponseURLs(ctx, c, account, []byte(payload))
				if err != nil {
					return nil, err
				}
				line = "data: " + string(rewritten)
			}
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read Suno stream: %w", err)
	}
	return out.Bytes(), nil
}
