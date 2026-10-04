//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeminiBillingIntegrity_MissingPriceCannotBecomeZeroCost(t *testing.T) {
	logs := &openAIRecordUsageLogRepoStub{inserted: true}
	billing := &openAIRecordUsageBillingRepoStub{}
	atomic := &openAIVideoRefundAtomicBillingRepoStub{openAIRecordUsageBillingRepoStub: billing, usageRepo: logs}
	svc := newGatewayRecordUsageServiceWithBillingRepoForTest(logs, atomic, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
	const model = "unpriced-custom-native-model"
	key := &APIKey{ID: 1}
	require.ErrorIs(t, svc.ValidateTokenPricing(context.Background(), key, model, "", "", model), ErrModelPricingUnavailable)
	err := svc.RecordUsage(context.Background(), &RecordUsageInput{
		Result: &ForwardResult{RequestID: "missing-price", Model: model, Usage: ClaudeUsage{InputTokens: 1000, OutputTokens: 500}},
		APIKey: key, User: &User{ID: 2}, Account: &Account{ID: 3, Platform: PlatformGemini, Type: AccountTypeAPIKey},
	})
	require.ErrorIs(t, err, ErrModelPricingUnavailable)
	require.Zero(t, logs.calls)
	require.Zero(t, billing.calls)
}

func TestGeminiBillingIntegrity_MappedAndExplicitFreePricingRemainValid(t *testing.T) {
	for _, multiplier := range []float64{1, 0} {
		logs := &openAIRecordUsageLogRepoStub{inserted: true}
		billing := &openAIRecordUsageBillingRepoStub{}
		atomic := &openAIVideoRefundAtomicBillingRepoStub{openAIRecordUsageBillingRepoStub: billing, usageRepo: logs}
		svc := newGatewayRecordUsageServiceWithBillingRepoForTest(logs, atomic, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
		group := &Group{ID: 4, Platform: PlatformGemini, RateMultiplier: multiplier}
		key := &APIKey{ID: 1, GroupID: &group.ID, Group: group}
		svc.resolver = newOpenAITokenImageChannelPricingResolverForTest(t, group.ID, "priced-upstream")
		require.NoError(t, svc.ValidateTokenPricing(context.Background(), key, "public-alias", "channel-alias", BillingModelSourceRequested, "priced-upstream"))
		err := svc.RecordUsage(context.Background(), &RecordUsageInput{
			Result: &ForwardResult{RequestID: "mapped-price", Model: "public-alias", UpstreamModel: "priced-upstream", Usage: ClaudeUsage{InputTokens: 1000, OutputTokens: 500}},
			APIKey: key, User: &User{ID: 2}, Account: &Account{ID: 3, Platform: PlatformGemini, Type: AccountTypeAPIKey},
		})
		require.NoError(t, err)
		require.Equal(t, 1, logs.calls)
		require.Positive(t, logs.lastLog.TotalCost)
		if multiplier == 0 {
			require.Zero(t, logs.lastLog.ActualCost)
		} else {
			require.Positive(t, logs.lastLog.ActualCost)
		}
		require.Equal(t, QuantizeUsageBillingAmount(logs.lastLog.ActualCost), billing.lastCmd.BalanceCost)
	}
}
