package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// a6e51622a fixes an upstream-only BPS pool partition. The local selector
// already keeps mixed candidates: verify that property without adding a new
// BPS priority or overriding the administrator's existing account priorities.
func TestBasispointsMixedPoolKeepsAvailableCapacity(t *testing.T) {
	for _, tc := range []struct {
		name                           string
		bpsLoad                        int
		nativeAcquired, nativeExcluded bool
		wantAccount                    int64
		wantWait                       bool
	}{
		{name: "full BPS uses available native", bpsLoad: 100, nativeAcquired: true, wantAccount: 2},
		{name: "lost BPS slot uses available native", nativeAcquired: true, wantAccount: 2},
		{name: "both busy keep bounded wait", bpsLoad: 100, wantWait: true},
		{name: "excluded native stays excluded", bpsLoad: 100, nativeAcquired: true, nativeExcluded: true, wantAccount: 1, wantWait: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accounts := []Account{
				{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 0,
					Extra: map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis}},
				{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 10},
			}
			acquired, released := []int64{}, []int64{}
			cache := schedulerTestConcurrencyCache{
				loadMap: map[int64]*AccountLoadInfo{
					1: {AccountID: 1, LoadRate: tc.bpsLoad, CurrentConcurrency: tc.bpsLoad / 100},
					2: {AccountID: 2},
				},
				acquireResults: map[int64]bool{1: false, 2: tc.nativeAcquired},
				acquiredIDs:    &acquired, releasedIDs: &released,
			}
			cfg := &config.Config{}
			// Unlike upstream's explicit BPS/native partitions, local TopK is
			// shared. Keep both candidates within the configured probe budget.
			cfg.Gateway.OpenAIWS.LBTopK = 2
			svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts}, cfg: cfg, concurrencyService: NewConcurrencyService(cache)}
			req := OpenAIAccountScheduleRequest{Platform: PlatformOpenAI, RequestedModel: "gpt-6-astra"}
			if tc.nativeExcluded {
				req.ExcludedIDs = map[int64]struct{}{2: {}}
			}
			selection, decision, err := newDefaultOpenAIAccountScheduler(svc, nil).Select(context.Background(), req)
			require.NoError(t, err)
			if !tc.nativeExcluded {
				require.Equal(t, 2, decision.CandidateCount)
			}
			require.NotNil(t, selection)
			if tc.wantAccount != 0 {
				require.Equal(t, tc.wantAccount, selection.Account.ID, "acquired=%v released=%v", acquired, released)
			} else {
				// No BPS-only preference is imposed on the local weighted wait plan.
				require.Contains(t, []int64{1, 2}, selection.Account.ID)
			}
			if tc.wantWait {
				require.NotNil(t, selection.WaitPlan)
				require.False(t, selection.Acquired)
			} else {
				require.Nil(t, selection.WaitPlan)
				require.True(t, selection.Acquired)
				require.NotNil(t, selection.ReleaseFunc)
				selection.ReleaseFunc()
				require.Equal(t, []int64{tc.wantAccount}, released)
			}
			if tc.nativeExcluded {
				require.NotContains(t, acquired, int64(2))
			}
		})
	}
}
