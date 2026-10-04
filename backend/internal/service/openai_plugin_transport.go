package service

import (
	"fmt"
	"net/http"
)

func (s *OpenAIGatewayService) SetPluginManager(manager *PluginManager) {
	s.pluginManager = manager
}

// doOpenAIUpstream 只在 OpenAI OAuth 能力绑定已启用时把真实请求交给插件。
// 插件返回标准 http.Response，响应解析、错误映射、SSE 和计费仍由现有核心链处理。
func (s *OpenAIGatewayService) doOpenAIUpstream(request *http.Request, proxyURL string, account *Account) (*http.Response, error) {
	return accountTrafficController(s.httpUpstream).DoHTTP(withCodexTrafficRequest(request, account), func(req *http.Request) (*http.Response, error) {
		response, err := s.doOpenAIUpstreamUncontrolled(req, proxyURL, account)
		if err == nil && response != nil && response.StatusCode >= 200 && response.StatusCode < 300 {
			s.observeCodexSignal(account, response.Header)
		}
		return response, err
	})
}

func (s *OpenAIGatewayService) doOpenAIUpstreamUncontrolled(request *http.Request, proxyURL string, account *Account) (*http.Response, error) {
	profile, err := codexProtectionTLSProfile(codexHTTPProtectionAccount(request, account), s.cfg)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		if s.pluginManager != nil && s.pluginManager.ShouldRouteOpenAIOAuth(account) {
			return nil, fmt.Errorf("Codex TLS profile cannot be applied to a plugin-managed transport")
		}
		upstream, ok := s.httpUpstream.(codexTLSUpstream)
		if !ok {
			return nil, fmt.Errorf("Codex TLS transport unavailable")
		}
		return upstream.DoWithCodexTLS(request, proxyURL, account.ID, account.Concurrency, profile)
	}
	if s.pluginManager != nil {
		response, handled, err := s.pluginManager.RoundTripOpenAIOAuth(request.Context(), request, proxyURL, account)
		if handled {
			return response, err
		}
	}
	return s.httpUpstream.Do(request, proxyURL, account.ID, account.Concurrency)
}

// doOpenAIAccountTestUpstream 让 OpenAI OAuth 账号测试与真实转发使用同一插件路径。
// API Key 和未命中插件的账号保持各自原有的 HTTPUpstream 行为。
func (s *AccountTestService) doOpenAIAccountTestUpstream(
	request *http.Request,
	proxyURL string,
	account *Account,
	useTLSFallback bool,
) (*http.Response, error) {
	return accountTrafficController(s.httpUpstream).DoHTTP(withCodexTrafficRequest(request, account), func(req *http.Request) (*http.Response, error) {
		response, err := s.doOpenAIAccountTestUncontrolled(req, proxyURL, account, useTLSFallback)
		if err == nil && response != nil && response.StatusCode >= 200 && response.StatusCode < 300 {
			s.openaiGatewayService.observeCodexSignal(account, response.Header)
		}
		return response, err
	})
}

func (s *AccountTestService) doOpenAIAccountTestUncontrolled(request *http.Request, proxyURL string, account *Account, useTLSFallback bool) (*http.Response, error) {
	profile, err := codexProtectionTLSProfile(codexHTTPProtectionAccount(request, account), s.cfg)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		if s.pluginManager != nil && s.pluginManager.ShouldRouteOpenAIOAuth(account) {
			return nil, fmt.Errorf("Codex TLS profile cannot be applied to a plugin-managed transport")
		}
		upstream, ok := s.httpUpstream.(codexTLSUpstream)
		if !ok {
			return nil, fmt.Errorf("Codex TLS transport unavailable")
		}
		return upstream.DoWithCodexTLS(request, proxyURL, account.ID, account.Concurrency, profile)
	}
	if s.pluginManager != nil {
		response, handled, err := s.pluginManager.RoundTripOpenAIOAuth(request.Context(), request, proxyURL, account)
		if handled {
			return response, err
		}
	}
	if useTLSFallback {
		return s.httpUpstream.DoWithTLS(
			request,
			proxyURL,
			account.ID,
			account.Concurrency,
			s.tlsFPProfileService.ResolveTLSProfile(account),
		)
	}
	return s.httpUpstream.Do(request, proxyURL, account.ID, account.Concurrency)
}
