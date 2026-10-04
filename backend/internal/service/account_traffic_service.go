package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type AccountTrafficService struct{ cache AccountTrafficCache }

func NewAccountTrafficService(cache AccountTrafficCache) *AccountTrafficService {
	return &AccountTrafficService{cache: cache}
}

func AccountTrafficFailover(err error) *UpstreamFailoverError {
	var limited *AccountTrafficLimitError
	if errors.As(err, &limited) {
		return limited.FailoverError()
	}
	return nil
}
func accountTrafficController(upstream HTTPUpstream) *AccountTrafficService {
	if provider, ok := upstream.(AccountTrafficProvider); ok {
		return provider.AccountTrafficController()
	}
	return nil
}
func beginAccountTrafficTurn(ctx context.Context, upstream HTTPUpstream, a *Account) (context.Context, *AccountTrafficPermit, error) {
	plan, err := AccountTrafficPlanFor(a)
	if err != nil {
		return ctx, nil, err
	}
	if !plan.Policy.Enabled() {
		return ctx, nil, nil
	}
	if err := ctx.Err(); err != nil {
		return ctx, nil, err
	}
	// Native WS can drain an accepted response after downstream cancellation.
	// Cancel admission with the caller, but retain its lease until the turn exits.
	operationCtx, cancelOperation := context.WithCancel(ctx)
	lifetimeCtx, cancelLifetime := context.WithCancel(context.WithoutCancel(ctx))
	stopAdmissionCancel := context.AfterFunc(ctx, cancelLifetime)
	permitCtx, permit, err := accountTrafficController(upstream).Begin(lifetimeCtx, plan, cancelOperation)
	stopAdmissionCancel()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	if err != nil || permit == nil {
		permit.Finish(0)
		cancelLifetime()
		cancelOperation()
	} else {
		context.AfterFunc(permitCtx, func() { cancelLifetime(); cancelOperation() })
	}
	if limited := AccountTrafficFailover(err); limited != nil {
		return ctx, nil, limited
	}
	if err != nil {
		return ctx, nil, err
	}
	if permit == nil {
		return ctx, nil, nil
	}
	return operationCtx, permit, nil
}
func finishAccountTrafficTurn(permit *AccountTrafficPermit, err error) {
	if permit == nil {
		return
	}
	status := 0
	if err != nil {
		var dialErr *openAIWSDialError
		if errors.As(err, &dialErr) && dialErr != nil && (dialErr.StatusCode == http.StatusTooManyRequests || (dialErr.StatusCode >= 500 && dialErr.StatusCode < 600)) {
			status = dialErr.StatusCode
		}
		var upstream *UpstreamFailoverError
		if status == 0 && errors.As(err, &upstream) && upstream.Reason != "account_traffic_limit" {
			status = upstream.StatusCode
		}
	}
	permit.Finish(status)
}

type accountTrafficFailureRecorder interface {
	RecordFailure(context.Context, AccountTrafficPlan, int) error
}

// recordAccountTrafficWSDialFailure observes only an actual upstream WS
// handshake rejection. Local admission errors, cancellation, network errors,
// and authentication failures must not lower the account's adaptive limit.
func recordAccountTrafficWSDialFailure(ctx context.Context, upstream HTTPUpstream, account *Account, err error) {
	var dialErr *openAIWSDialError
	if !errors.As(err, &dialErr) || dialErr == nil {
		return
	}
	status := dialErr.StatusCode
	if status != http.StatusTooManyRequests && (status < 500 || status >= 600) {
		return
	}
	controller := accountTrafficController(upstream)
	if controller == nil || account == nil {
		return
	}
	plan, planErr := AccountTrafficPlanFor(account)
	if planErr != nil || !plan.Policy.Enabled() {
		return
	}
	recorder, ok := controller.cache.(accountTrafficFailureRecorder)
	if !ok {
		return
	}
	observeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if observeErr := recorder.RecordFailure(observeCtx, plan, status); observeErr != nil {
		slog.Warn("account_traffic_ws_failure_observe_failed", "account_id", plan.AccountID, "status", status, "error", observeErr)
	}
}

type AccountTrafficPermit struct {
	ctx            context.Context
	service        *AccountTrafficService
	plan           AccountTrafficPlan
	id             string
	started        time.Time
	cancel         context.CancelFunc
	stop           chan struct{}
	once           sync.Once
	terminalStatus atomic.Int32
}

func (s *AccountTrafficService) Begin(ctx context.Context, plan AccountTrafficPlan, onLeaseLost ...func()) (context.Context, *AccountTrafficPermit, error) {
	if err := ctx.Err(); err != nil {
		return ctx, nil, err
	}
	if !plan.Policy.Enabled() || ctx.Value(accountTrafficCoveredKey{}) == plan.AccountID {
		return ctx, nil, nil
	}
	if s == nil || s.cache == nil {
		return ctx, nil, &AccountTrafficLimitError{Reason: "流量控制暂不可用，请稍后重试", Status: 503, RetryAfter: time.Second}
	}
	id := uuid.NewString()
	admission, err := s.cache.Acquire(ctx, plan, id)
	deadline := time.Now().Add(time.Duration(plan.Policy.WaitSeconds) * time.Second)
	for err == nil && !admission.Allowed && time.Now().Before(deadline) {
		wait := admission.RetryAfter
		if wait < 100*time.Millisecond {
			wait = 100 * time.Millisecond
		}
		if remaining := time.Until(deadline); wait > remaining {
			wait = remaining
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx, nil, ctx.Err()
		case <-timer.C:
		}
		admission, err = s.cache.Acquire(ctx, plan, id)
	}
	if err != nil {
		slog.Warn("account_traffic_acquire_failed", "account_id", plan.AccountID, "error", err)
		return ctx, nil, &AccountTrafficLimitError{Reason: "流量控制暂不可用，请稍后重试", Status: 503, RetryAfter: time.Second}
	}
	if !admission.Allowed {
		return ctx, nil, &AccountTrafficLimitError{Reason: admission.Reason, RetryAfter: admission.RetryAfter}
	}
	child, cancel := context.WithCancel(context.WithValue(ctx, accountTrafficCoveredKey{}, plan.AccountID))
	p := &AccountTrafficPermit{ctx: child, service: s, plan: plan, id: id, started: time.Now(), cancel: cancel, stop: make(chan struct{})}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-p.stop:
				return
			case <-child.Done():
				p.Finish(0)
				return
			case <-ticker.C:
				refreshCtx, end := context.WithTimeout(context.Background(), 5*time.Second)
				ok, err := s.cache.Refresh(refreshCtx, plan.AccountID, id)
				end()
				if err != nil || !ok {
					p.loseLease(onLeaseLost)
					return
				}
			}
		}
	}()
	return child, p, nil
}

func (p *AccountTrafficPermit) Finish(status int) {
	if p == nil {
		return
	}
	if observed := p.terminalStatus.Load(); observed != 0 {
		status = int(observed)
	}
	p.finishObservation(status)
	p.cancel()
}

func (p *AccountTrafficPermit) loseLease(callbacks []func()) {
	for _, callback := range callbacks {
		callback()
	}
	p.Finish(0)
}

func (p *AccountTrafficPermit) ObserveEvent(payload []byte) {
	if p == nil {
		return
	}
	if status, terminal := accountTrafficEventStatus(payload); terminal {
		p.terminalStatus.Store(int32(status))
	}
}

func (p *AccountTrafficPermit) finishObservation(status int) {
	p.once.Do(func() {
		close(p.stop)
		p.cancel()
		ctx, end := context.WithTimeout(context.Background(), 5*time.Second)
		defer end()
		if err := p.service.cache.Finish(ctx, p.plan, p.id, status, time.Since(p.started).Milliseconds()); err != nil {
			slog.Warn("account_traffic_finish_failed", "account_id", p.plan.AccountID, "error", err)
		}
	})
}

type accountTrafficAdmissionContextKey struct{}

func withAccountTrafficAdmissionContext(req *http.Request, clientCtx context.Context) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), accountTrafficAdmissionContextKey{}, clientCtx))
}

// DoHTTP covers the complete response-body lifetime. Nested plugin/decorator
// transports see the covered marker and cannot double charge a request.
func (s *AccountTrafficService) DoHTTP(req *http.Request, send func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	if req == nil {
		return send(req)
	}
	if _, ok := req.Context().Value(accountTrafficConfigErrorKey{}).(error); ok {
		return nil, &AccountTrafficLimitError{Reason: "账号流量控制配置无效，请管理员检查", Status: 400}
	}
	plan, ok := req.Context().Value(accountTrafficPlanKey{}).(AccountTrafficPlan)
	if !ok || !plan.Policy.Enabled() || req.Context().Value(accountTrafficCoveredKey{}) == plan.AccountID {
		return send(req)
	}
	clientCtx, ok := req.Context().Value(accountTrafficAdmissionContextKey{}).(context.Context)
	if !ok {
		clientCtx = req.Context()
	}
	if err := clientCtx.Err(); err != nil {
		return nil, err
	}
	// Only admission follows client cancellation. Once admitted, retain the
	// detached upstream lifetime so a disconnect cannot interrupt usage drain.
	admissionCtx, cancelAdmission := context.WithCancel(req.Context())
	stopClientCancel := context.AfterFunc(clientCtx, cancelAdmission)
	ctx, permit, err := s.Begin(admissionCtx, plan)
	stopClientCancel()
	if clientErr := clientCtx.Err(); clientErr != nil {
		err = clientErr
	}
	if err != nil {
		permit.Finish(0)
		cancelAdmission()
		return nil, err
	}
	if permit == nil {
		cancelAdmission()
		return send(req)
	}
	context.AfterFunc(ctx, cancelAdmission)
	resp, err := send(req.WithContext(ctx))
	if err != nil || resp == nil {
		permit.Finish(0)
		return resp, err
	}
	if resp.Body == nil {
		permit.Finish(resp.StatusCode)
		return resp, nil
	}
	resp.Body = &accountTrafficBody{ReadCloser: resp.Body, permit: permit, status: resp.StatusCode, sse: strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream"), expectedBytes: resp.ContentLength}
	return resp, nil
}

type accountTrafficBody struct {
	io.ReadCloser
	permit         *AccountTrafficPermit
	status         int
	sse            bool
	line           []byte
	discardLine    bool
	terminalStatus atomic.Int32
	expectedBytes  int64
	readBytes      atomic.Int64
}

func (b *accountTrafficBody) Read(data []byte) (int, error) {
	n, err := b.ReadCloser.Read(data)
	b.readBytes.Add(int64(n))
	if b.sse {
		b.inspect(data[:n])
	}
	if err != nil {
		status := b.status
		if value := b.terminalStatus.Load(); value != 0 {
			status = int(value)
		} else if b.sse && status < 400 {
			status = 0
		}
		if err != io.EOF && status < 400 {
			status = 0
		}
		b.permit.Finish(status)
	}
	return n, err
}
func (b *accountTrafficBody) Close() error {
	err := b.ReadCloser.Close()
	status := b.status
	// Without EOF a 2xx stream might have been abandoned; it is not recovery evidence.
	if value := b.terminalStatus.Load(); value != 0 {
		status = int(value)
	} else if status < 400 && (b.sse || b.expectedBytes <= 0 || b.readBytes.Load() < b.expectedBytes) {
		status = 0
	}
	b.permit.Finish(status)
	return err
}
func (b *accountTrafficBody) inspect(data []byte) {
	for _, ch := range data {
		if ch == '\n' {
			if !b.discardLine {
				line := strings.TrimSpace(string(b.line))
				if strings.HasPrefix(line, "data:") {
					raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
					status, terminal := accountTrafficEventStatus([]byte(raw))
					if raw == "[DONE]" {
						status, terminal = 200, true
					}
					previous := int(b.terminalStatus.Load())
					if terminal && (previous < 400 || status == 429 || status >= 500) {
						b.terminalStatus.Store(int32(status))
					}
				}
			}
			b.line = nil
			b.discardLine = false
		} else if !b.discardLine {
			if len(b.line) >= 64<<10 {
				b.line = nil
				b.discardLine = true
			} else {
				b.line = append(b.line, ch)
			}
		}
	}
}
