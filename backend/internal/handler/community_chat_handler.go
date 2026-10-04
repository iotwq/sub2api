package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	communityChatUploadDirMode  = 0755
	communityChatFileMode       = 0644
	communityChatUploadOverhead = 1 << 20
	communityChatWSProtocol     = "sub2api-chat"
	communityChatWSJWTPrefix    = "jwt."

	communityChatAttachmentCookieName   = "community_chat_attachment_session"
	communityChatAttachmentCookiePath   = "/api/v1/community-chat/uploads"
	communityChatAttachmentCookieMaxAge = 5 * 60
)

var communityChatAllowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var communityChatAllowedFileExtensions = map[string]string{
	".csv":      "text/csv",
	".doc":      "application/msword",
	".docx":     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".md":       "text/markdown",
	".markdown": "text/markdown",
	".mp4":      "video/mp4",
	".pdf":      "application/pdf",
	".ppt":      "application/vnd.ms-powerpoint",
	".pptx":     "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".rtf":      "application/rtf",
	".txt":      "text/plain",
	".xls":      "application/vnd.ms-excel",
	".xlsx":     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

type CommunityChatHandler struct {
	chatService *service.CommunityChatService
	authService *service.AuthService
	userService *service.UserService
	settingSvc  *service.SettingService
	uploadsDir  string
	upgrader    websocket.Upgrader
}

func NewCommunityChatHandler(chatService *service.CommunityChatService, authService *service.AuthService, userService *service.UserService, settingSvc *service.SettingService, cfg *config.Config) *CommunityChatHandler {
	dataDir := ""
	var allowedOrigins []string
	if cfg != nil {
		dataDir = cfg.Pricing.DataDir
		allowedOrigins = append([]string(nil), cfg.CORS.AllowedOrigins...)
	}
	uploadsDir := filepath.Join(dataDir, "chat", "uploads")
	_ = os.MkdirAll(uploadsDir, communityChatUploadDirMode)
	return &CommunityChatHandler{
		chatService: chatService,
		authService: authService,
		userService: userService,
		settingSvc:  settingSvc,
		uploadsDir:  uploadsDir,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return isAllowedCommunityChatOrigin(r, allowedOrigins)
			},
			Subprotocols: []string{communityChatWSProtocol},
		},
	}
}

type communityChatCreateMessageRequest struct {
	Content          string `json:"content"`
	ReplyToMessageID *int64 `json:"reply_to_message_id"`
}

type communityChatCreateDirectMessageRequest struct {
	Content string `json:"content"`
	UserID  int64  `json:"user_id"`
}

type communityChatMarkDirectReadRequest struct {
	UserID        int64 `json:"user_id"`
	LastMessageID int64 `json:"last_message_id" binding:"required,gt=0"`
}

func (h *CommunityChatHandler) ListMessages(c *gin.Context) {
	beforeID, afterID, limit, cursorRequested, ok := parseCommunityChatCursor(c)
	if !ok {
		return
	}
	if cursorRequested {
		if strings.TrimSpace(c.Query("search")) != "" {
			response.BadRequest(c, "Search cannot be combined with message cursors")
			return
		}
		var (
			messages []service.CommunityChatMessage
			err      error
		)
		if beforeID > 0 {
			messages, err = h.chatService.ListMessagesBefore(c.Request.Context(), beforeID, limit)
		} else {
			messages, err = h.chatService.ListMessagesAfter(c.Request.Context(), afterID, limit)
		}
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, messages)
		return
	}
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	search := strings.TrimSpace(c.Query("search"))
	var (
		messages []service.CommunityChatMessage
		result   *pagination.PaginationResult
		err      error
	)
	if search != "" {
		messages, result, err = h.chatService.SearchMessages(c.Request.Context(), search, params)
	} else {
		messages, result, err = h.chatService.ListMessages(c.Request.Context(), params)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, messages, result.Total, result.Page, result.PageSize)
}

func (h *CommunityChatHandler) ListDirectConversations(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	conversations, result, err := h.chatService.ListDirectConversations(c.Request.Context(), user, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, conversations, result.Total, result.Page, result.PageSize)
}

func (h *CommunityChatHandler) SearchDirectUsers(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	users, result, err := h.chatService.SearchDirectUsers(c.Request.Context(), user, c.Query("search"), pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, users, result.Total, result.Page, result.PageSize)
}

func (h *CommunityChatHandler) ListDirectMessages(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	conversationUserID, err := parseOptionalInt64Query(c, "user_id")
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	beforeID, afterID, limit, cursorRequested, ok := parseCommunityChatCursor(c)
	if !ok {
		return
	}
	if cursorRequested {
		var messages []service.CommunityChatMessage
		if beforeID > 0 {
			messages, err = h.chatService.ListDirectMessagesBefore(c.Request.Context(), user, conversationUserID, beforeID, limit)
		} else {
			messages, err = h.chatService.ListDirectMessagesAfter(c.Request.Context(), user, conversationUserID, afterID, limit)
		}
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, messages)
		return
	}
	messages, result, err := h.chatService.ListDirectMessages(c.Request.Context(), user, conversationUserID, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, messages, result.Total, result.Page, result.PageSize)
}

func (h *CommunityChatHandler) DirectUnread(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	summary, err := h.chatService.DirectUnread(c.Request.Context(), user)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, summary)
}

func (h *CommunityChatHandler) MarkDirectRead(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	var req communityChatMarkDirectReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	lastMessageID, err := h.chatService.MarkDirectRead(c.Request.Context(), user, req.UserID, req.LastMessageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"last_message_id": lastMessageID})
}

func (h *CommunityChatHandler) CreateTextMessage(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}

	var req communityChatCreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	msg, err := h.chatService.CreateMessage(c.Request.Context(), service.CommunityChatCreateMessageInput{
		User:             user,
		MessageType:      service.CommunityChatMessageTypeText,
		Content:          req.Content,
		ReplyToMessageID: req.ReplyToMessageID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, msg)
}

func (h *CommunityChatHandler) CreateDirectTextMessage(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}

	var req communityChatCreateDirectMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	msg, err := h.chatService.CreateDirectMessage(c.Request.Context(), service.CommunityChatCreateDirectMessageInput{
		Actor:              user,
		ConversationUserID: req.UserID,
		Content:            req.Content,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, msg)
}

func (h *CommunityChatHandler) UploadImage(c *gin.Context) {
	h.UploadFile(c)
}

func (h *CommunityChatHandler) UploadFile(c *gin.Context) {
	h.uploadFile(c, false)
}

func (h *CommunityChatHandler) UploadDirectFile(c *gin.Context) {
	h.uploadFile(c, true)
}

func (h *CommunityChatHandler) uploadFile(c *gin.Context, direct bool) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.CommunityChatMaxAttachmentBytes+communityChatUploadOverhead)
	if err := c.Request.ParseMultipartForm(service.CommunityChatMaxAttachmentBytes); err != nil {
		response.BadRequest(c, "Invalid file upload")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		file, header, err = c.Request.FormFile("image")
		if err != nil {
			response.BadRequest(c, "File is required")
			return
		}
	}
	defer file.Close()

	content := strings.TrimSpace(c.PostForm("content"))
	originalFileName := normalizeCommunityChatDownloadName(header.Filename)
	if content == "" {
		content = originalFileName
	}
	if content == "" {
		content = "附件"
	}
	var conversationUserID int64
	if direct {
		rawUserID := strings.TrimSpace(c.PostForm("user_id"))
		if rawUserID != "" {
			conversationUserID, err = strconv.ParseInt(rawUserID, 10, 64)
			if err != nil || conversationUserID <= 0 {
				response.BadRequest(c, "Invalid user ID")
				return
			}
		}
	}
	var replyToMessageID *int64
	if rawReplyTo := strings.TrimSpace(c.PostForm("reply_to_message_id")); !direct && rawReplyTo != "" {
		parsed, err := strconv.ParseInt(rawReplyTo, 10, 64)
		if err != nil || parsed <= 0 {
			response.BadRequest(c, "Invalid reply message ID")
			return
		}
		replyToMessageID = &parsed
	}

	limited := io.LimitReader(file, service.CommunityChatMaxAttachmentBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		response.InternalError(c, "Failed to read file")
		return
	}
	if int64(len(data)) > service.CommunityChatMaxAttachmentBytes {
		response.ErrorFrom(c, service.ErrCommunityChatAttachmentTooLarge)
		return
	}

	mimeType := http.DetectContentType(data)
	messageType := service.CommunityChatMessageTypeFile
	ext, ok := communityChatAllowedImageTypes[mimeType]
	if ok {
		messageType = service.CommunityChatMessageTypeImage
	} else {
		ext = strings.ToLower(filepath.Ext(header.Filename))
		defaultMIME, ok := communityChatAllowedFileExtensions[ext]
		if !ok {
			response.ErrorFrom(c, service.ErrCommunityChatInvalidAttachment)
			return
		}
		mimeType = normalizeCommunityChatFileMIME(mimeType, defaultMIME)
	}
	if strings.TrimSpace(ext) == "" {
		response.ErrorFrom(c, service.ErrCommunityChatInvalidAttachment)
		return
	}

	fileName, err := h.newUploadFileName(ext)
	if err != nil {
		response.InternalError(c, "Failed to prepare file")
		return
	}
	filePath := filepath.Join(h.uploadsDir, fileName)
	if err := os.WriteFile(filePath, data, communityChatFileMode); err != nil {
		response.InternalError(c, "Failed to save file")
		return
	}

	fileURL := "/api/v1/community-chat/uploads/" + fileName
	if originalFileName != "" {
		fileURL += "?name=" + url.QueryEscape(originalFileName)
	}
	var msg *service.CommunityChatMessage
	if direct {
		msg, err = h.chatService.CreateDirectMessage(c.Request.Context(), service.CommunityChatCreateDirectMessageInput{
			Actor:              user,
			ConversationUserID: conversationUserID,
			MessageType:        messageType,
			Content:            content,
			ImageURL:           fileURL,
			ImageMIMEType:      mimeType,
			ImageSizeBytes:     int64(len(data)),
		})
	} else {
		msg, err = h.chatService.CreateMessage(c.Request.Context(), service.CommunityChatCreateMessageInput{
			User:             user,
			MessageType:      messageType,
			Content:          content,
			ImageURL:         fileURL,
			ImageMIMEType:    mimeType,
			ImageSizeBytes:   int64(len(data)),
			ReplyToMessageID: replyToMessageID,
		})
	}
	if err != nil {
		_ = os.Remove(filePath)
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, msg)
}

func (h *CommunityChatHandler) ServeUpload(c *gin.Context) {
	user, ok := h.authenticateAttachmentRequest(c)
	if !ok {
		return
	}

	fileName := strings.TrimSpace(c.Param("filename"))
	if fileName == "" {
		c.Status(http.StatusNotFound)
		return
	}
	fileName = strings.TrimPrefix(fileName, "/")
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") || strings.Contains(fileName, "..") {
		c.Status(http.StatusNotFound)
		return
	}

	filePath := filepath.Join(h.uploadsDir, fileName)
	cleanUploadsDir := filepath.Clean(h.uploadsDir)
	cleanFilePath := filepath.Clean(filePath)
	if !isPathInside(cleanFilePath, cleanUploadsDir) {
		c.Status(http.StatusNotFound)
		return
	}
	active, err := h.chatService.CanAccessAttachment(c.Request.Context(), fileName, user.ID, user.IsAdmin())
	if err != nil {
		response.InternalError(c, "Failed to load attachment")
		return
	}
	if !active {
		c.Status(http.StatusNotFound)
		return
	}
	if info, err := os.Stat(cleanFilePath); err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	if c.Query("download") == "1" {
		downloadName := normalizeCommunityChatDownloadName(c.Query("name"))
		if downloadName == "" {
			downloadName = fileName
		}
		c.FileAttachment(cleanFilePath, downloadName)
		return
	}
	c.File(cleanFilePath)
}

func (h *CommunityChatHandler) EstablishAttachmentSession(c *gin.Context) {
	if _, ok := middleware2.GetAuthSubjectFromContext(c); !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	token := extractBearerToken(c.Request)
	if token == "" {
		response.Unauthorized(c, "Authorization header is required")
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     communityChatAttachmentCookieName,
		Value:    encodeCookieValue(token),
		Path:     communityChatAttachmentCookiePath,
		MaxAge:   communityChatAttachmentCookieMaxAge,
		Expires:  time.Now().Add(time.Duration(communityChatAttachmentCookieMaxAge) * time.Second),
		HttpOnly: true,
		Secure:   isRequestHTTPS(c),
		SameSite: http.SameSiteLaxMode,
	})
	response.Success(c, gin.H{"expires_in": communityChatAttachmentCookieMaxAge})
}

func (h *CommunityChatHandler) authenticateAttachmentRequest(c *gin.Context) (*service.User, bool) {
	token := extractBearerToken(c.Request)
	if token == "" {
		cookie, err := c.Request.Cookie(communityChatAttachmentCookieName)
		if err == nil {
			token, _ = decodeCookieValue(cookie.Value)
		}
	}
	if token == "" || h.authService == nil || h.userService == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil, false
	}

	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil, false
	}
	user, err := h.userService.GetProfile(c.Request.Context(), claims.UserID)
	if err != nil || user == nil || !user.IsActive() || user.TokenVersion != claims.TokenVersion {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil, false
	}
	if h.settingSvc != nil && h.settingSvc.IsBackendModeEnabled(c.Request.Context()) && !user.IsAdmin() {
		c.AbortWithStatus(http.StatusForbidden)
		return nil, false
	}
	return user, true
}

func extractBearerToken(r *http.Request) string {
	if r == nil {
		return ""
	}
	parts := strings.SplitN(strings.TrimSpace(r.Header.Get("Authorization")), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func normalizeCommunityChatFileMIME(detected, fallback string) string {
	detected = strings.TrimSpace(strings.ToLower(detected))
	if detected == "" || detected == "application/octet-stream" || detected == "application/zip" {
		return fallback
	}
	return detected
}

func (h *CommunityChatHandler) DeleteMessage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid message ID")
		return
	}

	msg, err := h.chatService.DeleteMessage(c.Request.Context(), id, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if fileName := communityChatLocalAttachmentFilename(msg.ImageURL); fileName != "" {
		filePath := filepath.Join(h.uploadsDir, fileName)
		if isPathInside(filepath.Clean(filePath), filepath.Clean(h.uploadsDir)) {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				slog.Error("remove deleted community chat attachment", "message_id", id, "filename", fileName, "error", err)
			}
		}
	}
	response.Success(c, msg)
}

func (h *CommunityChatHandler) DeleteDirectMessage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid message ID")
		return
	}
	msg, err := h.chatService.DeleteDirectMessage(c.Request.Context(), id, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if fileName := communityChatLocalAttachmentFilename(msg.ImageURL); fileName != "" {
		filePath := filepath.Join(h.uploadsDir, fileName)
		if isPathInside(filepath.Clean(filePath), filepath.Clean(h.uploadsDir)) {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				slog.Error("remove deleted community direct chat attachment", "message_id", id, "filename", fileName, "error", err)
			}
		}
	}
	response.Success(c, msg)
}

func communityChatLocalAttachmentFilename(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return ""
	}
	const prefix = "/api/v1/community-chat/uploads/"
	if !strings.HasPrefix(parsed.Path, prefix) {
		return ""
	}
	fileName := strings.TrimPrefix(parsed.Path, prefix)
	if fileName == "" || strings.ContainsAny(fileName, `/\\`) || strings.Contains(fileName, "..") {
		return ""
	}
	return fileName
}

func (h *CommunityChatHandler) WebSocket(c *gin.Context) {
	token := extractCommunityChatWebSocketJWT(c.Request)
	if token == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user, err := h.userService.GetProfile(c.Request.Context(), claims.UserID)
	if err != nil || user == nil || !user.IsActive() || user.TokenVersion != claims.TokenVersion {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if h.settingSvc != nil && h.settingSvc.IsBackendModeEnabled(c.Request.Context()) && !user.IsAdmin() {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	events, unsubscribe := h.chatService.Subscribe()
	defer unsubscribe()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-done:
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if !communityChatEventVisibleToUser(event, user) {
				continue
			}
			if err := conn.WriteJSON(event); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}

func extractCommunityChatWebSocketJWT(r *http.Request) string {
	if r == nil {
		return ""
	}
	for _, part := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		protocol := strings.TrimSpace(part)
		if strings.HasPrefix(protocol, communityChatWSJWTPrefix) {
			return strings.TrimSpace(strings.TrimPrefix(protocol, communityChatWSJWTPrefix))
		}
	}
	return ""
}

func isAllowedCommunityChatOrigin(r *http.Request, allowedOrigins []string) bool {
	if r == nil {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	if strings.EqualFold(parsed.Host, strings.TrimSpace(r.Host)) {
		return true
	}
	for _, allowedOrigin := range allowedOrigins {
		allowedOrigin = strings.TrimSpace(allowedOrigin)
		if allowedOrigin == "*" || strings.EqualFold(allowedOrigin, origin) {
			return true
		}
	}
	return false
}

func (h *CommunityChatHandler) currentUser(c *gin.Context) (*service.User, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return nil, false
	}
	user, err := h.userService.GetProfile(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return nil, false
	}
	return user, true
}

func (h *CommunityChatHandler) newUploadFileName(ext string) (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), hex.EncodeToString(buf[:]), ext), nil
}

func parseOptionalInt64Query(c *gin.Context, key string) (int64, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

func parseCommunityChatCursor(c *gin.Context) (beforeID, afterID int64, limit int, requested, ok bool) {
	var err error
	beforeID, err = parseOptionalInt64Query(c, "before_id")
	if err != nil {
		response.BadRequest(c, "Invalid before message ID")
		return 0, 0, 0, false, false
	}
	afterID, err = parseOptionalInt64Query(c, "after_id")
	if err != nil {
		response.BadRequest(c, "Invalid after message ID")
		return 0, 0, 0, false, false
	}
	if beforeID > 0 && afterID > 0 {
		response.BadRequest(c, "before_id and after_id cannot be used together")
		return 0, 0, 0, false, false
	}
	requested = beforeID > 0 || afterID > 0
	limit = 80
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			response.BadRequest(c, "Invalid limit")
			return 0, 0, 0, false, false
		}
		limit = parsed
	}
	return beforeID, afterID, limit, requested, true
}

func communityChatEventVisibleToUser(event service.CommunityChatEvent, user *service.User) bool {
	if event.Type != service.CommunityChatEventDirectMessageCreated && event.Type != service.CommunityChatEventDirectMessageDeleted {
		return true
	}
	if user == nil {
		return false
	}
	if user.IsAdmin() {
		return true
	}
	if event.ConversationUserID == user.ID {
		return true
	}
	if event.Message != nil && event.Message.UserID == user.ID {
		return true
	}
	return false
}

func isPathInside(path, base string) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func normalizeCommunityChatDownloadName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimSpace(filepath.Base(name))
	if name == "." || name == "/" {
		return ""
	}
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || r == '/' || r == '\\' {
			return -1
		}
		return r
	}, name)
	return strings.TrimSpace(name)
}
