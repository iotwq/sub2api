package routes

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/middleware"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterCommunityChatRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth servermiddleware.JWTAuthMiddleware,
	settingService *service.SettingService,
	redisClient *redis.Client,
) {
	rateLimiter := middleware.NewRateLimiter(redisClient)
	chat := v1.Group("/community-chat")
	{
		chat.GET("/uploads/:filename", h.CommunityChat.ServeUpload)
	}

	authenticated := chat.Group("")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(servermiddleware.BackendModeUserGuard(settingService))
	{
		authenticated.GET("/messages", h.CommunityChat.ListMessages)
		authenticated.POST("/attachments/session", h.CommunityChat.EstablishAttachmentSession)
		authenticated.POST("/messages", rateLimiter.Limit("community-chat-message", 30, time.Minute), h.CommunityChat.CreateTextMessage)
		authenticated.POST("/files", rateLimiter.Limit("community-chat-upload", 6, time.Minute), h.CommunityChat.UploadFile)
		authenticated.POST("/images", rateLimiter.Limit("community-chat-upload", 6, time.Minute), h.CommunityChat.UploadImage)
		authenticated.DELETE("/messages/:id", rateLimiter.Limit("community-chat-delete", 30, time.Minute), h.CommunityChat.DeleteMessage)
		authenticated.GET("/direct/conversations", h.CommunityChat.ListDirectConversations)
		authenticated.GET("/direct/users", h.CommunityChat.SearchDirectUsers)
		authenticated.GET("/direct/messages", h.CommunityChat.ListDirectMessages)
		authenticated.GET("/direct/unread", h.CommunityChat.DirectUnread)
		authenticated.POST("/direct/read", h.CommunityChat.MarkDirectRead)
		authenticated.POST("/direct/messages", rateLimiter.Limit("community-chat-direct-message", 30, time.Minute), h.CommunityChat.CreateDirectTextMessage)
		authenticated.POST("/direct/files", rateLimiter.Limit("community-chat-direct-upload", 6, time.Minute), h.CommunityChat.UploadDirectFile)
		authenticated.DELETE("/direct/messages/:id", rateLimiter.Limit("community-chat-direct-delete", 30, time.Minute), h.CommunityChat.DeleteDirectMessage)
	}

	chat.GET("/ws", h.CommunityChat.WebSocket)
}
