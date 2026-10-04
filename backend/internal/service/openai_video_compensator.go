package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

type OpenAIVideoRecoveryAPIKeyService interface {
	APIKeyQuotaUpdater
	GetByID(ctx context.Context, id int64) (*APIKey, error)
}

const (
	openAIVideoCompensationPending   = "pending"
	openAIVideoCompensationCompleted = "completed"
	openAIVideoCompensationRefunded  = "refunded"

	openAIVideoCompensationInterval     = 30 * time.Second
	openAIVideoCompensationInitialDelay = 30 * time.Second
	openAIVideoCompensationLease        = 2 * time.Minute
	openAIVideoCompensationTimeout      = 20 * time.Second
	openAIVideoCompensationBatchSize    = 10
)

type openAIVideoTaskState int

const (
	openAIVideoTaskPending openAIVideoTaskState = iota
	openAIVideoTaskSucceeded
	openAIVideoTaskFailed
)

func (s *OpenAIGatewayService) StartOpenAIVideoFailureCompensator(apiKeyService OpenAIVideoRecoveryAPIKeyService) {
	if s == nil || s.openAIVideoTaskBindingRepo == nil || s.httpUpstream == nil || s.accountRepo == nil {
		return
	}
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		return
	}
	s.openAIVideoCompensatorStartOnce.Do(func() {
		s.openAIVideoCompensatorAPIKeyService = apiKeyService
		s.openAIVideoCompensatorStopCh = make(chan struct{})
		s.openAIVideoCompensatorWG.Add(1)
		go s.runOpenAIVideoFailureCompensator()
	})
}

func (s *OpenAIGatewayService) StopOpenAIVideoFailureCompensator() {
	if s == nil || s.openAIVideoCompensatorStopCh == nil {
		return
	}
	s.openAIVideoCompensatorStopOnce.Do(func() {
		close(s.openAIVideoCompensatorStopCh)
		s.openAIVideoCompensatorWG.Wait()
	})
}

func (s *OpenAIGatewayService) ArmOpenAIVideoTaskCompensation(ctx context.Context, groupID *int64, userID int64, taskID string, apiKey *APIKey) error {
	if s == nil || s.openAIVideoTaskBindingRepo == nil || apiKey == nil {
		return nil
	}
	return s.openAIVideoTaskBindingRepo.ArmOpenAIVideoTaskCompensation(
		ctx,
		derefGroupID(groupID),
		userID,
		taskID,
		apiKey,
		time.Now().Add(openAIVideoCompensationInitialDelay),
	)
}

func (s *OpenAIGatewayService) CompleteOpenAIVideoTaskCompensation(ctx context.Context, groupID *int64, userID int64, taskID, status string) error {
	if s == nil || s.openAIVideoTaskBindingRepo == nil {
		return nil
	}
	binding, err := s.openAIVideoTaskBindingRepo.GetOpenAIVideoTaskBinding(ctx, derefGroupID(groupID), userID, taskID)
	if err != nil || binding == nil {
		return err
	}
	return s.openAIVideoTaskBindingRepo.UpdateOpenAIVideoTaskCompensation(ctx, binding.ID, status, nil, "")
}

func (s *OpenAIGatewayService) runOpenAIVideoFailureCompensator() {
	defer s.openAIVideoCompensatorWG.Done()
	ticker := time.NewTicker(openAIVideoCompensationInterval)
	defer ticker.Stop()

	s.reconcileOpenAIVideoTasks()
	for {
		select {
		case <-ticker.C:
			s.reconcileOpenAIVideoTasks()
		case <-s.openAIVideoCompensatorStopCh:
			return
		}
	}
}

func (s *OpenAIGatewayService) reconcileOpenAIVideoTasks() {
	s.reconcilePendingImageSettlements()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
	bindings, err := s.openAIVideoTaskBindingRepo.ClaimOpenAIVideoTaskCompensations(
		ctx,
		now,
		now.Add(openAIVideoCompensationLease),
		openAIVideoCompensationBatchSize,
	)
	cancel()
	if err != nil {
		logger.LegacyPrintf("service.openai_video_compensator", "claim failed: %v", err)
	} else {
		for i := range bindings {
			s.reconcileOpenAIVideoTask(bindings[i])
		}
	}
	s.reconcileMiniMaxVideoRecoveries(now)
}

func (s *OpenAIGatewayService) reconcileMiniMaxVideoRecoveries(now time.Time) {
	repo, ok := s.openAIVideoTaskBindingRepo.(OpenAIVideoRecoveryRepository)
	if !ok || repo == nil || s.openAIVideoCompensatorAPIKeyService == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
	bindings, err := repo.ClaimOpenAIVideoRecoveries(ctx, now, now.Add(openAIVideoRecoveryLease), openAIVideoRecoveryBatchSize)
	cancel()
	if err != nil {
		logger.LegacyPrintf("service.openai_video_compensator", "claim MiniMax-H3 recoveries failed: %v", err)
		return
	}
	for i := range bindings {
		s.reconcileMiniMaxVideoRecovery(bindings[i])
	}
}

func (s *OpenAIGatewayService) reconcileMiniMaxVideoRecovery(binding OpenAIVideoTaskBinding) {
	ctx, cancel := context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
	defer cancel()
	repo, ok := s.openAIVideoTaskBindingRepo.(OpenAIVideoRecoveryRepository)
	if !ok || repo == nil {
		return
	}

	upstreamTaskID := strings.TrimSpace(binding.UpstreamTaskID)
	if upstreamTaskID == "" && (binding.RecoveryExpiresAt == nil || !time.Now().Before(*binding.RecoveryExpiresAt)) {
		s.failMiniMaxVideoRecovery(ctx, repo, binding, openAIVideoRecoveryFailed, "MiniMax-H3 task submission could not be confirmed before the recovery window expired")
		return
	}

	if upstreamTaskID == "" {
		account, err := s.accountRepo.GetByID(ctx, binding.AccountID)
		if err != nil || account == nil || !AccountUsesMiniMaxV2Video(account) {
			s.rescheduleMiniMaxVideoRecovery(repo, binding, firstNonNilError(err, errors.New("MiniMax-H3 recovery account is unavailable")))
			return
		}
		var baseline []string
		var signature MiniMaxVideoRecoverySignature
		if err := json.Unmarshal(binding.RecoveryBaseline, &baseline); err != nil {
			s.failMiniMaxVideoRecovery(ctx, repo, binding, openAIVideoRecoveryFailed, "MiniMax-H3 recovery baseline is invalid")
			return
		}
		if err := json.Unmarshal(binding.RecoverySignature, &signature); err != nil {
			s.failMiniMaxVideoRecovery(ctx, repo, binding, openAIVideoRecoveryFailed, "MiniMax-H3 recovery signature is invalid")
			return
		}
		tasks, err := s.fetchMiniMaxVideoTaskList(ctx, account)
		if err != nil {
			s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
			return
		}
		candidates := findMiniMaxVideoRecoveryCandidates(baseline, signature, tasks)
		switch len(candidates) {
		case 0:
			s.rescheduleMiniMaxVideoRecovery(repo, binding, nil)
			return
		case 1:
			upstreamTaskID = candidates[0].ID
			nextCheckAt := time.Now().Add(openAIVideoRecoveryRetryDelay)
			if err := repo.UpdateOpenAIVideoRecovery(ctx, binding.ID, openAIVideoRecoveryIdentified, upstreamTaskID, &nextCheckAt, ""); err != nil {
				s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
				return
			}
			binding.UpstreamTaskID = upstreamTaskID
			binding.RecoveryStatus = openAIVideoRecoveryIdentified
		case 2:
			fallthrough
		default:
			s.failMiniMaxVideoRecovery(ctx, repo, binding, openAIVideoRecoveryAmbiguous, "MiniMax-H3 task was submitted, but multiple matching upstream tasks were found; automatic association was stopped to protect task ownership")
			return
		}
	}

	if err := s.ensureMiniMaxVideoRecoveryBilled(ctx, binding); err != nil {
		s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
		return
	}
	apiKey, err := s.openAIVideoCompensatorAPIKeyService.GetByID(ctx, binding.APIKeyID)
	if err != nil || apiKey == nil {
		s.rescheduleMiniMaxVideoRecovery(repo, binding, firstNonNilError(err, errors.New("MiniMax-H3 recovery API key is unavailable")))
		return
	}
	if err := s.ArmOpenAIVideoTaskCompensation(ctx, videoCompensationGroupIDPtr(binding.GroupID), binding.UserID, binding.TaskID, apiKey); err != nil {
		s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
		return
	}
	if err := repo.UpdateOpenAIVideoRecovery(ctx, binding.ID, openAIVideoRecoveryMatched, upstreamTaskID, nil, ""); err != nil {
		logger.LegacyPrintf("service.openai_video_compensator", "complete MiniMax-H3 recovery failed: %v", err)
		s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
	}
}

func (s *OpenAIGatewayService) ensureMiniMaxVideoRecoveryBilled(ctx context.Context, binding OpenAIVideoTaskBinding) error {
	requestID := OpenAIVideoUsageRequestID(firstNonEmptyString(binding.BillingTaskID, binding.TaskID))
	if existing, err := s.usageLogRepo.GetByRequestIDAndAPIKey(ctx, requestID, binding.APIKeyID); err == nil && existing != nil {
		return nil
	} else if err != nil && !errors.Is(err, ErrUsageLogNotFound) {
		return err
	}

	var snapshot OpenAIVideoRecoveryBillingSnapshot
	if err := json.Unmarshal(binding.RecoveryBilling, &snapshot); err != nil {
		return fmt.Errorf("decode MiniMax-H3 recovery billing snapshot: %w", err)
	}
	if snapshot.Cost == nil {
		return fmt.Errorf("MiniMax-H3 recovery has no saved price: %w", ErrUsageBillingRequestConflict)
	}
	apiKey, err := s.openAIVideoCompensatorAPIKeyService.GetByID(ctx, binding.APIKeyID)
	if err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, binding.UserID)
	if err != nil {
		return err
	}
	account, err := s.accountRepo.GetByID(ctx, binding.AccountID)
	if err != nil {
		return err
	}
	var subscription *UserSubscription
	if snapshot.SubscriptionID > 0 {
		subscription, err = s.userSubRepo.GetByID(ctx, snapshot.SubscriptionID)
		if err != nil {
			return err
		}
	}
	result := &OpenAIForwardResult{
		RequestID:                 requestID,
		ResponseID:                binding.TaskID,
		BillingTaskID:             firstNonEmptyString(binding.BillingTaskID, binding.TaskID),
		Model:                     snapshot.RequestModel,
		UpstreamModel:             snapshot.UpstreamModel,
		UpstreamEndpoint:          snapshot.UpstreamEndpoint,
		BillingModel:              snapshot.BillingModel,
		UseResultRequestID:        true,
		RequestCount:              1,
		MediaType:                 "video",
		VideoCount:                1,
		VideoResolution:           snapshot.VideoResolution,
		VideoDurationSeconds:      snapshot.VideoDurationSeconds,
		VideoInputDurationSeconds: snapshot.VideoInputDurationSeconds,
	}
	if err := s.RecordUsage(ctx, &OpenAIRecordUsageInput{
		Result:              result,
		APIKey:              apiKey,
		User:                user,
		Account:             account,
		Subscription:        subscription,
		InboundEndpoint:     snapshot.InboundEndpoint,
		UpstreamEndpoint:    snapshot.UpstreamEndpoint,
		UserAgent:           snapshot.UserAgent,
		IPAddress:           snapshot.IPAddress,
		RequestPayloadHash:  snapshot.RequestPayloadHash,
		APIKeyService:       s.openAIVideoCompensatorAPIKeyService,
		QuotaPlatform:       snapshot.QuotaPlatform,
		PricingAt:           snapshot.PricingAt,
		MediaBalanceHold:    snapshot.Hold,
		BillingCostSnapshot: snapshot.Cost,
		ChannelUsageFields:  snapshot.ChannelUsageFields,
	}); err != nil {
		return err
	}
	return nil
}

func (s *OpenAIGatewayService) failMiniMaxVideoRecovery(
	ctx context.Context,
	repo OpenAIVideoRecoveryRepository,
	binding OpenAIVideoTaskBinding,
	status string,
	message string,
) {
	requestID := OpenAIVideoUsageRequestID(firstNonEmptyString(binding.BillingTaskID, binding.TaskID))
	if existing, err := s.usageLogRepo.GetByRequestIDAndAPIKey(ctx, requestID, binding.APIKeyID); err == nil && existing != nil {
		apiKey := &APIKey{ID: binding.APIKeyID}
		if binding.APIKeyQuotaLimited {
			apiKey.Quota = 1
		}
		if binding.APIKeyRateLimited {
			apiKey.RateLimit5h = 1
		}
		if err := s.RefundFailedOpenAIVideoTaskWithBillingID(ctx, apiKey, binding.TaskID, binding.BillingTaskID, s.openAIVideoCompensatorAPIKeyService); err != nil {
			s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
			return
		}
	} else if err != nil && !errors.Is(err, ErrUsageLogNotFound) {
		s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
		return
	} else {
		var snapshot OpenAIVideoRecoveryBillingSnapshot
		if err := json.Unmarshal(binding.RecoveryBilling, &snapshot); err != nil {
			s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
			return
		}
		if err := s.ReleaseOpenAIMediaBalance(ctx, snapshot.Hold); err != nil {
			s.rescheduleMiniMaxVideoRecovery(repo, binding, err)
			return
		}
	}
	if err := repo.UpdateOpenAIVideoRecovery(ctx, binding.ID, status, binding.UpstreamTaskID, nil, message); err != nil {
		logger.LegacyPrintf("service.openai_video_compensator", "fail MiniMax-H3 recovery update failed: %v", err)
	}
}

func (s *OpenAIGatewayService) rescheduleMiniMaxVideoRecovery(repo OpenAIVideoRecoveryRepository, binding OpenAIVideoTaskBinding, cause error) {
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	ctx, cancel := context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
	defer cancel()
	status := openAIVideoRecoveryPending
	if strings.TrimSpace(binding.UpstreamTaskID) != "" {
		status = openAIVideoRecoveryIdentified
	}
	next := time.Now().Add(openAIVideoRecoveryRetryDelay)
	nextCheckAt := &next
	if errors.Is(cause, ErrMediaBalanceHoldInconsistent) || errors.Is(cause, ErrUsageBillingRequestConflict) {
		// Retain the task owner and billing evidence. Retrying an inconsistent
		// reservation cannot repair it and must never consume another hold.
		status, nextCheckAt = openAIVideoRecoveryBillingReview, nil
		logger.LegacyPrintf("service.openai_video_compensator", "billing review required task=%s upstream=%s: %v", binding.TaskID, binding.UpstreamTaskID, cause)
	}
	if err := repo.UpdateOpenAIVideoRecovery(ctx, binding.ID, status, binding.UpstreamTaskID, nextCheckAt, message); err != nil {
		logger.LegacyPrintf("service.openai_video_compensator", "reschedule MiniMax-H3 recovery failed: %v", err)
	}
}

func videoCompensationGroupIDPtr(groupID int64) *int64 {
	if groupID <= 0 {
		return nil
	}
	return &groupID
}

func firstNonNilError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

func (s *OpenAIGatewayService) reconcileOpenAIVideoTask(binding OpenAIVideoTaskBinding) {
	ctx, cancel := context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
	defer cancel()

	state, err := s.probeOpenAIVideoTaskState(ctx, &binding)
	if err != nil {
		s.rescheduleOpenAIVideoTask(ctx, binding, err)
		return
	}
	switch state {
	case openAIVideoTaskSucceeded:
		_ = s.openAIVideoTaskBindingRepo.UpdateOpenAIVideoTaskCompensation(ctx, binding.ID, openAIVideoCompensationCompleted, nil, "")
	case openAIVideoTaskFailed:
		apiKey := &APIKey{ID: binding.APIKeyID, Key: binding.APIKeyKey}
		if binding.APIKeyQuotaLimited {
			apiKey.Quota = 1
		}
		if binding.APIKeyRateLimited {
			apiKey.RateLimit5h = 1
		}
		if err := s.RefundFailedOpenAIVideoTaskWithBillingID(ctx, apiKey, binding.TaskID, binding.BillingTaskID, s.openAIVideoCompensatorAPIKeyService); err != nil {
			s.rescheduleOpenAIVideoTask(ctx, binding, err)
			return
		}
		_ = s.openAIVideoTaskBindingRepo.UpdateOpenAIVideoTaskCompensation(ctx, binding.ID, openAIVideoCompensationRefunded, nil, "")
	default:
		s.rescheduleOpenAIVideoTask(ctx, binding, nil)
	}
}

func (s *OpenAIGatewayService) rescheduleOpenAIVideoTask(ctx context.Context, binding OpenAIVideoTaskBinding, cause error) {
	delay := openAIVideoCompensationInitialDelay
	for attempt := 1; attempt < binding.CheckAttempts && delay < 15*time.Minute; attempt++ {
		delay *= 2
	}
	if delay > 15*time.Minute {
		delay = 15 * time.Minute
	}
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	nextCheckAt := time.Now().Add(delay)
	_ = s.openAIVideoTaskBindingRepo.UpdateOpenAIVideoTaskCompensation(ctx, binding.ID, openAIVideoCompensationPending, &nextCheckAt, message)
}

func (s *OpenAIGatewayService) probeOpenAIVideoTaskState(ctx context.Context, binding *OpenAIVideoTaskBinding) (openAIVideoTaskState, error) {
	account, err := s.accountRepo.GetByID(ctx, binding.AccountID)
	if err != nil {
		return openAIVideoTaskPending, err
	}
	if account == nil || account.Type != AccountTypeAPIKey || !account.IsOpenAI() {
		return openAIVideoTaskPending, fmt.Errorf("video compensation account is unavailable")
	}
	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return openAIVideoTaskPending, err
	}
	baseURL, err := s.validateUpstreamBaseURL(account.GetOpenAIBaseURL())
	if err != nil {
		return openAIVideoTaskPending, err
	}
	upstreamTaskID := strings.TrimSpace(binding.UpstreamTaskID)
	if upstreamTaskID == "" {
		upstreamTaskID = binding.TaskID
	}
	targetURL := buildOpenAIVideoEndpointURLForUpstream(baseURL, OpenAIVideoEndpointStatus, upstreamTaskID)
	if AccountUsesMiniMaxV2Video(account) {
		targetURL = buildMiniMaxV2VideoEndpointURL(baseURL, OpenAIVideoEndpointStatus, upstreamTaskID)
	}
	req, err := http.NewRequestWithContext(WithHTTPUpstreamProfile(ctx, HTTPUpstreamProfileOpenAI), http.MethodGet, targetURL, nil)
	if err != nil {
		return openAIVideoTaskPending, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	account.ApplyHeaderOverrides(req.Header)
	if userAgent := account.GetOpenAIUserAgent(); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return openAIVideoTaskPending, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return openAIVideoTaskPending, err
	}
	if OpenAIVideoStatusFailed(body) {
		return openAIVideoTaskFailed, nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return openAIVideoTaskPending, fmt.Errorf("video status probe returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if OpenAIVideoStatusSucceeded(body) {
		return openAIVideoTaskSucceeded, nil
	}
	return openAIVideoTaskPending, nil
}
