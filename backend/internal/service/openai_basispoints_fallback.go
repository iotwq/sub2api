package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
)

const openAIBasispointsAutoDisableOn403Key = "openai_basispoints_auto_disable_on_403"

// Account tests and real requests must update the same BPS account state.
// Returns true only when a generic upstream HTTP 403 switched the protocol.
func (s *OpenAIGatewayService) observeBasispointsUpstreamResponse(ctx context.Context, account *Account, resp *http.Response, raw []byte) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.UpdateCodexUsageSnapshotFromHeaders(ctx, account.ID, resp.Header)
	}
	// BPS endpoint throttling does not represent Codex quota exhaustion.
	s.handleBasispointsUnauthorized(ctx, account, resp.StatusCode, resp.Header, raw)
	if resp.StatusCode == http.StatusForbidden && gjson.GetBytes(raw, "error.code").String() != "basispoints_model_access_changed" {
		// Group routing and protocol fallback are independent, opt-in actions.
		s.moveBasispointsOn403(ctx, account)
		return s.disableBasispointsOn403(ctx, account)
	}
	return false
}

func (s *OpenAIGatewayService) moveBasispointsOn403(ctx context.Context, account *Account) bool {
	if _, enabled := account.Basispoints403GroupTarget(); !enabled {
		return false
	}
	repo, ok := s.accountRepo.(AccountBasispointsGroupRepository)
	if !ok {
		return false
	}
	stateCtx, cancel := openAIAccountStateContext(ctx)
	defer cancel()
	changed, err := repo.MoveBasispointsOn403(stateCtx, account)
	if err != nil {
		logger.LegacyPrintf("service.openai_basispoints", "automatic group action failed: account_id=%d error_type=%T", account.ID, err)
		return false
	}
	if changed {
		target, _ := account.Basispoints403GroupTarget()
		logger.LegacyPrintf("service.openai_basispoints", "changed groups after upstream HTTP 403: account_id=%d target_group_id=%d", account.ID, target)
	}
	return changed
}

// Preserve the existing OAuth authentication policy without storing arbitrary
// upstream messages, which may contain request content or credentials.
func (s *OpenAIGatewayService) handleBasispointsUnauthorized(ctx context.Context, account *Account, status int, headers http.Header, raw []byte) {
	if status != http.StatusUnauthorized || s.rateLimitService == nil {
		return
	}
	fields := map[string]string{"message": "Basispoints authentication failed"}
	code := extractUpstreamErrorCode(raw)
	if code == "token_invalidated" || code == "token_revoked" {
		fields["code"] = code
	}
	authError := map[string]any{"error": fields}
	if gjson.GetBytes(raw, "detail").String() == "Unauthorized" {
		authError["detail"] = "Unauthorized"
	}
	body, _ := json.Marshal(authError)
	stateCtx, cancel := openAIAccountStateContext(ctx)
	defer cancel()
	s.rateLimitService.HandleUpstreamError(stateCtx, account, status, headers, body)
}

// AccountBasispointsRepository is an optional atomic protocol-switch capability.
type AccountBasispointsRepository interface {
	DisableBasispointsOn403(context.Context, *Account) (bool, error)
}

func (a *Account) UsesBasispointsResponses() bool {
	return openAIOAuthResponsesEndpointMode(a) == openAIOAuthResponsesEndpointBasis
}

func (a *Account) IsBasispointsAutoDisableOn403Enabled() bool {
	if !a.UsesBasispointsResponses() || a.ParentAccountID != nil || a.IsOpenAIPersonalAccessToken() {
		return false
	}
	enabled, _ := a.Extra[openAIBasispointsAutoDisableOn403Key].(bool)
	return enabled
}

func (s *OpenAIGatewayService) disableBasispointsOn403(ctx context.Context, account *Account) bool {
	if !account.IsBasispointsAutoDisableOn403Enabled() {
		return false
	}
	repo, ok := s.accountRepo.(AccountBasispointsRepository)
	if !ok {
		return false
	}
	stateCtx, cancel := openAIAccountStateContext(ctx)
	defer cancel()
	changed, err := repo.DisableBasispointsOn403(stateCtx, account)
	if err != nil {
		// Never log credentials, upstream bodies or database query arguments.
		logger.LegacyPrintf("service.openai_basispoints", "auto-disable failed: account_id=%d error_type=%T", account.ID, err)
		return false
	}
	if changed {
		logger.LegacyPrintf("service.openai_basispoints", "switched to Codex after upstream HTTP 403: account_id=%d", account.ID)
	}
	return changed
}
