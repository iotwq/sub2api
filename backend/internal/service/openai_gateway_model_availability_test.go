//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func (r schedulerTestOpenAIAccountRepo) ListModelAvailabilityCandidates(_ context.Context, _ *int64, platforms []string, _ bool) ([]Account, error) {
	allowed := make(map[string]struct{}, len(platforms))
	for _, platform := range platforms {
		allowed[platform] = struct{}{}
	}
	result := make([]Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		if _, ok := allowed[account.Platform]; ok {
			result = append(result, account)
		}
	}
	return result, nil
}

func TestOpenAIGatewayServiceDiagnoseModelAvailability_PassthroughAllowsAnyModel(t *testing.T) {
	repo := schedulerTestOpenAIAccountRepo{
		accounts: []Account{
			{
				ID:          1,
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Extra:       map[string]any{"openai_passthrough": true},
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-4o": "gpt-4o"},
				},
			},
		},
	}
	svc := &OpenAIGatewayService{accountRepo: repo, cfg: testConfig()}

	diag := svc.DiagnoseModelAvailabilityForPlatform(context.Background(), nil, "gpt-5.5", PlatformOpenAI)

	require.True(t, diag.HasAccountsInPool)
	require.True(t, diag.HasModelSupport)
}

func TestOpenAIAccountScheduler_PassthroughAllowsUnmappedModel(t *testing.T) {
	account := &Account{
		ID:          1,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Extra:       map[string]any{"openai_passthrough": true},
		Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-4o": "gpt-4o"},
		},
	}
	scheduler := &defaultOpenAIAccountScheduler{}
	req := OpenAIAccountScheduleRequest{
		Platform:       PlatformOpenAI,
		RequestedModel: "gpt-5.5",
	}

	require.True(t, scheduler.isAccountRequestCompatible(context.Background(), account, req))
}
