package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
)

func codexProtectionEnabled(a *Account) bool {
	plan, err := AccountTrafficPlanFor(a)
	return err == nil && plan.Policy.Active
}

func codexProtectionTLSProfile(a *Account, cfg *config.Config) (*tlsfingerprint.Profile, error) {
	plan, err := AccountTrafficPlanFor(a)
	if err != nil || !plan.Policy.Active || plan.Policy.TLSProfile == "standard" {
		return nil, err
	}
	if cfg == nil || !cfg.Gateway.TLSFingerprint.Enabled {
		return nil, fmt.Errorf("Codex TLS profile requires the global TLS fingerprint switch")
	}
	return tlsfingerprint.BuiltinProfile(plan.Policy.TLSProfile), nil
}

type codexTLSUpstream interface {
	DoWithCodexTLS(*http.Request, string, int64, int, *tlsfingerprint.Profile) (*http.Response, error)
}

// Discovery, token refresh, image/video and other provider routes are unchanged.
func withCodexTrafficRequest(req *http.Request, a *Account) *http.Request {
	if !isCodexResponseRequest(req) {
		return req
	}
	return WithAccountTrafficRequest(req, a)
}

func isCodexResponseRequest(req *http.Request) bool {
	if req == nil || req.URL == nil || req.Method != http.MethodPost {
		return false
	}
	path := strings.TrimRight(req.URL.Path, "/")
	return strings.HasSuffix(path, "/responses") || strings.HasSuffix(path, "/responses/compact")
}

func codexHTTPProtectionAccount(req *http.Request, a *Account) *Account {
	if !isCodexResponseRequest(req) {
		return nil
	}
	return a
}
