package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type communityChatReliabilityRepo struct {
	CommunityChatRepository
	directMessage     *CommunityChatMessage
	createdDirect     *CommunityChatMessage
	deletedBy         int64
	unread            []CommunityChatDirectUnreadConversation
	attachmentAllowed bool
	searchQuery       string
	searchParams      pagination.PaginationParams
}

func (r *communityChatReliabilityRepo) SearchMessages(_ context.Context, query string, params pagination.PaginationParams) ([]CommunityChatMessage, *pagination.PaginationResult, error) {
	r.searchQuery = query
	r.searchParams = params
	return []CommunityChatMessage{{ID: 7, ImageURL: "/api/v1/community-chat/uploads/demo.pdf?name=demo.pdf", ImageMIMEType: "application/pdf", ImageSizeBytes: 128}}, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 1}, nil
}

func (r *communityChatReliabilityRepo) CreateDirectMessage(_ context.Context, msg *CommunityChatMessage) error {
	copy := *msg
	copy.ID = 101
	copy.CreatedAt = time.Now()
	copy.UpdatedAt = copy.CreatedAt
	*msg = copy
	r.createdDirect = &copy
	return nil
}

func (r *communityChatReliabilityRepo) CanAccessAttachment(_ context.Context, _ string, _ int64, _ bool) (bool, error) {
	return r.attachmentAllowed, nil
}

func (r *communityChatReliabilityRepo) GetDirectMessage(_ context.Context, id int64) (*CommunityChatMessage, error) {
	if r.directMessage == nil || r.directMessage.ID != id {
		return nil, ErrCommunityChatMessageNotFound
	}
	copy := *r.directMessage
	return &copy, nil
}

func (r *communityChatReliabilityRepo) DeleteDirectMessage(ctx context.Context, id, actorID int64) (*CommunityChatMessage, error) {
	msg, err := r.GetDirectMessage(ctx, id)
	if err != nil {
		return nil, err
	}
	r.deletedBy = actorID
	now := time.Now()
	msg.DeletedAt = &now
	msg.DeletedBy = &actorID
	return msg, nil
}

func (r *communityChatReliabilityRepo) ListDirectUnread(_ context.Context, _, _ int64, _ bool) ([]CommunityChatDirectUnreadConversation, error) {
	return r.unread, nil
}

type communityChatReliabilityUserRepo struct {
	UserRepository
	users       map[int64]*User
	listedUsers []User
	listFilters UserListFilters
}

func (r *communityChatReliabilityUserRepo) GetByID(_ context.Context, id int64) (*User, error) {
	if user := r.users[id]; user != nil {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (r *communityChatReliabilityUserRepo) GetUserAvatar(_ context.Context, _ int64) (*UserAvatar, error) {
	return nil, nil
}

func (r *communityChatReliabilityUserRepo) ListWithFilters(_ context.Context, params pagination.PaginationParams, filters UserListFilters) ([]User, *pagination.PaginationResult, error) {
	r.listFilters = filters
	return r.listedUsers, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: int64(len(r.listedUsers))}, nil
}

func TestCreateDirectMessageSupportsAttachments(t *testing.T) {
	repo := &communityChatReliabilityRepo{}
	userRepo := &communityChatReliabilityUserRepo{users: map[int64]*User{
		42: {ID: 42, Role: RoleUser, Username: "alice"},
	}}
	svc := NewCommunityChatService(repo, userRepo)

	message, err := svc.CreateDirectMessage(context.Background(), CommunityChatCreateDirectMessageInput{
		Actor:          userRepo.users[42],
		MessageType:    CommunityChatMessageTypeFile,
		Content:        "季度报告",
		ImageURL:       "/api/v1/community-chat/uploads/report.pdf?name=report.pdf",
		ImageMIMEType:  "application/pdf",
		ImageSizeBytes: 2048,
	})
	if err != nil {
		t.Fatalf("CreateDirectMessage() error = %v", err)
	}
	if repo.createdDirect == nil || repo.createdDirect.MessageType != CommunityChatMessageTypeFile {
		t.Fatalf("direct attachment was not persisted: %#v", repo.createdDirect)
	}
	if message.FileName != "report.pdf" || message.FileSizeBytes != 2048 {
		t.Fatalf("attachment aliases missing: %#v", message)
	}
}

func TestCreateDirectMessageRejectsInvalidAttachments(t *testing.T) {
	user := &User{ID: 42, Role: RoleUser, Username: "alice"}
	svc := NewCommunityChatService(&communityChatReliabilityRepo{}, &communityChatReliabilityUserRepo{users: map[int64]*User{42: user}})

	_, err := svc.CreateDirectMessage(context.Background(), CommunityChatCreateDirectMessageInput{
		Actor:         user,
		MessageType:   CommunityChatMessageTypeImage,
		Content:       "missing image",
		ImageMIMEType: "image/png",
	})
	if err != ErrCommunityChatInvalidImage {
		t.Fatalf("missing image error = %v, want %v", err, ErrCommunityChatInvalidImage)
	}

	_, err = svc.CreateDirectMessage(context.Background(), CommunityChatCreateDirectMessageInput{
		Actor:          user,
		MessageType:    CommunityChatMessageTypeFile,
		Content:        "large file",
		ImageURL:       "/api/v1/community-chat/uploads/large.pdf",
		ImageMIMEType:  "application/pdf",
		ImageSizeBytes: CommunityChatMaxAttachmentBytes + 1,
	})
	if err != ErrCommunityChatAttachmentTooLarge {
		t.Fatalf("large attachment error = %v, want %v", err, ErrCommunityChatAttachmentTooLarge)
	}
}

func TestCreateDirectMessageRequiresAdminTargetAndIgnoresUserTarget(t *testing.T) {
	repo := &communityChatReliabilityRepo{}
	userRepo := &communityChatReliabilityUserRepo{users: map[int64]*User{
		1:  {ID: 1, Role: RoleAdmin},
		42: {ID: 42, Role: RoleUser},
	}}
	svc := NewCommunityChatService(repo, userRepo)

	if _, err := svc.CreateDirectMessage(context.Background(), CommunityChatCreateDirectMessageInput{Actor: userRepo.users[1], Content: "hello"}); err != ErrCommunityChatDirectConversationRequired {
		t.Fatalf("admin missing target error = %v, want %v", err, ErrCommunityChatDirectConversationRequired)
	}
	message, err := svc.CreateDirectMessage(context.Background(), CommunityChatCreateDirectMessageInput{Actor: userRepo.users[42], ConversationUserID: 99, Content: "hello"})
	if err != nil {
		t.Fatalf("user CreateDirectMessage() error = %v", err)
	}
	if message.ConversationUserID != 42 {
		t.Fatalf("user selected conversation = %d, want own user ID", message.ConversationUserID)
	}
}

func TestSearchDirectUsersIsAdminOnlyAndReturnsActiveUsers(t *testing.T) {
	userRepo := &communityChatReliabilityUserRepo{listedUsers: []User{{ID: 42, Username: "alice", AvatarURL: "/avatar"}}}
	svc := NewCommunityChatService(&communityChatReliabilityRepo{}, userRepo)

	if _, _, err := svc.SearchDirectUsers(context.Background(), &User{ID: 7, Role: RoleUser}, "ali", pagination.PaginationParams{}); err != ErrCommunityChatForbidden {
		t.Fatalf("non-admin search error = %v, want %v", err, ErrCommunityChatForbidden)
	}
	users, result, err := svc.SearchDirectUsers(context.Background(), &User{ID: 1, Role: RoleAdmin}, " ali ", pagination.PaginationParams{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("SearchDirectUsers() error = %v", err)
	}
	if len(users) != 1 || users[0].ID != 42 || result.Total != 1 {
		t.Fatalf("unexpected users/result: %#v %#v", users, result)
	}
	if userRepo.listFilters.Search != "ali" || userRepo.listFilters.Status != StatusActive || userRepo.listFilters.Role != RoleUser {
		t.Fatalf("unexpected user filters: %#v", userRepo.listFilters)
	}
}

func TestSearchCommunityChatMessagesTrimsQueryAndNormalizesPagination(t *testing.T) {
	repo := &communityChatReliabilityRepo{}
	svc := NewCommunityChatService(repo, &communityChatReliabilityUserRepo{})

	messages, result, err := svc.SearchMessages(context.Background(), "  模型价格  ", pagination.PaginationParams{Page: 0, PageSize: 999})
	if err != nil {
		t.Fatalf("SearchMessages() error = %v", err)
	}
	if repo.searchQuery != "模型价格" || repo.searchParams.Page != 1 || repo.searchParams.PageSize != 100 {
		t.Fatalf("unexpected normalized search: query=%q params=%#v", repo.searchQuery, repo.searchParams)
	}
	if len(messages) != 1 || messages[0].FileName != "demo.pdf" || result.Total != 1 {
		t.Fatalf("unexpected messages/result: %#v %#v", messages, result)
	}
}

func TestNormalizeCommunityChatAttachmentAliasesUsesOriginalFileNameFromURL(t *testing.T) {
	msg := &CommunityChatMessage{
		Content:        "这里是用户写的说明",
		ImageURL:       "/api/v1/community-chat/uploads/generated.mp4?name=demo%20video.mp4",
		ImageMIMEType:  "video/mp4",
		ImageSizeBytes: 1024,
	}

	normalizeCommunityChatAttachmentAliases(msg)

	if msg.FileName != "demo video.mp4" {
		t.Fatalf("expected original file name, got %q", msg.FileName)
	}
	if msg.FileURL != msg.ImageURL {
		t.Fatalf("expected file URL alias to match image URL")
	}
}

func TestNormalizeCommunityChatAttachmentAliasesDoesNothingWithoutAttachmentURL(t *testing.T) {
	msg := &CommunityChatMessage{
		Content:        "fallback.pdf",
		ImageURL:       "",
		ImageMIMEType:  "application/pdf",
		ImageSizeBytes: 1024,
	}

	normalizeCommunityChatAttachmentAliases(msg)

	if msg.FileName != "" {
		t.Fatalf("expected no file name without attachment URL, got %q", msg.FileName)
	}
}

func TestDeleteDirectMessageAllowsSenderAndBroadcastsDeletion(t *testing.T) {
	repo := &communityChatReliabilityRepo{directMessage: &CommunityChatMessage{ID: 9, ConversationUserID: 42, UserID: 42}}
	userRepo := &communityChatReliabilityUserRepo{users: map[int64]*User{42: {ID: 42, Role: RoleUser}}}
	svc := NewCommunityChatService(repo, userRepo)
	events, unsubscribe := svc.Subscribe()
	defer unsubscribe()

	deleted, err := svc.DeleteDirectMessage(context.Background(), 9, 42)
	if err != nil {
		t.Fatalf("DeleteDirectMessage() error = %v", err)
	}
	if repo.deletedBy != 42 || deleted.DeletedAt == nil {
		t.Fatalf("direct message was not soft deleted by sender: %#v", deleted)
	}
	select {
	case event := <-events:
		if event.Type != CommunityChatEventDirectMessageDeleted || event.ID != 9 || event.ConversationUserID != 42 {
			t.Fatalf("unexpected delete event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for direct deletion event")
	}
}

func TestDeleteDirectMessageRejectsAnotherUserAndAllowsAdmin(t *testing.T) {
	repo := &communityChatReliabilityRepo{directMessage: &CommunityChatMessage{ID: 9, ConversationUserID: 42, UserID: 42}}
	userRepo := &communityChatReliabilityUserRepo{users: map[int64]*User{
		7: {ID: 7, Role: RoleUser},
		1: {ID: 1, Role: RoleAdmin},
	}}
	svc := NewCommunityChatService(repo, userRepo)

	if _, err := svc.DeleteDirectMessage(context.Background(), 9, 7); err != ErrCommunityChatForbidden {
		t.Fatalf("other user error = %v, want %v", err, ErrCommunityChatForbidden)
	}
	if repo.deletedBy != 0 {
		t.Fatalf("forbidden delete reached repository with actor %d", repo.deletedBy)
	}
	if _, err := svc.DeleteDirectMessage(context.Background(), 9, 1); err != nil {
		t.Fatalf("admin DeleteDirectMessage() error = %v", err)
	}
	if repo.deletedBy != 1 {
		t.Fatalf("admin delete actor = %d, want 1", repo.deletedBy)
	}
}

func TestDirectUnreadSummarizesIndependentConversations(t *testing.T) {
	repo := &communityChatReliabilityRepo{unread: []CommunityChatDirectUnreadConversation{
		{UserID: 42, UnreadCount: 2, LatestMessageID: 12},
		{UserID: 84, UnreadCount: 3, LatestMessageID: 18},
	}}
	svc := NewCommunityChatService(repo, &communityChatReliabilityUserRepo{})

	summary, err := svc.DirectUnread(context.Background(), &User{ID: 1, Role: RoleAdmin})
	if err != nil {
		t.Fatalf("DirectUnread() error = %v", err)
	}
	if summary.Total != 5 || len(summary.Conversations) != 2 {
		t.Fatalf("unexpected unread summary: %#v", summary)
	}
}

func TestCommunityChatRedisBroadcastReachesOtherInstanceWithoutSelfDuplicate(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	first := NewCommunityChatService(nil, nil, redisClient)
	second := NewCommunityChatService(nil, nil, redisClient)
	first.Start()
	second.Start()
	deadline := time.Now().Add(time.Second)
	for mr.PubSubNumSub(communityChatRedisChannel)[communityChatRedisChannel] < 2 {
		if time.Now().After(deadline) {
			t.Fatal("redis subscribers did not start")
		}
		time.Sleep(time.Millisecond)
	}

	firstEvents, unsubscribeFirst := first.Subscribe()
	defer unsubscribeFirst()
	secondEvents, unsubscribeSecond := second.Subscribe()
	defer unsubscribeSecond()
	first.Broadcast(CommunityChatEvent{Type: CommunityChatEventMessageDeleted, ID: 77})

	for name, events := range map[string]<-chan CommunityChatEvent{"first": firstEvents, "second": secondEvents} {
		select {
		case event := <-events:
			if event.ID != 77 {
				t.Fatalf("%s instance event = %#v", name, event)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %s instance", name)
		}
	}
	select {
	case event := <-firstEvents:
		t.Fatalf("origin instance received duplicate event: %#v", event)
	case <-time.After(30 * time.Millisecond):
	}
}
