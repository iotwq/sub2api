package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

var communityDirectMessageColumns = []string{
	"id", "conversation_user_id", "user_id", "username", "avatar_url", "message_type", "content",
	"image_url", "image_mime_type", "image_size_bytes", "deleted_at", "deleted_by", "sent_at", "created_at", "updated_at",
}

func TestCommunityChatDirectReadBoundedByDisplayedMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}
	// A concurrent later message must not become the read boundary.
	mock.ExpectQuery(`(?s)WHERE conversation_user_id = \$2 AND deleted_at IS NULL AND id <= \$3.*GREATEST`).
		WithArgs(int64(1), int64(42), int64(19)).
		WillReturnRows(sqlmock.NewRows([]string{"last_message_id"}).AddRow(19))
	id, err := repo.MarkDirectRead(context.Background(), 1, 42, 19)
	if err != nil || id != 19 {
		t.Fatalf("read cursor = %d, error = %v", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

var communityMessageColumns = []string{
	"id", "user_id", "username", "avatar_url", "message_type", "content", "image_url", "image_mime_type", "image_size_bytes",
	"reply_to_message_id", "reply_to_username", "reply_to_content", "reply_to_message_type", "reply_to_image_url", "deleted_at", "deleted_by", "sent_at", "created_at", "updated_at",
}

func TestCommunityChatRepositorySearchMessagesUsesLiteralCaseInsensitiveContains(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}
	now := time.Now()

	mock.ExpectQuery(`(?s)SELECT COUNT\(\*\).*deleted_at IS NULL.*STRPOS\(LOWER\(content\), LOWER\(\$1\)\).*STRPOS\(LOWER\(username\), LOWER\(\$1\)\)`).
		WithArgs("100%_可用").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`(?s)FROM community_chat_messages.*deleted_at IS NULL.*STRPOS\(LOWER\(content\), LOWER\(\$1\)\).*ORDER BY created_at DESC, id DESC`).
		WithArgs("100%_可用", 20, 0).
		WillReturnRows(sqlmock.NewRows(communityMessageColumns).AddRow(
			int64(7), int64(42), "Alice", "", "text", "100%_可用", "", "", int64(0), nil, "", "", "", "", nil, nil, now, now, now,
		))

	messages, result, err := repo.SearchMessages(context.Background(), "100%_可用", pagination.PaginationParams{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("SearchMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].ID != 7 || result.Total != 1 {
		t.Fatalf("unexpected messages/result: %#v %#v", messages, result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommunityChatRepositoryCreateDirectMessagePersistsAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}
	now := time.Now()
	message := &service.CommunityChatMessage{
		ConversationUserID: 42,
		UserID:             1,
		Username:           "admin",
		MessageType:        service.CommunityChatMessageTypeFile,
		Content:            "report",
		ImageURL:           "/api/v1/community-chat/uploads/report.pdf",
		ImageMIMEType:      "application/pdf",
		ImageSizeBytes:     2048,
	}

	mock.ExpectQuery(`(?s)INSERT INTO community_chat_direct_messages.*VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, NOW\(\)\)`).
		WithArgs(int64(42), int64(1), "admin", "", "file", "report", message.ImageURL, message.ImageMIMEType, int64(2048)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sent_at", "created_at", "updated_at"}).AddRow(int64(9), now, now, now))

	if err := repo.CreateDirectMessage(context.Background(), message); err != nil {
		t.Fatalf("CreateDirectMessage() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommunityChatRepositoryAttachmentAccessIncludesOwnedDirectMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}

	mock.ExpectQuery(`(?s)SELECT EXISTS.*community_chat_messages.*community_chat_direct_messages.*conversation_user_id = \$2 OR \$3`).
		WithArgs("/api/v1/community-chat/uploads/private.pdf", int64(42), false).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	allowed, err := repo.CanAccessAttachment(context.Background(), "private.pdf", 42, false)
	if err != nil || !allowed {
		t.Fatalf("CanAccessAttachment() = %v, %v", allowed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommunityChatRepositoryAttachmentAccessRejectsUnrelatedUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}

	mock.ExpectQuery(`(?s)SELECT EXISTS.*community_chat_messages.*community_chat_direct_messages.*conversation_user_id = \$2 OR \$3`).
		WithArgs("/api/v1/community-chat/uploads/private.pdf", int64(84), false).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	allowed, err := repo.CanAccessAttachment(context.Background(), "private.pdf", 84, false)
	if err != nil || allowed {
		t.Fatalf("CanAccessAttachment() = %v, %v", allowed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommunityChatRepositoryListDirectMessagesExcludesDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM community_chat_direct_messages WHERE conversation_user_id = $1 AND deleted_at IS NULL`)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`(?s)FROM community_chat_direct_messages\s+WHERE conversation_user_id = \$1 AND deleted_at IS NULL`).
		WithArgs(int64(42), 20, 0).
		WillReturnRows(sqlmock.NewRows(communityDirectMessageColumns).AddRow(
			int64(7), int64(42), int64(42), "user", "", "text", "hello", "", "", int64(0), nil, nil, now, now, now,
		))

	messages, result, err := repo.ListDirectMessages(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListDirectMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].ID != 7 || result.Total != 1 {
		t.Fatalf("unexpected direct messages: %#v result=%#v", messages, result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommunityChatRepositoryDirectUnreadIsGroupedByConversation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}

	mock.ExpectQuery(`(?s)m\.deleted_at IS NULL.*m\.id > COALESCE\(r\.last_message_id, 0\).*m\.user_id = m\.conversation_user_id.*GROUP BY m\.conversation_user_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"conversation_user_id", "count", "max"}).
			AddRow(int64(42), int64(2), int64(12)).
			AddRow(int64(84), int64(1), int64(15)))

	items, err := repo.ListDirectUnread(context.Background(), 1, 0, true)
	if err != nil {
		t.Fatalf("ListDirectUnread() error = %v", err)
	}
	if len(items) != 2 || items[0].UserID != 42 || items[0].UnreadCount != 2 || items[1].UserID != 84 {
		t.Fatalf("unexpected unread conversations: %#v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommunityChatRepositoryDeleteDirectMessageSoftDeletesVisibleMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	repo := &communityChatRepository{db: db}
	now := time.Now()

	mock.ExpectQuery(`(?s)UPDATE community_chat_direct_messages.*SET deleted_at = NOW\(\), deleted_by = \$2.*WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows(communityDirectMessageColumns).AddRow(
			int64(7), int64(42), int64(42), "user", "", "text", "hello", "", "", int64(0), now, int64(42), now, now, now,
		))

	message, err := repo.DeleteDirectMessage(context.Background(), 7, 42)
	if err != nil {
		t.Fatalf("DeleteDirectMessage() error = %v", err)
	}
	if message.DeletedAt == nil || message.DeletedBy == nil || *message.DeletedBy != 42 {
		t.Fatalf("unexpected deleted message: %#v", message)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
