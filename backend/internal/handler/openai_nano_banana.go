package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NanoBanana handles Visionary-compatible Nano Banana Pro image requests.
// POST /v1/api/nano-banana
func (h *OpenAIGatewayHandler) NanoBanana(c *gin.Context) {
	streamStarted := false
	defer h.recoverResponsesPanic(c, &streamStarted)

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

	reqLog := requestLogger(
		c,
		"handler.openai_gateway.nano_banana",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)
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
	parsed, err := service.ParseOpenAINanoBananaRequest(body)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	requestModel := parsed.Model
	reqLog = reqLog.With(zap.String("model", requestModel), zap.String("image_size", parsed.ImageSize))

	setOpsRequestContext(c, requestModel, false)
	setOpsEndpointContext(c, "", int16(service.RequestTypeFromLegacy(false, false)))

	if !service.GroupAllowsImageGeneration(apiKey.Group) {
		h.errorResponse(c, http.StatusForbidden, "permission_error", service.ImageGenerationPermissionMessage())
		return
	}
	if decision := h.checkSecurityAudit(c, reqLog, apiKey, subject, service.ContentModerationProtocolOpenAIImages, requestModel, parsed.ModerationBody()); decision != nil && !decision.AllowNextStage {
		h.openAISecurityAuditError(c, decision)
		return
	}

	imageReleaseFunc, acquired := h.acquireImageGenerationSlot(c, streamStarted)
	if !acquired {
		return
	}
	if imageReleaseFunc != nil {
		defer imageReleaseFunc()
	}

	channelMapping, _ := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, requestModel)
	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	requestPayloadHash := service.HashUsageRequestPayload(body)
	estimatedCost, err := h.gatewayService.EstimateOpenAIImagesCost(
		c.Request.Context(),
		apiKey,
		apiKey.User,
		requestModel,
		channelMapping.MappedModel,
		channelMapping.BillingModelSource,
		parsed.N,
		parsed.ImageSize,
	)
	if err != nil {
		reqLog.Warn("openai.nano_banana.estimate_cost_failed", zap.Error(err))
		status, errType, message := pricingPreflightErrorDetails(requestModel, err)
		h.handleStreamingAwareError(c, status, errType, message, streamStarted)
		return
	}
	if estimatedCost != nil && estimatedCost.ActualCost > 0 {
		if err := h.billingCacheService.CheckEstimatedCostCoverage(c.Request.Context(), apiKey.User, apiKey.Group, subscription, estimatedCost.ActualCost); err != nil {
			reqLog.Info("openai.nano_banana.estimated_cost_check_failed",
				zap.Error(err),
				zap.Float64("estimated_actual_cost", estimatedCost.ActualCost),
				zap.Float64("estimated_total_cost", estimatedCost.TotalCost),
			)
			status, code, message, retryAfter := billingErrorDetails(err)
			if retryAfter > 0 {
				c.Header("Retry-After", strconv.Itoa(retryAfter))
			}
			h.handleStreamingAwareError(c, status, code, message, streamStarted)
			return
		}
	}
	service.SetOpsLatencyMs(c, service.OpsAuthLatencyMsKey, time.Since(requestStart).Milliseconds())

	userReleaseFunc, acquired := h.acquireResponsesUserSlot(c, subject.UserID, subject.Concurrency, false, &streamStarted, reqLog)
	if !acquired {
		return
	}
	if userReleaseFunc != nil {
		defer userReleaseFunc()
	}

	if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription, service.QuotaPlatform(c.Request.Context(), apiKey)); err != nil {
		reqLog.Info("openai.nano_banana.billing_eligibility_check_failed", zap.Error(err))
		status, code, message, retryAfter := billingErrorDetails(err)
		if retryAfter > 0 {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
		}
		h.handleStreamingAwareError(c, status, code, message, streamStarted)
		return
	}
	isSubscriptionBilling := subscription != nil && apiKey.Group != nil && apiKey.Group.IsSubscriptionType()
	var mediaHold *service.OpenAIMediaBalanceHold
	if !isSubscriptionBilling && estimatedCost != nil && estimatedCost.ActualCost > 0 {
		mediaHold, err = h.gatewayService.ReserveOpenAIMediaBalance(c.Request.Context(), apiKey, apiKey.User, estimatedCost.ActualCost, requestPayloadHash)
		if err != nil {
			reqLog.Info("openai.nano_banana.balance_hold_failed",
				zap.Error(err),
				zap.Float64("estimated_actual_cost", estimatedCost.ActualCost),
			)
			status, code, message, retryAfter := billingErrorDetails(err)
			if retryAfter > 0 {
				c.Header("Retry-After", strconv.Itoa(retryAfter))
			}
			h.handleStreamingAwareError(c, status, code, message, streamStarted)
			return
		}
	}
	defer func() {
		if mediaHold == nil {
			return
		}
		if err := h.gatewayService.ReleaseOpenAIMediaBalance(context.Background(), mediaHold); err != nil {
			reqLog.Error("openai.nano_banana.balance_hold_release_failed", zap.Error(err))
		}
	}()

	sessionHash := h.gatewayService.GenerateExplicitSessionHash(c, []byte(parsed.StickySessionSeed()))
	requestCtx := service.WithOpenAIImageGenerationIntent(c.Request.Context())
	routingStart := time.Now()
	maxAccountSwitches := h.maxAccountSwitches
	switchCount := 0
	profitVetoCount := 0
	failedAccountIDs := make(map[int64]struct{})
	var lastFailoverErr *service.UpstreamFailoverError

	for {
		reqLog.Debug("openai.nano_banana.account_selecting", zap.Int("excluded_account_count", len(failedAccountIDs)))
		selection, scheduleDecision, err := h.gatewayService.SelectAccountWithSchedulerForImages(
			requestCtx,
			apiKey.GroupID,
			sessionHash,
			requestModel,
			failedAccountIDs,
			service.OpenAIImagesCapabilityNative,
		)
		if err != nil {
			reqLog.Warn("openai.nano_banana.account_select_failed", zap.Error(err), zap.Int("excluded_account_count", len(failedAccountIDs)))
			if len(failedAccountIDs) == 0 {
				markOpsRoutingCapacityLimitedIfNoAvailable(c, err)
				h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available compatible accounts", streamStarted)
				return
			}
			if lastFailoverErr != nil {
				h.handleFailoverExhausted(c, lastFailoverErr, streamStarted)
			} else {
				h.handleFailoverExhaustedSimple(c, 502, streamStarted)
			}
			return
		}
		if selection == nil || selection.Account == nil {
			markOpsRoutingCapacityLimited(c)
			h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available compatible accounts", streamStarted)
			return
		}

		reqLog.Debug("openai.nano_banana.account_schedule_decision",
			zap.String("layer", scheduleDecision.Layer),
			zap.Bool("sticky_session_hit", scheduleDecision.StickySessionHit),
			zap.Int("candidate_count", scheduleDecision.CandidateCount),
			zap.Int("top_k", scheduleDecision.TopK),
			zap.Int64("latency_ms", scheduleDecision.LatencyMs),
			zap.Float64("load_skew", scheduleDecision.LoadSkew),
		)

		account := selection.Account
		if !service.AccountSupportsNanoBananaEndpoint(account, requestModel) {
			baseURL := strings.TrimSpace(account.GetCredential("base_url"))
			reqLog.Warn("openai.nano_banana.account_skipped",
				zap.Int64("account_id", account.ID),
				zap.String("account_type", string(account.Type)),
				zap.Bool("has_custom_base_url", baseURL != ""),
				zap.String("base_url", service.NanoBananaEndpointURLForBase(baseURL)),
			)
			failedAccountIDs[account.ID] = struct{}{}
			continue
		}
		sessionHash = ensureOpenAIPoolModeSessionHash(sessionHash, account)
		setOpsSelectedAccount(c, account.ID, account.Platform)
		accountReleaseFunc, slotResult := h.acquireResponsesAccountSlot(c, apiKey.GroupID, sessionHash, selection, false, &streamStarted, reqLog)
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
		baseURL := strings.TrimSpace(account.GetCredential("base_url"))
		selectedUpstreamModel := service.NormalizeNanoBananaImageModel(account.GetMappedModel(parsed.Model))
		reqLog.Info("openai.nano_banana.account_selected",
			zap.Int64("account_id", account.ID),
			zap.String("account_name", account.Name),
			zap.String("account_type", string(account.Type)),
			zap.String("base_url", service.NanoBananaEndpointURLForBase(baseURL)),
			zap.String("upstream_model", selectedUpstreamModel),
			zap.String("image_size", parsed.ImageSize),
		)

		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(routingStart).Milliseconds())
		forwardStart := time.Now()
		writerSizeBeforeForward := c.Writer.Size()
		result, err := func() (*service.OpenAIForwardResult, error) {
			defer func() {
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
			}()
			return h.gatewayService.ForwardNanoBanana(requestCtx, c, account, parsed)
		}()
		forwardDurationMs := time.Since(forwardStart).Milliseconds()
		upstreamLatencyMs, _ := getContextInt64(c, service.OpsUpstreamLatencyMsKey)
		responseLatencyMs := forwardDurationMs
		if upstreamLatencyMs > 0 && forwardDurationMs > upstreamLatencyMs {
			responseLatencyMs = forwardDurationMs - upstreamLatencyMs
		}
		service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, responseLatencyMs)

		if err != nil {
			var imageUpstreamErr *service.OpenAIImagesUpstreamError
			if errors.As(err, &imageUpstreamErr) {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account, selectedUpstreamModel, true, nil)
				reqLog.Warn("openai.nano_banana.upstream_user_error",
					zap.Int64("account_id", account.ID),
					zap.Int("status_code", imageUpstreamErr.StatusCode),
					zap.String("error_type", imageUpstreamErr.ErrorType),
					zap.String("error_code", imageUpstreamErr.Code),
					zap.Error(err),
				)
				return
			}
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account, selectedUpstreamModel, false, nil)
				h.gatewayService.RecordOpenAIAccountSwitch()
				failedAccountIDs[account.ID] = struct{}{}
				lastFailoverErr = failoverErr
				if switchCount >= maxAccountSwitches {
					h.handleFailoverExhausted(c, failoverErr, streamStarted)
					return
				}
				switchCount++
				reqLog.Warn("openai.nano_banana.upstream_failover_switching",
					zap.Int64("account_id", account.ID),
					zap.Int("upstream_status", failoverErr.StatusCode),
					zap.Int("switch_count", switchCount),
					zap.Int("max_switches", maxAccountSwitches),
				)
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account, selectedUpstreamModel, false, nil)
			upstreamErrorAlreadyCommunicated := openAIForwardErrorAlreadyCommunicated(c, writerSizeBeforeForward, err)
			wroteFallback := false
			if !upstreamErrorAlreadyCommunicated {
				wroteFallback = h.ensureForwardErrorResponse(c, streamStarted)
			}
			fields := []zap.Field{
				zap.Int64("account_id", account.ID),
				zap.String("account_type", string(account.Type)),
				zap.Bool("fallback_error_response_written", wroteFallback),
				zap.Bool("upstream_error_response_already_written", upstreamErrorAlreadyCommunicated),
				zap.Error(err),
			}
			if shouldLogOpenAIForwardFailureAsWarn(c, wroteFallback) {
				reqLog.Warn("openai.nano_banana.forward_failed", fields...)
				return
			}
			reqLog.Error("openai.nano_banana.forward_failed", fields...)
			return
		}

		if result != nil {
			h.gatewayService.ReportOpenAIAccountScheduleResult(account, selectedUpstreamModel, true, result.FirstTokenMs)
		} else {
			h.gatewayService.ReportOpenAIAccountScheduleResult(account, selectedUpstreamModel, true, nil)
		}

		userAgent := c.GetHeader("User-Agent")
		clientIP := ip.GetClientIP(c)
		inboundEndpoint := GetInboundEndpoint(c)
		upstreamEndpoint := "/v1/api/nano-banana"
		upstreamModel := requestModel
		if result != nil && result.UpstreamModel != "" {
			upstreamModel = result.UpstreamModel
		}
		// Upstream work has completed. Settlement owns the reservation from here;
		// a bookkeeping error must not turn a successful generation into free usage.
		settlementHold := mediaHold
		mediaHold = nil
		if err := h.gatewayService.RecordUsage(c.Request.Context(), &service.OpenAIRecordUsageInput{
			Result:             result,
			APIKey:             apiKey,
			User:               apiKey.User,
			Account:            account,
			Subscription:       subscription,
			InboundEndpoint:    inboundEndpoint,
			UpstreamEndpoint:   upstreamEndpoint,
			UserAgent:          userAgent,
			IPAddress:          clientIP,
			RequestPayloadHash: requestPayloadHash,
			APIKeyService:      h.apiKeyService,
			MediaBalanceHold:   settlementHold,
			ChannelUsageFields: channelMapping.ToUsageFields(requestModel, upstreamModel),
		}); err != nil {
			logger.L().With(
				zap.String("component", "handler.openai_gateway.nano_banana"),
				zap.Int64("user_id", subject.UserID),
				zap.Int64("api_key_id", apiKey.ID),
				zap.Any("group_id", apiKey.GroupID),
				zap.String("model", requestModel),
				zap.Int64("account_id", account.ID),
			).Error("openai.nano_banana.record_usage_failed", zap.Error(err))
		}

		reqLog.Debug("openai.nano_banana.request_completed",
			zap.Int64("account_id", account.ID),
			zap.String("base_url", service.NanoBananaEndpointURLForBase(baseURL)),
			zap.Int("switch_count", switchCount),
		)
		return
	}
}
