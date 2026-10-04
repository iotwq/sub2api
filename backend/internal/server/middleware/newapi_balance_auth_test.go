package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type newAPIBalanceAuthRepoStub struct {
	subject *service.NewAPIBalanceAccessSubject
}

func (r *newAPIBalanceAuthRepoStub) Rotate(context.Context, int64, string) error { return nil }
func (r *newAPIBalanceAuthRepoStub) FindSubjectByHash(context.Context, string) (*service.NewAPIBalanceAccessSubject, error) {
	if r.subject == nil {
		return nil, service.ErrNewAPIBalanceAccessTokenInvalid
	}
	return r.subject, nil
}

func TestNewAPIBalanceAuthMiddlewareAcceptsBalanceToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &newAPIBalanceAuthRepoStub{subject: &service.NewAPIBalanceAccessSubject{
		UserID: 31, Role: service.RoleUser, Email: "user@example.com", Concurrency: 4,
	}}
	tokens := service.NewNewAPIBalanceAccessTokenService(repo)
	jwtCalled := false
	jwt := JWTAuthMiddleware(func(c *gin.Context) {
		jwtCalled = true
		c.AbortWithStatus(http.StatusUnauthorized)
	})
	router := gin.New()
	router.GET("/api/user/self", NewAPIBalanceAuthMiddleware(jwt, tokens), func(c *gin.Context) {
		subject, ok := GetAuthSubjectFromContext(c)
		require.True(t, ok)
		c.JSON(http.StatusOK, gin.H{"user_id": subject.UserID})
	})
	req := httptest.NewRequest(http.MethodGet, "/api/user/self", nil)
	req.Header.Set("Authorization", "Bearer "+service.NewAPIBalanceAccessTokenPrefix+"test")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.False(t, jwtCalled)
	require.Contains(t, rec.Body.String(), `"user_id":31`)
}

func TestNewAPIBalanceAuthMiddlewareFallsBackToJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtCalled := false
	jwt := JWTAuthMiddleware(func(c *gin.Context) {
		jwtCalled = true
		c.Set(string(ContextKeyUser), AuthSubject{UserID: 32})
		c.Next()
	})
	router := gin.New()
	router.GET("/api/user/self", NewAPIBalanceAuthMiddleware(jwt, nil), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/user/self", nil)
	req.Header.Set("Authorization", "Bearer dashboard.jwt.token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.True(t, jwtCalled)
}
