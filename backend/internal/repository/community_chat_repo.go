package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type communityChatRepository struct {
	db *sql.DB
}

func NewCommunityChatRepository(db *sql.DB) service.CommunityChatRepository {
	return &communityChatRepository{db: db}
}

func (r *communityChatRepository) CreateMessage(ctx context.Context, msg *service.CommunityChatMessage) error {
	var sentAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO community_chat_messages (
			user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes,
			reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, sent_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
		RETURNING id, sent_at, created_at, updated_at
	`,
		msg.UserID,
		msg.Username,
		msg.AvatarURL,
		msg.MessageType,
		msg.Content,
		msg.ImageURL,
		msg.ImageMIMEType,
		msg.ImageSizeBytes,
		msg.ReplyToMessageID,
		msg.ReplyToUsername,
		msg.ReplyToContent,
		msg.ReplyToMessageType,
		msg.ReplyToImageURL,
	).Scan(&msg.ID, &sentAt, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create community chat message: %w", err)
	}
	if sentAt.Valid {
		msg.SentAt = &sentAt.Time
	}
	return nil
}

func (r *communityChatRepository) ListMessages(ctx context.Context, params pagination.PaginationParams) ([]service.CommunityChatMessage, *pagination.PaginationResult, error) {
	total := int64(0)
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_chat_messages WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count community chat messages: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM (
			SELECT id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
			FROM community_chat_messages
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
			LIMIT $1 OFFSET $2
		) AS recent
		ORDER BY created_at ASC, id ASC
	`, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, fmt.Errorf("list community chat messages: %w", err)
	}
	defer rows.Close()

	messages, err := scanCommunityChatMessages(rows)
	if err != nil {
		return nil, nil, err
	}
	return messages, paginationResultFromTotal(total, params), nil
}

func (r *communityChatRepository) SearchMessages(ctx context.Context, query string, params pagination.PaginationParams) ([]service.CommunityChatMessage, *pagination.PaginationResult, error) {
	total := int64(0)
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM community_chat_messages
		WHERE deleted_at IS NULL
			AND (STRPOS(LOWER(content), LOWER($1)) > 0 OR STRPOS(LOWER(username), LOWER($1)) > 0)
	`, query).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count community chat search results: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM community_chat_messages
		WHERE deleted_at IS NULL
			AND (STRPOS(LOWER(content), LOWER($1)) > 0 OR STRPOS(LOWER(username), LOWER($1)) > 0)
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, query, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, fmt.Errorf("search community chat messages: %w", err)
	}
	defer rows.Close()

	messages, err := scanCommunityChatMessages(rows)
	if err != nil {
		return nil, nil, err
	}
	return messages, paginationResultFromTotal(total, params), nil
}
func (r *communityChatRepository) ListMessagesBefore(ctx context.Context, beforeID int64, limit int) ([]service.CommunityChatMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM (
			SELECT id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
			FROM community_chat_messages
			WHERE deleted_at IS NULL AND id < $1
			ORDER BY id DESC
			LIMIT $2
		) AS older
		ORDER BY id ASC
	`, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("list older community chat messages: %w", err)
	}
	defer rows.Close()
	return scanCommunityChatMessages(rows)
}

func (r *communityChatRepository) ListMessagesAfter(ctx context.Context, afterID int64, limit int) ([]service.CommunityChatMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM community_chat_messages
		WHERE deleted_at IS NULL AND id > $1
		ORDER BY id ASC
		LIMIT $2
	`, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list newer community chat messages: %w", err)
	}
	defer rows.Close()
	return scanCommunityChatMessages(rows)
}

func (r *communityChatRepository) GetMessage(ctx context.Context, id int64) (*service.CommunityChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM community_chat_messages
		WHERE id = $1
	`, id)
	msg, err := scanCommunityChatMessage(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrCommunityChatMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get community chat message: %w", err)
	}
	return msg, nil
}

func (r *communityChatRepository) CanAccessAttachment(ctx context.Context, filename string, userID int64, isAdmin bool) (bool, error) {
	var allowed bool
	attachmentURL := "/api/v1/community-chat/uploads/" + filename
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM community_chat_messages
			WHERE deleted_at IS NULL
				AND split_part(image_url, '?', 1) = $1
			UNION ALL
			SELECT 1
			FROM community_chat_direct_messages
			WHERE deleted_at IS NULL
				AND split_part(image_url, '?', 1) = $1
				AND (conversation_user_id = $2 OR $3)
		)
	`, attachmentURL, userID, isAdmin).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("check community chat attachment access: %w", err)
	}
	return allowed, nil
}

func (r *communityChatRepository) DeleteMessage(ctx context.Context, id, actorID int64) (*service.CommunityChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE community_chat_messages
		SET deleted_at = COALESCE(deleted_at, NOW()), deleted_by = COALESCE(deleted_by, $2), updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, reply_to_message_id, reply_to_username, reply_to_content, reply_to_message_type, reply_to_image_url, deleted_at, deleted_by, sent_at, created_at, updated_at
	`, id, actorID)
	msg, err := scanCommunityChatMessage(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrCommunityChatMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("delete community chat message: %w", err)
	}
	return msg, nil
}

func (r *communityChatRepository) CreateDirectMessage(ctx context.Context, msg *service.CommunityChatMessage) error {
	var sentAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO community_chat_direct_messages (
			conversation_user_id, user_id, username, avatar_url, message_type, content,
			image_url, image_mime_type, image_size_bytes, sent_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, sent_at, created_at, updated_at
	`,
		msg.ConversationUserID,
		msg.UserID,
		msg.Username,
		msg.AvatarURL,
		msg.MessageType,
		msg.Content,
		msg.ImageURL,
		msg.ImageMIMEType,
		msg.ImageSizeBytes,
	).Scan(&msg.ID, &sentAt, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create community direct chat message: %w", err)
	}
	if sentAt.Valid {
		msg.SentAt = &sentAt.Time
	}
	return nil
}

func (r *communityChatRepository) ListDirectMessages(ctx context.Context, conversationUserID int64, params pagination.PaginationParams) ([]service.CommunityChatMessage, *pagination.PaginationResult, error) {
	total := int64(0)
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_chat_direct_messages WHERE conversation_user_id = $1 AND deleted_at IS NULL`, conversationUserID).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count community direct chat messages: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM (
			SELECT id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
			FROM community_chat_direct_messages
			WHERE conversation_user_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
			LIMIT $2 OFFSET $3
		) AS recent
		ORDER BY created_at ASC, id ASC
	`, conversationUserID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, fmt.Errorf("list community direct chat messages: %w", err)
	}
	defer rows.Close()

	messages, err := scanCommunityDirectChatMessages(rows)
	if err != nil {
		return nil, nil, err
	}
	return messages, paginationResultFromTotal(total, params), nil
}

func (r *communityChatRepository) ListDirectMessagesBefore(ctx context.Context, conversationUserID, beforeID int64, limit int) ([]service.CommunityChatMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM (
			SELECT id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
			FROM community_chat_direct_messages
			WHERE conversation_user_id = $1 AND deleted_at IS NULL AND id < $2
			ORDER BY id DESC
			LIMIT $3
		) AS older
		ORDER BY id ASC
	`, conversationUserID, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("list older community direct chat messages: %w", err)
	}
	defer rows.Close()
	return scanCommunityDirectChatMessages(rows)
}

func (r *communityChatRepository) ListDirectMessagesAfter(ctx context.Context, conversationUserID, afterID int64, limit int) ([]service.CommunityChatMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM community_chat_direct_messages
		WHERE conversation_user_id = $1 AND deleted_at IS NULL AND id > $2
		ORDER BY id ASC
		LIMIT $3
	`, conversationUserID, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list newer community direct chat messages: %w", err)
	}
	defer rows.Close()
	return scanCommunityDirectChatMessages(rows)
}

func (r *communityChatRepository) GetDirectMessage(ctx context.Context, id int64) (*service.CommunityChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
		FROM community_chat_direct_messages
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	msg, err := scanCommunityDirectChatMessage(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrCommunityChatMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get community direct chat message: %w", err)
	}
	return msg, nil
}

func (r *communityChatRepository) DeleteDirectMessage(ctx context.Context, id, actorID int64) (*service.CommunityChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE community_chat_direct_messages
		SET deleted_at = NOW(), deleted_by = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
	`, id, actorID)
	msg, err := scanCommunityDirectChatMessage(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrCommunityChatMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("delete community direct chat message: %w", err)
	}
	return msg, nil
}

func (r *communityChatRepository) ListDirectUnread(ctx context.Context, readerUserID, conversationUserID int64, readerIsAdmin bool) ([]service.CommunityChatDirectUnreadConversation, error) {
	query := `
		SELECT m.conversation_user_id, COUNT(*), MAX(m.id)
		FROM community_chat_direct_messages m
		LEFT JOIN community_chat_direct_read_states r
			ON r.reader_user_id = $1 AND r.conversation_user_id = m.conversation_user_id
		WHERE m.deleted_at IS NULL
			AND m.id > COALESCE(r.last_message_id, 0)
	`
	args := []any{readerUserID}
	if readerIsAdmin {
		query += ` AND m.user_id = m.conversation_user_id`
	} else {
		query += ` AND m.conversation_user_id = $2 AND m.user_id <> m.conversation_user_id`
		args = append(args, conversationUserID)
	}
	query += ` GROUP BY m.conversation_user_id ORDER BY MAX(m.id) DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list community direct unread messages: %w", err)
	}
	defer rows.Close()

	items := make([]service.CommunityChatDirectUnreadConversation, 0)
	for rows.Next() {
		var item service.CommunityChatDirectUnreadConversation
		if err := rows.Scan(&item.UserID, &item.UnreadCount, &item.LatestMessageID); err != nil {
			return nil, fmt.Errorf("scan community direct unread messages: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan community direct unread messages: %w", err)
	}
	return items, nil
}

func (r *communityChatRepository) MarkDirectRead(ctx context.Context, readerUserID, conversationUserID, visibleMessageID int64) (int64, error) {
	var lastMessageID int64
	err := r.db.QueryRowContext(ctx, `
		WITH latest AS (
			SELECT COALESCE(MAX(id), 0) AS id
			FROM community_chat_direct_messages
			WHERE conversation_user_id = $2 AND deleted_at IS NULL AND id <= $3
		), saved AS (
			INSERT INTO community_chat_direct_read_states (reader_user_id, conversation_user_id, last_message_id, updated_at)
			SELECT $1, $2, id, NOW() FROM latest
			ON CONFLICT (reader_user_id, conversation_user_id) DO UPDATE
			SET last_message_id = GREATEST(community_chat_direct_read_states.last_message_id, EXCLUDED.last_message_id),
				updated_at = NOW()
			RETURNING last_message_id
		)
		SELECT last_message_id FROM saved
	`, readerUserID, conversationUserID, visibleMessageID).Scan(&lastMessageID)
	if err != nil {
		return 0, fmt.Errorf("mark community direct conversation read: %w", err)
	}
	return lastMessageID, nil
}

func (r *communityChatRepository) ListDirectConversations(ctx context.Context, params pagination.PaginationParams) ([]service.CommunityChatDirectConversation, *pagination.PaginationResult, error) {
	total := int64(0)
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT conversation_user_id) FROM community_chat_direct_messages WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count community direct chat conversations: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (conversation_user_id)
				id, conversation_user_id, user_id, username, avatar_url, message_type, content, image_url, image_mime_type, image_size_bytes, deleted_at, deleted_by, sent_at, created_at, updated_at
			FROM community_chat_direct_messages
			WHERE deleted_at IS NULL
			ORDER BY conversation_user_id, created_at DESC, id DESC
		)
		SELECT
			latest.id, latest.conversation_user_id, latest.user_id, latest.username, latest.avatar_url, latest.message_type,
			latest.content, latest.image_url, latest.image_mime_type, latest.image_size_bytes, latest.deleted_at, latest.deleted_by, latest.sent_at, latest.created_at, latest.updated_at,
			COALESCE(NULLIF(users.username, ''), users.email, latest.username), COALESCE(user_avatars.url, '')
		FROM latest
		JOIN users ON users.id = latest.conversation_user_id
		LEFT JOIN user_avatars ON user_avatars.user_id = users.id
		ORDER BY latest.created_at DESC, latest.id DESC
		LIMIT $1 OFFSET $2
	`, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, fmt.Errorf("list community direct chat conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]service.CommunityChatDirectConversation, 0)
	for rows.Next() {
		msg, username, avatarURL, err := scanCommunityDirectChatConversationRow(rows)
		if err != nil {
			return nil, nil, err
		}
		updatedAt := msg.CreatedAt
		if msg.SentAt != nil {
			updatedAt = *msg.SentAt
		}
		conversations = append(conversations, service.CommunityChatDirectConversation{
			UserID:      msg.ConversationUserID,
			Username:    username,
			AvatarURL:   avatarURL,
			LastMessage: msg,
			UpdatedAt:   updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("scan community direct chat conversations: %w", err)
	}
	return conversations, paginationResultFromTotal(total, params), nil
}

type communityChatRowScanner interface {
	Scan(dest ...any) error
}

func scanCommunityChatMessage(row communityChatRowScanner) (*service.CommunityChatMessage, error) {
	var msg service.CommunityChatMessage
	var sentAt sql.NullTime
	if err := row.Scan(
		&msg.ID,
		&msg.UserID,
		&msg.Username,
		&msg.AvatarURL,
		&msg.MessageType,
		&msg.Content,
		&msg.ImageURL,
		&msg.ImageMIMEType,
		&msg.ImageSizeBytes,
		&msg.ReplyToMessageID,
		&msg.ReplyToUsername,
		&msg.ReplyToContent,
		&msg.ReplyToMessageType,
		&msg.ReplyToImageURL,
		&msg.DeletedAt,
		&msg.DeletedBy,
		&sentAt,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if sentAt.Valid {
		msg.SentAt = &sentAt.Time
	}
	return &msg, nil
}

func scanCommunityChatMessages(rows *sql.Rows) ([]service.CommunityChatMessage, error) {
	messages := make([]service.CommunityChatMessage, 0)
	for rows.Next() {
		msg, err := scanCommunityChatMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan community chat message: %w", err)
		}
		messages = append(messages, *msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan community chat messages: %w", err)
	}
	return messages, nil
}

func scanCommunityDirectChatMessage(row communityChatRowScanner) (*service.CommunityChatMessage, error) {
	var msg service.CommunityChatMessage
	var sentAt sql.NullTime
	if err := row.Scan(
		&msg.ID,
		&msg.ConversationUserID,
		&msg.UserID,
		&msg.Username,
		&msg.AvatarURL,
		&msg.MessageType,
		&msg.Content,
		&msg.ImageURL,
		&msg.ImageMIMEType,
		&msg.ImageSizeBytes,
		&msg.DeletedAt,
		&msg.DeletedBy,
		&sentAt,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if sentAt.Valid {
		msg.SentAt = &sentAt.Time
	}
	return &msg, nil
}

func scanCommunityDirectChatMessages(rows *sql.Rows) ([]service.CommunityChatMessage, error) {
	messages := make([]service.CommunityChatMessage, 0)
	for rows.Next() {
		msg, err := scanCommunityDirectChatMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan community direct chat message: %w", err)
		}
		messages = append(messages, *msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan community direct chat messages: %w", err)
	}
	return messages, nil
}

func scanCommunityDirectChatConversationRow(row communityChatRowScanner) (*service.CommunityChatMessage, string, string, error) {
	var msg service.CommunityChatMessage
	var sentAt sql.NullTime
	var username string
	var avatarURL string
	if err := row.Scan(
		&msg.ID,
		&msg.ConversationUserID,
		&msg.UserID,
		&msg.Username,
		&msg.AvatarURL,
		&msg.MessageType,
		&msg.Content,
		&msg.ImageURL,
		&msg.ImageMIMEType,
		&msg.ImageSizeBytes,
		&msg.DeletedAt,
		&msg.DeletedBy,
		&sentAt,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&username,
		&avatarURL,
	); err != nil {
		return nil, "", "", fmt.Errorf("scan community direct chat conversation: %w", err)
	}
	if sentAt.Valid {
		msg.SentAt = &sentAt.Time
	}
	return &msg, username, avatarURL, nil
}
