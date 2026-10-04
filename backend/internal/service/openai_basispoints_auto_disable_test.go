package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type basispointsAutoDisableRepo struct {
	AccountRepository
	disable func(context.Context, *Account) (bool, error)
}

func (r *basispointsAutoDisableRepo) DisableBasispointsOn403(ctx context.Context, account *Account) (bool, error) {
	return r.disable(ctx, account)
}

func TestBasispointsAutoDisableOn403(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		optIn      bool
		modelError bool
		changed    bool
		writeErr   error
		wantCalls  int
	}{
		{name: "default off", status: 403},
		{name: "enabled", status: 403, optIn: true, changed: true, wantCalls: 1},
		{name: "settings already changed", status: 403, optIn: true, wantCalls: 1},
		{name: "write failed", status: 403, optIn: true, writeErr: errors.New("write failed"), wantCalls: 1},
		{name: "model access denied", status: 403, optIn: true, modelError: true},
		{name: "bad request", status: 400, optIn: true},
		{name: "unauthorized", status: 401, optIn: true},
		{name: "rate limited", status: 429, optIn: true},
		{name: "server error", status: 500, optIn: true},
	} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%v", tc.name, stream), func(t *testing.T) {
				account := basispointsAccountForTest()
				if tc.optIn {
					account.Extra["openai_basispoints_auto_disable_on_403"] = true
				}
				raw := `{"error":{"code":"permission_denied","message":"PRIVATE_UPSTREAM"}}`
				wantCode := "basispoints_upstream_error"
				if tc.status == 429 {
					wantCode = "basispoints_rate_limited"
				}
				if tc.modelError {
					raw = `{"error":{"code":"basispoints_model_access_changed"}}`
					wantCode = "basispoints_model_access_changed"
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(raw))}}
				svc := openAIClientToolsTestService(upstream)
				calls := 0
				svc.accountRepo = &basispointsAutoDisableRepo{disable: func(ctx context.Context, got *Account) (bool, error) {
					calls++
					require.NoError(t, ctx.Err())
					require.Same(t, account, got)
					return tc.changed, tc.writeErr
				}}
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				_, err := svc.Forward(context.Background(), c, account, []byte(fmt.Sprintf(`{"model":"gpt-6-astra","input":"test","stream":%v}`, stream)))
				if tc.status >= 500 {
					var failover *UpstreamFailoverError
					require.ErrorAs(t, err, &failover)
					require.False(t, IsResponseCommitted(c), "preserve existing bounded failover")
				} else if tc.status == http.StatusTooManyRequests {
					var failover *UpstreamFailoverError
					require.ErrorAs(t, err, &failover)
					require.Equal(t, http.StatusTooManyRequests, failover.StatusCode)
					require.False(t, IsResponseCommitted(c))
				} else {
					require.EqualError(t, err, "Basispoints: "+wantCode)
					require.Equal(t, tc.status, rec.Code)
					require.Equal(t, wantCode, gjson.Get(rec.Body.String(), "error.code").String())
				}
				require.Equal(t, tc.changed && tc.writeErr == nil, strings.Contains(rec.Body.String(), "now uses ChatGPT Codex"))
				require.NotContains(t, rec.Body.String(), "PRIVATE_UPSTREAM")
				require.Equal(t, tc.wantCalls, calls)
				require.Len(t, upstream.requests, 1, "do not replay the failed request")
				require.True(t, account.UsesBasispointsResponses(), "do not mutate a shared scheduler snapshot")
				require.True(t, account.Schedulable)
				require.Equal(t, StatusActive, account.Status)
			})
		}
	}
}

func TestBasispointsAutoDisableIgnoresTransportAndStreamErrors(t *testing.T) {
	for _, transport := range []bool{false, true} {
		t.Run(fmt.Sprint(transport), func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader("event: error\ndata: {\"type\":\"error\",\"error\":{\"code\":\"basispoints_upstream_error\",\"message\":\"403 forbidden\"}}\n\n"))}}
			if transport {
				upstream.err = errors.New("connection failed")
			}
			account := basispointsAccountForTest()
			account.Extra["openai_basispoints_auto_disable_on_403"] = true
			svc := openAIClientToolsTestService(upstream)
			svc.accountRepo = &basispointsAutoDisableRepo{disable: func(context.Context, *Account) (bool, error) {
				t.Fatal("only an upstream HTTP 403 may disable BPS")
				return false, nil
			}}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			_, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-6-astra","input":"test","stream":true}`))
			require.Error(t, err)
			require.True(t, account.UsesBasispointsResponses())
		})
	}
}

func TestBasispointsAutoDisableUsesBoundedDetachedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	account := basispointsAccountForTest()
	account.Extra["openai_basispoints_auto_disable_on_403"] = true
	var writeContext context.Context
	svc := &OpenAIGatewayService{accountRepo: &basispointsAutoDisableRepo{disable: func(ctx context.Context, _ *Account) (bool, error) {
		writeContext = ctx
		require.NoError(t, ctx.Err())
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(openAIAccountStateUpdateTimeout), deadline, time.Second)
		return true, nil
	}}}
	require.True(t, svc.disableBasispointsOn403(ctx, account))
	require.ErrorIs(t, writeContext.Err(), context.Canceled)
}

func TestAccount_IsBasispointsAutoDisableOn403Enabled(t *testing.T) {
	require.False(t, (*Account)(nil).IsBasispointsAutoDisableOn403Enabled())
	for _, mutate := range []func(*Account){
		func(a *Account) { delete(a.Extra, "openai_basispoints_auto_disable_on_403") },
		func(a *Account) { a.Extra["openai_basispoints_auto_disable_on_403"] = false },
		func(a *Account) { a.Extra["openai_basispoints_auto_disable_on_403"] = "true" },
		func(a *Account) { a.Extra[openAIOAuthResponsesEndpointExtraKey] = openAIOAuthResponsesEndpointChatGPT },
		func(a *Account) { a.Type = AccountTypeAPIKey },
		func(a *Account) { a.Platform = PlatformAnthropic },
		func(a *Account) { parent := int64(1); a.ParentAccountID = &parent },
		func(a *Account) { a.Credentials["auth_mode"] = OpenAIAuthModeAgentIdentity },
		func(a *Account) { a.Credentials["auth_mode"] = "personal_access_token" },
	} {
		account := basispointsAccountForTest()
		account.Extra["openai_basispoints_auto_disable_on_403"] = true
		require.True(t, account.IsBasispointsAutoDisableOn403Enabled())
		mutate(account)
		require.False(t, account.IsBasispointsAutoDisableOn403Enabled())
	}
}

func TestBasispointsAccountTestObservesAutoDisable(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		optIn      bool
		modelError bool
		wantCalls  int
	}{
		{name: "default off", status: 403},
		{name: "enabled", status: 403, optIn: true, wantCalls: 1},
		{name: "model access denied", status: 403, optIn: true, modelError: true},
		{name: "unauthorized", status: 401, optIn: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := basispointsAccountForTest()
			account.Extra[openAIBasispointsAutoDisableOn403Key] = tc.optIn
			raw := `{"error":{"code":"permission_denied","message":"PRIVATE_UPSTREAM"}}`
			if tc.modelError {
				raw = `{"error":{"code":"basispoints_model_access_changed"}}`
			}
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(raw))}}
			gateway := openAIClientToolsTestService(upstream)
			calls := 0
			gateway.accountRepo = &basispointsAutoDisableRepo{disable: func(context.Context, *Account) (bool, error) {
				calls++
				return true, nil
			}}
			svc := &AccountTestService{openaiGatewayService: gateway}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/admin/accounts/300/test", nil)
			require.Error(t, svc.testOpenAIBasispointsAccount(c, account, "gpt-6-astra", "hi"))
			require.Equal(t, tc.wantCalls, calls)
			require.Equal(t, tc.wantCalls == 1, strings.Contains(rec.Body.String(), "now uses ChatGPT Codex"))
			require.NotContains(t, rec.Body.String(), "PRIVATE_UPSTREAM")
			require.Len(t, upstream.requests, 1, "account tests must not replay a rejected request")
			require.True(t, account.UsesBasispointsResponses())
		})
	}
}
