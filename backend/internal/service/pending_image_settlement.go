package service

import (
	"context"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// MediaSettlementError reports whether a media balance hold is already owned
// by a durable pending-settlement record. Callers must release the hold when
// this is false and leave it untouched when this is true.
type MediaSettlementError struct {
	Err          error
	HoldRetained bool
}

func (e *MediaSettlementError) Error() string {
	if e == nil || e.Err == nil {
		return "media settlement failed"
	}
	return e.Err.Error()
}

func (e *MediaSettlementError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func mediaSettlementError(err error, holdRetained bool) error {
	if err == nil {
		return nil
	}
	return &MediaSettlementError{Err: err, HoldRetained: holdRetained}
}

// MediaBalanceHoldRetained reports whether a failed settlement has a durable
// retry record that owns the hold. A false result means the request path must
// release its in-memory hold.
func MediaBalanceHoldRetained(err error) bool {
	var settlementErr *MediaSettlementError
	return errors.As(err, &settlementErr) && settlementErr.HoldRetained
}

// PendingImageSettlement preserves the exact command and usage after successful
// image generation. Retrying never recalculates prices from current settings.
type PendingImageSettlement struct {
	Command  *UsageBillingCommand
	Usage    *UsageLog
	Platform string
	Attempts int
}

type PendingImageSettlementRepository interface {
	SavePendingImageSettlement(context.Context, *PendingImageSettlement) error
	ClaimPendingImageSettlements(context.Context, time.Time) ([]PendingImageSettlement, error)
	FinishPendingImageSettlement(context.Context, *PendingImageSettlement, error) error
}

func (s *OpenAIGatewayService) reconcilePendingImageSettlements() {
	repo, ok := s.usageBillingRepo.(PendingImageSettlementRepository)
	atomicRepo, atomicOK := s.usageBillingRepo.(AtomicUsageBillingRepository)
	if !ok || !atomicOK || s.openAIVideoCompensatorAPIKeyService == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
	items, err := repo.ClaimPendingImageSettlements(ctx, time.Now())
	cancel()
	if err != nil {
		logger.LegacyPrintf("service.image_settlement", "claim pending settlements failed: %v", err)
		return
	}
	for i := range items {
		item := &items[i]
		ctx, cancel := context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
		err := s.replayImageSettlement(ctx, atomicRepo, item)
		cancel()
		ctx, cancel = context.WithTimeout(context.Background(), openAIVideoCompensationTimeout)
		if finishErr := repo.FinishPendingImageSettlement(ctx, item, err); finishErr != nil {
			logger.LegacyPrintf("service.image_settlement", "save settlement result request=%s: %v", item.Command.RequestID, finishErr)
		}
		cancel()
		if err != nil {
			logger.LegacyPrintf("service.image_settlement", "settlement pending request=%s: %v", item.Command.RequestID, err)
		}
	}
}

func (s *OpenAIGatewayService) replayImageSettlement(ctx context.Context, repo AtomicUsageBillingRepository, item *PendingImageSettlement) error {
	cmd, usage := item.Command, item.Usage
	key, err := s.openAIVideoCompensatorAPIKeyService.GetByID(ctx, cmd.APIKeyID)
	if err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	account, err := s.accountRepo.GetByID(ctx, cmd.AccountID)
	if err != nil {
		return err
	}
	result, err := repo.ApplyWithUsageLog(ctx, cmd, usage)
	if err != nil {
		return err
	}
	// Invalidate even on a duplicate: the previous process may have committed
	// successfully and exited before refreshing its caches.
	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateUserBalance(ctx, cmd.UserID)
		_ = s.billingCacheService.InvalidateAPIKeyRateLimit(ctx, cmd.APIKeyID)
		if usage.GroupID != nil && usage.SubscriptionID != nil {
			_ = s.billingCacheService.InvalidateSubscription(ctx, cmd.UserID, *usage.GroupID)
		}
	}
	if invalidator, ok := s.openAIVideoCompensatorAPIKeyService.(apiKeyAuthCacheInvalidator); ok && key != nil {
		invalidator.InvalidateAuthCacheByKey(ctx, key.Key)
	}
	if result != nil && result.Applied {
		finalizePostUsageBilling(ctx, &postUsageBillingParams{
			Cost: &CostBreakdown{TotalCost: usage.TotalCost, ActualCost: usage.ActualCost},
			User: user, APIKey: key, Account: account,
			IsSubscriptionBill:   usage.BillingType == BillingTypeSubscription,
			SkipBalanceDeduction: true, Platform: item.Platform,
			BillingCachesInvalidated: true,
		}, s.billingDeps(), result)
	}
	return nil
}
