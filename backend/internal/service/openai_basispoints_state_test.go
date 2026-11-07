package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestBasispointsSharedStateRequiresStoreOnlyForStableSessions(t *testing.T) {
	for _, stable := range []bool{true, false} {
		upstream := &httpUpstreamRecorder{resp: bpsCompletionResponse(200, "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"output\":[]}}\n\n")}
		svc := openAIClientToolsTestService(upstream)
		svc.cfg.Gateway.ExcelBPSStateRedisEnabled = true
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
		if stable {
			c.Request.Header.Set("thread-id", "synthetic-thread")
		}
		_, err := svc.Forward(context.Background(), c, basispointsAccountForTest(), []byte(`{"model":"gpt-6-astra","input":"test","tools":[{"type":"function","name":"shell"}]}`))
		if stable {
			require.Error(t, err)
			require.Equal(t, 503, rec.Code)
			require.Empty(t, upstream.requests)
			require.Contains(t, rec.Body.String(), "basispoints_state_unavailable")
		} else {
			require.NoError(t, err)
			require.Len(t, upstream.requests, 1)
		}
	}
}
