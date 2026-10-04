package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// ExcelBPSRateLimitedReason identifies a Basispoints 429 that can be retried
// on another OAuth account without changing the account's Codex quota state.
const ExcelBPSRateLimitedReason = GatewayFailureReason("basispoints_rate_limited")

const (
	excelBPSRateLimitedFilterReason  = "excel_bps_rate_limited"
	excelBPSRateLimitedClientMessage = "Excel BPS rate limit exceeded, please retry later"
)

func newExcelBPSRateLimitedFailoverError(retryAfter string) *UpstreamFailoverError {
	err := &UpstreamFailoverError{
		StatusCode:        http.StatusTooManyRequests,
		Stage:             GatewayFailureStageInference,
		Scope:             GatewayFailureScopeAccount,
		Reason:            ExcelBPSRateLimitedReason,
		NextAccountAction: NextAccountRetry,
		ClientStatusCode:  http.StatusTooManyRequests,
		ClientMessage:     excelBPSRateLimitedClientMessage,
	}
	if _, ok := excelBPSRetryAfter(retryAfter, time.Now()); ok {
		err.ResponseHeaders = http.Header{"Retry-After": {strings.TrimSpace(retryAfter)}}
	}
	return err
}

// coolDownExcelBPS keeps a 429'd account out of the BPS route only. It never
// writes Codex quota fields or the account's global scheduler cooldown.
func (s *OpenAIGatewayService) coolDownExcelBPS(ctx context.Context, account *Account, retryAfter string) {
	if s == nil || account == nil {
		return
	}
	cooldown, ok := excelBPSRetryAfter(retryAfter, time.Now())
	if !ok {
		cooldown, ok = s.excelBPS429FallbackCooldown(ctx, account)
		if !ok {
			return
		}
	}
	if cooldown < time.Second {
		cooldown = time.Second
	}
	maxCooldown := time.Duration(maxRateLimit429CooldownSeconds) * time.Second
	if cooldown > maxCooldown {
		cooldown = maxCooldown
	}
	until := time.Now().Add(cooldown)
	for {
		current, loaded := s.excelBPSCooldownUntil.LoadOrStore(account.ID, until)
		if !loaded {
			break
		}
		currentUntil, valid := current.(time.Time)
		if valid && !until.After(currentUntil) {
			return
		}
		if s.excelBPSCooldownUntil.CompareAndSwap(account.ID, current, until) {
			break
		}
	}
	logger.LegacyPrintf("service.openai_basispoints", "rate limited; skipping account for BPS requests: account_id=%d cooldown=%s", account.ID, cooldown)
}

func (s *OpenAIGatewayService) excelBPS429FallbackCooldown(ctx context.Context, account *Account) (time.Duration, bool) {
	if s.rateLimitService == nil {
		return time.Duration(defaultRateLimit429CooldownSeconds) * time.Second, true
	}
	stateCtx, cancel := openAIAccountStateContext(ctx)
	defer cancel()
	return s.rateLimitService.get429FallbackCooldown(stateCtx, account)
}

// excelBPSRetryAfter accepts both delta-seconds and an HTTP date.
func excelBPSRetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return 0, false
	}
	if seconds, err := strconv.ParseUint(value, 10, 32); err == nil {
		return time.Duration(seconds) * time.Second, true
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		return retryAt.Sub(now), true
	}
	return 0, false
}

// isExcelBPSCoolingDown applies a local cooldown only when this account would
// actually route the requested model through Basispoints.
func (s *OpenAIGatewayService) isExcelBPSCoolingDown(account *Account, requestedModel string) bool {
	if s == nil || account == nil || !account.UsesBasispointsResponses() {
		return false
	}
	if strings.TrimSpace(requestedModel) != "" && !account.IsModelSupported(requestedModel) {
		return false
	}
	value, ok := s.excelBPSCooldownUntil.Load(account.ID)
	if !ok {
		return false
	}
	until, valid := value.(time.Time)
	if valid && time.Now().Before(until) {
		return true
	}
	s.excelBPSCooldownUntil.CompareAndDelete(account.ID, value)
	return false
}
