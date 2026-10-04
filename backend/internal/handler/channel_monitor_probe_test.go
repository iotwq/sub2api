package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResolveChannelMonitorProbePolicy(t *testing.T) {
	newContext := func(header string) *gin.Context {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
		if header != "" {
			c.Request.Header.Set(service.ChannelMonitorProbeAttemptsHeader, header)
		}
		return c
	}

	t.Run("ordinary request keeps configured failover", func(t *testing.T) {
		policy := resolveChannelMonitorProbePolicy(newContext(""), 10)
		require.Equal(t, 10, policy.maxAccountSwitches)
		require.Equal(t, 0, policy.maxAccounts)
		require.True(t, policy.allowsAnotherAccount(100))
		require.True(t, policy.allowsSelectionExhaustedRetry())
	})

	t.Run("monitor probes at most five distinct accounts", func(t *testing.T) {
		policy := resolveChannelMonitorProbePolicy(
			newContext(strconv.Itoa(service.ChannelMonitorProbeAttempts)),
			10,
		)
		require.Equal(t, 4, policy.maxAccountSwitches)
		require.Equal(t, 5, policy.maxAccounts)
		require.True(t, policy.allowsAnotherAccount(4))
		require.False(t, policy.allowsAnotherAccount(5))
		require.False(t, policy.allowsSelectionExhaustedRetry())
	})

	t.Run("configured lower switch limit remains authoritative", func(t *testing.T) {
		policy := resolveChannelMonitorProbePolicy(
			newContext(strconv.Itoa(service.ChannelMonitorProbeAttempts)),
			1,
		)
		require.Equal(t, 1, policy.maxAccountSwitches)
		require.Equal(t, 5, policy.maxAccounts)
	})

	t.Run("unrecognized value cannot change routing", func(t *testing.T) {
		policy := resolveChannelMonitorProbePolicy(newContext("99"), 10)
		require.Equal(t, 10, policy.maxAccountSwitches)
		require.Equal(t, 0, policy.maxAccounts)
	})
}

func TestChannelMonitorProbePolicyCapsFailoverAtFiveAccounts(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set(service.ChannelMonitorProbeAttemptsHeader, strconv.Itoa(service.ChannelMonitorProbeAttempts))

	policy := resolveChannelMonitorProbePolicy(c, 10)
	state := NewFailoverState(policy.maxAccountSwitches, false)
	failoverErr := &service.UpstreamFailoverError{StatusCode: http.StatusServiceUnavailable}
	unscheduler := &mockTempUnscheduler{}

	for accountID := int64(1); accountID <= service.ChannelMonitorProbeAttempts; accountID++ {
		action := state.HandleFailoverError(c.Request.Context(), unscheduler, accountID, service.PlatformAnthropic, 0, failoverErr)
		if accountID < service.ChannelMonitorProbeAttempts {
			require.Equal(t, FailoverContinue, action)
			continue
		}
		require.Equal(t, FailoverExhausted, action)
	}

	require.Equal(t, service.ChannelMonitorProbeAttempts-1, state.SwitchCount)
	require.Len(t, state.FailedAccountIDs, service.ChannelMonitorProbeAttempts)
}

func TestChannelMonitorProbeLatencyUsesCurrentAccountAttempt(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Request.Header.Set(service.ChannelMonitorProbeAttemptsHeader, strconv.Itoa(service.ChannelMonitorProbeAttempts))

	resolveChannelMonitorProbePolicy(c, 10)
	startChannelMonitorProbeAttempt(c, time.Now().Add(-5*time.Second))
	startChannelMonitorProbeAttempt(c, time.Now().Add(-20*time.Millisecond))
	c.JSON(http.StatusOK, gin.H{"ok": true})

	latencyMs, err := strconv.Atoi(rec.Header().Get(service.ChannelMonitorProbeLatencyHeader))
	require.NoError(t, err)
	require.GreaterOrEqual(t, latencyMs, 0)
	require.Less(t, latencyMs, 500, "previous failed account latency must not be accumulated")
}

func TestOrdinaryRequestDoesNotExposeChannelMonitorProbeLatency(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	resolveChannelMonitorProbePolicy(c, 10)
	startChannelMonitorProbeAttempt(c, time.Now().Add(-20*time.Millisecond))
	c.JSON(http.StatusOK, gin.H{"ok": true})

	require.Empty(t, rec.Header().Get(service.ChannelMonitorProbeLatencyHeader))
}
