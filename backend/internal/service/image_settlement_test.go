//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageSettlementUsesResolvedOutputCost(t *testing.T) {
	logs := &openAIRecordUsageLogRepoStub{inserted: true}
	billing := &openAIRecordUsageBillingRepoStub{}
	atomic := &openAIVideoRefundAtomicBillingRepoStub{openAIRecordUsageBillingRepoStub: billing, usageRepo: logs}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(logs, atomic, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
	groupID := int64(4)
	low, high := 1.0, 2.0
	key := &APIKey{ID: 1, GroupID: &groupID, Group: &Group{
		ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 1, ImageRateMultiplier: 1,
		ImagePrice1K: &low, ImagePrice2K: &high,
	}}
	user := &User{ID: 2}
	result := &OpenAIForwardResult{
		RequestID: "audit-image", Model: "gpt-image-2", ImageCount: 1, ImageSize: "1K",
		ImageInputSize: "1024x1024", ImageOutputSizes: []string{"2048x2048"},
	}
	// This is the handler's post-response estimate, before RecordUsage resolves output size.
	quoted, err := svc.EstimateOpenAIImagesCost(context.Background(), key, user, result.Model, "", "", result.ImageCount, result.ImageSize)
	require.NoError(t, err)
	require.InDelta(t, 1, quoted.ActualCost, 1e-8)
	err = svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: result, APIKey: key, User: user, Account: &Account{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
		MediaBalanceHold: &OpenAIMediaBalanceHold{ID: "audit-hold", UserID: user.ID, APIKeyID: key.ID, Amount: 1},
	})
	require.NoError(t, err)
	require.InDelta(t, 2, billing.lastCmd.CapturedBalanceHold.ActualAmount, 1e-8)
	require.Zero(t, billing.lastCmd.BalanceCost, "capture skips a separate balance debit")
	require.InDelta(t, 2, logs.lastLog.ActualCost, 1e-8)
	require.True(t, billing.lastCmd.CapturedBalanceHold.CompletedMedia)
}
