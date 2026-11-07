//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlazaAccountAliasesUseMappedChannelPrices(t *testing.T) {
	mapping := map[string]any{
		"Midjourney-V7":                  "midjourney-D",
		"gemini-3-pro-image-preview":     "nano-banana2-pro",
		"gemini-3.1-flash-image-preview": "nano-banana2",
		"seedream-5.0-pro":               "seedream-5.0-Flash（满血）",
		"gpt-5.5":                        "gpt-5.5", // whitelist entry
	}
	channel := plazaPricedChannel(1, "media", []int64{1}, PlatformOpenAI, "gpt-5.5")
	for _, name := range []string{"midjourney-D", "nano-banana2-pro", "nano-banana2", "seedream-5.0-Flash（满血）"} {
		channel.ModelPricing = append(channel.ModelPricing, ChannelModelPricing{
			Platform: PlatformOpenAI, Models: []string{name}, BillingMode: BillingModePerRequest, PerRequestPrice: testPtrFloat64(6),
		})
	}
	svc := newPlazaServiceWithBilling([]Channel{channel}, []Group{{ID: 1, Platform: PlatformOpenAI}}, map[int64]string{1: PlatformOpenAI}, nil)
	svc.accountRepo = plazaAccountRepo{byGroup: map[int64][]Account{1: {{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": mapping}}}}}
	out, err := svc.ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].Models, len(mapping), "仅展示映射左侧和白名单，不附加右侧渠道内部名称")
	for _, m := range out[0].Models {
		require.Contains(t, mapping, m.Name, "保持配置的大小写")
		require.NotNil(t, m.Pricing, m.Name)
		if m.Name == "gpt-5.5" {
			require.InDelta(t, 3e-6, *m.Pricing.InputPrice, 1e-15)
			continue
		}
		require.Equal(t, BillingModePerRequest, m.Pricing.BillingMode, m.Name)
		require.Equal(t, float64(6), *m.Pricing.PerRequestPrice, m.Name)
		require.Equal(t, "image", m.MediaType, m.Name)
	}
}

func TestPlazaAliasKeepsTargetExplicitlyWhitelistedInSameGroup(t *testing.T) {
	channel := plazaPricedChannel(1, "channel", []int64{1, 2}, PlatformOpenAI, "midjourney-D")
	channel.ModelPricing[0].BillingMode = BillingModeImage
	channel.ModelPricing[0].PerRequestPrice = testPtrFloat64(6)
	groups := []Group{{ID: 1, Platform: PlatformOpenAI}, {ID: 2, Platform: PlatformOpenAI}}
	svc := newPlazaServiceWithBilling([]Channel{channel}, groups, map[int64]string{1: PlatformOpenAI, 2: PlatformOpenAI}, nil)
	svc.accountRepo = plazaAccountRepo{byGroup: map[int64][]Account{
		1: {
			{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"Midjourney-V7": "midjourney-D"}}},
			{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"midjourney-D": "midjourney-D"}}},
		},
		2: {{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"midjourney-D": "midjourney-D"}}}},
	}}
	out, err := svc.ListGroups(context.Background())
	require.NoError(t, err)
	for _, g := range out {
		names := []string{}
		for _, m := range g.Models {
			names = append(names, m.Name)
			require.NotNil(t, m.Pricing, m.Name)
			require.Equal(t, float64(6), *m.Pricing.PerRequestPrice)
		}
		if g.ID == 1 {
			require.ElementsMatch(t, []string{"Midjourney-V7", "midjourney-D"}, names)
		} else {
			require.Equal(t, []string{"midjourney-D"}, names, "不同分组不受别组映射影响")
		}
	}
}

func TestPlazaMappedMediaRetainsIndependentRateClassification(t *testing.T) {
	require.Equal(t, "image", plazaMediaType(PlazaModel{
		Name: "public-image", mappedModel: "midjourney-D",
		Pricing: &ChannelModelPricing{BillingMode: BillingModePerRequest},
	}))
}

func TestPlazaAliasPriceSourceAndGroupAllowlist(t *testing.T) {
	for _, explicitPrice := range []bool{false, true} {
		t.Run(map[bool]string{false: "mapped default", true: "explicit alias channel"}[explicitPrice], func(t *testing.T) {
			catalog := newStubPricingServiceFromMap(map[string]*LiteLLMModelPricing{
				"public-alias":  {Mode: "chat", InputCostPerToken: 90e-6, OutputCostPerToken: 100e-6},
				"upstream-text": {Mode: "chat", InputCostPerToken: 2e-6, OutputCostPerToken: 8e-6, CacheCreationInputTokenCost: 3e-6, CacheReadInputTokenCost: 0.2e-6},
			})
			var channels []Channel
			if explicitPrice {
				channels = []Channel{plazaPricedChannel(1, "channel", []int64{1}, PlatformOpenAI, "public-alias")}
			}
			group := Group{ID: 1, Platform: PlatformOpenAI, ModelAllowlist: GroupModelAllowlist{Enabled: true, Models: []string{"public-alias"}}}
			svc := newPlazaServiceWithBilling(channels, []Group{group}, map[int64]string{1: PlatformOpenAI}, catalog)
			svc.accountRepo = plazaAccountRepo{byGroup: map[int64][]Account{1: {{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"public-alias": "upstream-text", "blocked": "upstream-text"}}}}}}
			out, err := svc.ListGroups(context.Background())
			require.NoError(t, err)
			require.Len(t, out[0].Models, 1)
			m := out[0].Models[0]
			require.Equal(t, "public-alias", m.Name)
			if explicitPrice {
				require.InDelta(t, 3e-6, *m.Pricing.InputPrice, 1e-15)
				require.InDelta(t, 15e-6, *m.Pricing.OutputPrice, 1e-15)
			} else {
				require.InDelta(t, 2e-6, *m.Pricing.InputPrice, 1e-15)
				require.InDelta(t, 8e-6, *m.Pricing.OutputPrice, 1e-15)
				require.InDelta(t, 3e-6, *m.Pricing.CacheWritePrice, 1e-15)
				require.InDelta(t, 0.2e-6, *m.Pricing.CacheReadPrice, 1e-15)
			}
		})
	}
}
