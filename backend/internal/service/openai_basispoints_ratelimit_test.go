package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type basispointsQuotaWrite struct {
	accountID int64
	updates   map[string]any
	resetAt   time.Time
	ctxErr    error
	deadline  time.Time
}

type basispointsQuotaRepo struct {
	AccountRepository
	writes    chan basispointsQuotaWrite
	updateErr error
	limitErr  error
}

type basispointsQuotaSettingsRepo struct {
	SettingRepository
	cooldown string
}

func (r *basispointsQuotaSettingsRepo) GetValue(_ context.Context, key string) (string, error) {
	if key == SettingKeyRateLimit429CooldownSettings {
		return r.cooldown, nil
	}
	return "", ErrSettingNotFound
}

func (r *basispointsQuotaRepo) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	copied := make(map[string]any, len(updates))
	for key, value := range updates {
		copied[key] = value
	}
	deadline, _ := ctx.Deadline()
	r.writes <- basispointsQuotaWrite{accountID: id, updates: copied, ctxErr: ctx.Err(), deadline: deadline}
	return r.updateErr
}

func (r *basispointsQuotaRepo) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	deadline, _ := ctx.Deadline()
	r.writes <- basispointsQuotaWrite{accountID: id, resetAt: resetAt, ctxErr: ctx.Err(), deadline: deadline}
	return r.limitErr
}

func nextBasispointsQuotaWrite(t *testing.T, repo *basispointsQuotaRepo) basispointsQuotaWrite {
	t.Helper()
	select {
	case write := <-repo.writes:
		return write
	case <-time.After(3 * time.Second):
		t.Fatal("expected an account quota update")
		return basispointsQuotaWrite{}
	}
}

func requireNoBasispointsQuotaWrite(t *testing.T, repo *basispointsQuotaRepo) {
	t.Helper()
	select {
	case write := <-repo.writes:
		t.Fatalf("unexpected account quota update: %+v", write)
	case <-time.After(20 * time.Millisecond):
	}
}

func basispointsQuotaHeaders(used5h, used7d string) http.Header {
	headers := http.Header{}
	headers.Set("x-codex-primary-used-percent", used7d)
	headers.Set("x-codex-primary-reset-after-seconds", "604800")
	headers.Set("x-codex-primary-window-minutes", "10080")
	headers.Set("x-codex-secondary-used-percent", used5h)
	headers.Set("x-codex-secondary-reset-after-seconds", "18000")
	headers.Set("x-codex-secondary-window-minutes", "300")
	return headers
}

func TestBasispointsAccountTestObservesQuotaState(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusTooManyRequests} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			wire := "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_bps_test\",\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"hi\"}]}]}}\n\n"
			used5h := "63"
			if status == http.StatusTooManyRequests {
				wire, used5h = `{"error":{"type":"rate_limit_error"}}`, "100"
			}
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: status, Header: basispointsQuotaHeaders(used5h, "28"), Body: io.NopCloser(strings.NewReader(wire))}}
			gateway := openAIClientToolsTestService(upstream)
			repo := &basispointsQuotaRepo{writes: make(chan basispointsQuotaWrite, 4)}
			gateway.accountRepo = repo
			gateway.codexSnapshotThrottle = newAccountWriteThrottle(time.Hour)
			gateway.rateLimitService = NewRateLimitService(repo, nil, gateway.cfg, nil, nil)
			svc := &AccountTestService{openaiGatewayService: gateway}
			account := basispointsAccountForTest()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/admin/accounts/300/test", nil)
			err := svc.testOpenAIBasispointsAccount(c, account, "gpt-6-astra", "hi")
			if status == http.StatusOK {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			if status == http.StatusOK {
				write := nextBasispointsQuotaWrite(t, repo)
				require.Equal(t, account.ID, write.accountID)
				require.Equal(t, 28.0, write.updates["codex_7d_used_percent"])
			}
			requireNoBasispointsQuotaWrite(t, repo)
			require.Len(t, upstream.requests, 1)
			require.Nil(t, account.RateLimitResetAt)
		})
	}
}

func TestBasispointsQuotaSnapshotAtResponseBoundary(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, brokenStream := range []bool{false, true} {
			t.Run(fmt.Sprintf("stream=%t/broken=%t", stream, brokenStream), func(t *testing.T) {
				wire := "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_bps_quota\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":10,\"output_tokens\":2}}}\n\n"
				if brokenStream {
					wire = ""
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: http.StatusOK, Header: basispointsQuotaHeaders("63", "28"), Body: io.NopCloser(strings.NewReader(wire)),
				}}
				svc := openAIClientToolsTestService(upstream)
				svc.codexSnapshotThrottle = newAccountWriteThrottle(time.Hour)
				repo := &basispointsQuotaRepo{writes: make(chan basispointsQuotaWrite, 4)}
				svc.accountRepo = repo
				account := basispointsAccountForTest()
				body := []byte(fmt.Sprintf("{\"model\":\"gpt-6-astra\",\"input\":\"test\",\"stream\":%t}", stream))
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				_, err := svc.Forward(context.Background(), c, account, body)
				if brokenStream {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				write := nextBasispointsQuotaWrite(t, repo)
				require.Equal(t, account.ID, write.accountID)
				require.Equal(t, 63.0, write.updates["codex_5h_used_percent"])
				require.Equal(t, 28.0, write.updates["codex_7d_used_percent"])
				updatedAtValue, ok := write.updates["codex_usage_updated_at"].(string)
				require.True(t, ok)
				updatedAt, err := time.Parse(time.RFC3339, updatedAtValue)
				require.NoError(t, err)
				require.WithinDuration(t, time.Now(), updatedAt, 5*time.Second)
				require.NoError(t, write.ctxErr)
				require.True(t, write.resetAt.IsZero())
				require.NotContains(t, account.Extra, "codex_usage_updated_at", "do not mutate a shared account snapshot")
				// A second response still succeeds but uses the existing snapshot throttle.
				upstream.resp.Body = io.NopCloser(strings.NewReader(wire))
				c, _ = gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				_, _ = svc.Forward(context.Background(), c, account, body)
				requireNoBasispointsQuotaWrite(t, repo)
			})
		}
	}
}

func TestBasispointsQuotaSnapshotRequiresHeaders(t *testing.T) {
	for _, headers := range []http.Header{nil, {"X-Codex-Primary-Used-Percent": {"invalid"}}} {
		upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: headers, Body: io.NopCloser(strings.NewReader(""))}}
		svc := openAIClientToolsTestService(upstream)
		svc.codexSnapshotThrottle = newAccountWriteThrottle(time.Hour)
		repo := &basispointsQuotaRepo{writes: make(chan basispointsQuotaWrite, 4)}
		svc.accountRepo = repo
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
		_, _ = svc.Forward(context.Background(), c, basispointsAccountForTest(), []byte("{\"model\":\"gpt-6-astra\",\"input\":\"test\"}"))
		requireNoBasispointsQuotaWrite(t, repo)
	}
}

type basispointsCancelReader struct {
	io.Reader
	cancel context.CancelFunc
}

func (r *basispointsCancelReader) Read(p []byte) (int, error) {
	r.cancel()
	return r.Reader.Read(p)
}

func TestBasispoints429FailsOverWithoutChangingCodexQuotaOrScheduling(t *testing.T) {
	for _, tc := range []struct {
		name         string
		headers      http.Header
		raw          string
		cancelOnRead bool
	}{
		{name: "quota headers", headers: basispointsQuotaHeaders("100", "100")},
		{name: "unexhausted headers", headers: basispointsQuotaHeaders("30", "20")},
		{name: "body reset timestamp", raw: fmt.Sprintf(`{"error":{"type":"usage_limit_reached","resets_at":%d}}`, time.Now().Add(2*time.Hour).Unix())},
		{name: "body reset duration", raw: `{"error":{"type":"usage_limit_reached","resets_in_seconds":7200}}`},
		{name: "generic rate limit"},
		{name: "malformed body", raw: "not json"},
		{name: "client cancellation", headers: basispointsQuotaHeaders("100", "100"), cancelOnRead: true},
	} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", tc.name, stream), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				raw := tc.raw
				if raw == "" {
					raw = `{"error":{"type":"rate_limit_error","code":"rate_limit_exceeded","message":"PRIVATE_UPSTREAM"}}`
				}
				var reader io.Reader = strings.NewReader(raw)
				if tc.cancelOnRead {
					reader = &basispointsCancelReader{Reader: reader, cancel: cancel}
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusTooManyRequests, Header: tc.headers, Body: io.NopCloser(reader)}}
				svc := openAIClientToolsTestService(upstream)
				repo := &basispointsQuotaRepo{writes: make(chan basispointsQuotaWrite, 4)}
				svc.accountRepo = repo
				svc.rateLimitService = NewRateLimitService(repo, nil, svc.cfg, nil, nil)
				svc.rateLimitService.SetSettingService(NewSettingService(&basispointsQuotaSettingsRepo{cooldown: `{"enabled":true,"cooldown_seconds":11}`}, svc.cfg))
				svc.rateLimitService.SetAccountRuntimeBlocker(svc)
				account := basispointsAccountForTest()
				account.Extra["openai_basispoints_auto_disable_on_403"] = true
				retryStarted := time.Now()
				svc.openaiOAuth429RetryStartedAt.Store(account.ID, retryStarted)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
				body := []byte(fmt.Sprintf(`{"model":"gpt-6-astra","input":"test","stream":%t}`, stream))

				_, err := svc.Forward(ctx, c, account, body)

				var failover *UpstreamFailoverError
				if tc.cancelOnRead {
					require.ErrorIs(t, err, context.Canceled)
					require.NotErrorAs(t, err, &failover)
				} else {
					require.ErrorAs(t, err, &failover)
					require.Equal(t, ExcelBPSRateLimitedReason, failover.Reason)
					require.Equal(t, http.StatusTooManyRequests, failover.StatusCode)
					require.Equal(t, http.StatusTooManyRequests, failover.ClientStatusCode)
				}
				require.Equal(t, tc.cancelOnRead, IsResponseCommitted(c))
				require.Empty(t, rec.Body.String())
				require.NotContains(t, rec.Body.String(), "PRIVATE_UPSTREAM")
				require.Len(t, upstream.requests, 1)
				requireNoBasispointsQuotaWrite(t, repo)
				require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
				require.True(t, account.IsSchedulable())
				require.True(t, account.UsesBasispointsResponses())
				require.Nil(t, account.RateLimitResetAt)
				require.True(t, svc.isExcelBPSCoolingDown(account, "gpt-6-astra"))
				retryState, ok := svc.openaiOAuth429RetryStartedAt.Load(account.ID)
				require.True(t, ok)
				require.Equal(t, retryStarted, retryState)
			})
		}
	}
}

func TestBasispointsNon429DoesNotChangeQuotaState(t *testing.T) {
	for _, status := range []int{400, 403, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: status, Header: basispointsQuotaHeaders("100", "100"), Body: io.NopCloser(strings.NewReader("{}"))}}
			svc := openAIClientToolsTestService(upstream)
			repo := &basispointsQuotaRepo{writes: make(chan basispointsQuotaWrite, 4)}
			svc.accountRepo = repo
			svc.rateLimitService = NewRateLimitService(repo, nil, svc.cfg, nil, nil)
			svc.rateLimitService.SetAccountRuntimeBlocker(svc)
			account := basispointsAccountForTest()
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			_, err := svc.Forward(context.Background(), c, account, []byte("{\"model\":\"gpt-6-astra\",\"input\":\"test\"}"))
			require.Error(t, err)
			if status < 500 {
				require.Equal(t, status, rec.Code)
			}
			requireNoBasispointsQuotaWrite(t, repo)
			require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
			require.True(t, account.IsSchedulable())
		})
	}
}

func TestExcelBPSRetryAfterAcceptsSecondsAndHTTPDate(t *testing.T) {
	now := time.Date(2026, 9, 28, 15, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name  string
		value string
		want  time.Duration
		ok    bool
	}{
		{name: "delta seconds", value: "60", want: time.Minute, ok: true},
		{name: "http date", value: now.Add(2 * time.Minute).Format(http.TimeFormat), want: 2 * time.Minute, ok: true},
		{name: "invalid", value: "not-a-retry-after", ok: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := excelBPSRetryAfter(tc.value, now)
			require.Equal(t, tc.ok, ok)
			if ok {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestExcelBPSCooldownOnlyAppliesToBasispointsAccounts(t *testing.T) {
	svc := openAIClientToolsTestService(nil)
	account := basispointsAccountForTest()
	svc.coolDownExcelBPS(context.Background(), account, "60")
	require.True(t, svc.isExcelBPSCoolingDown(account, "gpt-6-astra"))

	account.Extra[openAIOAuthResponsesEndpointExtraKey] = openAIOAuthResponsesEndpointChatGPT
	require.False(t, svc.isExcelBPSCoolingDown(account, "gpt-6-astra"))
}
