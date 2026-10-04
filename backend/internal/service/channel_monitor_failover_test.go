//go:build unit

package service

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func channelMonitorTestContext() *gin.Context {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))
	c.Request.Header.Set(ChannelMonitorProbeAttemptsHeader, strconv.Itoa(ChannelMonitorProbeAttempts))
	return c
}

func TestShouldFailoverChannelMonitorProbeError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "policy rejection code",
			body: `{"error":{"code":"probe_request_rejected","message":"request blocked by gateway policy"}}`,
			want: true,
		},
		{
			name: "wrapped policy rejection",
			body: `{"detail":"{\"error\":{\"code\":\"probe_request_rejected\",\"message\":\"request blocked by gateway policy\"}}"}`,
			want: true,
		},
		{
			name: "ordinary invalid request",
			body: `{"error":{"type":"invalid_request_error","message":"missing messages"}}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shouldFailoverChannelMonitorProbeError(channelMonitorTestContext(), http.StatusBadRequest, []byte(tt.body)))
		})
	}
}

func TestShouldFailoverChannelMonitorProbeErrorRequiresInternalHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))
	body := []byte(`{"error":{"code":"probe_request_rejected","message":"request blocked by gateway policy"}}`)

	require.False(t, shouldFailoverChannelMonitorProbeError(c, http.StatusBadRequest, body))
	require.False(t, shouldFailoverChannelMonitorProbeError(channelMonitorTestContext(), http.StatusUnauthorized, body))
}
