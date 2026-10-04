package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func signalHeader(length int) http.Header {
	h := http.Header{}
	h.Set(openAICodexTurnStateHeader, "gAAAAA"+strings.Repeat("x", length-6))
	return h
}

func signalAccount() *Account {
	return &Account{ID: 41, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1,
		Extra: map[string]any{"codex_ticket_enabled": true, "codex_ticket_type": "team"}}
}

func TestCodexSignalObservationEligibilityAndMissingHeaders(t *testing.T) {
	d := &DeferredService{}
	s := &OpenAIGatewayService{deferredService: d}
	a := signalAccount()
	s.observeCodexSignal(a, signalHeader(312))
	require.Equal(t, 312, d.codexSignals[a.ID].Length)
	first := d.codexSignals[a.ID]
	for _, h := range []http.Header{nil, {"X-Codex-Turn-State": {strings.Repeat("x", 312)}}, {"X-Codex-Turn-State": {"gAAAAA invalid"}}} {
		s.observeCodexSignal(a, h)
		require.Equal(t, first, d.codexSignals[a.ID], "missing/malformed state must not clear an observation")
	}
	for _, account := range []*Account{nil, {ID: 42, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, {ID: 43, Platform: PlatformAnthropic, Type: AccountTypeOAuth}, {ID: 44, Platform: PlatformOpenAI, Type: AccountTypeOAuth, ParentAccountID: &a.ID}} {
		s.observeCodexSignal(account, signalHeader(312))
	}
	require.Len(t, d.codexSignals, 1)
	s.observeCodexSignal(a, signalHeader(332))
	require.Equal(t, 332, d.codexSignals[a.ID].Length)
	require.Equal(t, first.Last312At, d.codexSignals[a.ID].Last312At)
	a.Type = AccountTypeSetupToken
	s.observeCodexSignal(a, signalHeader(292))
	require.Equal(t, 292, d.codexSignals[a.ID].Length)
}

type signalHTTPUpstream struct {
	HTTPUpstream
	response *http.Response
	seen     http.Header
}

func (u *signalHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.seen = req.Header.Clone()
	return u.response, nil
}

func TestCodexSignalHTTPAndAccountTestObserveResponseOnly(t *testing.T) {
	for _, accountTest := range []bool{false, true} {
		for _, code := range []int{200, 503} {
			d := &DeferredService{}
			body := io.NopCloser(strings.NewReader("unchanged response"))
			u := &signalHTTPUpstream{response: &http.Response{StatusCode: code, Header: signalHeader(312), Body: body}}
			s := &OpenAIGatewayService{deferredService: d, httpUpstream: u}
			a := signalAccount()
			require.False(t, s.isOpenAIAccountRequestRuntimeBlocked(a, "gpt-6-astra", false), "legacy ticket opt-in must not gate scheduling")
			req, err := http.NewRequest(http.MethodPost, "https://upstream.test/responses", nil)
			require.NoError(t, err)
			req.Header = signalHeader(292)
			var response *http.Response
			if accountTest {
				tester := &AccountTestService{openaiGatewayService: s, httpUpstream: u}
				response, err = tester.doOpenAIAccountTestUpstream(req, "", a, false)
			} else {
				response, err = s.doOpenAIUpstream(req, "", a)
			}
			require.NoError(t, err)
			require.Same(t, u.response, response)
			require.Equal(t, body, response.Body)
			content, readErr := io.ReadAll(response.Body)
			require.NoError(t, readErr)
			require.Equal(t, "unchanged response", string(content))
			require.Equal(t, req.Header, u.seen)
			if code == 200 {
				require.Equal(t, 312, d.codexSignals[a.ID].Length)
			} else {
				require.Empty(t, d.codexSignals)
			}
			u.response.Header = nil
			delete(d.codexSignals, a.ID)
			_, err = s.doOpenAIUpstream(req, "", a)
			require.NoError(t, err)
			require.Empty(t, d.codexSignals, "request header is not evidence")
		}
	}
}

type signalWSDialer struct{ conn openAIWSClientConn }

func (d *signalWSDialer) Dial(context.Context, string, http.Header, string) (openAIWSClientConn, int, http.Header, error) {
	if d.conn != nil {
		return d.conn, http.StatusSwitchingProtocols, signalHeader(312), nil
	}
	return &openAIWSFakeConn{}, http.StatusSwitchingProtocols, signalHeader(312), nil
}

func TestCodexSignalNativeWSPassthrough(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	upstream := newStagedPassthroughConn()
	cfg := passthroughLifecycleConfig()
	cfg.Gateway.OpenAIWS.OAuthEnabled = true
	s := newPassthroughLifecycleService(cfg, upstream)
	s.deferredService = &DeferredService{}
	s.openaiWSPassthroughDialer = &signalWSDialer{conn: upstream}
	a := signalAccount()
	a.Credentials = map[string]any{"access_token": "sk-test", "chatgpt_account_id": "test-account"}
	a.Extra["openai_oauth_responses_websockets_v2_mode"] = OpenAIWSIngressModePassthrough
	server, serverErr := startPassthroughLifecycleServer(t, ctx, s, a)
	defer server.Close()
	client := dialPassthroughLifecycleClient(t, server)
	defer func() { _ = client.CloseNow() }()
	select {
	case <-upstream.writes:
	case err := <-serverErr:
		t.Fatalf("passthrough failed before forwarding: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("passthrough did not forward")
	}
	s.deferredService.codexSignalMu.Lock()
	observed := s.deferredService.codexSignals[a.ID]
	s.deferredService.codexSignalMu.Unlock()
	require.Equal(t, 312, observed.Length)
	upstream.Send(`{"type":"response.completed","response":{"id":"resp_signal","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`)
	frame, err := readPassthroughLifecycleFrame(t, client, time.Second)
	require.NoError(t, err)
	require.Contains(t, string(frame), "resp_signal")
	_ = client.CloseNow()
	cancel()
	select {
	case <-serverErr:
	case <-time.After(3 * time.Second):
		t.Fatal("passthrough did not exit")
	}
}

func TestCodexSignalPoolObservesFreshHandshakeOnly(t *testing.T) {
	d := &DeferredService{}
	s := &OpenAIGatewayService{cfg: &config.Config{}, deferredService: d}
	p := s.getOpenAIWSConnPool()
	t.Cleanup(p.Close)
	p.clientDialer = &signalWSDialer{}
	a := signalAccount()
	request := openAIWSAcquireRequest{Account: a, WSURL: "wss://upstream.test/responses"}
	lease, err := p.Acquire(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, 312, d.codexSignals[a.ID].Length)
	first := d.codexSignals[a.ID]
	lease.Release()
	lease, err = p.Acquire(context.Background(), request)
	require.NoError(t, err)
	require.True(t, lease.Reused())
	require.Equal(t, first, d.codexSignals[a.ID], "reused handshake must not refresh observation time")
	lease.Release()
}

type signalRepo struct {
	AccountRepository
	fail     bool
	onRecord func()
	signal   CodexSignalStatus
}

func (r *signalRepo) RecordCodexSignal(_ context.Context, _ int64, signal CodexSignalStatus) error {
	if r.onRecord != nil {
		r.onRecord()
	}
	if r.fail {
		return errors.New("test storage failure")
	}
	r.signal = MergeCodexSignals(r.signal, signal)
	return nil
}

func TestCodexSignalCoalescingRetryAndConcurrentUpdates(t *testing.T) {
	r := &signalRepo{fail: true}
	d := &DeferredService{accountRepo: r}
	start := time.Now().UTC()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ts := start.Add(time.Duration(i) * time.Second)
			signal := CodexSignalStatus{Length: 292, ObservedAt: ts}
			if i == 50 {
				signal.Length = 312
				signal.Last312At = &ts
			}
			d.scheduleCodexSignal(41, signal)
		}(i)
	}
	wg.Wait()
	r.onRecord = func() {
		d.scheduleCodexSignal(41, CodexSignalStatus{Length: 332, ObservedAt: start.Add(100 * time.Second)})
	}
	d.flushCodexSignals()
	require.Equal(t, 332, d.codexSignals[41].Length)
	require.Equal(t, start.Add(50*time.Second), *d.codexSignals[41].Last312At)
	r.fail = false
	r.onRecord = nil
	d.flushCodexSignals()
	require.Empty(t, d.codexSignals)
	require.Equal(t, 332, r.signal.Length)
	require.Equal(t, start.Add(50*time.Second), *r.signal.Last312At)
}
