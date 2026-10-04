package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/redis/go-redis/v9"
)

const (
	CommunityChatMessageTypeText  = "text"
	CommunityChatMessageTypeImage = "image"
	CommunityChatMessageTypeFile  = "file"

	CommunityChatEventMessageCreated       = "message_created"
	CommunityChatEventMessageDeleted       = "message_deleted"
	CommunityChatEventDirectMessageCreated = "direct_message_created"
	CommunityChatEventDirectMessageDeleted = "direct_message_deleted"

	CommunityChatMaxTextLength            = 2000
	CommunityChatMaxReplyPreview          = 120
	CommunityChatMaxAttachmentBytes int64 = 15 << 20
	CommunityChatMaxImageBytes            = CommunityChatMaxAttachmentBytes
)

var (
	ErrCommunityChatMessageNotFound            = infraerrors.NotFound("COMMUNITY_CHAT_MESSAGE_NOT_FOUND", "message not found")
	ErrCommunityChatEmptyMessage               = infraerrors.BadRequest("COMMUNITY_CHAT_EMPTY_MESSAGE", "message cannot be empty")
	ErrCommunityChatMessageTooLong             = infraerrors.BadRequest("COMMUNITY_CHAT_MESSAGE_TOO_LONG", "message is too long")
	ErrCommunityChatInvalidImage               = infraerrors.BadRequest("COMMUNITY_CHAT_INVALID_IMAGE", "image is invalid")
	ErrCommunityChatImageTooLarge              = infraerrors.BadRequest("COMMUNITY_CHAT_IMAGE_TOO_LARGE", "image must be 15MB or smaller")
	ErrCommunityChatInvalidAttachment          = infraerrors.BadRequest("COMMUNITY_CHAT_INVALID_ATTACHMENT", "file is invalid or unsupported")
	ErrCommunityChatAttachmentTooLarge         = infraerrors.BadRequest("COMMUNITY_CHAT_ATTACHMENT_TOO_LARGE", "file must be 15MB or smaller")
	ErrCommunityChatForbidden                  = infraerrors.Forbidden("COMMUNITY_CHAT_FORBIDDEN", "only admins can perform this action")
	ErrCommunityChatDirectConversationRequired = infraerrors.BadRequest("COMMUNITY_CHAT_DIRECT_CONVERSATION_REQUIRED", "conversation user is required")
)

type CommunityChatMessage struct {
	ID                 int64      `json:"id"`
	ConversationUserID int64      `json:"conversation_user_id,omitempty"`
	UserID             int64      `json:"user_id"`
	Username           string     `json:"username"`
	AvatarURL          string     `json:"avatar_url"`
	MessageType        string     `json:"message_type"`
	Content            string     `json:"content"`
	ImageURL           string     `json:"image_url"`
	ImageMIMEType      string     `json:"image_mime_type"`
	ImageSizeBytes     int64      `json:"image_size_bytes"`
	FileURL            string     `json:"file_url,omitempty"`
	FileMIMEType       string     `json:"file_mime_type,omitempty"`
	FileSizeBytes      int64      `json:"file_size_bytes,omitempty"`
	FileName           string     `json:"file_name,omitempty"`
	ReplyToMessageID   *int64     `json:"reply_to_message_id,omitempty"`
	ReplyToUsername    string     `json:"reply_to_username"`
	ReplyToContent     string     `json:"reply_to_content"`
	ReplyToMessageType string     `json:"reply_to_message_type"`
	ReplyToImageURL    string     `json:"reply_to_image_url"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	DeletedBy          *int64     `json:"deleted_by,omitempty"`
	SentAt             *time.Time `json:"sent_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CommunityChatCreateMessageInput struct {
	User             *User
	MessageType      string
	Content          string
	ImageURL         string
	ImageMIMEType    string
	ImageSizeBytes   int64
	ReplyToMessageID *int64
}

type CommunityChatCreateDirectMessageInput struct {
	Actor              *User
	ConversationUserID int64
	MessageType        string
	Content            string
	ImageURL           string
	ImageMIMEType      string
	ImageSizeBytes     int64
}

type CommunityChatDirectUser struct {
	ID        int64  `json:"user_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type CommunityChatDirectConversation struct {
	UserID      int64                 `json:"user_id"`
	Username    string                `json:"username"`
	AvatarURL   string                `json:"avatar_url"`
	LastMessage *CommunityChatMessage `json:"last_message,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at"`
	UnreadCount int64                 `json:"unread_count"`
}

type CommunityChatDirectUnreadConversation struct {
	UserID          int64 `json:"user_id"`
	UnreadCount     int64 `json:"unread_count"`
	LatestMessageID int64 `json:"latest_message_id"`
}

type CommunityChatDirectUnreadSummary struct {
	Total         int64                                   `json:"total"`
	Conversations []CommunityChatDirectUnreadConversation `json:"conversations"`
}

type CommunityChatEvent struct {
	Type               string                `json:"type"`
	Message            *CommunityChatMessage `json:"message,omitempty"`
	ID                 int64                 `json:"id,omitempty"`
	ConversationUserID int64                 `json:"conversation_user_id,omitempty"`
}

type CommunityChatRepository interface {
	CreateMessage(ctx context.Context, msg *CommunityChatMessage) error
	ListMessages(ctx context.Context, params pagination.PaginationParams) ([]CommunityChatMessage, *pagination.PaginationResult, error)
	SearchMessages(ctx context.Context, query string, params pagination.PaginationParams) ([]CommunityChatMessage, *pagination.PaginationResult, error)
	ListMessagesBefore(ctx context.Context, beforeID int64, limit int) ([]CommunityChatMessage, error)
	ListMessagesAfter(ctx context.Context, afterID int64, limit int) ([]CommunityChatMessage, error)
	GetMessage(ctx context.Context, id int64) (*CommunityChatMessage, error)
	CanAccessAttachment(ctx context.Context, filename string, userID int64, isAdmin bool) (bool, error)
	DeleteMessage(ctx context.Context, id, actorID int64) (*CommunityChatMessage, error)
	CreateDirectMessage(ctx context.Context, msg *CommunityChatMessage) error
	ListDirectMessages(ctx context.Context, conversationUserID int64, params pagination.PaginationParams) ([]CommunityChatMessage, *pagination.PaginationResult, error)
	ListDirectMessagesBefore(ctx context.Context, conversationUserID, beforeID int64, limit int) ([]CommunityChatMessage, error)
	ListDirectMessagesAfter(ctx context.Context, conversationUserID, afterID int64, limit int) ([]CommunityChatMessage, error)
	ListDirectConversations(ctx context.Context, params pagination.PaginationParams) ([]CommunityChatDirectConversation, *pagination.PaginationResult, error)
	GetDirectMessage(ctx context.Context, id int64) (*CommunityChatMessage, error)
	DeleteDirectMessage(ctx context.Context, id, actorID int64) (*CommunityChatMessage, error)
	ListDirectUnread(ctx context.Context, readerUserID, conversationUserID int64, readerIsAdmin bool) ([]CommunityChatDirectUnreadConversation, error)
	MarkDirectRead(ctx context.Context, readerUserID, conversationUserID, lastMessageID int64) (int64, error)
}

func (s *CommunityChatService) CanAccessAttachment(ctx context.Context, filename string, userID int64, isAdmin bool) (bool, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" || userID <= 0 {
		return false, nil
	}
	return s.repo.CanAccessAttachment(ctx, filename, userID, isAdmin)
}

type CommunityChatService struct {
	repo        CommunityChatRepository
	userRepo    UserRepository
	subscribers map[chan CommunityChatEvent]struct{}
	mu          sync.RWMutex
	redis       *redis.Client
	instanceID  string
	startOnce   sync.Once
}

func NewCommunityChatService(repo CommunityChatRepository, userRepo UserRepository, redisClients ...*redis.Client) *CommunityChatService {
	svc := &CommunityChatService{
		repo:        repo,
		userRepo:    userRepo,
		subscribers: make(map[chan CommunityChatEvent]struct{}),
	}
	if len(redisClients) > 0 {
		svc.redis = redisClients[0]
		svc.instanceID = fmt.Sprintf("%d-%p", time.Now().UnixNano(), svc)
	}
	return svc
}

const communityChatRedisChannel = "sub2api:community-chat:events"

type communityChatRedisEvent struct {
	Origin string             `json:"origin"`
	Event  CommunityChatEvent `json:"event"`
}

func (s *CommunityChatService) Start() {
	if s == nil || s.redis == nil {
		return
	}
	s.startOnce.Do(func() {
		go s.subscribeRedisEvents()
	})
}

func (s *CommunityChatService) subscribeRedisEvents() {
	pubsub := s.redis.Subscribe(context.Background(), communityChatRedisChannel)
	defer func() { _ = pubsub.Close() }()
	for message := range pubsub.Channel() {
		if message == nil {
			continue
		}
		var envelope communityChatRedisEvent
		if err := json.Unmarshal([]byte(message.Payload), &envelope); err != nil || envelope.Origin == s.instanceID {
			continue
		}
		s.broadcastLocal(envelope.Event)
	}
}

func (s *CommunityChatService) ListMessages(ctx context.Context, params pagination.PaginationParams) ([]CommunityChatMessage, *pagination.PaginationResult, error) {
	params = normalizeCommunityChatPagination(params)
	messages, result, err := s.repo.ListMessages(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	for i := range messages {
		normalizeCommunityChatAttachmentAliases(&messages[i])
	}
	return messages, result, nil
}

func (s *CommunityChatService) SearchMessages(ctx context.Context, query string, params pagination.PaginationParams) ([]CommunityChatMessage, *pagination.PaginationResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []CommunityChatMessage{}, &pagination.PaginationResult{Page: 1, PageSize: normalizeCommunityChatPagination(params).PageSize}, nil
	}
	params = normalizeCommunityChatPagination(params)
	messages, result, err := s.repo.SearchMessages(ctx, query, params)
	if err != nil {
		return nil, nil, err
	}
	for i := range messages {
		normalizeCommunityChatAttachmentAliases(&messages[i])
	}
	return messages, result, nil
}

func (s *CommunityChatService) ListMessagesBefore(ctx context.Context, beforeID int64, limit int) ([]CommunityChatMessage, error) {
	messages, err := s.repo.ListMessagesBefore(ctx, beforeID, normalizeCommunityChatCursorLimit(limit))
	for i := range messages {
		normalizeCommunityChatAttachmentAliases(&messages[i])
	}
	return messages, err
}

func (s *CommunityChatService) ListMessagesAfter(ctx context.Context, afterID int64, limit int) ([]CommunityChatMessage, error) {
	messages, err := s.repo.ListMessagesAfter(ctx, afterID, normalizeCommunityChatCursorLimit(limit))
	for i := range messages {
		normalizeCommunityChatAttachmentAliases(&messages[i])
	}
	return messages, err
}

func (s *CommunityChatService) ListDirectConversations(ctx context.Context, actor *User, params pagination.PaginationParams) ([]CommunityChatDirectConversation, *pagination.PaginationResult, error) {
	if actor == nil || actor.ID <= 0 {
		return nil, nil, ErrUserNotFound
	}
	if !actor.IsAdmin() {
		return nil, nil, ErrCommunityChatForbidden
	}
	params = normalizeCommunityChatPagination(params)
	conversations, result, err := s.repo.ListDirectConversations(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	unread, err := s.repo.ListDirectUnread(ctx, actor.ID, 0, true)
	if err != nil {
		return nil, nil, err
	}
	unreadByUser := make(map[int64]int64, len(unread))
	for _, item := range unread {
		unreadByUser[item.UserID] = item.UnreadCount
	}
	for i := range conversations {
		conversations[i].UnreadCount = unreadByUser[conversations[i].UserID]
		normalizeCommunityChatAttachmentAliases(conversations[i].LastMessage)
	}
	return conversations, result, nil
}

func (s *CommunityChatService) SearchDirectUsers(ctx context.Context, actor *User, search string, params pagination.PaginationParams) ([]CommunityChatDirectUser, *pagination.PaginationResult, error) {
	if actor == nil || actor.ID <= 0 {
		return nil, nil, ErrUserNotFound
	}
	if !actor.IsAdmin() {
		return nil, nil, ErrCommunityChatForbidden
	}
	params = normalizeCommunityChatPagination(params)
	includeSubscriptions := false
	users, result, err := s.userRepo.ListWithFilters(ctx, params, UserListFilters{
		Status:               StatusActive,
		Role:                 RoleUser,
		Search:               strings.TrimSpace(search),
		IncludeSubscriptions: &includeSubscriptions,
	})
	if err != nil {
		return nil, nil, err
	}
	items := make([]CommunityChatDirectUser, 0, len(users))
	for i := range users {
		avatar, avatarErr := s.userRepo.GetUserAvatar(ctx, users[i].ID)
		if avatarErr == nil {
			applyUserAvatar(&users[i], avatar)
		}
		items = append(items, CommunityChatDirectUser{
			ID:        users[i].ID,
			Username:  communityChatDisplayName(&users[i]),
			AvatarURL: strings.TrimSpace(users[i].AvatarURL),
		})
	}
	return items, result, nil
}

func (s *CommunityChatService) ListDirectMessages(ctx context.Context, actor *User, conversationUserID int64, params pagination.PaginationParams) ([]CommunityChatMessage, *pagination.PaginationResult, error) {
	targetUserID, err := s.resolveDirectConversationUserID(ctx, actor, conversationUserID)
	if err != nil {
		return nil, nil, err
	}
	params = normalizeCommunityChatPagination(params)
	messages, result, err := s.repo.ListDirectMessages(ctx, targetUserID, params)
	for i := range messages {
		normalizeCommunityChatAttachmentAliases(&messages[i])
	}
	return messages, result, err
}

func (s *CommunityChatService) ListDirectMessagesBefore(ctx context.Context, actor *User, conversationUserID, beforeID int64, limit int) ([]CommunityChatMessage, error) {
	targetUserID, err := s.resolveDirectConversationUserID(ctx, actor, conversationUserID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repo.ListDirectMessagesBefore(ctx, targetUserID, beforeID, normalizeCommunityChatCursorLimit(limit))
	for i := range messages {
		normalizeCommunityChatAttachmentAliases(&messages[i])
	}
	return messages, err
}

func (s *CommunityChatService) ListDirectMessagesAfter(ctx context.Context, actor *User, conversationUserID, afterID int64, limit int) ([]CommunityChatMessage, error) {
	targetUserID, err := s.resolveDirectConversationUserID(ctx, actor, conversationUserID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repo.ListDirectMessagesAfter(ctx, targetUserID, afterID, normalizeCommunityChatCursorLimit(limit))
	for i := range messages {
		normalizeCommunityChatAttachmentAliases(&messages[i])
	}
	return messages, err
}

func (s *CommunityChatService) DirectUnread(ctx context.Context, actor *User) (*CommunityChatDirectUnreadSummary, error) {
	if actor == nil || actor.ID <= 0 {
		return nil, ErrUserNotFound
	}
	conversationUserID := actor.ID
	if actor.IsAdmin() {
		conversationUserID = 0
	}
	items, err := s.repo.ListDirectUnread(ctx, actor.ID, conversationUserID, actor.IsAdmin())
	if err != nil {
		return nil, err
	}
	summary := &CommunityChatDirectUnreadSummary{Conversations: items}
	for _, item := range items {
		summary.Total += item.UnreadCount
	}
	return summary, nil
}

func (s *CommunityChatService) MarkDirectRead(ctx context.Context, actor *User, conversationUserID, lastMessageID int64) (int64, error) {
	targetUserID, err := s.resolveDirectConversationUserID(ctx, actor, conversationUserID)
	if err != nil {
		return 0, err
	}
	return s.repo.MarkDirectRead(ctx, actor.ID, targetUserID, lastMessageID)
}

func (s *CommunityChatService) CreateMessage(ctx context.Context, input CommunityChatCreateMessageInput) (*CommunityChatMessage, error) {
	if input.User == nil || input.User.ID <= 0 {
		return nil, ErrUserNotFound
	}

	messageType := strings.TrimSpace(input.MessageType)
	if messageType == "" {
		messageType = CommunityChatMessageTypeText
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrCommunityChatEmptyMessage
	}
	if len([]rune(content)) > CommunityChatMaxTextLength {
		return nil, ErrCommunityChatMessageTooLong
	}

	msg := &CommunityChatMessage{
		UserID:        input.User.ID,
		Username:      communityChatDisplayName(input.User),
		AvatarURL:     strings.TrimSpace(input.User.AvatarURL),
		MessageType:   messageType,
		Content:       content,
		ImageURL:      strings.TrimSpace(input.ImageURL),
		ImageMIMEType: strings.TrimSpace(input.ImageMIMEType),
	}
	if input.ImageSizeBytes > 0 {
		msg.ImageSizeBytes = input.ImageSizeBytes
	}
	normalizeCommunityChatAttachmentAliases(msg)
	if input.ReplyToMessageID != nil && *input.ReplyToMessageID > 0 {
		replyTo, err := s.repo.GetMessage(ctx, *input.ReplyToMessageID)
		if err != nil {
			return nil, err
		}
		msg.ReplyToMessageID = &replyTo.ID
		msg.ReplyToUsername = replyTo.Username
		msg.ReplyToContent = truncateCommunityChatPreview(replyTo.Content)
		msg.ReplyToMessageType = replyTo.MessageType
		msg.ReplyToImageURL = replyTo.ImageURL
	}

	switch messageType {
	case CommunityChatMessageTypeText:
		msg.ImageURL = ""
		msg.ImageMIMEType = ""
		msg.ImageSizeBytes = 0
	case CommunityChatMessageTypeImage:
		if msg.ImageURL == "" || msg.ImageMIMEType == "" || msg.ImageSizeBytes <= 0 {
			return nil, ErrCommunityChatInvalidImage
		}
		if msg.ImageSizeBytes > CommunityChatMaxImageBytes {
			return nil, ErrCommunityChatImageTooLarge
		}
	case CommunityChatMessageTypeFile:
		if msg.ImageURL == "" || msg.ImageMIMEType == "" || msg.ImageSizeBytes <= 0 {
			return nil, ErrCommunityChatInvalidAttachment
		}
		if msg.ImageSizeBytes > CommunityChatMaxAttachmentBytes {
			return nil, ErrCommunityChatAttachmentTooLarge
		}
	default:
		return nil, ErrCommunityChatEmptyMessage
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	normalizeCommunityChatAttachmentAliases(msg)
	s.Broadcast(CommunityChatEvent{Type: CommunityChatEventMessageCreated, Message: msg})
	return msg, nil
}

func (s *CommunityChatService) CreateDirectMessage(ctx context.Context, input CommunityChatCreateDirectMessageInput) (*CommunityChatMessage, error) {
	targetUserID, err := s.resolveDirectConversationUserID(ctx, input.Actor, input.ConversationUserID)
	if err != nil {
		return nil, err
	}

	messageType := strings.TrimSpace(input.MessageType)
	if messageType == "" {
		messageType = CommunityChatMessageTypeText
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrCommunityChatEmptyMessage
	}
	if len([]rune(content)) > CommunityChatMaxTextLength {
		return nil, ErrCommunityChatMessageTooLong
	}

	msg := &CommunityChatMessage{
		ConversationUserID: targetUserID,
		UserID:             input.Actor.ID,
		Username:           communityChatDisplayName(input.Actor),
		AvatarURL:          strings.TrimSpace(input.Actor.AvatarURL),
		MessageType:        messageType,
		Content:            content,
		ImageURL:           strings.TrimSpace(input.ImageURL),
		ImageMIMEType:      strings.TrimSpace(input.ImageMIMEType),
		ImageSizeBytes:     input.ImageSizeBytes,
	}
	normalizeCommunityChatAttachmentAliases(msg)
	switch messageType {
	case CommunityChatMessageTypeText:
		msg.ImageURL = ""
		msg.ImageMIMEType = ""
		msg.ImageSizeBytes = 0
	case CommunityChatMessageTypeImage:
		if msg.ImageURL == "" || msg.ImageMIMEType == "" || msg.ImageSizeBytes <= 0 {
			return nil, ErrCommunityChatInvalidImage
		}
		if msg.ImageSizeBytes > CommunityChatMaxImageBytes {
			return nil, ErrCommunityChatImageTooLarge
		}
	case CommunityChatMessageTypeFile:
		if msg.ImageURL == "" || msg.ImageMIMEType == "" || msg.ImageSizeBytes <= 0 {
			return nil, ErrCommunityChatInvalidAttachment
		}
		if msg.ImageSizeBytes > CommunityChatMaxAttachmentBytes {
			return nil, ErrCommunityChatAttachmentTooLarge
		}
	default:
		return nil, ErrCommunityChatEmptyMessage
	}
	if err := s.repo.CreateDirectMessage(ctx, msg); err != nil {
		return nil, err
	}
	normalizeCommunityChatAttachmentAliases(msg)
	s.Broadcast(CommunityChatEvent{
		Type:               CommunityChatEventDirectMessageCreated,
		Message:            msg,
		ConversationUserID: targetUserID,
	})
	return msg, nil
}

func (s *CommunityChatService) DeleteMessage(ctx context.Context, id, actorID int64) (*CommunityChatMessage, error) {
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	msg, err := s.repo.GetMessage(ctx, id)
	if err != nil {
		return nil, err
	}
	if msg.UserID != actorID && !actor.IsAdmin() {
		return nil, ErrCommunityChatForbidden
	}

	msg, err = s.repo.DeleteMessage(ctx, id, actorID)
	if err != nil {
		return nil, err
	}
	normalizeCommunityChatAttachmentAliases(msg)
	s.Broadcast(CommunityChatEvent{Type: CommunityChatEventMessageDeleted, ID: id, Message: msg})
	return msg, nil
}

func (s *CommunityChatService) DeleteDirectMessage(ctx context.Context, id, actorID int64) (*CommunityChatMessage, error) {
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	msg, err := s.repo.GetDirectMessage(ctx, id)
	if err != nil {
		return nil, err
	}
	if msg.UserID != actorID && !actor.IsAdmin() {
		return nil, ErrCommunityChatForbidden
	}
	msg, err = s.repo.DeleteDirectMessage(ctx, id, actorID)
	if err != nil {
		return nil, err
	}
	normalizeCommunityChatAttachmentAliases(msg)
	s.Broadcast(CommunityChatEvent{
		Type:               CommunityChatEventDirectMessageDeleted,
		ID:                 id,
		Message:            msg,
		ConversationUserID: msg.ConversationUserID,
	})
	return msg, nil
}

func normalizeCommunityChatAttachmentAliases(msg *CommunityChatMessage) {
	if msg == nil || msg.ImageURL == "" {
		return
	}
	msg.FileURL = msg.ImageURL
	msg.FileMIMEType = msg.ImageMIMEType
	msg.FileSizeBytes = msg.ImageSizeBytes
	msg.FileName = communityChatFileNameFromURL(msg.ImageURL)
	if msg.FileName == "" {
		msg.FileName = msg.Content
	}
}

func communityChatFileNameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil {
		if name := strings.TrimSpace(parsed.Query().Get("name")); name != "" {
			return name
		}
		base := strings.TrimSpace(path.Base(parsed.Path))
		if base != "." && base != "/" {
			return base
		}
	}
	base := strings.TrimSpace(path.Base(rawURL))
	if base == "." || base == "/" {
		return ""
	}
	return base
}

func communityChatDisplayName(user *User) string {
	username := strings.TrimSpace(user.Username)
	if username != "" {
		return username
	}
	email := strings.TrimSpace(user.Email)
	if email != "" {
		return email
	}
	return "User"
}

func normalizeCommunityChatPagination(params pagination.PaginationParams) pagination.PaginationParams {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 50
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	return params
}

func normalizeCommunityChatCursorLimit(limit int) int {
	if limit <= 0 {
		return 80
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func (s *CommunityChatService) resolveDirectConversationUserID(ctx context.Context, actor *User, requestedUserID int64) (int64, error) {
	if actor == nil || actor.ID <= 0 {
		return 0, ErrUserNotFound
	}
	if !actor.IsAdmin() {
		return actor.ID, nil
	}
	if requestedUserID <= 0 {
		return 0, ErrCommunityChatDirectConversationRequired
	}
	target, err := s.userRepo.GetByID(ctx, requestedUserID)
	if err != nil {
		return 0, err
	}
	return target.ID, nil
}

func truncateCommunityChatPreview(content string) string {
	runes := []rune(strings.TrimSpace(content))
	if len(runes) <= CommunityChatMaxReplyPreview {
		return string(runes)
	}
	return string(runes[:CommunityChatMaxReplyPreview]) + "..."
}

func (s *CommunityChatService) Subscribe() (<-chan CommunityChatEvent, func()) {
	ch := make(chan CommunityChatEvent, 32)
	s.mu.Lock()
	s.subscribers[ch] = struct{}{}
	s.mu.Unlock()

	unsubscribe := func() {
		s.mu.Lock()
		if _, ok := s.subscribers[ch]; ok {
			delete(s.subscribers, ch)
			close(ch)
		}
		s.mu.Unlock()
	}
	return ch, unsubscribe
}

func (s *CommunityChatService) Broadcast(event CommunityChatEvent) {
	s.broadcastLocal(event)
	if s.redis == nil {
		return
	}
	payload, err := json.Marshal(communityChatRedisEvent{Origin: s.instanceID, Event: event})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.redis.Publish(ctx, communityChatRedisChannel, payload).Err(); err != nil {
		slog.Warn("publish community chat event", "error", err)
	}
}

func (s *CommunityChatService) broadcastLocal(event CommunityChatEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}
