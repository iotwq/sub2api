package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
)

const (
	openAIImageResultDownloadTimeout = 60 * time.Second
	openAIImageResultMaxRedirects    = 5
)

func (s *OpenAIGatewayService) normalizeOpenAIImageResponseURLs(
	ctx context.Context,
	account *Account,
	proxyURL string,
	body []byte,
) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		return nil, errors.New("invalid upstream image response")
	}
	data, ok := payload["data"].([]any)
	if !ok {
		return body, nil
	}

	changed := false
	for _, rawItem := range data {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		if encoded, ok := item["b64_json"].(string); ok && strings.TrimSpace(encoded) != "" {
			if _, exists := item["url"]; exists {
				delete(item, "url")
				changed = true
			}
			continue
		}

		rawURL, exists := item["url"]
		if !exists {
			continue
		}
		imageURL, ok := rawURL.(string)
		if !ok || strings.TrimSpace(imageURL) == "" {
			return nil, errors.New("upstream returned an invalid image URL")
		}
		imageBytes, err := s.downloadOpenAIImageResult(ctx, account, proxyURL, imageURL)
		if err != nil {
			return nil, errors.New("failed to retrieve upstream image")
		}
		item["b64_json"] = base64.StdEncoding.EncodeToString(imageBytes)
		delete(item, "url")
		changed = true
	}
	if !changed {
		return body, nil
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("failed to normalize upstream image response")
	}
	return normalized, nil
}

func (s *OpenAIGatewayService) downloadOpenAIImageResult(
	ctx context.Context,
	account *Account,
	proxyURL string,
	rawURL string,
) ([]byte, error) {
	if imageBytes, ok := decodeOpenAIImageDataURL(rawURL); ok {
		return imageBytes, nil
	}
	if s == nil || s.httpUpstream == nil || account == nil {
		return nil, errors.New("image downloader is unavailable")
	}

	downloadCtx, cancel := context.WithTimeout(ctx, openAIImageResultDownloadTimeout)
	defer cancel()
	currentURL := strings.TrimSpace(rawURL)
	for redirects := 0; redirects <= openAIImageResultMaxRedirects; redirects++ {
		validatedURL, err := s.validateOpenAIImageResultURL(currentURL)
		if err != nil {
			return nil, errors.New("image URL is not allowed")
		}
		requestCtx := WithHTTPUpstreamRedirectsDisabled(downloadCtx)
		req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, validatedURL, nil)
		if err != nil {
			return nil, errors.New("failed to build image download request")
		}
		req.Header.Set("Accept", "image/png,image/jpeg,image/webp,image/*;q=0.8")
		resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
		if err != nil {
			return nil, errors.New("image download request failed")
		}
		if resp == nil {
			return nil, errors.New("image download returned no response")
		}
		if isOpenAIImageDownloadRedirect(resp.StatusCode) {
			location := strings.TrimSpace(resp.Header.Get("Location"))
			closeOpenAIImageResultBody(resp)
			if location == "" || redirects == openAIImageResultMaxRedirects {
				return nil, errors.New("image download redirect is invalid")
			}
			nextURL, err := resolveOpenAIImageDownloadRedirect(validatedURL, location)
			if err != nil {
				return nil, errors.New("image download redirect is invalid")
			}
			currentURL = nextURL
			continue
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			closeOpenAIImageResultBody(resp)
			return nil, errors.New("image download returned an error")
		}
		imageBytes, err := readOpenAIImageResultBytes(resp)
		closeOpenAIImageResultBody(resp)
		return imageBytes, err
	}
	return nil, errors.New("too many image download redirects")
}

func closeOpenAIImageResultBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func (s *OpenAIGatewayService) validateOpenAIImageResultURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.User != nil {
		return "", errors.New("invalid image URL")
	}
	if s != nil && s.cfg != nil && !s.cfg.Security.URLAllowlist.Enabled {
		if _, err := urlvalidator.ValidateURLFormat(trimmed, s.cfg.Security.URLAllowlist.AllowInsecureHTTP); err != nil {
			return "", err
		}
		return trimmed, nil
	}
	if s == nil || s.cfg == nil {
		if _, err := urlvalidator.ValidateURLFormat(trimmed, false); err != nil {
			return "", err
		}
		return trimmed, nil
	}
	if _, err := urlvalidator.ValidateHTTPSURL(trimmed, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	}); err != nil {
		return "", err
	}
	return trimmed, nil
}

func decodeOpenAIImageDataURL(raw string) ([]byte, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(strings.ToLower(trimmed), "data:image/") {
		return nil, false
	}
	metadata, encoded, ok := strings.Cut(trimmed, ",")
	if !ok || !strings.Contains(strings.ToLower(metadata), ";base64") {
		return nil, false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(decoded) == 0 || int64(len(decoded)) > openAIImageMaxDownloadBytes {
		return nil, false
	}
	return decoded, true
}

func readOpenAIImageResultBytes(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, errors.New("image response body is empty")
	}
	imageBytes, err := io.ReadAll(io.LimitReader(resp.Body, openAIImageMaxDownloadBytes+1))
	if err != nil {
		return nil, errors.New("failed to read image response")
	}
	if len(imageBytes) == 0 || int64(len(imageBytes)) > openAIImageMaxDownloadBytes {
		return nil, errors.New("image response size is invalid")
	}
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
		contentType = mediaType
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		contentType = http.DetectContentType(imageBytes)
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return nil, errors.New("downloaded content is not an image")
	}
	return imageBytes, nil
}

func isOpenAIImageDownloadRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

func resolveOpenAIImageDownloadRedirect(currentURL string, location string) (string, error) {
	base, err := url.Parse(currentURL)
	if err != nil {
		return "", err
	}
	next, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(next).String(), nil
}
