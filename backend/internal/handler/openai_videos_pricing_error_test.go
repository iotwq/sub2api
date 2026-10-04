package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestVideosRejectsUnpricedMiniMaxH3BeforeUpstream(t *testing.T) {
	handler, billingRepo, _, upstream, apiKey := newNanoBananaBillingTestHandler(t, 100)
	body := `{"model":"MiniMax-H3","content":[{"type":"text","text":"cinematic waves"}],"resolution":"2K","duration":5,"ratio":"16:9"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: apiKey.User.ID})

	handler.Videos(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "invalid_request_error", gjson.GetBytes(rec.Body.Bytes(), "error.type").String())
	message := gjson.GetBytes(rec.Body.Bytes(), "error.message").String()
	require.Contains(t, message, "MiniMax-H3")
	require.Contains(t, message, "not priced")
	require.Contains(t, message, "cannot be used")
	require.Zero(t, upstream.calls)
	require.Nil(t, billingRepo.reserveCmd)
	require.Nil(t, billingRepo.applyCmd)
}
