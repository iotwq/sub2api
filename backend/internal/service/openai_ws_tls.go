package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type openAIWSTLSContextKey struct{}

type openAIWSTLSConfig struct {
	accountID int64
	profile   *tlsfingerprint.Profile
}

func withOpenAIWSTLSProfile(ctx context.Context, accountID int64, profile *tlsfingerprint.Profile) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if profile == nil {
		return ctx
	}
	return context.WithValue(ctx, openAIWSTLSContextKey{}, openAIWSTLSConfig{accountID: accountID, profile: profile.Clone()})
}

// Include the full target and proxy in the digest without keeping credentials in
// cache keys or logs. Profile contents, rather than its display name, define TLS.
func openAIWSTransportKey(accountID int64, target, proxy string, profile *tlsfingerprint.Profile) string {
	if profile == nil {
		return ""
	}
	value := fmt.Sprintf("%d\x00%s\x00%s\x00%s", accountID, strings.TrimSpace(target), strings.TrimSpace(proxy), profile.CacheKey())
	return fmt.Sprintf("mode1:%x", sha256.Sum256([]byte(value)))
}

func (d *coderOpenAIWSClientDialer) fingerprintHTTPClient(cfg openAIWSTLSConfig, target, proxy string) (*http.Client, error) {
	if d == nil || cfg.accountID <= 0 || cfg.profile == nil {
		return nil, errors.New("invalid mode1 websocket TLS configuration")
	}
	var proxyURL *url.URL
	if strings.TrimSpace(proxy) != "" {
		var err error
		proxyURL, err = url.Parse(strings.TrimSpace(proxy))
		if err != nil {
			return nil, errors.New("invalid mode1 websocket proxy URL")
		}
	}
	key := openAIWSTransportKey(cfg.accountID, target, proxy, cfg.profile)
	now := time.Now().UnixNano()
	d.proxyMu.Lock()
	defer d.proxyMu.Unlock()
	if entry := d.proxyClients[key]; entry != nil && entry.client != nil {
		entry.lastUsedUnixNano = now
		d.proxyHits.Add(1)
		return entry.client, nil
	}
	d.cleanupProxyClientsLocked(now)
	transport, err := tlsfingerprint.NewHTTPTransport(cfg.profile, proxyURL, tlsfingerprint.TransportOptions{
		RootCAs: d.tlsRootCAs, ProxyRootCAs: d.tlsProxyRootCAs, HandshakeTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("mode1 websocket TLS transport: %w", err)
	}
	transport.MaxIdleConns = openAIWSProxyTransportMaxIdleConns
	transport.MaxIdleConnsPerHost = openAIWSProxyTransportMaxIdleConnsPerHost
	transport.IdleConnTimeout = openAIWSProxyTransportIdleConnTimeout
	client := &http.Client{Transport: transport}
	if d.proxyClients == nil {
		d.proxyClients = make(map[string]*openAIWSProxyClientEntry)
	}
	d.proxyClients[key] = &openAIWSProxyClientEntry{client: client, lastUsedUnixNano: now}
	d.ensureProxyClientCapacityLocked()
	d.proxyMisses.Add(1)
	return client, nil
}
