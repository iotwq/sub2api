package middleware

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func NewAPIBalanceAuthMiddleware(
	jwtAuth JWTAuthMiddleware,
	tokens *service.NewAPIBalanceAccessTokenService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := newAPIBearerToken(c.GetHeader("Authorization"))
		if !ok || !strings.HasPrefix(token, service.NewAPIBalanceAccessTokenPrefix) {
			gin.HandlerFunc(jwtAuth)(c)
			return
		}

		subject, err := tokens.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid access token",
			})
			return
		}
		c.Set(string(ContextKeyUser), AuthSubject{
			UserID:      subject.UserID,
			Concurrency: subject.Concurrency,
		})
		c.Set(string(ContextKeyUserRole), subject.Role)
		c.Set(ContextKeyAuthEmail, subject.Email)
		c.Next()
	}
}

func newAPIBearerToken(header string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	return token, token != ""
}
