package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type cancellationTrafficCache struct {
	AccountTrafficCache
	cancel       context.CancelFunc
	grantOnFirst bool
	calls        atomic.Int32
	finishes     atomic.Int32
}

func (c *cancellationTrafficCache) Acquire(context.Context, AccountTrafficPlan, string) (AccountTrafficAdmission, error) {
	first := c.calls.Add(1) == 1
	if first {
		c.cancel()
	}
	return AccountTrafficAdmission{Allowed: !first || c.grantOnFirst, RetryAfter: time.Millisecond}, nil
}

func (c *cancellationTrafficCache) Finish(context.Context, AccountTrafficPlan, string, int, int64) error {
	c.finishes.Add(1)
	return nil
}

func TestCodexProtectionHTTPForwardCanceledAdmission(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, grant := range []bool{false, true} {
			t.Run(fmt.Sprintf("passthrough=%v/granted=%v", passthrough, grant), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				cache := &cancellationTrafficCache{cancel: cancel, grantOnFirst: grant}
				upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
				account := protectionAccount()
				account.Credentials["access_token"] = "test-token"
				account.Extra["openai_passthrough"] = passthrough
				account.Extra["openai_oauth_responses_websockets_v2_mode"] = OpenAIWSIngressModeOff
				account.Extra[AccountTrafficPolicyKey].(map[string]any)["wait_seconds"] = 1
				body := []byte(`{"model":"gpt-5.4","input":"hello","stream":false}`)
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body)).WithContext(ctx)
				_, err := svc.Forward(ctx, c, account, body)
				require.Zero(t, upstream.calls, "canceling admission must not start an upstream request")
				require.ErrorIs(t, err, context.Canceled)
				require.EqualValues(t, 1, cache.calls.Load(), "cancellation must stop capacity polling")
				if grant {
					require.Eventually(t, func() bool { return cache.finishes.Load() == 1 }, time.Second, time.Millisecond,
						"a lease granted concurrently with cancellation must be released")
				} else {
					require.Zero(t, cache.finishes.Load())
				}
			})
		}
	}
}

func TestCodexProtectionHTTPAcceptedRequestRetainsUsageDrain(t *testing.T) {
	clientCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	upstreamCtx, release := detachUpstreamContext(clientCtx)
	defer release()
	req := httptest.NewRequest(http.MethodPost, "https://example.test/responses", nil).WithContext(upstreamCtx)
	req = withAccountTrafficAdmissionContext(WithAccountTrafficRequest(req, protectionAccount()), clientCtx)
	cache := &protectionTestCache{}
	control := NewAccountTrafficService(cache)
	var sentCtx context.Context
	const completed = "data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"output_tokens\":42}}}\n\n"
	resp, err := control.DoHTTP(req, func(r *http.Request) (*http.Response, error) {
		sentCtx = r.Context()
		cancel()
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/event-stream"}},
			Body: io.NopCloser(strings.NewReader(completed))}, nil
	})
	require.NoError(t, err)
	require.NoError(t, sentCtx.Err(), "client disconnect after admission must not cancel usage drain")
	require.Zero(t, cache.finishes.Load(), "response headers do not release the lease")
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, completed, string(body))
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 1, cache.finishes.Load())
	require.EqualValues(t, 200, cache.status.Load())
}

func TestCodexProtectionHTTPAdmissionAlreadyCanceledAndScope(t *testing.T) {
	for _, mode := range []string{"protected", "disabled", "api-key", "image"} {
		t.Run(mode, func(t *testing.T) {
			clientCtx, cancel := context.WithCancel(context.Background())
			cancel()
			account := protectionAccount()
			path := "/responses"
			switch mode {
			case "disabled":
				account.Extra[AccountTrafficPolicyKey].(map[string]any)["enabled"] = false
			case "api-key":
				account.Type = AccountTypeAPIKey
			case "image":
				path = "/images/generations"
			}
			cache := &protectionTestCache{}
			up := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
			svc := &OpenAIGatewayService{httpUpstream: up}
			upstreamCtx, release := detachUpstreamContext(clientCtx)
			defer release()
			req := httptest.NewRequest(http.MethodPost, "https://example.test"+path, nil).WithContext(upstreamCtx)
			resp, err := svc.doOpenAIUpstream(withAccountTrafficAdmissionContext(req, clientCtx), "", account)
			if mode == "protected" {
				require.ErrorIs(t, err, context.Canceled)
				require.Zero(t, up.calls)
			} else {
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())
				require.Equal(t, 1, up.calls, "unprotected traffic keeps its original detached lifetime")
			}
			require.Zero(t, cache.calls.Load())
		})
	}
}
