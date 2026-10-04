package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service/basispoints"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var openAIBasispointsReplay basispoints.ReplayCache
var openAIBasispointsCatalog basispoints.CatalogCache

func prepareOpenAIBasispointsBody(c *gin.Context, account *Account, body []byte) ([]byte, *basispoints.Bridge, error) {
	return prepareOpenAIBasispointsBodyWithImages(c, account, body, nil)
}

func prepareOpenAIBasispointsBodyWithImages(c *gin.Context, account *Account, body []byte, relay *basispoints.ImageRelay, native ...*basispoints.NativeImages) ([]byte, *basispoints.Bridge, error) {
	identity, _ := resolveOpenAIWSExecutionScope(c, body, getAPIKeyIDFromContext(c))
	var replay *basispoints.ReplayCache
	var catalog *basispoints.CatalogCache
	if identity != "" {
		var err error
		body, err = sjson.SetBytes(body, "prompt_cache_key", identity)
		if err != nil {
			return nil, nil, err
		}
		replay = &openAIBasispointsReplay
		catalog = &openAIBasispointsCatalog
	}
	// Without a stable client thread identity, require complete history rather
	// than sharing cached tool calls across unrelated conversations.
	scope := fmt.Sprintf("account:%d/key:%d/thread:%s", account.ID, getAPIKeyIDFromContext(c), identity)
	var err error
	body, err = relay.Rewrite(body, scope)
	if err != nil {
		return nil, nil, err
	}
	if isOpenAIResponsesCompactPath(c) {
		body, err = basispointsCompactInput(body)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(native) > 0 && native[0] != nil {
		_, bridge, err := native[0].PrepareWithCatalog(scope, replay, catalog)
		if err != nil {
			return nil, nil, err
		}
		return bridge.Reprepare(body)
	}
	return basispoints.PrepareWithCatalog(body, scope, replay, catalog)
}

func basispointsCompactInput(body []byte) ([]byte, error) {
	var request map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil {
		return nil, err
	}
	var input []any
	switch value := request["input"].(type) {
	case []any:
		input = value
	case string:
		input = []any{map[string]any{"role": "user", "content": value}}
	default:
		return nil, fmt.Errorf("Basispoints compact requires input")
	}
	request["input"] = append(input, map[string]any{"type": "compaction_trigger"})
	request["tool_choice"] = "none"
	return json.Marshal(request)
}

// Use the existing proxy and account concurrency controller, but never invoke
// Codex plugins/TLS identity injection for the separate Basispoints host.
func (s *OpenAIGatewayService) doOpenAIBasispoints(req *http.Request, account *Account) (*http.Response, error) {
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	return accountTrafficController(s.httpUpstream).DoHTTP(withCodexTrafficRequest(req, account), func(req *http.Request) (*http.Response, error) {
		return s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	})
}

func (s *OpenAIGatewayService) forwardOpenAIBasispoints(ctx context.Context, c *gin.Context, account *Account, body []byte, start time.Time) (*OpenAIForwardResult, error) {
	return s.forwardOpenAIBasispointsWithChatResponse(ctx, c, account, body, start, nil)
}

// Chat requests are already mapped and converted to Responses before entering
// the shared BPS pipeline. Only the downstream response format differs.
type basispointsChatResponse struct {
	originalModel  string
	billingModel   string
	upstreamModel  string
	requestBodyLen int
}

func (s *OpenAIGatewayService) forwardOpenAIBasispointsWithChatResponse(ctx context.Context, c *gin.Context, account *Account, body []byte, start time.Time, chat *basispointsChatResponse) (*OpenAIForwardResult, error) {
	fail := func(status int, code, message string, param ...string) (*OpenAIForwardResult, error) {
		committed := StopOpenAICompactSSEKeepaliveCommitted(c)
		MarkResponseCommitted(c)
		if committed {
			failureParam := ""
			if len(param) > 0 {
				failureParam = param[0]
			}
			writeOpenAICompactSSEFailureMessageParam(c, status, code, message, failureParam)
		} else {
			errorType := "invalid_request_error"
			if status >= 500 {
				errorType = "server_error"
			}
			errorBody := gin.H{"type": errorType, "code": code, "message": message}
			if len(param) > 0 && param[0] != "" {
				errorBody["param"] = param[0]
			}
			c.JSON(status, gin.H{"error": errorBody})
		}
		return nil, fmt.Errorf("Basispoints: %s", code)
	}
	originalModel := gjson.GetBytes(body, "model").String()
	var model string
	if chat != nil {
		originalModel, model = chat.originalModel, chat.upstreamModel
	} else {
		model = account.GetMappedModel(originalModel)
	}
	stream := gjson.GetBytes(body, "stream").Bool()
	clientCanceled := func() (*OpenAIForwardResult, error) {
		StopOpenAICompactSSEKeepaliveCommitted(c)
		MarkResponseCommitted(c)
		MarkOpsClientCancellation(c, stream)
		return nil, context.Canceled
	}
	failoverRateLimited := func(retryAfter string) (*OpenAIForwardResult, error) {
		s.coolDownExcelBPS(ctx, account, retryAfter)
		if isBasispointsClientCancellation(c, ctx.Err()) {
			return clientCanceled()
		}
		return nil, newExcelBPSRateLimitedFailoverError(retryAfter)
	}
	if isBasispointsClientCancellation(c, ctx.Err()) {
		return clientCanceled()
	}
	if c.GetBool(bpsAccountProbeRequiredContextKey) && basispoints.NativeFallbackReason(body) != "" {
		return nil, errors.New("bps probe path is unavailable")
	}
	imageSettings, err := s.settingService.GetExcelBPSImageRelaySettings(ctx)
	if err != nil {
		if isBasispointsClientCancellation(c, err) {
			return clientCanceled()
		}
		return fail(503, "basispoints_image_relay_unavailable", "Basispoints image settings are unavailable")
	}
	if !imageSettings.Enabled && account.IsBasispointsIgnoreImagesEnabled() {
		body, err = basispoints.StripInputImages(body)
		if err != nil {
			return fail(400, "basispoints_request_invalid", err.Error())
		}
	}
	if account.IsBasispointsIgnoreEncryptedContentEnabled() {
		body, err = basispoints.StripEncryptedContent(body)
		if err != nil {
			return fail(400, "basispoints_request_invalid", err.Error())
		}
	}
	if imageSettings.Enabled {
		s.excelBPSImageAdmission.SetAdmissionLimits(imageSettings.BodyLimitMiB, imageSettings.BudgetMiB, imageSettings.MaxRequests)
		releaseImages, admissionErr := s.excelBPSImageAdmission.AdmitRequest(ctx, len(body))
		if admissionErr != nil {
			if isBasispointsClientCancellation(c, admissionErr) {
				return clientCanceled()
			}
			if errors.Is(admissionErr, basispoints.ErrImageRelayRequestTooLarge) {
				return fail(413, "basispoints_image_request_too_large", admissionErr.Error())
			}
			c.Header("Retry-After", "1")
			return fail(503, "basispoints_image_relay_unavailable", "Basispoints image relay is busy or unavailable")
		}
		defer releaseImages()
	}
	body, err = sjson.SetBytes(body, "model", model)
	if err != nil {
		return fail(400, "basispoints_request_invalid", "Invalid model request")
	}
	var images *basispoints.NativeImages
	var relay *basispoints.ImageRelay
	if imageSettings.Enabled && imageSettings.Mode == ExcelBPSImageModeNative {
		images, err = basispoints.PrepareNativeImagesWithLimit(body, imageSettings.Limits.MaxImages)
		if err == nil {
			body, err = images.Body()
		}
		if err != nil {
			return fail(400, "basispoints_request_invalid", err.Error())
		}
	} else {
		relay, err = s.excelBPSImageRelayForSettings(imageSettings)
		if err != nil {
			return fail(503, "basispoints_image_relay_unavailable", "Basispoints image relay is unavailable")
		}
	}
	wireBody, bridge, err := prepareOpenAIBasispointsBodyWithImages(c, account, body, relay, images)
	if err != nil {
		if errors.Is(err, basispoints.ErrImageRelayFull) || errors.Is(err, basispoints.ErrImageRelayStorage) {
			return fail(503, "basispoints_image_relay_unavailable", err.Error())
		}
		var contentErr *basispoints.ContentValidationError
		if errors.As(err, &contentErr) {
			return fail(400, "basispoints_request_invalid", err.Error(), contentErr.Path)
		}
		return fail(400, "basispoints_request_invalid", err.Error())
	}
	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		if isBasispointsClientCancellation(c, err) {
			return clientCanceled()
		}
		return fail(502, "basispoints_auth_unavailable", "Account OAuth credential is unavailable")
	}
	// Complete local validation before uploading anything. Credential and thread
	// scope prevent attachment reuse across users, sessions, or rotated tokens.
	if images != nil && images.HasImages() {
		identity, _ := resolveOpenAIWSExecutionScope(c, body, getAPIKeyIDFromContext(c))
		attachmentScope := ""
		if identity != "" {
			attachmentScope = fmt.Sprintf("account:%d/key:%d/thread:%s\x00%s\x00%s", account.ID, getAPIKeyIDFromContext(c), identity, account.GetChatGPTAccountID(), token)
		}
		body, err = images.Upload(ctx, &s.excelBPSAttachments, attachmentScope, func(uploadCtx context.Context, img basispoints.InlineAttachment) (string, error) {
			SetActualOpenAIUpstreamEndpoint(c, "/basispoints/api/attachments")
			return s.uploadBasispointsAttachment(uploadCtx, account, token, img)
		})
		if err != nil {
			if isBasispointsClientCancellation(c, err) {
				return clientCanceled()
			}
			status, code := http.StatusBadGateway, "basispoints_attachment_error"
			var uploadError *basispointsAttachmentError
			if errors.As(err, &uploadError) {
				status = uploadError.status
			}
			if errors.Is(err, basispoints.ErrAttachmentBusy) {
				status = http.StatusServiceUnavailable
			}
			if status == http.StatusTooManyRequests {
				retryAfter := ""
				if uploadError != nil {
					retryAfter = uploadError.retryAfter
				}
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
					ProxyID: opsUpstreamProxyID(account), ProxyName: opsUpstreamProxyName(account),
					UpstreamStatusCode: status, UpstreamURL: basispoints.AttachmentsURL,
					Kind: "failover", Message: "Basispoints attachment upload was rate limited",
				})
				return failoverRateLimited(retryAfter)
			}
			setOpsUpstreamError(c, status, "Basispoints attachment upload failed", "")
			if status == http.StatusUnauthorized {
				return fail(status, code, "Basispoints attachment authentication failed; request was not replayed")
			}
			return fail(status, code, "Basispoints attachment upload failed; request was not replayed")
		}
		if identity != "" {
			body, err = sjson.SetBytes(body, "prompt_cache_key", identity)
		}
		if err == nil && isOpenAIResponsesCompactPath(c) {
			body, err = basispointsCompactInput(body)
		}
		if err == nil {
			wireBody, bridge, err = bridge.Reprepare(body)
		}
		if err != nil {
			return fail(400, "basispoints_request_invalid", err.Error())
		}
	}
	upstreamCtx, release := detachUpstreamContext(ctx)
	defer release()
	req, err := newOpenAIBasispointsRequest(upstreamCtx, s.accountRepo, account, wireBody, token)
	if err != nil {
		return fail(400, "basispoints_account_id_missing", "Basispoints requires a valid chatgpt_account_id")
	}
	SetActualOpenAIUpstreamEndpoint(c, "/basispoints/api/responses")
	SetOpsUpstreamModel(c, model)
	sent := time.Now()
	resp, err := s.doOpenAIBasispoints(withAccountTrafficAdmissionContext(req, ctx), account)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(sent).Milliseconds())
	if err != nil {
		if isBasispointsClientCancellation(c, err) {
			return clientCanceled()
		}
		return nil, s.handleOpenAIUpstreamTransportError(ctx, c, account, err, false)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	recoveredEncrypted := false
	if resp.StatusCode == http.StatusBadRequest && !c.GetBool(bpsAccountProbeRequiredContextKey) {
		const maxRejectionBytes = 512 << 10
		raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxRejectionBytes+1))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(raw))
		if retryBody, retry := prepareBasispointsInvalidEncryptedRetry(wireBody, raw); retry && readErr == nil && len(raw) <= maxRejectionBytes && ctx.Err() == nil {
			retryReq, retryErr := newOpenAIBasispointsRequest(upstreamCtx, s.accountRepo, account, retryBody, token)
			if retryErr != nil {
				return fail(502, "basispoints_transport_error", "Basispoints recovery request could not be prepared")
			}
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
				ProxyID: opsUpstreamProxyID(account), ProxyName: opsUpstreamProxyName(account),
				UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
				UpstreamURL: basispointsResponsesURL, Kind: "invalid_encrypted_content_retry",
				Message: "Basispoints rejected encrypted reasoning; retrying once without opaque reasoning on the same route",
			})
			recoveredEncrypted = true
			resp, err = s.doOpenAIBasispoints(withAccountTrafficAdmissionContext(retryReq, ctx), account)
			SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(sent).Milliseconds())
			if err != nil {
				if isBasispointsClientCancellation(c, err) {
					return clientCanceled()
				}
				return fail(502, "basispoints_transport_error", "Basispoints recovery connection failed; request was not replayed again")
			}
			wireBody = retryBody
		}
	}
	if isBasispointsClientCancellation(c, ctx.Err()) {
		return clientCanceled()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, openAIUpstreamErrorBodyReadLimit))
		disabled := s.observeBasispointsUpstreamResponse(ctx, account, resp, raw)
		message := fmt.Sprintf("Basispoints returned HTTP %d", resp.StatusCode)
		code := "basispoints_upstream_error"
		if resp.StatusCode == http.StatusBadRequest && isBasispointsInvalidEncryptedContent(raw) {
			code = "invalid_encrypted_content"
			message = "Basispoints could not verify encrypted conversation state; resend the original plaintext history or start a new conversation"
		}
		if resp.StatusCode == http.StatusUnauthorized {
			message = "Basispoints authentication failed; request was not replayed"
		}
		if gjson.GetBytes(raw, "error.code").String() == "basispoints_model_access_changed" {
			code = "basispoints_model_access_changed"
			message = "This model is not available on the account's Basispoints endpoint"
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			code = "basispoints_rate_limited"
			message = "Basispoints rate limit exceeded; account will be retried through another route"
		}
		// Do not send raw upstream errors to clients: they may echo credentials
		// or full request content. Preserve status and request ID for diagnosis.
		kind := "http_error"
		if resp.StatusCode == http.StatusTooManyRequests {
			kind = "failover"
		} else {
			setOpsUpstreamError(c, resp.StatusCode, message, "")
		}
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
			ProxyID: opsUpstreamProxyID(account), ProxyName: opsUpstreamProxyName(account),
			UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
			UpstreamURL: basispointsResponsesURL, Kind: kind, Message: message,
		})
		if resp.StatusCode == http.StatusTooManyRequests {
			return failoverRateLimited(resp.Header.Get("Retry-After"))
		}
		if resp.StatusCode >= 500 && !recoveredEncrypted {
			safeBody, _ := json.Marshal(gin.H{"error": gin.H{"code": code, "message": message}})
			return nil, s.newOpenAIAccountFailoverError(account, resp.StatusCode, resp.Header, safeBody, message, false, false)
		}
		if disabled {
			message = "Basispoints returned HTTP 403; this account now uses ChatGPT Codex; this request was not replayed"
		}
		return fail(resp.StatusCode, code, message)
	}
	// Quota belongs to the account even if downstream streaming later fails.
	s.observeBasispointsUpstreamResponse(ctx, account, resp, nil)
	if c.GetBool(bpsAccountProbeRequiredContextKey) {
		// The diagnostic probe promises at most three requests and must expose
		// a malformed tool call, not silently heal it with additional requests.
		resp.Body = bridge.Stream(resp.Body)
	} else {
		resp.Body = s.basispointsStreamWithRepairs(ctx, account, wireBody, token, bridge, resp.Body)
	}
	if resp.Header == nil {
		resp.Header = make(http.Header)
	}
	resp.Header.Set("Content-Type", "text/event-stream")
	result := &OpenAIForwardResult{
		RequestID: resp.Header.Get("x-request-id"), UpstreamHeaders: resp.Header,
		Model: originalModel, UpstreamModel: model, UpstreamEndpoint: "/basispoints/api/responses",
		Stream: stream, ReasoningEffort: &bridge.Effort,
		RequestedReasoningEffort: coalesceRequestedReasoningEffort(RequestedReasoningEffortFromContext(ctx), &bridge.RequestedEffort),
	}
	if chat != nil {
		var parsed *OpenAIForwardResult
		if stream {
			parsed, err = s.handleChatStreamingResponse(resp, c, account, originalModel, chat.billingModel, model, start, chat.requestBodyLen)
		} else {
			parsed, err = s.handleChatBufferedStreamingResponse(resp, c, account, originalModel, chat.billingModel, model, start)
		}
		result.BillingModel = chat.billingModel
		if parsed != nil {
			result.Usage, result.ResponseID, result.FirstTokenMs = parsed.Usage, parsed.ResponseID, parsed.FirstTokenMs
			result.ClientDisconnect = parsed.ClientDisconnect
			result.UpstreamResponseServiceTier = parsed.UpstreamResponseServiceTier
		}
	} else if stream {
		parsed, parseErr := s.handleStreamingResponseWithReasoning(ctx, resp, c, account, start, originalModel, model, bridge.Effort)
		err = parseErr
		if parsed != nil {
			result.ResponseID, result.FirstTokenMs = parsed.responseID, parsed.firstTokenMs
			if parsed.usage != nil {
				result.Usage = *parsed.usage
			}
		}
	} else {
		parsed, parseErr := s.handleNonStreamingResponse(ctx, resp, c, account, originalModel, model)
		err = parseErr
		if parsed != nil {
			result.ResponseID = parsed.responseID
			if parsed.usage != nil {
				result.Usage = *parsed.usage
			}
		}
	}
	result.Duration = time.Since(start)
	result.UpstreamResponseModel = observedUpstreamResponseModel(c)
	result.UpstreamResponseModelConflict = observedUpstreamResponseModelConflict(c)
	if terminal := c.GetString(basispointsFailurePayloadKey); terminal != "" {
		result.UpstreamTerminalEvent = gjson.Get(terminal, "type").String()
		// Preserve the existing usage/no-usage return policy below while keeping
		// accepted failures distinct from pre-send account failover errors.
		err = basispoints.ParseUpstreamFailure([]byte(terminal))
	}
	if err != nil {
		if isBasispointsClientCancellation(c, ctx.Err()) {
			MarkOpsClientCancellation(c, stream)
			result.ClientDisconnect = true
			return result, context.Canceled
		}
		if !OpenAIForwardResultHasBillableUsage(result) {
			return nil, err
		}
		result.ClientDisconnect = ctx.Err() != nil
		return result, err
	}
	s.bindHTTPResponseAccount(ctx, c, account, result.ResponseID)
	return result, nil
}

func (s *AccountTestService) testOpenAIBasispointsAccount(c *gin.Context, account *Account, model, prompt string) error {
	ctx := c.Request.Context()
	if strings.TrimSpace(prompt) == "" {
		prompt = "hi"
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")
	s.sendEvent(c, TestEvent{Type: "test_start", Model: model})
	body, _ := json.Marshal(gin.H{"model": model, "input": prompt, "stream": true, "reasoning": gin.H{"effort": "low"}, "prompt_cache_key": "bps-account-test-" + uuid.NewString()})
	wireBody, bridge, err := prepareOpenAIBasispointsBody(c, account, body)
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}
	gateway := s.openaiGatewayService
	if gateway == nil {
		gateway = &OpenAIGatewayService{accountRepo: s.accountRepo, httpUpstream: s.httpUpstream}
	}
	token, _, err := gateway.GetAccessToken(ctx, account)
	if err != nil {
		return s.sendErrorAndEnd(c, "Account OAuth credential is unavailable")
	}
	req, err := newOpenAIBasispointsRequest(ctx, s.accountRepo, account, wireBody, token)
	if err != nil {
		return s.sendErrorAndEnd(c, "Basispoints requires a valid chatgpt_account_id")
	}
	resp, err := gateway.doOpenAIBasispoints(req, account)
	if err != nil {
		return s.sendErrorAndEnd(c, "Basispoints connection failed")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, openAIUpstreamErrorBodyReadLimit))
		if gateway.observeBasispointsUpstreamResponse(ctx, account, resp, raw) {
			return s.sendErrorAndEnd(c, "Basispoints returned HTTP 403; this account now uses ChatGPT Codex; this test was not replayed")
		}
		return s.sendErrorAndEnd(c, fmt.Sprintf("Basispoints returned HTTP %d (request ID: %s)", resp.StatusCode, resp.Header.Get("x-request-id")))
	}
	gateway.observeBasispointsUpstreamResponse(ctx, account, resp, nil)
	resp.Body = bridge.Stream(resp.Body)
	return s.processOpenAIStream(c, resp.Body)
}
