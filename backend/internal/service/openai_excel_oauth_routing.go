package service

import (
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

// Excel primary credentials are valid only on the BPS route.
func (a *Account) IsExcelOAuth() bool {
	return a != nil && a.IsOpenAIOAuthLike() && strings.TrimSpace(a.GetCredential("client_id")) == openai.ExcelClientID
}

var errExcelOAuthRouteUnavailable = errors.New("Excel OAuth credentials require the Basispoints route; native Codex is unavailable")
