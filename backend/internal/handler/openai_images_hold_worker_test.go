package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestImagesBillingWorkerOwnsHoldAfterHandlerReturns(t *testing.T) {
	for _, outcome := range []string{"success", "durable pending", "pending write failure"} {
		t.Run(outcome, func(t *testing.T) {
			h, billing, _, upstream, key := newNanoBananaBillingTestHandler(t, 100)
			upstream.responseBody = `{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aSfoAAAAASUVORK5CYII="}]}`
			if outcome == "durable pending" {
				billing.applyErr = errors.New("settlement temporarily unavailable")
			} else if outcome == "pending write failure" {
				billing.pendingErr = errors.New("pending record unavailable")
			}
			pool := newUsageRecordTestPool(t)
			blocked := make(chan struct{})
			var unblock sync.Once
			// Release the blocked worker before the pool cleanup, including on assertion failure.
			t.Cleanup(func() { unblock.Do(func() { close(blocked) }) })
			pool.Submit(func(context.Context) { <-blocked })
			h.usageRecordWorkerPool = pool
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"model":"gpt-image-2-D","prompt":"draw","size":"2048x2048"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set(string(middleware.ContextKeyAPIKey), key)
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: key.User.ID})
			h.Images(c)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.NotNil(t, billing.reserveCmd)
			require.Nil(t, billing.releaseCmd, "handler must not release a hold owned by queued billing")
			require.Nil(t, billing.captureCmd)
			unblock.Do(func() { close(blocked) })
			pool.Stop()
			switch outcome {
			case "success":
				require.NotNil(t, billing.captureCmd)
				require.Nil(t, billing.releaseCmd)
			case "durable pending":
				require.NotNil(t, billing.pending)
				require.Nil(t, billing.releaseCmd)
			case "pending write failure":
				require.NotNil(t, billing.releaseCmd)
				require.Nil(t, billing.pending)
			}
		})
	}
}
