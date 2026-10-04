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
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const openAINanoBananaEndpoint = "/v1/api/nano-banana"

type OpenAINanoBananaRequest struct {
	Model     string
	ImageSize string
	N         int
	Body      []byte
}

func ParseOpenAINanoBananaRequest(body []byte) (*OpenAINanoBananaRequest, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("request body is empty")
	}
	if !gjson.ValidBytes(body) {
		return nil, fmt.Errorf("failed to parse request body")
	}
	model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if !isNanoBananaImageModel(model) {
		if model == "" {
			return nil, fmt.Errorf("nano-banana endpoint requires model")
		}
		return nil, fmt.Errorf("nano-banana endpoint supports nano-banana-pro or nano-banana-pro-fast, got %q", model)
	}
	prompt := strings.TrimSpace(gjson.GetBytes(body, "prompt").String())
	if prompt == "" {
		return nil, fmt.Errorf("nano-banana endpoint requires prompt")
	}
	imageCount := 1
	if requestedCount := gjson.GetBytes(body, "n"); requestedCount.Exists() {
		if requestedCount.Type != gjson.Number || requestedCount.Int() <= 0 || requestedCount.Float() != float64(requestedCount.Int()) {
			return nil, fmt.Errorf("nano-banana endpoint requires n to be a positive integer")
		}
		imageCount = int(requestedCount.Int())
	}

	return &OpenAINanoBananaRequest{
		Model:     model,
		ImageSize: extractNanoBananaImageSize(body),
		N:         imageCount,
		Body:      body,
	}, nil
}

func (r *OpenAINanoBananaRequest) StickySessionSeed() string {
	if r == nil {
		return ""
	}
	prompt := strings.TrimSpace(gjson.GetBytes(r.Body, "prompt").String())
	aspectRatio := strings.TrimSpace(gjson.GetBytes(r.Body, "aspectRatio").String())
	return strings.Join([]string{"nano-banana", r.Model, r.ImageSize, fmt.Sprintf("%d", r.N), aspectRatio, prompt}, "|")
}

func (r *OpenAINanoBananaRequest) ModerationBody() []byte {
	if r == nil || len(r.Body) == 0 {
		return nil
	}
	payload := []byte(`{}`)
	if prompt := strings.TrimSpace(gjson.GetBytes(r.Body, "prompt").String()); prompt != "" {
		payload, _ = sjson.SetBytes(payload, "prompt", prompt)
	}
	if images := gjson.GetBytes(r.Body, "images"); images.Exists() && images.IsArray() {
		payload, _ = sjson.SetRawBytes(payload, "images", []byte(images.Raw))
	}
	return payload
}

func (s *OpenAIGatewayService) ForwardNanoBanana(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *OpenAINanoBananaRequest,
) (*OpenAIForwardResult, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed nano-banana request is required")
	}
	if account == nil {
		return nil, fmt.Errorf("openai account is required")
	}
	if account.Type != AccountTypeAPIKey {
		return nil, fmt.Errorf("nano-banana endpoint requires an OpenAI APIKey account")
	}

	upstreamModel := normalizeNanoBananaImageModel(account.GetMappedModel(parsed.Model))
	if !isNanoBananaImageModel(upstreamModel) {
		return nil, fmt.Errorf("nano-banana upstream model must be nano-banana-pro or nano-banana-pro-fast, got %q", upstreamModel)
	}

	if !AccountSupportsNanoBananaEndpoint(account, parsed.Model) {
		return nil, fmt.Errorf("openai account is not configured for nano-banana endpoint")
	}

	body, err := prepareNanoBananaUpstreamBody(parsed.Body, upstreamModel)
	if err != nil {
		return nil, err
	}

	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	if apiKey == "" {
		return nil, fmt.Errorf("openai api_key not configured")
	}
	baseURL := strings.TrimSpace(account.GetCredential("base_url"))
	if baseURL == "" {
		return nil, fmt.Errorf("nano-banana endpoint requires an OpenAI APIKey account with custom base_url")
	}
	validatedURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	targetURL := buildNanoBananaEndpointURL(validatedURL)

	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	upstreamReq = upstreamReq.WithContext(WithHTTPUpstreamProfile(upstreamReq.Context(), HTTPUpstreamProfileOpenAI))
	upstreamReq.Header.Set("Authorization", "Bearer "+apiKey)
	upstreamReq.Header.Set("Content-Type", "application/json")
	if customUA := account.GetOpenAIUserAgent(); customUA != "" {
		upstreamReq.Header.Set("User-Agent", customUA)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	startTime := time.Now()
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(startTime).Milliseconds())
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
			Kind:               "request_error",
			Message:            safeErr,
		})
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	if resp.StatusCode >= 400 {
		respBody := s.readUpstreamErrorBody(resp)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		if s.shouldFailoverOpenAIUpstreamResponse(account, resp.StatusCode, upstreamMsg, respBody) {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
				Kind:               "failover",
				Message:            upstreamMsg,
			})
			s.handleFailoverSideEffects(ctx, resp, account, respBody, upstreamModel)
			return nil, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				RetryableOnSameAccount: account.IsPoolMode() && account.IsPoolModeRetryableStatus(resp.StatusCode),
			}
		}
		return s.handleOpenAIImagesErrorResponse(ctx, resp, c, account, upstreamModel)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	if err != nil {
		return nil, err
	}
	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := "application/json"
	if s.cfg != nil && !s.cfg.Security.ResponseHeaders.Enabled {
		if upstreamType := resp.Header.Get("Content-Type"); upstreamType != "" {
			contentType = upstreamType
		}
	}
	c.Data(resp.StatusCode, contentType, bodyBytes)

	imageCount := extractNanoBananaImageCountFromJSONBytes(bodyBytes)
	if imageCount <= 0 {
		imageCount = parsed.N
	}
	return &OpenAIForwardResult{
		RequestID:       resp.Header.Get("x-request-id"),
		Model:           parsed.Model,
		UpstreamModel:   upstreamModel,
		ResponseHeaders: resp.Header.Clone(),
		Duration:        time.Since(startTime),
		ImageCount:      imageCount,
		ImageSize:       NormalizeImageBillingTierOrDefault(parsed.ImageSize),
		ImageInputSize:  parsed.ImageSize,
	}, nil
}

func isNanoBananaImageModel(model string) bool {
	switch normalizeNanoBananaImageModel(model) {
	case "nano-banana-pro", "nano-banana-pro-fast":
		return true
	default:
		return false
	}
}

func normalizeNanoBananaImageModel(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return normalized
}

func NormalizeNanoBananaImageModel(model string) string {
	return normalizeNanoBananaImageModel(model)
}

func AccountSupportsNanoBananaEndpoint(account *Account, requestedModel string) bool {
	if account == nil || account.Type != AccountTypeAPIKey || !account.IsOpenAI() {
		return false
	}
	baseURL := strings.TrimSpace(account.GetCredential("base_url"))
	if baseURL == "" {
		return false
	}
	if isKnownNanoBananaBaseURL(baseURL) {
		return true
	}
	return false
}

func isKnownNanoBananaBaseURL(baseURL string) bool {
	normalized := strings.ToLower(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	return strings.Contains(normalized, "visionary.beer") ||
		strings.HasSuffix(normalized, "/v1/api/nano-banana") ||
		strings.HasSuffix(normalized, "/api/nano-banana")
}

func prepareNanoBananaUpstreamBody(body []byte, upstreamModel string) ([]byte, error) {
	if !isNanoBananaImageModel(upstreamModel) {
		return nil, fmt.Errorf("nano-banana upstream model must be nano-banana-pro or nano-banana-pro-fast, got %q", upstreamModel)
	}
	rewritten, err := sjson.SetBytes(body, "model", normalizeNanoBananaImageModel(upstreamModel))
	if err != nil {
		return nil, fmt.Errorf("rewrite nano-banana model: %w", err)
	}

	aspectRatio := strings.TrimSpace(gjson.GetBytes(rewritten, "aspectRatio").String())
	ratio := strings.TrimSpace(gjson.GetBytes(rewritten, "ratio").String())
	if aspectRatio == "" && ratio != "" {
		rewritten, err = sjson.SetBytes(rewritten, "aspectRatio", ratio)
		if err != nil {
			return nil, fmt.Errorf("rewrite nano-banana aspectRatio: %w", err)
		}
		aspectRatio = ratio
	}
	if ratio == "" && aspectRatio != "" {
		rewritten, err = sjson.SetBytes(rewritten, "ratio", aspectRatio)
		if err != nil {
			return nil, fmt.Errorf("rewrite nano-banana ratio: %w", err)
		}
	}

	imageSize := strings.TrimSpace(gjson.GetBytes(rewritten, "imageSize").String())
	if imageSize == "" {
		for _, path := range []string{"image_size", "generationConfig.imageConfig.imageSize"} {
			if value := strings.TrimSpace(gjson.GetBytes(rewritten, path).String()); value != "" {
				rewritten, err = sjson.SetBytes(rewritten, "imageSize", value)
				if err != nil {
					return nil, fmt.Errorf("rewrite nano-banana imageSize: %w", err)
				}
				break
			}
		}
	}

	return rewritten, nil
}

func extractNanoBananaImageSize(body []byte) string {
	for _, path := range []string{
		"imageSize",
		"image_size",
		"generationConfig.imageConfig.imageSize",
	} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ImageBillingSize2K
}

func extractNanoBananaImageCountFromJSONBytes(body []byte) int {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return 0
	}
	return countNanoBananaImageValue(gjson.ParseBytes(body))
}

func countNanoBananaImageValue(value gjson.Result) int {
	if value.IsArray() {
		count := 0
		for _, item := range value.Array() {
			if itemCount := countNanoBananaImageValue(item); itemCount > 0 {
				count += itemCount
			}
		}
		return count
	}
	if value.Type == gjson.String {
		payload := strings.TrimSpace(value.String())
		if strings.HasPrefix(payload, "http://") || strings.HasPrefix(payload, "https://") || strings.HasPrefix(payload, "data:image/") || len(payload) >= 64 {
			return 1
		}
		return 0
	}
	if !value.IsObject() {
		return 0
	}
	for _, key := range []string{"url", "image_url", "imageUrl", "b64_json", "base64", "image"} {
		if payload := value.Get(key); payload.Exists() && strings.TrimSpace(payload.String()) != "" {
			return 1
		}
	}
	for _, key := range []string{"results", "images", "data", "result", "output"} {
		if child := value.Get(key); child.Exists() {
			if count := countNanoBananaImageValue(child); count > 0 {
				return count
			}
		}
	}
	return 0
}

func buildNanoBananaEndpointURL(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, openAINanoBananaEndpoint) || strings.HasSuffix(normalized, "/api/nano-banana") {
		return normalized
	}
	if openAIBaseURLHasVersionSuffix(normalized) {
		return normalized + "/api/nano-banana"
	}
	return normalized + openAINanoBananaEndpoint
}

func NanoBananaEndpointURLForBase(base string) string {
	if strings.TrimSpace(base) == "" {
		return ""
	}
	return buildNanoBananaEndpointURL(base)
}
