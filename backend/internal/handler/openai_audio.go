package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// AudioSpeech handles POST /v1/audio/speech.
func (h *OpenAIGatewayHandler) AudioSpeech(c *gin.Context) {
	h.audio(c, service.OpenAIAudioEndpointSpeech)
}

// AudioTranscriptions handles POST /v1/audio/transcriptions.
func (h *OpenAIGatewayHandler) AudioTranscriptions(c *gin.Context) {
	h.audio(c, service.OpenAIAudioEndpointTranscriptions)
}

// AudioTranslations handles POST /v1/audio/translations.
func (h *OpenAIGatewayHandler) AudioTranslations(c *gin.Context) {
	h.audio(c, service.OpenAIAudioEndpointTranslations)
}

func (h *OpenAIGatewayHandler) audio(c *gin.Context, endpoint service.OpenAIAudioEndpoint) {
	streamStarted := false
	requestStart := time.Now()
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	reqLog := requestLogger(c, "handler.openai_gateway.audio", zap.Int64("user_id", subject.UserID), zap.Int64("api_key_id", apiKey.ID), zap.Any("group_id", apiKey.GroupID))
	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}
	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.errorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}
	contentType := c.GetHeader("Content-Type")
	requestModel, err := parseOpenAIAudioModel(endpoint, body, contentType)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	reqLog = reqLog.With(zap.String("model", requestModel), zap.String("endpoint", string(endpoint)))
	setOpsRequestContext(c, requestModel, false)
	setOpsEndpointContext(c, endpoint.Path(), int16(service.RequestTypeSync))
	if decision := h.checkSecurityAudit(c, reqLog, apiKey, subject, "openai_audio", requestModel, body); decision != nil && !decision.AllowNextStage {
		h.openAISecurityAuditError(c, decision)
		return
	}
	channelMapping, _ := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, requestModel)
	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	service.SetOpsLatencyMs(c, service.OpsAuthLatencyMsKey, time.Since(requestStart).Milliseconds())
	userReleaseFunc, acquired := h.acquireResponsesUserSlot(c, subject.UserID, subject.Concurrency, false, &streamStarted, reqLog)
	if !acquired {
		return
	}
	if userReleaseFunc != nil {
		defer userReleaseFunc()
	}
	if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription, service.QuotaPlatform(c.Request.Context(), apiKey)); err != nil {
		status, code, message, retryAfter := billingErrorDetails(err)
		if retryAfter > 0 {
			c.Header("Retry-After", fmt.Sprint(retryAfter))
		}
		h.errorResponse(c, status, code, message)
		return
	}

	failedAccountIDs := make(map[int64]struct{})
	profitVetoCount := 0
	var lastFailoverErr *service.UpstreamFailoverError
	switchCount := 0
	maxAccountSwitches := h.maxAccountSwitches
	if maxAccountSwitches <= 0 {
		maxAccountSwitches = 3
	}
	routingStart := time.Now()
	for {
		selection, _, err := h.gatewayService.SelectAccountWithSchedulerForCapability(c.Request.Context(), apiKey.GroupID, "", "", requestModel, failedAccountIDs, service.OpenAIUpstreamTransportHTTPSSE, service.OpenAIEndpointCapabilityAudio, false, false, true)
		if err != nil || selection == nil || selection.Account == nil {
			if len(failedAccountIDs) == 0 {
				cls := classifyNoAccountErrorFromGin(c, h.gatewayService, apiKey, requestModel, requestModel, service.PlatformOpenAI)
				if !cls.ModelNotFound {
					markOpsRoutingCapacityLimitedIfNoAvailable(c, err)
				}
				h.errorResponse(c, cls.Status, cls.ErrType, cls.Message)
				return
			}
			if lastFailoverErr != nil {
				h.handleFailoverExhausted(c, lastFailoverErr, false)
			} else {
				h.errorResponse(c, http.StatusBadGateway, "api_error", "Upstream request failed")
			}
			return
		}
		account := selection.Account
		setOpsSelectedAccount(c, account.ID, account.Platform)
		pricing, pricingErr := h.gatewayService.EstimateOpenAIPerRequestCost(
			c.Request.Context(),
			apiKey,
			apiKey.User,
			requestModel,
			channelMapping.MappedModel,
			channelMapping.BillingModelSource,
			account.GetMappedModel(requestModel),
		)
		if pricingErr != nil || pricing == nil || pricing.ActualCost <= 0 {
			reqLog.Warn("openai_audio.pricing_unavailable", zap.Error(pricingErr))
			status, errType, message := pricingPreflightErrorDetails(requestModel, pricingErr)
			h.errorResponse(c, status, errType, message)
			return
		}
		if err := h.billingCacheService.CheckEstimatedCostCoverage(c.Request.Context(), apiKey.User, apiKey.Group, subscription, pricing.ActualCost); err != nil {
			status, code, message, retryAfter := billingErrorDetails(err)
			if retryAfter > 0 {
				c.Header("Retry-After", fmt.Sprint(retryAfter))
			}
			h.errorResponse(c, status, code, message)
			return
		}
		accountReleaseFunc, slotResult := h.acquireResponsesAccountSlot(c, apiKey.GroupID, "", selection, false, &streamStarted, reqLog)
		if slotResult == openAISlotAcquireProfitVetoed {
			if !recordOpenAIProfitVeto(failedAccountIDs, account.ID, &profitVetoCount) {
				h.handleOpenAIProfitVetoExhausted(c, streamStarted, reqLog, profitVetoCount)
				return
			}
			continue
		}
		if slotResult != openAISlotAcquireOK {
			return
		}
		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(routingStart).Milliseconds())
		forwardBody := body
		forwardContentType := contentType
		if channelMapping.Mapped {
			mappedModel := channelMapping.MappedModel
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "multipart/") {
				forwardBody, forwardContentType, err = rewriteAudioMultipartModel(body, contentType, mappedModel)
			} else {
				forwardBody = h.gatewayService.ReplaceModelInBody(body, mappedModel)
			}
			if err != nil {
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
				h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to rewrite model in multipart request")
				return
			}
		}
		writerSizeBeforeForward := c.Writer.Size()
		result, forwardErr := func() (*service.OpenAIForwardResult, error) {
			defer func() {
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
			}()
			return h.gatewayService.ForwardOpenAIAudio(c.Request.Context(), c, account, endpoint, forwardBody, forwardContentType, requestModel)
		}()
		if forwardErr != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(forwardErr, &failoverErr) {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account, account.GetMappedModel(requestModel), false, nil)
				failedAccountIDs[account.ID] = struct{}{}
				lastFailoverErr = failoverErr
				if switchCount >= maxAccountSwitches {
					h.handleFailoverExhausted(c, failoverErr, false)
					return
				}
				switchCount++
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account, account.GetMappedModel(requestModel), false, nil)
			if c.Writer.Size() == writerSizeBeforeForward {
				h.errorResponse(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
			}
			return
		}
		h.gatewayService.ReportOpenAIAccountScheduleResult(account, account.GetMappedModel(requestModel), true, nil)
		userAgent := c.GetHeader("User-Agent")
		clientIP := ip.GetClientIP(c)
		inboundEndpoint := GetInboundEndpoint(c)
		upstreamEndpoint := GetUpstreamEndpoint(c, account.Platform)
		quotaPlatform := service.QuotaPlatform(c.Request.Context(), apiKey)
		h.submitOpenAIUsageRecordTask(c.Request.Context(), result, func(ctx context.Context) {
			if err := h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
				Result: result, APIKey: apiKey, User: apiKey.User, Account: account, Subscription: subscription,
				InboundEndpoint: inboundEndpoint, UpstreamEndpoint: upstreamEndpoint, UserAgent: userAgent, IPAddress: clientIP,
				APIKeyService: h.apiKeyService, QuotaPlatform: quotaPlatform,
				ChannelUsageFields: channelMapping.ToUsageFields(requestModel, result.UpstreamModel),
			}); err != nil {
				reqLog.Warn("openai_audio.record_usage_failed", zap.Error(err))
			}
		})
		return
	}
}

func parseOpenAIAudioModel(endpoint service.OpenAIAudioEndpoint, body []byte, contentType string) (string, error) {
	if endpoint == service.OpenAIAudioEndpointSpeech {
		if !gjson.ValidBytes(body) {
			return "", errors.New("Failed to parse request body")
		}
		model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
		if model == "" {
			return "", errors.New("model is required")
		}
		if strings.TrimSpace(gjson.GetBytes(body, "input").String()) == "" {
			return "", errors.New("input is required")
		}
		return model, nil
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(strings.ToLower(mediaType), "multipart/") || strings.TrimSpace(params["boundary"]) == "" {
		return "", errors.New("multipart form data with a model and file is required")
	}
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	model := ""
	hasFile := false
	for {
		part, nextErr := reader.NextPart()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return "", errors.New("invalid multipart request")
		}
		fieldName := part.FormName()
		data, readErr := io.ReadAll(part)
		part.Close()
		if readErr != nil {
			return "", errors.New("invalid multipart request")
		}
		if fieldName == "model" {
			model = strings.TrimSpace(string(data))
			if model == "" {
				return "", errors.New("model is required")
			}
		}
		if fieldName == "file" {
			hasFile = true
		}
	}
	if model == "" {
		return "", errors.New("model is required")
	}
	if !hasFile {
		return "", errors.New("file is required")
	}
	return model, nil
}

func rewriteAudioMultipartModel(body []byte, contentType, mappedModel string) ([]byte, string, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || strings.TrimSpace(params["boundary"]) == "" {
		return nil, "", errors.New("invalid multipart content type")
	}
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	var out bytes.Buffer
	writer := multipart.NewWriter(&out)
	if err := writer.SetBoundary(params["boundary"]); err != nil {
		return nil, "", err
	}
	for {
		part, nextErr := reader.NextPart()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return nil, "", nextErr
		}
		fieldName := part.FormName()
		data, readErr := io.ReadAll(part)
		part.Close()
		if readErr != nil {
			return nil, "", readErr
		}
		hdr := part.Header
		partWriter, err := writer.CreatePart(hdr)
		if err != nil {
			return nil, "", err
		}
		if fieldName == "model" {
			data = []byte(mappedModel)
		}
		if _, err := partWriter.Write(data); err != nil {
			return nil, "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return out.Bytes(), mediaType + "; boundary=" + params["boundary"], nil
}
