package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

const openAIMediaBalanceHoldPrefix = "openai-media:"

type OpenAIMediaBalanceHold struct {
	ID          string
	APIKeyID    int64
	UserID      int64
	Amount      float64
	PayloadHash string
}

func OpenAIMediaBalanceHoldID(ctx context.Context, payloadHash string) string {
	if ctx != nil {
		if clientRequestID, _ := ctx.Value(ctxkey.ClientRequestID).(string); strings.TrimSpace(clientRequestID) != "" {
			return openAIMediaBalanceHoldPrefix + strings.TrimSpace(clientRequestID)
		}
		if requestID, _ := ctx.Value(ctxkey.RequestID).(string); strings.TrimSpace(requestID) != "" {
			return openAIMediaBalanceHoldPrefix + strings.TrimSpace(requestID)
		}
	}
	if payloadHash = strings.TrimSpace(payloadHash); payloadHash != "" {
		return openAIMediaBalanceHoldPrefix + payloadHash + ":" + generateRequestID()
	}
	return openAIMediaBalanceHoldPrefix + generateRequestID()
}

func (s *OpenAIGatewayService) ReserveOpenAIMediaBalance(ctx context.Context, apiKey *APIKey, user *User, amount float64, payloadHash string) (*OpenAIMediaBalanceHold, error) {
	if amount <= 0 {
		return nil, nil
	}
	if s != nil && s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		return nil, nil
	}
	if s == nil || s.usageBillingRepo == nil {
		return nil, ErrBatchImageBillingHoldFailed.WithCause(errors.New("usage billing repository is not configured"))
	}
	if apiKey == nil || apiKey.ID <= 0 {
		return nil, ErrBatchImageSettlementMissingAPIKeyID
	}
	if user == nil || user.ID <= 0 {
		return nil, ErrUserNotFound
	}

	hold := &OpenAIMediaBalanceHold{
		ID:          OpenAIMediaBalanceHoldID(ctx, payloadHash),
		APIKeyID:    apiKey.ID,
		UserID:      user.ID,
		Amount:      amount,
		PayloadHash: strings.TrimSpace(payloadHash),
	}
	cmd := openAIMediaBalanceHoldCommand(hold, BatchImageHoldRequestID(hold.ID), amount)
	if _, err := s.usageBillingRepo.ReserveBatchImageBalance(ctx, cmd); err != nil {
		if errors.Is(err, ErrBatchImageInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		return nil, ErrBatchImageBillingHoldFailed.WithCause(err)
	}
	s.invalidateOpenAIMediaBalanceCache(ctx, user.ID)
	return hold, nil
}

func (s *OpenAIGatewayService) CaptureOpenAIMediaBalance(ctx context.Context, hold *OpenAIMediaBalanceHold, actualAmount float64) error {
	if hold == nil || hold.Amount <= 0 {
		return nil
	}
	if s == nil || s.usageBillingRepo == nil {
		return ErrBatchImageSettlementBillingFailed.WithCause(errors.New("usage billing repository is not configured"))
	}
	if actualAmount < 0 {
		actualAmount = 0
	}
	cmd := openAIMediaBalanceHoldCommand(hold, BatchImageCaptureRequestID(hold.ID), actualAmount)
	if _, err := s.usageBillingRepo.CaptureBatchImageBalance(ctx, cmd); err != nil {
		return ErrBatchImageSettlementBillingFailed.WithCause(err)
	}
	s.invalidateOpenAIMediaBalanceCache(ctx, hold.UserID)
	return nil
}

func (s *OpenAIGatewayService) ReleaseOpenAIMediaBalance(ctx context.Context, hold *OpenAIMediaBalanceHold) error {
	if hold == nil || hold.Amount <= 0 {
		return nil
	}
	if s == nil || s.usageBillingRepo == nil {
		return nil
	}
	cmd := openAIMediaBalanceHoldCommand(hold, BatchImageReleaseRequestID(hold.ID), 0)
	if _, err := s.usageBillingRepo.ReleaseBatchImageBalance(ctx, cmd); err != nil {
		if errors.Is(err, ErrUsageBillingRequestConflict) {
			return nil
		}
		return ErrBatchImageBillingHoldFailed.WithCause(err)
	}
	s.invalidateOpenAIMediaBalanceCache(ctx, hold.UserID)
	return nil
}

func openAIMediaBalanceHoldCommand(hold *OpenAIMediaBalanceHold, requestID string, actualAmount float64) *BatchImageBalanceHoldCommand {
	return &BatchImageBalanceHoldCommand{
		RequestID:          requestID,
		APIKeyID:           hold.APIKeyID,
		UserID:             hold.UserID,
		BatchID:            hold.ID,
		HoldAmount:         hold.Amount,
		ActualAmount:       actualAmount,
		RequestPayloadHash: hold.PayloadHash,
		CompletedMedia:     true,
	}
}

func (s *OpenAIGatewayService) invalidateOpenAIMediaBalanceCache(ctx context.Context, userID int64) {
	if s == nil || s.billingCacheService == nil || userID <= 0 {
		return
	}
	_ = s.billingCacheService.InvalidateUserBalance(ctx, userID)
}
