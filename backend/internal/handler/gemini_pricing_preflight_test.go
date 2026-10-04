//go:build unit

package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiNativeMissingPricingStopsBeforeUpstream(t *testing.T) {
	for _, action := range []string{"generateContent", "streamGenerateContent"} {
		t.Run(action, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			group := &service.Group{ID: 2001, Hydrated: true, Platform: service.PlatformGemini, Status: service.StatusActive}
			account := &service.Account{ID: 1001, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey,
				Status: service.StatusActive, Schedulable: true, Concurrency: 1,
				Credentials:   map[string]any{"api_key": "unused-test-key", "model_mapping": map[string]any{"unpriced-native": "unpriced-native"}},
				AccountGroups: []service.AccountGroup{{AccountID: 1001, GroupID: group.ID}}}
			h, cleanup := newTestGatewayHandler(t, group, []*service.Account{account})
			defer cleanup()
			// No upstream client is installed: forwarding would panic if the guard regresses.
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/unpriced-native:"+action,
				bytes.NewBufferString(`{"contents":[{"parts":[{"text":"hi"}]}]}`))
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.Group, group))
			c.Params = gin.Params{{Key: "modelAction", Value: "unpriced-native:" + action}}
			key := &service.APIKey{ID: 3001, UserID: 4001, GroupID: &group.ID, Group: group, User: &service.User{ID: 4001, Balance: 100}}
			c.Set(string(middleware.ContextKeyAPIKey), key)
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: key.UserID, Concurrency: 10})
			h.GeminiV1BetaModels(c)
			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Contains(t, rec.Body.String(), "configure pricing")
			require.Equal(t, account.ID, c.GetInt64(opsAccountIDKey))
		})
	}
}
