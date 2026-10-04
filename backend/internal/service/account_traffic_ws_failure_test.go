package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type wsFailureObservationCache struct {
	AccountTrafficCache
	observed     atomic.Int32
	observations atomic.Int32
}

func (c *wsFailureObservationCache) RecordFailure(_ context.Context, _ AccountTrafficPlan, status int) error {
	c.observed.Store(int32(status))
	c.observations.Add(1)
	return nil
}

func TestCodexProtectionWSHandshakeFailureClassification(t *testing.T) {
	for _, status := range []int{429, 500, 502, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			cache := &protectionTestCache{}
			upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
			_, permit, err := beginAccountTrafficTurn(context.Background(), upstream, protectionAccount())
			require.NoError(t, err)
			failure := wrapOpenAIWSFallback("dial_failed", &openAIWSDialError{
				StatusCode: status, Err: errors.New("upstream handshake rejected"),
			})
			finishAccountTrafficTurn(permit, fmt.Errorf("forward: %w", failure))
			finishAccountTrafficTurn(permit, failure)
			require.EqualValues(t, status, cache.status.Load())
			require.EqualValues(t, 1, cache.finishes.Load(), "one handshake must only be counted once")
		})
	}
}

func TestCodexProtectionWSLocalErrorsAreNotUpstreamFailures(t *testing.T) {
	for _, failure := range []error{
		context.Canceled, context.DeadlineExceeded,
		(&AccountTrafficLimitError{Status: 503, Reason: "local budget"}).FailoverError(),
		wrapOpenAIWSFallback("dial_failed", &openAIWSDialError{StatusCode: 0, Err: errors.New("network timeout")}),
		wrapOpenAIWSFallback("auth_failed", &openAIWSDialError{StatusCode: 401, Err: errors.New("invalid token")}),
	} {
		cache := &protectionTestCache{}
		upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
		_, permit, err := beginAccountTrafficTurn(context.Background(), upstream, protectionAccount())
		require.NoError(t, err)
		finishAccountTrafficTurn(permit, failure)
		require.Zero(t, cache.status.Load(), failure.Error())
		require.EqualValues(t, 1, cache.finishes.Load())
	}
}

func TestCodexProtectionWSNativeHandshakeFailureObservation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "429", err: &openAIWSDialError{StatusCode: 429, Err: errors.New("rate limited")}, status: 429},
		{name: "503", err: &openAIWSDialError{StatusCode: 503, Err: errors.New("overloaded")}, status: 503},
		{name: "network", err: &openAIWSDialError{StatusCode: 0, Err: errors.New("timeout")}},
		{name: "auth", err: &openAIWSDialError{StatusCode: 401, Err: errors.New("invalid token")}},
		{name: "local", err: (&AccountTrafficLimitError{Status: 503, Reason: "local budget"}).FailoverError()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := &wsFailureObservationCache{}
			upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			recordAccountTrafficWSDialFailure(ctx, upstream, protectionAccount(), tc.err)
			if tc.status == 0 {
				require.Zero(t, cache.observations.Load())
			} else {
				require.EqualValues(t, 1, cache.observations.Load())
				require.EqualValues(t, tc.status, cache.observed.Load())
			}
		})
	}
}

func TestCodexProtectionWSNativeHandshakeObservationRequiresEnabledOAuth(t *testing.T) {
	accounts := []*Account{
		func() *Account { a := protectionAccount(); a.Extra = nil; return a }(),
		func() *Account { a := protectionAccount(); a.Type = AccountTypeAPIKey; return a }(),
		func() *Account { a := protectionAccount(); a.Platform = PlatformAnthropic; return a }(),
	}
	for _, account := range accounts {
		cache := &wsFailureObservationCache{}
		upstream := &protectionTestUpstream{control: NewAccountTrafficService(cache)}
		recordAccountTrafficWSDialFailure(context.Background(), upstream, account, &openAIWSDialError{StatusCode: 503, Err: errors.New("overloaded")})
		require.Zero(t, cache.observations.Load())
	}
}
