//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type pendingImageSettlementRepoStub struct {
	*openAIVideoRefundAtomicBillingRepoStub
	pending   *PendingImageSettlement
	lastError error
}

func (r *pendingImageSettlementRepoStub) SavePendingImageSettlement(_ context.Context, item *PendingImageSettlement) error {
	r.pending = item
	return nil
}

func (r *pendingImageSettlementRepoStub) ClaimPendingImageSettlements(context.Context, time.Time) ([]PendingImageSettlement, error) {
	if r.pending == nil {
		return nil, nil
	}
	return []PendingImageSettlement{*r.pending}, nil
}

func (r *pendingImageSettlementRepoStub) FinishPendingImageSettlement(_ context.Context, _ *PendingImageSettlement, cause error) error {
	r.lastError = cause
	if cause == nil {
		r.pending = nil
	}
	return nil
}

type mediaSettlementUserRepoStub struct {
	UserRepository
	user *User
}

func (r *mediaSettlementUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return r.user, nil
}

func TestImageSettlementQueuesAndReplaysExactCost(t *testing.T) {
	logs := &openAIRecordUsageLogRepoStub{inserted: true}
	billing := &openAIRecordUsageBillingRepoStub{err: errors.New("temporary database failure"), applied: map[string]bool{}}
	repo := &pendingImageSettlementRepoStub{openAIVideoRefundAtomicBillingRepoStub: &openAIVideoRefundAtomicBillingRepoStub{openAIRecordUsageBillingRepoStub: billing, usageRepo: logs}}
	user := &User{ID: 2}
	account := &Account{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(logs, repo, &mediaSettlementUserRepoStub{user: user}, &openAIRecordUsageSubRepoStub{}, nil)
	price := 2.0
	group := &Group{ID: 4, RateMultiplier: 1, ImageRateMultiplier: 1, ImagePrice2K: &price}
	key := &APIKey{ID: 1, GroupID: &group.ID, Group: group}
	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{RequestID: "pending-image", Model: "gpt-image-2", ImageCount: 1, ImageSize: "2K"},
		APIKey: key, User: user, Account: account,
		MediaBalanceHold: &OpenAIMediaBalanceHold{ID: "image-hold", APIKeyID: key.ID, UserID: user.ID, Amount: 1},
	})
	require.ErrorContains(t, err, "temporary database failure")
	require.NotNil(t, repo.pending)
	require.Zero(t, logs.calls)
	require.Error(t, repo.lastError)
	// The worker uses the original command after a pricing change.
	price = 99
	billing.err = nil
	svc.accountRepo = &openAIRecordUsageAccountRepoStub{account: account}
	svc.openAIVideoCompensatorAPIKeyService = &openAIVideoRecoveryAPIKeyServiceStub{apiKey: key}
	svc.reconcilePendingImageSettlements()
	require.Nil(t, repo.pending)
	require.Equal(t, 1, logs.calls)
	require.Equal(t, 2.0, logs.lastLog.ActualCost)
	require.Equal(t, 2.0, billing.lastCmd.CapturedBalanceHold.ActualAmount)
	svc.reconcilePendingImageSettlements()
	require.Equal(t, 2, billing.calls, "one failed attempt and one successful settlement")
}

func TestMiniMaxVideoMissingHoldStopsRetryingAndRetainsTaskOwner(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	binding.RecoveryStatus, binding.UpstreamTaskID = openAIVideoRecoveryIdentified, "verified-task"
	key := &APIKey{ID: binding.APIKeyID}
	user := &User{ID: binding.UserID}
	account := newMiniMaxVideoTestAccount()
	account.ID = binding.AccountID
	snapshot, err := json.Marshal(OpenAIVideoRecoveryBillingSnapshot{
		RequestModel: "MiniMax-H3", UpstreamModel: "MiniMax-H3", VideoDurationSeconds: 15, VideoResolution: "2K",
		Cost: &CostBreakdown{TotalCost: 18, ActualCost: 18, BillingMode: string(BillingModePerRequest)},
		Hold: &OpenAIMediaBalanceHold{ID: "original-hold", APIKeyID: key.ID, UserID: user.ID, Amount: 18},
	})
	require.NoError(t, err)
	binding.RecoveryBilling = snapshot
	recovery := &openAIVideoCompensationRepoStub{recoveryClaimed: []OpenAIVideoTaskBinding{binding}, persistRecoveryClaims: true}
	logs := &openAIRecordUsageLogRepoStub{}
	billing := &openAIRecordUsageBillingRepoStub{err: ErrMediaBalanceHoldInconsistent}
	atomic := &openAIVideoRefundAtomicBillingRepoStub{openAIRecordUsageBillingRepoStub: billing, usageRepo: logs}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(logs, atomic, &mediaSettlementUserRepoStub{user: user}, &openAIRecordUsageSubRepoStub{}, nil)
	svc.openAIVideoTaskBindingRepo = recovery
	svc.accountRepo = &openAIRecordUsageAccountRepoStub{account: account}
	svc.openAIVideoCompensatorAPIKeyService = &openAIVideoRecoveryAPIKeyServiceStub{apiKey: key}
	svc.reconcileMiniMaxVideoRecoveries(time.Now())
	require.Equal(t, openAIVideoRecoveryBillingReview, recovery.updatedStatus)
	require.Equal(t, "verified-task", recovery.updatedUpstreamTaskID)
	require.Nil(t, recovery.updatedNext)
	require.Contains(t, recovery.updatedError, "requires reconciliation")
	require.Zero(t, recovery.armCalls)
	require.Zero(t, logs.calls)
	svc.reconcileMiniMaxVideoRecoveries(time.Now().Add(time.Minute))
	require.Equal(t, 1, billing.calls, "accounting review is not a permanent 15-second retry loop")
}
