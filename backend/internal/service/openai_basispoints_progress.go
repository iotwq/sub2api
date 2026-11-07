package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service/basispoints"
	"github.com/gin-gonic/gin"
)

func (s *OpenAIGatewayService) guardBasispointsProgress(ctx context.Context, response *http.Response) {
	if response == nil || response.Body == nil {
		return
	}
	idle := 120 * time.Second
	if s.cfg != nil {
		idle = time.Duration(s.cfg.Gateway.ExcelBPSStreamDataIntervalTimeout) * time.Second
	}
	response.Body = basispoints.WithProgressTimeout(ctx, response.Body, idle)
}

// Keep partial usage when an accepted BPS response times out while buffering.
func (s *OpenAIGatewayService) handleBasispointsBufferedResponse(resp *http.Response, c *gin.Context, account *Account, originalModel, model string) (*openaiNonStreamingResult, error) {
	limit := resolveUpstreamResponseReadLimit(s.cfg)
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || int64(len(body)) > limit {
		usage := s.parseSSEUsageFromBody(string(body))
		return &openaiNonStreamingResult{OpenAIUsage: usage, usage: usage}, fmt.Errorf("Basispoints stream ended before completion")
	}
	observer := upstreamResponseModelObserverFromContext(c)
	if observer == nil {
		observer = beginUpstreamResponseModelObservation(c)
	}
	observeOpenAISSEBody(observer, string(body))
	return s.handleSSEToJSON(resp, c, account, body, originalModel, model)
}

func (s *OpenAIGatewayService) writeBasispointsIncomplete(c *gin.Context, resp *http.Response, stream, chat bool) {
	const code = "basispoints_stream_incomplete"
	const message = "Basispoints stream ended before completion; request was not replayed"
	setOpsUpstreamError(c, resp.StatusCode, message, "")
	MarkOpsStreamErrorValue(c, OpsStreamError{ErrType: "server_error", Code: code, Message: message, UpstreamStatus: resp.StatusCode, IntendedStatus: http.StatusBadGateway, CountTowardsSLA: true, NonStream: !stream})
	keepalive := StopOpenAICompactSSEKeepaliveCommitted(c)
	MarkResponseCommitted(c)
	if stream || keepalive {
		c.Header("Content-Type", "text/event-stream")
		if chat {
			_, _ = fmt.Fprintf(c.Writer, "data: {\"error\":{\"type\":\"server_error\",\"code\":\"%s\",\"message\":\"%s\"}}\n\n", code, message)
			c.Writer.Flush()
		} else {
			writeOpenAICompactSSEFailureMessage(c, http.StatusBadGateway, code, message)
		}
	} else {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"type": "server_error", "code": code, "message": message}})
	}
}
