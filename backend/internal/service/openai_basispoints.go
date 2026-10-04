package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	openAIOAuthResponsesEndpointExtraKey = "openai_oauth_responses_endpoint"
	openAIOAuthResponsesEndpointChatGPT  = "chatgpt_codex"
	openAIOAuthResponsesEndpointBasis    = "basispoints"
)

// openAIOAuthResponsesEndpointMode is intentionally opt-in. Missing, invalid,
// or non-OAuth values always preserve the existing ChatGPT Codex route.
func openAIOAuthResponsesEndpointMode(account *Account) string {
	if account == nil || !account.IsOpenAIOAuthLike() || account.IsOpenAIAgentIdentity() || account.Extra == nil {
		return openAIOAuthResponsesEndpointChatGPT
	}
	value, _ := account.Extra[openAIOAuthResponsesEndpointExtraKey].(string)
	if strings.EqualFold(strings.TrimSpace(value), openAIOAuthResponsesEndpointBasis) {
		return openAIOAuthResponsesEndpointBasis
	}
	return openAIOAuthResponsesEndpointChatGPT
}

func shouldForwardOpenAIBasispoints(ctx context.Context, c *gin.Context, account *Account, body []byte) bool {
	return openAIOAuthResponsesEndpointMode(account) == openAIOAuthResponsesEndpointBasis &&
		!isOpenAIImagesSelfBuiltRequest(ctx) && GetOpenAIClientTransport(c) != OpenAIClientTransportWS
}

// Basispoints is a different wire protocol, not a Codex URL override. Do not
// copy client headers or apply Codex fingerprints to this request.
func newOpenAIBasispointsRequest(ctx context.Context, repo AccountRepository, account *Account, body []byte, token string) (*http.Request, error) {
	ctx = WithHTTPUpstreamRedirectsDisabled(WithHTTPUpstreamProfile(ctx, HTTPUpstreamProfileExcelBPS))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, basispointsResponsesURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Origin", "https://bps.openai.com")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("x-openai-internal-basispoints-client-product", "basispoints-excel-plugin")
	req.Header.Set("x-openai-internal-basispoints-client-agent-profile", "excel")
	if err := applyOpenAIBasispointsHeaders(ctx, repo, account, token, req.Header); err != nil {
		return nil, fmt.Errorf("build Basispoints headers: %w", err)
	}
	return req, nil
}

func applyOpenAIBasispointsHeaders(ctx context.Context, repo AccountRepository, account *Account, token string, headers http.Header) error {
	if openAIOAuthResponsesEndpointMode(account) != openAIOAuthResponsesEndpointBasis {
		return nil
	}
	accountID, err := resolveOpenAIBasispointsAccountID(ctx, repo, account, token)
	if err != nil {
		return err
	}
	headers.Set("x-openai-account-id", accountID)
	headers.Set("chatgpt-account-id", accountID)
	headers.Set("x-basispoints-auth-mode", "chatgpt")
	return nil
}

func resolveOpenAIBasispointsAccountID(ctx context.Context, repo AccountRepository, account *Account, token string) (string, error) {
	credAccount, err := resolveCredentialAccount(ctx, repo, account)
	if err != nil {
		return "", err
	}
	stored := strings.TrimSpace(credAccount.GetChatGPTAccountID())
	// An explicitly selected workspace (e.g. Team) can differ from the token's
	// default account claim. Let upstream authenticate the selected membership.
	if stored != "" {
		return stored, nil
	}
	fromToken := openAIChatGPTAccountIDFromAccessToken(token)
	if fromToken != "" {
		return fromToken, nil
	}
	return "", errors.New("Basispoints requires chatgpt_account_id")
}

// openAIChatGPTAccountIDFromAccessToken only extracts a candidate. The token
// itself is still authenticated by the upstream; this value is never treated
// as a replacement for the bearer token or as a locally trusted JWT.
func openAIChatGPTAccountIDFromAccessToken(token string) string {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		if padded, padErr := base64.URLEncoding.DecodeString(parts[1]); padErr == nil {
			payload = padded
		} else {
			return ""
		}
	}
	var claims map[string]json.RawMessage
	if json.Unmarshal(payload, &claims) != nil {
		return ""
	}
	const authClaimKey = "https://api.openai.com/auth"
	var auth map[string]any
	if raw, ok := claims[authClaimKey]; ok && json.Unmarshal(raw, &auth) == nil {
		if value, ok := auth["chatgpt_account_id"].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
