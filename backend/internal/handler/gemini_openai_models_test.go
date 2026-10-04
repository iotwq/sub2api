package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type openAIGeminiModelsAccountRepo struct {
	service.AccountRepository
	account service.Account
}

func (r *openAIGeminiModelsAccountRepo) ListSchedulableByGroupIDAndPlatform(_ context.Context, _ int64, platform string) ([]service.Account, error) {
	if platform == service.PlatformOpenAI {
		return []service.Account{r.account}, nil
	}
	return nil, nil
}

func (r *openAIGeminiModelsAccountRepo) GetByID(context.Context, int64) (*service.Account, error) {
	return &r.account, nil
}

func TestGeminiModelsOpenAIGroupKeepsNativeCapabilityBoundary(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		name := "capability disabled"
		if enabled {
			name = "capability enabled"
		}
		t.Run(name, func(t *testing.T) {
			groupID := int64(47)
			capabilities := []any{"chat_completions"}
			if enabled {
				capabilities = append(capabilities, "gemini_native")
			}
			repo := &openAIGeminiModelsAccountRepo{account: service.Account{
				ID: 301, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
				Status: service.StatusActive, Schedulable: true,
				Credentials: map[string]any{"api_key": "test-key", "base_url": "https://upstream.example/v1", "openai_capabilities": capabilities},
			}}
			cfg := &config.Config{}
			upstream := &geminiMixedModelsUpstream{status: http.StatusOK, body: `{"models":[{"name":"models/gemini-native"}]}`}
			h := &GatewayHandler{
				openAIGatewayService: service.NewOpenAIGatewayService(repo, nil, nil, nil, nil, nil, nil, cfg, nil, nil, nil,
					nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
				// OpenAI groups must not query an Antigravity repository for their model list.
				geminiCompatService: service.NewGeminiMessagesCompatService(nil, nil, nil, nil, nil, nil, upstream, nil, cfg),
			}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1beta/models", nil)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{GroupID: &groupID, Group: &service.Group{ID: groupID, Platform: service.PlatformOpenAI}})
			h.GeminiV1BetaListModels(c)
			if enabled {
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				require.JSONEq(t, upstream.body, rec.Body.String())
			} else {
				require.Equal(t, http.StatusServiceUnavailable, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), "No available OpenAI accounts with Gemini native capability")
			}
		})
	}
}
