package service

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service/basispoints"
	"github.com/gin-gonic/gin"
)

const basispointsFailurePayloadKey = "basispoints_failure_payload"

// Preserve the safe bridge event before the shared Chat types discard the
// semantic status/type. Only BPS callers consume this request-local value.
func rememberBasispointsFailure(c *gin.Context, payload []byte) {
	if GetActualOpenAIUpstreamEndpoint(c) == "/basispoints/api/responses" && basispoints.ParseUpstreamFailure(payload) != nil {
		c.Set(basispointsFailurePayloadKey, string(payload))
	}
}

func (s *OpenAIGatewayService) recordBasispointsTerminalFailure(c *gin.Context, account *Account, resp *http.Response, terminal []byte, stream bool) *basispoints.UpstreamFailure {
	if len(terminal) == 0 {
		terminal = []byte(c.GetString(basispointsFailurePayloadKey))
	}
	failure := basispoints.ParseUpstreamFailure(terminal)
	if failure == nil {
		return nil
	}
	rememberBasispointsFailure(c, terminal)
	// A failure within an accepted HTTP response is not an HTTP rejection.
	// Do not invoke Codex account health, auth recovery, or failover policies.
	setOpsUpstreamError(c, resp.StatusCode, failure.Message, "")
	MarkOpsStreamErrorValue(c, OpsStreamError{
		ErrType: failure.Type, Code: failure.Code, Message: failure.Message,
		IntendedStatus: failure.Status, CountTowardsSLA: true, NonStream: !stream,
	})
	if failure.Status == http.StatusTooManyRequests {
		s.coolDownExcelBPS(c.Request.Context(), account, resp.Header.Get("Retry-After"))
	}
	return failure
}
