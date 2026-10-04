package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func extractMiniMaxVideoPrompt(body []byte) string {
	content := gjson.GetBytes(body, "content")
	if !content.IsArray() {
		return ""
	}
	for _, item := range content.Array() {
		if strings.EqualFold(strings.TrimSpace(item.Get("type").String()), "text") {
			if prompt := strings.TrimSpace(item.Get("text").String()); prompt != "" {
				return prompt
			}
		}
	}
	return ""
}

func miniMaxVideoLegacyMediaFieldsPresent(body []byte) bool {
	for _, field := range []string{"image", "images", "videos", "audios"} {
		if value := gjson.GetBytes(body, field); value.Exists() {
			return true
		}
	}
	return false
}

func miniMaxVideoHasBaseVideo(body []byte) bool {
	content := gjson.GetBytes(body, "content")
	if !content.IsArray() {
		return false
	}
	for _, item := range content.Array() {
		if strings.EqualFold(strings.TrimSpace(item.Get("type").String()), "video_url") &&
			strings.EqualFold(strings.TrimSpace(item.Get("role").String()), "base_video") {
			return true
		}
	}
	return false
}

func miniMaxVideoEndpointOnly(endpoint OpenAIVideoEndpoint) bool {
	switch endpoint {
	case OpenAIVideoEndpointList, OpenAIVideoEndpointDelete, OpenAIVideoEndpointContext, OpenAIVideoEndpointRegenerate:
		return true
	default:
		return false
	}
}

func IsMiniMaxOnlyOpenAIVideoEndpoint(endpoint OpenAIVideoEndpoint) bool {
	return miniMaxVideoEndpointOnly(endpoint)
}

func miniMaxVideoRequestModel(endpoint OpenAIVideoEndpoint, parsed *OpenAIVideoCreateRequest) string {
	if parsed != nil {
		return parsed.Model
	}
	if endpoint == OpenAIVideoEndpointList {
		return openAIVideoModelMiniMaxH3Canonical
	}
	return ""
}

func OpenAIVideoRequestModel(endpoint OpenAIVideoEndpoint, parsed *OpenAIVideoCreateRequest) string {
	return miniMaxVideoRequestModel(endpoint, parsed)
}

func buildMiniMaxV2VideoEndpointURL(base string, endpoint OpenAIVideoEndpoint, taskID string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	switch endpoint {
	case OpenAIVideoEndpointCreate:
		return base + "/v2/video_generation"
	case OpenAIVideoEndpointStatus:
		return base + "/v2/query/video_generation/" + url.PathEscape(strings.TrimSpace(taskID))
	case OpenAIVideoEndpointList:
		return base + "/v2/query/video_generation"
	case OpenAIVideoEndpointDelete:
		return base + "/v2/video_generation/" + url.PathEscape(strings.TrimSpace(taskID))
	case OpenAIVideoEndpointContext:
		return base + "/v2/h3_context_ir"
	case OpenAIVideoEndpointRegenerate:
		return base + "/v2/video_regeneration"
	default:
		return ""
	}
}

func miniMaxV2VideoEndpointPath(endpoint OpenAIVideoEndpoint, taskID string) string {
	return "/" + strings.TrimPrefix(buildMiniMaxV2VideoEndpointURL("", endpoint, taskID), "/")
}

func prepareMiniMaxV2VideoBody(endpoint OpenAIVideoEndpoint, body []byte, upstreamModel string) ([]byte, error) {
	if !endpoint.requiresRequestBody() {
		return body, nil
	}
	out, err := sjson.SetBytes(body, "model", canonicalOpenAIVideoModel(upstreamModel))
	if err != nil {
		return nil, fmt.Errorf("rewrite MiniMax video model: %w", err)
	}

	if endpoint != OpenAIVideoEndpointRegenerate && !gjson.GetBytes(out, "content").Exists() {
		prompt := strings.TrimSpace(gjson.GetBytes(out, "prompt").String())
		content, marshalErr := json.Marshal([]map[string]string{{"type": "text", "text": prompt}})
		if marshalErr != nil {
			return nil, fmt.Errorf("build MiniMax video content: %w", marshalErr)
		}
		out, err = sjson.SetRawBytes(out, "content", content)
		if err != nil {
			return nil, fmt.Errorf("set MiniMax video content: %w", err)
		}
	}

	if duration := gjson.GetBytes(out, "seconds"); !gjson.GetBytes(out, "duration").Exists() && duration.Exists() {
		seconds, parseErr := strconv.Atoi(strings.TrimSpace(duration.String()))
		if parseErr != nil {
			return nil, fmt.Errorf("MiniMax-H3 duration must be an integer from 4 to 15")
		}
		out, err = sjson.SetBytes(out, "duration", seconds)
		if err != nil {
			return nil, fmt.Errorf("set MiniMax video duration: %w", err)
		}
	}
	if ratio := strings.TrimSpace(gjson.GetBytes(out, "aspect_ratio").String()); ratio != "" && !gjson.GetBytes(out, "ratio").Exists() {
		out, err = sjson.SetBytes(out, "ratio", ratio)
		if err != nil {
			return nil, fmt.Errorf("set MiniMax video ratio: %w", err)
		}
	}
	for _, field := range []string{"prompt", "seconds", "aspect_ratio"} {
		if gjson.GetBytes(out, field).Exists() {
			out, err = sjson.DeleteBytes(out, field)
			if err != nil {
				return nil, fmt.Errorf("remove MiniMax video compatibility field %s: %w", field, err)
			}
		}
	}
	return out, nil
}

func miniMaxVideoTaskStatus(body []byte) string {
	return strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "task.status").String()))
}

func miniMaxVideoTaskFailed(body []byte) bool {
	switch miniMaxVideoTaskStatus(body) {
	case "failed", "cancelled":
		return true
	default:
		return false
	}
}

func miniMaxVideoContentURL(body []byte) (string, error) {
	status := miniMaxVideoTaskStatus(body)
	if status != "succeeded" {
		if status == "" {
			return "", fmt.Errorf("MiniMax video status response did not contain task.status")
		}
		return "", fmt.Errorf("MiniMax video task is %s", status)
	}
	rawURL := strings.TrimSpace(gjson.GetBytes(body, "task.content.url").String())
	if rawURL == "" {
		return "", fmt.Errorf("MiniMax video task succeeded without task.content.url")
	}
	return rawURL, nil
}

func validateMiniMaxVideoContentURL(raw string, allowPrivate bool) (string, error) {
	if allowPrivate {
		return urlvalidator.ValidateURLFormat(raw, true)
	}
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{AllowPrivate: false})
	if err != nil {
		return "", fmt.Errorf("MiniMax video content URL is not allowed: %w", err)
	}
	return normalized, nil
}

func rewriteMiniMaxVideoContentURLs(body []byte, endpoint OpenAIVideoEndpoint, clientTaskID string) []byte {
	out := body
	switch endpoint {
	case OpenAIVideoEndpointStatus:
		taskID := strings.TrimSpace(gjson.GetBytes(out, "task.id").String())
		if strings.TrimSpace(clientTaskID) != "" {
			taskID = strings.TrimSpace(clientTaskID)
			out, _ = sjson.SetBytes(out, "task.id", taskID)
		}
		if taskID != "" && strings.TrimSpace(gjson.GetBytes(out, "task.content.url").String()) != "" {
			out, _ = sjson.SetBytes(out, "task.content.url", "/v1/videos/"+url.PathEscape(taskID)+"/content")
		}
	case OpenAIVideoEndpointList:
		items := gjson.GetBytes(out, "items")
		if items.IsArray() {
			for index, item := range items.Array() {
				taskID := strings.TrimSpace(item.Get("id").String())
				if taskID != "" && strings.TrimSpace(item.Get("content.url").String()) != "" {
					out, _ = sjson.SetBytes(out, fmt.Sprintf("items.%d.content.url", index), "/v1/videos/"+url.PathEscape(taskID)+"/content")
				}
			}
		}
	}
	return out
}

func (s *OpenAIGatewayService) fetchMiniMaxVideoStatus(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	token string,
	validatedBaseURL string,
	taskID string,
) (*http.Response, []byte, error) {
	statusURL := buildMiniMaxV2VideoEndpointURL(validatedBaseURL, OpenAIVideoEndpointStatus, taskID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(WithHTTPUpstreamProfile(req.Context(), HTTPUpstreamProfileOpenAI))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	account.ApplyHeaderOverrides(req.Header)
	if customUA := account.GetOpenAIUserAgent(); customUA != "" {
		req.Header.Set("User-Agent", customUA)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	upstreamStart := time.Now()
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return resp, nil, nil
	}
	body, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	_ = resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

func (s *OpenAIGatewayService) forwardMiniMaxVideoContent(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	token string,
	validatedBaseURL string,
	taskID string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()
	statusResp, statusBody, err := s.fetchMiniMaxVideoStatus(ctx, c, account, token, validatedBaseURL, taskID)
	if err != nil {
		return nil, err
	}
	if statusResp == nil {
		return nil, fmt.Errorf("MiniMax video status response is empty")
	}
	if statusResp.StatusCode >= http.StatusBadRequest {
		statusURL := buildMiniMaxV2VideoEndpointURL(validatedBaseURL, OpenAIVideoEndpointStatus, taskID)
		return s.handleOpenAIVideoErrorResponse(ctx, statusResp, c, account, statusURL, openAIVideoModelMiniMaxH3Canonical, taskID, false)
	}

	result := &OpenAIForwardResult{
		ResponseID:          strings.TrimSpace(taskID),
		Model:               openAIVideoModelMiniMaxH3Canonical,
		UpstreamModel:       openAIVideoModelMiniMaxH3Canonical,
		UpstreamEndpoint:    "/v2/query/video_generation/{task_id}",
		ResponseHeaders:     statusResp.Header.Clone(),
		ResponseBody:        append([]byte(nil), statusBody...),
		TaskTerminalFailure: miniMaxVideoTaskFailed(statusBody),
	}
	contentURL, err := miniMaxVideoContentURL(statusBody)
	if err != nil {
		if c != nil && c.Writer != nil && !c.Writer.Written() {
			MarkResponseCommitted(c)
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"type": "invalid_request_error", "message": err.Error()}})
		}
		return result, err
	}
	allowPrivate := s.cfg != nil && s.cfg.Security.URLAllowlist.AllowPrivateHosts
	contentURL, err = validateMiniMaxVideoContentURL(contentURL, allowPrivate)
	if err != nil {
		return result, err
	}

	method := http.MethodGet
	if c != nil && c.Request != nil && c.Request.Method == http.MethodHead {
		method = http.MethodHead
	}
	contentReq, err := http.NewRequestWithContext(
		WithHTTPUpstreamRedirectsDisabled(WithHTTPUpstreamProfile(ctx, HTTPUpstreamProfileOpenAI)),
		method,
		contentURL,
		nil,
	)
	if err != nil {
		return result, err
	}
	contentReq.Header.Set("Accept", "*/*")
	if c != nil {
		for _, key := range []string{"Range", "If-Range"} {
			if value := strings.TrimSpace(c.GetHeader(key)); value != "" {
				contentReq.Header.Set(key, value)
			}
		}
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	upstreamStart := time.Now()
	contentResp, err := s.httpUpstream.Do(contentReq, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if err != nil {
		return result, err
	}
	defer func() { _ = contentResp.Body.Close() }()
	if contentResp.StatusCode >= http.StatusMultipleChoices && contentResp.StatusCode < http.StatusBadRequest {
		return result, fmt.Errorf("MiniMax video content redirect is not allowed")
	}
	if contentResp.StatusCode >= http.StatusBadRequest && contentResp.StatusCode != http.StatusRequestedRangeNotSatisfiable {
		body, readErr := io.ReadAll(io.LimitReader(contentResp.Body, 64<<10))
		if readErr != nil {
			return result, readErr
		}
		return result, fmt.Errorf("MiniMax video content returned HTTP %d: %s", contentResp.StatusCode, sanitizeUpstreamErrorMessage(string(body)))
	}
	if err := s.streamOpenAIVideoContentResponse(contentResp, c); err != nil {
		return result, err
	}
	result.ResponseHeaders = contentResp.Header.Clone()
	result.UpstreamEndpoint = "/v2/query/video_generation/{task_id} -> task.content.url"
	result.Duration = time.Since(startTime)
	return result, nil
}
