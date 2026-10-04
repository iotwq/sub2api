package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type channelMonitorProbePolicy struct {
	maxAccountSwitches int
	maxAccounts        int
}

type channelMonitorProbeResponseWriter struct {
	gin.ResponseWriter
	attemptStartedAt time.Time
}

func (w *channelMonitorProbeResponseWriter) WriteHeader(code int) {
	w.injectLatencyHeader()
	w.ResponseWriter.WriteHeader(code)
}

func (w *channelMonitorProbeResponseWriter) WriteHeaderNow() {
	w.injectLatencyHeader()
	w.ResponseWriter.WriteHeaderNow()
}

func (w *channelMonitorProbeResponseWriter) Write(data []byte) (int, error) {
	w.injectLatencyHeader()
	return w.ResponseWriter.Write(data)
}

func (w *channelMonitorProbeResponseWriter) WriteString(data string) (int, error) {
	w.injectLatencyHeader()
	return w.ResponseWriter.WriteString(data)
}

func (w *channelMonitorProbeResponseWriter) injectLatencyHeader() {
	if w == nil || w.Written() || w.attemptStartedAt.IsZero() {
		return
	}
	latencyMs := time.Since(w.attemptStartedAt).Milliseconds()
	w.Header().Set(service.ChannelMonitorProbeLatencyHeader, strconv.FormatInt(latencyMs, 10))
}

func (p channelMonitorProbePolicy) allowsAnotherAccount(completedAccounts int) bool {
	return p.maxAccounts <= 0 || completedAccounts < p.maxAccounts
}

func (p channelMonitorProbePolicy) allowsSelectionExhaustedRetry() bool {
	return p.maxAccounts <= 0
}

func resolveChannelMonitorProbePolicy(c *gin.Context, configuredMaxSwitches int) channelMonitorProbePolicy {
	policy := channelMonitorProbePolicy{maxAccountSwitches: configuredMaxSwitches}
	if c == nil {
		return policy
	}
	attempts, err := strconv.Atoi(strings.TrimSpace(c.GetHeader(service.ChannelMonitorProbeAttemptsHeader)))
	if err != nil || attempts != service.ChannelMonitorProbeAttempts {
		return policy
	}

	policy.maxAccounts = attempts
	if _, ok := c.Writer.(*channelMonitorProbeResponseWriter); !ok {
		c.Writer = &channelMonitorProbeResponseWriter{ResponseWriter: c.Writer}
	}
	probeMaxSwitches := attempts - 1
	if configuredMaxSwitches >= 0 && configuredMaxSwitches < probeMaxSwitches {
		probeMaxSwitches = configuredMaxSwitches
	}
	policy.maxAccountSwitches = probeMaxSwitches
	return policy
}

func startChannelMonitorProbeAttempt(c *gin.Context, startedAt time.Time) {
	if c == nil {
		return
	}
	writer, ok := c.Writer.(*channelMonitorProbeResponseWriter)
	if !ok {
		return
	}
	writer.attemptStartedAt = startedAt
}
