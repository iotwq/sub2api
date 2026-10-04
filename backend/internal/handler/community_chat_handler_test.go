package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type communityChatHandlerTestUserRepo struct {
	service.UserRepository
	user *service.User
}

func (r *communityChatHandlerTestUserRepo) GetByID(_ context.Context, id int64) (*service.User, error) {
	if r.user == nil || r.user.ID != id {
		return nil, errors.New("user not found")
	}
	return r.user, nil
}

func (r *communityChatHandlerTestUserRepo) GetUserAvatar(_ context.Context, _ int64) (*service.UserAvatar, error) {
	return nil, nil
}

type communityChatHandlerTestRepo struct {
	service.CommunityChatRepository
	message          *service.CommunityChatMessage
	directMessage    *service.CommunityChatMessage
	createdDirect    *service.CommunityChatMessage
	attachmentActive bool
}

func (r *communityChatHandlerTestRepo) CreateDirectMessage(_ context.Context, message *service.CommunityChatMessage) error {
	copy := *message
	copy.ID = 18
	copy.CreatedAt = time.Now()
	copy.UpdatedAt = copy.CreatedAt
	*message = copy
	r.createdDirect = &copy
	return nil
}

func (r *communityChatHandlerTestRepo) IsAttachmentActive(_ context.Context, _ string) (bool, error) {
	return r.attachmentActive, nil
}

func (r *communityChatHandlerTestRepo) CanAccessAttachment(_ context.Context, _ string, _ int64, _ bool) (bool, error) {
	return r.attachmentActive, nil
}

func (r *communityChatHandlerTestRepo) GetMessage(_ context.Context, id int64) (*service.CommunityChatMessage, error) {
	if r.message == nil || r.message.ID != id {
		return nil, service.ErrCommunityChatMessageNotFound
	}
	copy := *r.message
	return &copy, nil
}

func (r *communityChatHandlerTestRepo) DeleteMessage(_ context.Context, id, actorID int64) (*service.CommunityChatMessage, error) {
	msg, err := r.GetMessage(context.Background(), id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	msg.DeletedAt = &now
	msg.DeletedBy = &actorID
	return msg, nil
}

func (r *communityChatHandlerTestRepo) GetDirectMessage(_ context.Context, id int64) (*service.CommunityChatMessage, error) {
	if r.directMessage == nil || r.directMessage.ID != id {
		return nil, service.ErrCommunityChatMessageNotFound
	}
	copy := *r.directMessage
	return &copy, nil
}

func (r *communityChatHandlerTestRepo) DeleteDirectMessage(_ context.Context, id, actorID int64) (*service.CommunityChatMessage, error) {
	msg, err := r.GetDirectMessage(context.Background(), id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	msg.DeletedAt = &now
	msg.DeletedBy = &actorID
	return msg, nil
}

func newCommunityChatAttachmentTestHandler(t *testing.T, active bool) (*CommunityChatHandler, *service.User, string) {
	t.Helper()
	user := &service.User{
		ID:           42,
		Email:        "chat@example.com",
		Role:         "user",
		Status:       service.StatusActive,
		TokenVersion: 1,
	}
	userRepo := &communityChatHandlerTestUserRepo{user: user}
	cfg := &config.Config{}
	cfg.JWT.Secret = "community-chat-test-secret-32bytes"
	cfg.JWT.AccessTokenExpireMinutes = 60
	authService := service.NewAuthService(nil, userRepo, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil)
	token, err := authService.GenerateToken(context.Background(), user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	chatRepo := &communityChatHandlerTestRepo{attachmentActive: active}
	return &CommunityChatHandler{
		chatService: service.NewCommunityChatService(chatRepo, userRepo),
		authService: authService,
		userService: service.NewUserService(userRepo, nil, nil, nil),
		uploadsDir:  t.TempDir(),
	}, user, token
}

func TestNormalizeCommunityChatDownloadName(t *testing.T) {
	cases := map[string]string{
		"report.pdf":            "report.pdf",
		"../report.pdf":         "report.pdf",
		`C:\Users\me\video.mp4`: "video.mp4",
		"bad\nname.txt":         "badname.txt",
		"  demo markdown.md  ":  "demo markdown.md",
	}

	for input, expected := range cases {
		if got := normalizeCommunityChatDownloadName(input); got != expected {
			t.Fatalf("normalizeCommunityChatDownloadName(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestExtractCommunityChatWebSocketJWTUsesSubprotocolOnly(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/community-chat/ws?token=query-token", nil)
	req.Header.Set("Sec-WebSocket-Protocol", "sub2api-chat, jwt.header.payload.signature")

	if got := extractCommunityChatWebSocketJWT(req); got != "header.payload.signature" {
		t.Fatalf("extractCommunityChatWebSocketJWT() = %q", got)
	}

	req.Header.Del("Sec-WebSocket-Protocol")
	if got := extractCommunityChatWebSocketJWT(req); got != "" {
		t.Fatalf("query token must be ignored, got %q", got)
	}
}

func TestCommunityChatDirectDeleteEventIsVisibleOnlyToConversation(t *testing.T) {
	event := service.CommunityChatEvent{
		Type:               service.CommunityChatEventDirectMessageDeleted,
		ID:                 7,
		ConversationUserID: 42,
	}
	if !communityChatEventVisibleToUser(event, &service.User{ID: 42, Role: service.RoleUser}) {
		t.Fatal("conversation user should receive direct deletion event")
	}
	if communityChatEventVisibleToUser(event, &service.User{ID: 84, Role: service.RoleUser}) {
		t.Fatal("unrelated user received direct deletion event")
	}
	if !communityChatEventVisibleToUser(event, &service.User{ID: 1, Role: service.RoleAdmin}) {
		t.Fatal("admin should receive direct deletion event")
	}
}

func TestParseCommunityChatCursorRejectsConflictingDirections(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community-chat/messages?before_id=10&after_id=20", nil)

	_, _, _, _, ok := parseCommunityChatCursor(c)
	if ok || recorder.Code != http.StatusBadRequest {
		t.Fatalf("ok/status = %v/%d, want false/%d", ok, recorder.Code, http.StatusBadRequest)
	}
}

func TestIsAllowedCommunityChatOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    bool
	}{
		{name: "same host", origin: "https://chat.example.com", want: true},
		{name: "configured cross origin", origin: "https://app.example.com", allowed: []string{"https://app.example.com"}, want: true},
		{name: "unconfigured cross origin", origin: "https://evil.example.com", want: false},
		{name: "non browser client", origin: "", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "https://chat.example.com/api/v1/community-chat/ws", nil)
			req.Host = "chat.example.com"
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if got := isAllowedCommunityChatOrigin(req, tt.allowed); got != tt.want {
				t.Fatalf("isAllowedCommunityChatOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstablishCommunityChatAttachmentSessionSetsRestrictedCookie(t *testing.T) {
	handler, user, token := newCommunityChatAttachmentTestHandler(t, true)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/community-chat/attachments/session", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.EstablishAttachmentSession(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != communityChatAttachmentCookieName || cookie.Path != communityChatAttachmentCookiePath {
		t.Fatalf("unexpected cookie scope: %#v", cookie)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.MaxAge != communityChatAttachmentCookieMaxAge {
		t.Fatalf("unexpected cookie security attributes: %#v", cookie)
	}
	decoded, err := decodeCookieValue(cookie.Value)
	if err != nil || decoded != token {
		t.Fatalf("cookie token mismatch: decoded=%q err=%v", decoded, err)
	}
}

func TestCommunityChatDirectReadRequiresVisibleMessageID(t *testing.T) {
	handler, user, _ := newCommunityChatAttachmentTestHandler(t, true)
	for _, body := range []string{`{}`, `{"last_message_id":0}`, `{"last_message_id":-1}`} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/community-chat/direct/read", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})
		handler.MarkDirectRead(c)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("body %s: status = %d", body, recorder.Code)
		}
	}
}

func TestServeCommunityChatUploadRejectsUnauthenticatedRequest(t *testing.T) {
	handler, _, _ := newCommunityChatAttachmentTestHandler(t, true)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community-chat/uploads/demo.txt", nil)
	c.Params = gin.Params{{Key: "filename", Value: "demo.txt"}}

	handler.ServeUpload(c)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestServeCommunityChatUploadAllowsActiveAttachmentSession(t *testing.T) {
	handler, _, token := newCommunityChatAttachmentTestHandler(t, true)
	fileName := "active.txt"
	if err := os.WriteFile(filepath.Join(handler.uploadsDir, fileName), []byte("active attachment"), communityChatFileMode); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community-chat/uploads/"+fileName, nil)
	c.Request.AddCookie(&http.Cookie{Name: communityChatAttachmentCookieName, Value: encodeCookieValue(token)})
	c.Params = gin.Params{{Key: "filename", Value: fileName}}

	handler.ServeUpload(c)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "active attachment" {
		t.Fatalf("status/body = %d/%q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("Cache-Control = %q", recorder.Header().Get("Cache-Control"))
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", recorder.Header().Get("X-Content-Type-Options"))
	}
}

func TestServeCommunityChatUploadDeniesDeletedAttachment(t *testing.T) {
	handler, _, token := newCommunityChatAttachmentTestHandler(t, false)
	fileName := "deleted.txt"
	if err := os.WriteFile(filepath.Join(handler.uploadsDir, fileName), []byte("deleted attachment"), communityChatFileMode); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community-chat/uploads/"+fileName, nil)
	c.Request.AddCookie(&http.Cookie{Name: communityChatAttachmentCookieName, Value: encodeCookieValue(token)})
	c.Params = gin.Params{{Key: "filename", Value: fileName}}

	handler.ServeUpload(c)

	if c.Writer.Status() != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", c.Writer.Status(), http.StatusNotFound)
	}
}

func TestDeleteCommunityChatMessageRemovesLocalAttachment(t *testing.T) {
	handler, user, _ := newCommunityChatAttachmentTestHandler(t, true)
	fileName := "remove-me.pdf"
	filePath := filepath.Join(handler.uploadsDir, fileName)
	if err := os.WriteFile(filePath, []byte("attachment"), communityChatFileMode); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	repo := &communityChatHandlerTestRepo{message: &service.CommunityChatMessage{
		ID:       7,
		UserID:   user.ID,
		ImageURL: "/api/v1/community-chat/uploads/" + fileName + "?name=report.pdf",
	}}
	handler.chatService = service.NewCommunityChatService(repo, &communityChatHandlerTestUserRepo{user: user})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/community-chat/messages/7", nil)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.DeleteMessage(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("attachment still exists or stat failed unexpectedly: %v", err)
	}
}

func TestDeleteCommunityChatDirectMessageRemovesLocalAttachment(t *testing.T) {
	handler, user, _ := newCommunityChatAttachmentTestHandler(t, true)
	fileName := "remove-private.pdf"
	filePath := filepath.Join(handler.uploadsDir, fileName)
	if err := os.WriteFile(filePath, []byte("private attachment"), communityChatFileMode); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	repo := &communityChatHandlerTestRepo{directMessage: &service.CommunityChatMessage{
		ID:                 8,
		ConversationUserID: user.ID,
		UserID:             user.ID,
		ImageURL:           "/api/v1/community-chat/uploads/" + fileName + "?name=private.pdf",
	}}
	handler.chatService = service.NewCommunityChatService(repo, &communityChatHandlerTestUserRepo{user: user})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/community-chat/direct/messages/8", nil)
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.DeleteDirectMessage(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("direct attachment still exists or stat failed unexpectedly: %v", err)
	}
}

func TestUploadCommunityChatDirectFileCreatesAttachmentMessage(t *testing.T) {
	handler, user, _ := newCommunityChatAttachmentTestHandler(t, true)
	repo := &communityChatHandlerTestRepo{}
	handler.chatService = service.NewCommunityChatService(repo, &communityChatHandlerTestUserRepo{user: user})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "private.txt")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte("private message attachment")); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.WriteField("content", "private note"); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/community-chat/direct/files", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: user.ID})

	handler.UploadDirectFile(c)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status/body = %d/%s, want %d", recorder.Code, recorder.Body.String(), http.StatusCreated)
	}
	if repo.createdDirect == nil || repo.createdDirect.ConversationUserID != user.ID || repo.createdDirect.MessageType != service.CommunityChatMessageTypeFile {
		t.Fatalf("unexpected direct attachment: %#v", repo.createdDirect)
	}
	if repo.createdDirect.ImageURL == "" || repo.createdDirect.ImageSizeBytes == 0 {
		t.Fatalf("direct attachment metadata missing: %#v", repo.createdDirect)
	}
}

func TestCommunityChatLocalAttachmentFilenameRejectsExternalURL(t *testing.T) {
	if got := communityChatLocalAttachmentFilename("https://example.com/api/v1/community-chat/uploads/demo.txt"); got != "" {
		t.Fatalf("external URL filename = %q, want empty", got)
	}
}
