//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type newAPIAccessTokenHandlerRepoStub struct {
	rotatedUserID int64
	rotatedHash   string
}

func (r *newAPIAccessTokenHandlerRepoStub) Rotate(_ context.Context, userID int64, tokenHash string) error {
	r.rotatedUserID = userID
	r.rotatedHash = tokenHash
	return nil
}

func (r *newAPIAccessTokenHandlerRepoStub) FindSubjectByHash(context.Context, string) (*service.NewAPIBalanceAccessSubject, error) {
	return nil, service.ErrNewAPIBalanceAccessTokenInvalid
}

func TestAuthHandlerGetCurrentUserReturnsProfileCompatibilityFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	verifiedAt := time.Date(2026, 4, 20, 8, 30, 0, 0, time.UTC)
	repo := &userHandlerRepoStub{
		user: &service.User{
			ID:           31,
			Email:        "me@example.com",
			Username:     "linuxdo-handle",
			Role:         service.RoleUser,
			Status:       service.StatusActive,
			AvatarURL:    "https://cdn.example.com/linuxdo.png",
			AvatarSource: "remote_url",
		},
		identities: []service.UserAuthIdentityRecord{
			{
				ProviderType:    "linuxdo",
				ProviderKey:     "linuxdo",
				ProviderSubject: "linuxdo-subject-31",
				VerifiedAt:      &verifiedAt,
				Metadata: map[string]any{
					"username":   "linuxdo-handle",
					"avatar_url": "https://cdn.example.com/linuxdo.png",
				},
			},
		},
	}

	handler := &AuthHandler{
		userService: service.NewUserService(repo, nil, nil, nil),
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 31})

	handler.GetCurrentUser(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, true, resp.Data["email_bound"])
	require.Equal(t, true, resp.Data["linuxdo_bound"])
	require.Equal(t, "https://cdn.example.com/linuxdo.png", resp.Data["avatar_url"])

	authBindings, ok := resp.Data["auth_bindings"].(map[string]any)
	require.True(t, ok)
	linuxdoBinding, ok := authBindings["linuxdo"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, linuxdoBinding["bound"])

	avatarSource, ok := resp.Data["avatar_source"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "linuxdo", avatarSource["provider"])
	require.Equal(t, "linuxdo", avatarSource["source"])

	profileSources, ok := resp.Data["profile_sources"].(map[string]any)
	require.True(t, ok)
	usernameSource, ok := profileSources["username"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "linuxdo", usernameSource["provider"])
	require.Equal(t, "linuxdo", usernameSource["source"])
}

func TestAuthHandlerGetNewAPISelfReturnsQuotaInNewAPIUnits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &userHandlerRepoStub{user: &service.User{
		ID:             31,
		Username:       "newapi-user",
		Email:          "newapi@example.com",
		Status:         service.StatusActive,
		Balance:        12.345678,
		TotalRecharged: 20,
	}}
	handler := &AuthHandler{userService: service.NewUserService(repo, nil, nil, nil)}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/user/self", nil)
	c.Request.Header.Set("New-Api-User", "31")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 31})

	handler.GetNewAPISelf(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var resp struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.True(t, resp.Success)
	require.Equal(t, float64(6172839), resp.Data["quota"])
	require.Equal(t, float64(3827161), resp.Data["used_quota"])
	require.Equal(t, float64(31), resp.Data["id"])
}

func TestAuthHandlerGetNewAPISelfRejectsMismatchedUserHint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &AuthHandler{}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/user/self", nil)
	c.Request.Header.Set("New-Api-User", "32")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 31})

	handler.GetNewAPISelf(c)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "New-Api-User does not match")
}

func TestAuthHandlerGenerateNewAPIAccessTokenReturnsTokenOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &newAPIAccessTokenHandlerRepoStub{}
	handler := &AuthHandler{newAPIAccessTokens: service.NewNewAPIBalanceAccessTokenService(repo)}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/user/token", nil)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 31})

	handler.GenerateNewAPIAccessToken(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	var resp struct {
		Success bool   `json:"success"`
		Data    string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.True(t, resp.Success)
	require.Contains(t, resp.Data, service.NewAPIBalanceAccessTokenPrefix)
	require.Equal(t, int64(31), repo.rotatedUserID)
	require.Len(t, repo.rotatedHash, 64)
	require.NotContains(t, repo.rotatedHash, resp.Data)
}
