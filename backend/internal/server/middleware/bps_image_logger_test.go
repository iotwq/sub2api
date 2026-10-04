package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBPSImageRequestLogPathRedactsCapability(t *testing.T) {
	require.Equal(t, "/api/bps-images/[redacted]", requestLogPath("/api/bps-images/secret-capability"))
	require.Equal(t, "/api/bps-images/[redacted]", requestLogPath("/api/bps-images/secret/malformed"))
	require.Equal(t, "/v1/responses", requestLogPath("/v1/responses"))
	require.Equal(t, "/health", requestLogPath("/health"))
}
