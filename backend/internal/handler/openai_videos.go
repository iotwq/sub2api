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
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Videos creates an async OpenAI-compatible video generation task.
// POST /v1/videos
func (h *OpenAIGatewayHandler) Videos(c *gin.Context) {
	h.handleOpenAIVideo(c, service.OpenAIVideoEndpointCreate)
}

// VideoStatus retrieves an async OpenAI-compatible video generation task.
// GET /v1/videos/{task_id}
func (h *OpenAIGatewayHandler) VideoStatus(c *gin.Context) {
	h.handleOpenAIVideo(c, service.OpenAIVideoEndpointStatus)
}

// VideoContent streams generated video content.
// GET /v1/videos/{task_id}/content
func (h *OpenAIGatewayHandler) VideoContent(c *gin.Context) {
	h.handleOpenAIVideo(c, service.OpenAIVideoEndpointContent)
}

func (h *OpenAIGatewayHandler) VideoList(c *gin.Context) {
	h.handleOpenAIVideo(c, service.OpenAIVideoEndpointList)
}

func (h *OpenAIGatewayHandler) VideoDelete(c *gin.Context) {
	h.handleOpenAIVideo(c, service.OpenAIVideoEndpointDelete)
}

func (h *OpenAIGatewayHandler) VideoContextIR(c *gin.Context) {
	h.handleOpenAIVideo(c, service.OpenAIVideoEndpointContext)
}

func (h *OpenAIGatewayHandler) VideoRegeneration(c *gin.Context) {
	h.handleOpenAIVideo(c, service.OpenAIVideoEndpointRegenerate)
}

func (h *OpenAIGatewayHandler) handleOpenAIVideo(c *gin.Context, endpoint service.OpenAIVideoEndpoint) {
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
		"handler.openai_gateway.videos",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
		zap.String("endpoint", string(endpoint)),
	)
	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}

	var body []byte
	var parsed *service.OpenAIVideoCreateRequest
	var err error
	if endpoint.RequiresRequestBody() {
		contentType := c.GetHeader("Content-Type")
		body, err = pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
		if err != nil {
			if maxErr, ok := extractMaxBytesError(err); ok {
				h.errorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
				return
			}
			h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
			return
		}
		parsed, err = service.ParseOpenAIVideoOperationRequestWithContentType(endpoint, body, contentType)
		if err != nil {
			h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", err.Error())
			return
		}
	}

	requestModel := service.OpenAIVideoRequestModel(endpoint, parsed)
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		taskID = strings.TrimSpace(c.Param("request_id"))
	}
	requiresTaskID := endpoint == service.OpenAIVideoEndpointStatus || endpoint == service.OpenAIVideoEndpointContent || endpoint == service.OpenAIVideoEndpointDelete
	if requiresTaskID && taskID == "" {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
		return
	}
	upstreamTaskID := taskID
	var taskRoute *service.OpenAIVideoTaskBinding
	isMiniMaxRecoveryRoute := false
	if requiresTaskID {
		taskRoute, err = h.gatewayService.GetMiniMaxVideoRecovery(c.Request.Context(), apiKey.GroupID, subject.UserID, taskID)
		if err != nil {
			h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Video task recovery state is temporarily unavailable")
			return
		}
		if taskRoute != nil {
			recoveryStatus := strings.TrimSpace(taskRoute.RecoveryStatus)
			isMiniMaxRecoveryRoute = recoveryStatus != "" && recoveryStatus != service.OpenAIVideoRecoveryInactive
			switch recoveryStatus {
			case service.OpenAIVideoRecoveryPending, "submitting", service.OpenAIVideoRecoveryIdentified:
				// A unique upstream task can be queried while billing retries. Its
				// actual generation state must not be hidden behind local processing.
				if strings.TrimSpace(taskRoute.UpstreamTaskID) == "" {
					statusCode := http.StatusOK
					if endpoint == service.OpenAIVideoEndpointContent {
						statusCode = http.StatusAccepted
					}
					c.JSON(statusCode, gin.H{"task": gin.H{
						"id": taskID, "status": "processing", "progress": 0,
					}})
					return
				}
			case service.OpenAIVideoRecoveryBillingReview:
				if strings.TrimSpace(taskRoute.UpstreamTaskID) == "" {
					c.JSON(http.StatusOK, gin.H{"task": gin.H{
						"id": taskID, "status": "failed", "error": gin.H{"message": "Video task submission could not be confirmed"},
					}})
					return
				}
			case service.OpenAIVideoRecoveryAmbiguous, service.OpenAIVideoRecoveryFailed, service.OpenAIVideoRecoveryCancelled:
				message := "MiniMax-H3 task submission could not be confirmed safely"
				if taskRoute.RecoveryLastError != nil && strings.TrimSpace(*taskRoute.RecoveryLastError) != "" {
					message = strings.TrimSpace(*taskRoute.RecoveryLastError)
				}
				c.JSON(http.StatusOK, gin.H{"task": gin.H{
					"id": taskID, "status": "failed", "error": gin.H{"message": message},
				}})
				return
			}
			if strings.TrimSpace(taskRoute.UpstreamTaskID) != "" {
				upstreamTaskID = strings.TrimSpace(taskRoute.UpstreamTaskID)
			}
		}
	}

	boundAccountID := int64(0)
	if taskRoute != nil {
		boundAccountID = taskRoute.AccountID
	}
	reqLog = reqLog.With(
		zap.String("model", requestModel),
		zap.String("task_id", taskID),
		zap.String("upstream_task_id", upstreamTaskID),
		zap.Int64("bound_account_id", boundAccountID),
	)
	setOpsRequestContext(c, requestModel, false, body)
	setOpsEndpointContext(c, "", int16(service.RequestTypeSync))
	channelMapping, _ := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, requestModel)

	if endpoint.CreatesTask() {
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
	}

	if h.errorPassthroughService != nil {
		service.BindErrorPassthroughService(c, h.errorPassthroughService)
	}

	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	service.SetOpsLatencyMs(c, service.OpsAuthLatencyMsKey, time.Since(requestStart).Milliseconds())

	userReleaseFunc, acquired := h.acquireResponsesUserSlot(c, subject.UserID, subject.Concurrency, false, &streamStarted, reqLog)
	if !acquired {
		return
	}
	if userReleaseFunc != nil {
		defer userReleaseFunc()
	}

	requestPayloadHash := ""
	var estimatedCost *service.CostBreakdown
	inputVideoSeconds := 0.0
	billingModel := requestModel
	if channelMapping.BillingModelSource == service.BillingModelSourceChannelMapped && strings.TrimSpace(channelMapping.MappedModel) != "" {
		billingModel = strings.TrimSpace(channelMapping.MappedModel)
	}
	if endpoint.CreatesTask() {
		requestPayloadHash = service.HashUsageRequestPayload(body)
		estimatedCost, inputVideoSeconds, err = h.gatewayService.EstimateOpenAIVideoCreateBillingForModel(c.Request.Context(), apiKey, apiKey.User, requestModel, billingModel, parsed.BillingBody)
		if err != nil {
			reqLog.Warn("openai.videos.estimate_cost_failed", zap.Error(err))
			status, errType, message := pricingPreflightErrorDetails(requestModel, err)
			h.handleStreamingAwareError(c, status, errType, message, streamStarted)
			return
		}
		if estimatedCost != nil && estimatedCost.ActualCost > 0 {
			if err := h.billingCacheService.CheckEstimatedCostCoverage(c.Request.Context(), apiKey.User, apiKey.Group, subscription, estimatedCost.ActualCost); err != nil {
				reqLog.Info("openai.videos.estimated_cost_check_failed",
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
	}
	quotaPlatform := service.QuotaPlatform(c.Request.Context(), apiKey)
	if endpoint.CreatesTask() {
		if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription, quotaPlatform); err != nil {
			reqLog.Info("openai.videos.billing_eligibility_check_failed", zap.Error(err))
			status, code, message, retryAfter := billingErrorDetails(err)
			if retryAfter > 0 {
				c.Header("Retry-After", strconv.Itoa(retryAfter))
			}
			h.handleStreamingAwareError(c, status, code, message, streamStarted)
			return
		}
	}
	isSubscriptionBilling := subscription != nil && apiKey.Group != nil && apiKey.Group.IsSubscriptionType()
	var mediaHold *service.OpenAIMediaBalanceHold
	if endpoint.CreatesTask() && !isSubscriptionBilling && estimatedCost != nil && estimatedCost.ActualCost > 0 {
		mediaHold, err = h.gatewayService.ReserveOpenAIMediaBalance(c.Request.Context(), apiKey, apiKey.User, estimatedCost.ActualCost, requestPayloadHash)
		if err != nil {
			reqLog.Info("openai.videos.balance_hold_failed",
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
			reqLog.Error("openai.videos.balance_hold_release_failed", zap.Error(err))
		}
	}()

	sessionHash := h.gatewayService.GenerateExplicitSessionHash(c, body)
	boundTaskID := taskID
	if parsed != nil && strings.TrimSpace(parsed.SourceTaskID) != "" {
		boundTaskID = strings.TrimSpace(parsed.SourceTaskID)
	}
	if boundTaskID != "" {
		sessionHash = service.OpenAIVideoTaskSessionHash(subject.UserID, boundTaskID)
	}
	requestCtx := c.Request.Context()
	if endpoint.CreatesTask() {
		requestCtx = service.WithOpenAIImageGenerationIntent(requestCtx)
	}
	if taskRoute != nil && taskRoute.AccountID > 0 {
		requestCtx = service.WithRequiredOpenAIAccount(requestCtx, taskRoute.AccountID)
	}
	if boundTaskID != "" {
		accountID, restoreErr := h.gatewayService.RestoreOpenAIVideoTaskStickySession(requestCtx, apiKey.GroupID, subject.UserID, boundTaskID)
		if restoreErr != nil {
			reqLog.Warn("openai.videos.restore_task_binding_failed",
				zap.String("task_id", boundTaskID),
				zap.Error(restoreErr),
			)
		} else if accountID > 0 {
			reqLog.Debug("openai.videos.restore_task_binding_hit",
				zap.String("task_id", boundTaskID),
				zap.Int64("account_id", accountID),
			)
		}
	}
	routingStart := time.Now()
	maxAccountSwitches := h.maxAccountSwitches
	switchCount := 0
	profitVetoCount := 0
	failedAccountIDs := make(map[int64]struct{})
	var lastFailoverErr *service.UpstreamFailoverError

	for {
		var selection *service.AccountSelectionResult
		var scheduleDecision service.OpenAIAccountScheduleDecision
		var err error
		if service.IsMiniMaxOnlyOpenAIVideoEndpoint(endpoint) || service.IsMiniMaxH3VideoModel(requestModel) || isMiniMaxRecoveryRoute {
			schedulingModel := requestModel
			if isMiniMaxRecoveryRoute {
				schedulingModel = "MiniMax-H3"
			}
			selection, scheduleDecision, err = h.gatewayService.SelectAccountWithSchedulerForCapability(
				requestCtx,
				apiKey.GroupID,
				"",
				sessionHash,
				schedulingModel,
				failedAccountIDs,
				service.OpenAIUpstreamTransportHTTPSSE,
				service.OpenAIEndpointCapabilityMiniMaxVideo,
				false,
				false,
				false,
			)
		} else {
			selection, scheduleDecision, err = h.gatewayService.SelectAccountWithSchedulerForImages(
				requestCtx,
				apiKey.GroupID,
				sessionHash,
				requestModel,
				failedAccountIDs,
				service.OpenAIImagesCapabilityNative,
			)
		}
		if err != nil {
			reqLog.Warn("openai.videos.account_select_failed", zap.Error(err), zap.Int("excluded_account_count", len(failedAccountIDs)))
			if len(failedAccountIDs) == 0 {
				cls := classifyNoAccountErrorFromGin(c, h.gatewayService, apiKey, requestModel, requestModel, service.PlatformOpenAI)
				if !cls.ModelNotFound {
					markOpsRoutingCapacityLimitedIfNoAvailable(c, err)
				}
				message := cls.Message
				if !cls.ModelNotFound {
					message = "No available compatible accounts"
				}
				h.handleStreamingAwareError(c, cls.Status, cls.ErrType, message, streamStarted)
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

		reqLog.Debug("openai.videos.account_schedule_decision",
			zap.String("layer", scheduleDecision.Layer),
			zap.Bool("sticky_session_hit", scheduleDecision.StickySessionHit),
			zap.Int("candidate_count", scheduleDecision.CandidateCount),
			zap.Int("top_k", scheduleDecision.TopK),
			zap.Int64("latency_ms", scheduleDecision.LatencyMs),
			zap.Float64("load_skew", scheduleDecision.LoadSkew),
		)

		account := selection.Account
		selectedUpstreamModel := account.GetMappedModel(requestModel)
		if endpoint.CreatesTask() && !service.AccountSupportsOpenAIVideoEndpoint(account, requestModel) {
			reqLog.Warn("openai.videos.account_skipped",
				zap.Int64("account_id", account.ID),
				zap.String("account_type", string(account.Type)),
				zap.String("base_url", service.OpenAIVideoEndpointURLForBase(account.GetCredential("base_url"), endpoint, taskID)),
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

		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(routingStart).Milliseconds())
		forwardStart := time.Now()
		writerSizeBeforeForward := c.Writer.Size()
		var recoveryTask *service.MiniMaxVideoRecoveryTask
		result, err := func() (*service.OpenAIForwardResult, error) {
			defer func() {
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
			}()
			forwardClientTaskID := taskID
			if endpoint == service.OpenAIVideoEndpointCreate && service.IsMiniMaxH3VideoModel(selectedUpstreamModel) && service.AccountUsesMiniMaxV2Video(account) {
				subscriptionID := int64(0)
				if subscription != nil {
					subscriptionID = subscription.ID
				}
				videoResolution, videoDurationSeconds := service.OpenAIVideoBillingParameters(requestModel, parsed.BillingBody)
				recoveryTask, err = h.gatewayService.PrepareMiniMaxVideoRecovery(
					requestCtx,
					apiKey.GroupID,
					subject.UserID,
					apiKey,
					account,
					body,
					service.OpenAIVideoRecoveryBillingSnapshot{
						Cost:                      estimatedCost,
						Hold:                      mediaHold,
						RequestModel:              requestModel,
						UpstreamModel:             selectedUpstreamModel,
						BillingModel:              billingModel,
						VideoResolution:           videoResolution,
						VideoDurationSeconds:      videoDurationSeconds,
						VideoInputDurationSeconds: inputVideoSeconds,
						RequestPayloadHash:        requestPayloadHash,
						SubscriptionID:            subscriptionID,
						InboundEndpoint:           GetInboundEndpoint(c),
						UpstreamEndpoint:          "/v2/video_generation",
						UserAgent:                 c.GetHeader("User-Agent"),
						IPAddress:                 ip.GetClientIP(c),
						QuotaPlatform:             quotaPlatform,
						PricingAt:                 requestStart,
						ChannelUsageFields:        channelMapping.ToUsageFields(requestModel, selectedUpstreamModel),
					},
				)
				if err != nil {
					return nil, err
				}
				forwardClientTaskID = recoveryTask.Binding.TaskID
			}
			return h.gatewayService.ForwardOpenAIVideo(requestCtx, c, account, endpoint, upstreamTaskID, forwardClientTaskID, body, c.GetHeader("Content-Type"), requestModel)
		}()
		forwardDurationMs := time.Since(forwardStart).Milliseconds()
		upstreamLatencyMs, _ := getContextInt64(c, service.OpsUpstreamLatencyMsKey)
		responseLatencyMs := forwardDurationMs
		if upstreamLatencyMs > 0 && forwardDurationMs > upstreamLatencyMs {
			responseLatencyMs = forwardDurationMs - upstreamLatencyMs
		}
		service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, responseLatencyMs)

		terminalFailure := result != nil && !result.SubmissionUncertain && (result.TaskTerminalFailure || service.OpenAIVideoStatusFailed(result.ResponseBody))
		billingUnresolved := taskRoute != nil && (taskRoute.RecoveryStatus == service.OpenAIVideoRecoveryIdentified || taskRoute.RecoveryStatus == service.OpenAIVideoRecoveryBillingReview)
		if !billingUnresolved && (endpoint == service.OpenAIVideoEndpointStatus || endpoint == service.OpenAIVideoEndpointContent || endpoint == service.OpenAIVideoEndpointDelete) && terminalFailure {
			refundTaskID := strings.TrimSpace(result.ResponseID)
			if refundTaskID == "" {
				refundTaskID = taskID
			}
			refundBillingTaskID := refundTaskID
			if taskRoute != nil && strings.TrimSpace(taskRoute.BillingTaskID) != "" {
				refundBillingTaskID = strings.TrimSpace(taskRoute.BillingTaskID)
			}
			h.submitMandatoryUsageRecordTask(c.Request.Context(), func(ctx context.Context) {
				if err := h.gatewayService.RefundFailedOpenAIVideoTaskWithBillingID(ctx, apiKey, refundTaskID, refundBillingTaskID, h.apiKeyService); err != nil {
					reqLog.Warn("openai.videos.refund_failed_task_failed",
						zap.Int64("account_id", account.ID),
						zap.String("task_id", refundTaskID),
						zap.Error(err),
					)
					return
				}
				if err := h.gatewayService.CompleteOpenAIVideoTaskCompensation(ctx, apiKey.GroupID, subject.UserID, refundTaskID, "refunded"); err != nil {
					reqLog.Warn("openai.videos.mark_refunded_task_failed",
						zap.String("task_id", refundTaskID),
						zap.Error(err),
					)
				}
			})
		} else if !billingUnresolved && endpoint == service.OpenAIVideoEndpointStatus && result != nil && service.OpenAIVideoStatusSucceeded(result.ResponseBody) {
			completedTaskID := strings.TrimSpace(result.ResponseID)
			if completedTaskID == "" {
				completedTaskID = taskID
			}
			if err := h.gatewayService.CompleteOpenAIVideoTaskCompensation(requestCtx, apiKey.GroupID, subject.UserID, completedTaskID, "completed"); err != nil {
				reqLog.Warn("openai.videos.mark_completed_task_failed",
					zap.String("task_id", completedTaskID),
					zap.Error(err),
				)
			}
		}

		if err != nil {
			if recoveryTask != nil {
				if recoveryErr := h.gatewayService.MarkMiniMaxVideoRecovery(requestCtx, recoveryTask, service.OpenAIVideoRecoveryCancelled, "", err.Error()); recoveryErr != nil {
					reqLog.Warn("openai.videos.cancel_submission_recovery_failed", zap.Error(recoveryErr))
				}
			}
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) && endpoint.CreatesTask() {
				h.gatewayService.ReportOpenAIAccountScheduleResult(account, selectedUpstreamModel, false, nil)
				h.gatewayService.RecordOpenAIAccountSwitch()
				failedAccountIDs[account.ID] = struct{}{}
				lastFailoverErr = failoverErr
				if switchCount >= maxAccountSwitches {
					h.handleFailoverExhausted(c, failoverErr, streamStarted)
					return
				}
				switchCount++
				reqLog.Warn("openai.videos.upstream_failover_switching",
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
				reqLog.Warn("openai.videos.forward_failed", fields...)
				return
			}
			reqLog.Error("openai.videos.forward_failed", fields...)
			return
		}

		h.gatewayService.ReportOpenAIAccountScheduleResult(account, selectedUpstreamModel, !terminalFailure, nil)
		if endpoint.CreatesTask() && result != nil && strings.TrimSpace(result.ResponseID) != "" {
			if terminalFailure {
				if recoveryTask != nil {
					_ = h.gatewayService.MarkMiniMaxVideoRecovery(requestCtx, recoveryTask, service.OpenAIVideoRecoveryCancelled, "", "upstream returned a terminal failure while creating the task")
				}
				reqLog.Info("openai.videos.create_terminal_failure_not_billed",
					zap.Int64("account_id", account.ID),
					zap.String("task_id", result.ResponseID),
				)
				return
			}
			billingTaskID := strings.TrimSpace(result.BillingTaskID)
			if billingTaskID == "" {
				billingTaskID = strings.TrimSpace(result.ResponseID)
			}
			if recoveryTask != nil {
				recoveryStatus := service.OpenAIVideoRecoveryIdentified
				upstreamID := result.ResponseID
				if result.SubmissionUncertain {
					recoveryStatus = service.OpenAIVideoRecoveryPending
					upstreamID = ""
				}
				if recoveryErr := h.gatewayService.MarkMiniMaxVideoRecovery(requestCtx, recoveryTask, recoveryStatus, upstreamID, ""); recoveryErr != nil {
					reqLog.Warn("openai.videos.mark_submission_recovery_failed", zap.Error(recoveryErr))
				}
			}

			bindErr := error(nil)
			if result.SubmissionUncertain {
				bindErr = h.gatewayService.BindOpenAIVideoTaskAccount(requestCtx, apiKey.GroupID, subject.UserID, result.ResponseID, account.ID)
			} else {
				bindErr = h.gatewayService.BindOpenAIVideoTaskRoute(requestCtx, apiKey.GroupID, subject.UserID, result.ResponseID, result.ResponseID, billingTaskID, account.ID)
			}
			if bindErr != nil {
				reqLog.Warn("openai.videos.bind_task_account_failed",
					zap.Int64("account_id", account.ID),
					zap.String("task_id", result.ResponseID),
					zap.Error(bindErr),
				)
			}
			if !result.SubmissionUncertain {
				if err := h.gatewayService.ArmOpenAIVideoTaskCompensation(requestCtx, apiKey.GroupID, subject.UserID, result.ResponseID, apiKey); err != nil {
					reqLog.Error("openai.videos.arm_task_compensation_failed",
						zap.Int64("account_id", account.ID),
						zap.String("task_id", result.ResponseID),
						zap.Error(err),
					)
					return
				}
			}
			userAgent := c.GetHeader("User-Agent")
			clientIP := ip.GetClientIP(c)
			inboundEndpoint := GetInboundEndpoint(c)
			upstreamEndpoint := strings.TrimSpace(result.UpstreamEndpoint)
			if upstreamEndpoint == "" {
				upstreamEndpoint = service.OpenAIVideoUpstreamEndpointPath(endpoint, result.UpstreamModel)
			}
			billingResult := *result
			billingResult.RequestID = service.OpenAIVideoUsageRequestID(billingTaskID)
			billingResult.UseResultRequestID = true
			billingResult.RequestCount = 1
			billingResult.MediaType = "video"
			billingResult.BillingModel = billingModel
			billingResult.VideoInputDurationSeconds = inputVideoSeconds
			usageRecorded := false
			if result.SubmissionUncertain {
				// The recovery snapshot owns the hold until a unique upstream task is
				// identified. Do not create a charge for an unconfirmed submission.
				mediaHold = nil
			} else {
				if err := h.gatewayService.RecordUsage(c.Request.Context(), &service.OpenAIRecordUsageInput{
					Result:              &billingResult,
					APIKey:              apiKey,
					User:                apiKey.User,
					Account:             account,
					Subscription:        subscription,
					InboundEndpoint:     inboundEndpoint,
					UpstreamEndpoint:    upstreamEndpoint,
					UserAgent:           userAgent,
					IPAddress:           clientIP,
					RequestPayloadHash:  requestPayloadHash,
					APIKeyService:       h.apiKeyService,
					QuotaPlatform:       quotaPlatform,
					PricingAt:           requestStart,
					MediaBalanceHold:    mediaHold,
					BillingCostSnapshot: estimatedCost,
					ChannelUsageFields:  channelMapping.ToUsageFields(requestModel, result.UpstreamModel),
				}); err != nil {
					reqLog.Error("openai.videos.record_usage_failed",
						zap.Int64("account_id", account.ID),
						zap.String("task_id", result.ResponseID),
						zap.Error(err),
					)
					if recoveryTask == nil {
						if completeErr := h.gatewayService.CompleteOpenAIVideoTaskCompensation(requestCtx, apiKey.GroupID, subject.UserID, result.ResponseID, "completed"); completeErr != nil {
							reqLog.Warn("openai.videos.disarm_task_compensation_failed",
								zap.String("task_id", result.ResponseID),
								zap.Error(completeErr),
							)
						}
					}
					if recoveryTask != nil {
						mediaHold = nil
					}
				} else {
					usageRecorded = true
					if mediaHold != nil {
						mediaHold = nil
					}
				}
			}
			if recoveryTask != nil && !result.SubmissionUncertain && usageRecorded {
				if recoveryErr := h.gatewayService.MarkMiniMaxVideoRecovery(requestCtx, recoveryTask, service.OpenAIVideoRecoveryMatched, result.ResponseID, ""); recoveryErr != nil {
					reqLog.Warn("openai.videos.complete_submission_recovery_failed", zap.Error(recoveryErr))
				}
			}
			if result.SubmissionUncertain && !c.Writer.Written() {
				c.JSON(http.StatusOK, gin.H{
					"task_id": result.ResponseID,
					"status":  "processing",
				})
			}
		}
		reqLog.Debug("openai.videos.request_completed",
			zap.Int64("account_id", account.ID),
			zap.String("task_id", taskID),
			zap.Int("switch_count", switchCount),
		)
		return
	}
}
