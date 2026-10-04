package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
)

type OpenAIAudioEndpoint string

const (
	OpenAIAudioEndpointSpeech         OpenAIAudioEndpoint = "speech"
	OpenAIAudioEndpointTranscriptions OpenAIAudioEndpoint = "transcriptions"
	OpenAIAudioEndpointTranslations   OpenAIAudioEndpoint = "translations"
)

func (e OpenAIAudioEndpoint) path() string {
	switch e {
	case OpenAIAudioEndpointSpeech:
		return "/v1/audio/speech"
	case OpenAIAudioEndpointTranscriptions:
		return "/v1/audio/transcriptions"
	case OpenAIAudioEndpointTranslations:
		return "/v1/audio/translations"
	default:
		return ""
	}
}

func (e OpenAIAudioEndpoint) Path() string { return e.path() }

func (s *OpenAIGatewayService) ForwardOpenAIAudio(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	endpoint OpenAIAudioEndpoint,
	body []byte,
	contentType string,
	requestModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()
	if account == nil || !account.IsOpenAIApiKey() {
		return nil, fmt.Errorf("audio endpoint requires an OpenAI APIKey account")
	}
	upstreamModel := account.GetMappedModel(requestModel)
	path := endpoint.path()
	if path == "" {
		return nil, fmt.Errorf("unsupported audio endpoint")
	}
	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}
	baseURL := account.GetOpenAIBaseURL()
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com"
	}
	validatedURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base_url: %w", err)
	}
	targetURL := buildOpenAIEndpointURL(validatedURL, path)
	upstreamCtx, releaseUpstreamCtx := detachUpstreamContext(ctx)
	upstreamReq, err := http.NewRequestWithContext(upstreamCtx, http.MethodPost, targetURL, bytes.NewReader(body))
	releaseUpstreamCtx()
	if err != nil {
		return nil, fmt.Errorf("build upstream request: %w", err)
	}
	upstreamReq = upstreamReq.WithContext(WithHTTPUpstreamProfile(upstreamReq.Context(), HTTPUpstreamProfileOpenAI))
	upstreamReq.Header.Set("Authorization", "Bearer "+token)
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/json"
	}
	upstreamReq.Header.Set("Content-Type", contentType)
	upstreamReq.Header.Set("Accept", "application/json")
	if endpoint == OpenAIAudioEndpointSpeech {
		upstreamReq.Header.Set("Accept", "audio/mpeg, application/json")
	}
	for key, values := range c.Request.Header {
		lowerKey := strings.ToLower(key)
		if openaiCCRawAllowedHeaders[lowerKey] && lowerKey != "content-length" && lowerKey != "authorization" {
			upstreamReq.Header.Del(key)
			for _, value := range values {
				upstreamReq.Header.Add(key, value)
			}
		}
	}
	upstreamReq.Header.Set("Authorization", "Bearer "+token)
	if customUA := account.GetOpenAIUserAgent(); customUA != "" {
		upstreamReq.Header.Set("User-Agent", customUA)
	}
	account.ApplyHeaderOverrides(upstreamReq.Header)
	proxyURL := ""
	if account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return nil, s.handleOpenAIUpstreamTransportError(ctx, c, account, err, false)
	}
	if resp.StatusCode >= 400 {
		respBody := s.readUpstreamErrorBody(resp)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		upstreamMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(respBody)))
		if failoverErr := s.failoverOpenAIUpstreamHTTPError(ctx, c, account, resp, respBody, upstreamMsg, upstreamModel); failoverErr != nil {
			return nil, failoverErr
		}
		writeOpenAIAudioResponse(c, resp, respBody, s.responseHeaderFilter)
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	result := &OpenAIForwardResult{
		RequestID:        firstNonEmptyString(resp.Header.Get("x-request-id"), resp.Header.Get("request-id")),
		Model:            requestModel,
		UpstreamModel:    upstreamModel,
		UpstreamEndpoint: path,
		ResponseHeaders:  resp.Header.Clone(),
		Duration:         time.Since(startTime),
		RequestCount:     1,
		MediaType:        "audio",
	}
	if endpoint == OpenAIAudioEndpointSpeech && !strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json") {
		if err := streamOpenAIAudioResponse(resp, c, s.responseHeaderFilter); err != nil {
			return nil, err
		}
		return result, nil
	}
	respBody, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	if err != nil {
		return nil, fmt.Errorf("read upstream body: %w", err)
	}
	writeOpenAIAudioResponse(c, resp, respBody, s.responseHeaderFilter)
	result.ResponseBody = append([]byte(nil), respBody...)
	result.ResponseID = extractOpenAIResponseIDFromJSONBytes(respBody)
	if usage, ok := extractOpenAIUsageFromJSONBytes(respBody); ok {
		result.Usage = usage
	}
	return result, nil
}

func streamOpenAIAudioResponse(resp *http.Response, c *gin.Context, filter *responseheaders.CompiledHeaderFilter) error {
	if resp == nil || c == nil {
		return fmt.Errorf("audio response context is nil")
	}
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, filter)
	if c.Writer.Header().Get("Content-Type") == "" {
		c.Writer.Header().Set("Content-Type", "audio/mpeg")
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err := io.Copy(c.Writer, resp.Body)
	return err
}

func writeOpenAIAudioResponse(c *gin.Context, resp *http.Response, body []byte, filter *responseheaders.CompiledHeaderFilter) {
	if c == nil || resp == nil || c.Writer.Written() {
		return
	}
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, filter)
	if c.Writer.Header().Get("Content-Type") == "" {
		c.Writer.Header().Set("Content-Type", "application/json")
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(body)
}
