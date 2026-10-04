package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service/basispoints"
	"github.com/gin-gonic/gin"
)

// The bridge closes the preceding response before acquiring another account
// slot. Corrections use the same account, proxy and cancellation policy; they
// never trigger account failover or replay an HTTP rejection.
func (s *OpenAIGatewayService) basispointsStreamWithRepairs(ctx context.Context, account *Account, wireBody []byte, token string, bridge *basispoints.Bridge, source io.ReadCloser) io.ReadCloser {
	send := func(repairCtx context.Context, body []byte) (io.ReadCloser, error) {
		if err := repairCtx.Err(); err != nil {
			return nil, err
		}
		req, err := newOpenAIBasispointsRequest(repairCtx, s.accountRepo, account, body, token)
		if err != nil {
			return nil, err
		}
		resp, err := s.doOpenAIBasispoints(withAccountTrafficAdmissionContext(req, repairCtx), account)
		if err != nil {
			if repairCtx.Err() != nil {
				return nil, repairCtx.Err()
			}
			return nil, fmt.Errorf("Basispoints tool correction connection failed")
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if resp.StatusCode == http.StatusTooManyRequests {
				// The original response already produced output; cool the BPS
				// route but never replay the user request after a repair 429.
				s.coolDownExcelBPS(repairCtx, account, resp.Header.Get("Retry-After"))
			}
			stop := context.AfterFunc(repairCtx, func() { _ = resp.Body.Close() })
			raw, _ := io.ReadAll(io.LimitReader(resp.Body, openAIUpstreamErrorBodyReadLimit))
			stop()
			_ = resp.Body.Close()
			s.observeBasispointsUpstreamResponse(repairCtx, account, resp, raw)
			return nil, fmt.Errorf("Basispoints tool correction returned HTTP %d", resp.StatusCode)
		}
		s.observeBasispointsUpstreamResponse(repairCtx, account, resp, nil)
		return resp.Body, nil
	}
	return bridge.StreamWithRepairs(ctx, source, func(repairCtx context.Context, failed map[string]any, validation error) (map[string]any, error) {
		corrected, err := basispoints.BuildToolRepairRequest(wireBody, failed, validation)
		if err != nil {
			return nil, err
		}
		body, err := send(repairCtx, corrected)
		if err != nil {
			return nil, err
		}
		defer body.Close()
		stop := context.AfterFunc(repairCtx, func() { _ = body.Close() })
		defer stop()
		wireBody = corrected
		return basispoints.ReadToolRepairResponse(body)
	}, func(repairCtx context.Context) (io.ReadCloser, error) {
		repaired, err := basispoints.RepairRequest(wireBody)
		if err != nil {
			return nil, err
		}
		return send(repairCtx, repaired)
	})
}

// A completed BPS HTTP exchange may contain a failed correction with billable
// usage. Preserve it without entering Codex terminal failover/health policies.
func (s *OpenAIGatewayService) handleBasispointsTerminalFailure(c *gin.Context, account *Account, resp *http.Response, wire, terminal []byte) (*openaiNonStreamingResult, error) {
	usage := s.parseSSEUsageFromBody(string(wire))
	result := &openaiNonStreamingResult{OpenAIUsage: usage, usage: usage, responseID: extractOpenAIResponseIDFromJSONBytes(terminal)}
	failure := s.recordBasispointsTerminalFailure(c, account, resp, terminal, false)
	committed := StopOpenAICompactSSEKeepaliveCommitted(c)
	MarkResponseCommitted(c)
	if committed {
		writeOpenAICompactSSEFailureMessage(c, failure.Status, failure.Code, failure.Message)
	} else {
		c.JSON(failure.Status, gin.H{"error": failure.Details()})
	}
	return result, failure
}
