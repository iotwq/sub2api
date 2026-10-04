package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

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
