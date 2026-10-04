package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type openAIVideoTaskBindingRepository struct {
	db *sql.DB
}

func NewOpenAIVideoTaskBindingRepository(db *sql.DB) service.OpenAIVideoTaskBindingRepository {
	return &openAIVideoTaskBindingRepository{db: db}
}

func (r *openAIVideoTaskBindingRepository) UpsertOpenAIVideoTaskBinding(ctx context.Context, binding service.OpenAIVideoTaskBinding) error {
	if r == nil || r.db == nil {
		return nil
	}
	taskID := strings.TrimSpace(binding.TaskID)
	if taskID == "" || binding.UserID <= 0 || binding.AccountID <= 0 {
		return nil
	}
	if binding.ExpiresAt.IsZero() {
		return fmt.Errorf("openai video task binding expires_at is required")
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO openai_video_task_bindings (
			group_id, user_id, task_id, account_id, expires_at
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, group_id, task_id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
	`, binding.GroupID, binding.UserID, taskID, binding.AccountID, binding.ExpiresAt)
	if err != nil {
		return fmt.Errorf("upsert openai video task binding: %w", err)
	}
	return nil
}

func (r *openAIVideoTaskBindingRepository) GetOpenAIVideoTaskBinding(ctx context.Context, groupID, userID int64, taskID string) (*service.OpenAIVideoTaskBinding, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" || userID <= 0 {
		return nil, nil
	}

	binding := &service.OpenAIVideoTaskBinding{}
	err := r.db.QueryRowContext(ctx, `
			SELECT id, group_id, user_id, task_id, account_id, expires_at, created_at, updated_at
		FROM openai_video_task_bindings
		WHERE group_id = $1
			AND user_id = $2
			AND task_id = $3
			AND expires_at > NOW()
	`, groupID, userID, taskID).Scan(
		&binding.ID,
		&binding.GroupID,
		&binding.UserID,
		&binding.TaskID,
		&binding.AccountID,
		&binding.ExpiresAt,
		&binding.CreatedAt,
		&binding.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get openai video task binding: %w", err)
	}
	return binding, nil
}

func (r *openAIVideoTaskBindingRepository) ArmOpenAIVideoTaskCompensation(ctx context.Context, groupID, userID int64, taskID string, apiKey *service.APIKey, nextCheckAt time.Time) error {
	if r == nil || r.db == nil || apiKey == nil || apiKey.ID <= 0 {
		return nil
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE openai_video_task_bindings
		SET api_key_id = $4,
			api_key_quota_limited = $5,
			api_key_rate_limited = $6,
			compensation_status = 'pending',
			next_check_at = $7,
			check_attempts = 0,
			last_check_error = NULL,
			checked_at = NULL,
			updated_at = NOW()
		WHERE group_id = $1 AND user_id = $2 AND task_id = $3
	`, groupID, userID, strings.TrimSpace(taskID), apiKey.ID, apiKey.Quota > 0, apiKey.HasRateLimits(), nextCheckAt)
	if err != nil {
		return fmt.Errorf("arm openai video task compensation: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read openai video compensation update result: %w", err)
	}
	if updated == 0 {
		return fmt.Errorf("openai video task binding not found while arming compensation")
	}
	return nil
}

func (r *openAIVideoTaskBindingRepository) ClaimOpenAIVideoTaskCompensations(ctx context.Context, now, leaseUntil time.Time, limit int) ([]service.OpenAIVideoTaskBinding, error) {
	if r == nil || r.db == nil || limit <= 0 {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		WITH candidates AS (
			SELECT id
			FROM openai_video_task_bindings AS task
			WHERE compensation_status = 'pending'
				AND next_check_at <= $1
				AND expires_at > $1
				AND NOT EXISTS (
					SELECT 1 FROM openai_video_task_bindings AS recovery
					WHERE recovery.user_id = task.user_id
						AND recovery.account_id = task.account_id
						AND recovery.billing_task_id = task.billing_task_id
						AND recovery.recovery_status IN ('submitting', 'pending', 'identified', 'billing_review')
				)
			ORDER BY next_check_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $2
		), claimed AS (
			UPDATE openai_video_task_bindings AS binding
			SET next_check_at = $3,
				check_attempts = binding.check_attempts + 1,
				checked_at = $1,
				updated_at = NOW()
			FROM candidates
			WHERE binding.id = candidates.id
			RETURNING binding.*
		)
		SELECT claimed.id, claimed.group_id, claimed.user_id, claimed.task_id,
			claimed.account_id, COALESCE(claimed.api_key_id, 0), COALESCE(api_keys.key, ''),
			claimed.api_key_quota_limited, claimed.api_key_rate_limited,
			claimed.compensation_status, claimed.next_check_at, claimed.check_attempts,
			claimed.last_check_error, claimed.checked_at, claimed.expires_at,
			claimed.created_at, claimed.updated_at,
			claimed.upstream_task_id, claimed.billing_task_id
		FROM claimed
		LEFT JOIN api_keys ON api_keys.id = claimed.api_key_id
	`, now, limit, leaseUntil)
	if err != nil {
		return nil, fmt.Errorf("claim openai video task compensations: %w", err)
	}
	defer rows.Close()

	bindings := make([]service.OpenAIVideoTaskBinding, 0, limit)
	for rows.Next() {
		var binding service.OpenAIVideoTaskBinding
		var nextCheckAt sql.NullTime
		var lastCheckError sql.NullString
		var checkedAt sql.NullTime
		if err := rows.Scan(
			&binding.ID,
			&binding.GroupID,
			&binding.UserID,
			&binding.TaskID,
			&binding.AccountID,
			&binding.APIKeyID,
			&binding.APIKeyKey,
			&binding.APIKeyQuotaLimited,
			&binding.APIKeyRateLimited,
			&binding.CompensationStatus,
			&nextCheckAt,
			&binding.CheckAttempts,
			&lastCheckError,
			&checkedAt,
			&binding.ExpiresAt,
			&binding.CreatedAt,
			&binding.UpdatedAt,
			&binding.UpstreamTaskID,
			&binding.BillingTaskID,
		); err != nil {
			return nil, fmt.Errorf("scan openai video task compensation: %w", err)
		}
		if nextCheckAt.Valid {
			binding.NextCheckAt = &nextCheckAt.Time
		}
		if lastCheckError.Valid {
			binding.LastCheckError = &lastCheckError.String
		}
		if checkedAt.Valid {
			binding.CheckedAt = &checkedAt.Time
		}
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate openai video task compensations: %w", err)
	}
	return bindings, nil
}

func (r *openAIVideoTaskBindingRepository) UpdateOpenAIVideoTaskCompensation(ctx context.Context, id int64, status string, nextCheckAt *time.Time, lastError string) error {
	if r == nil || r.db == nil || id <= 0 {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE openai_video_task_bindings
		SET compensation_status = $2,
			next_check_at = $3,
			last_check_error = NULLIF($4, ''),
			checked_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND compensation_status = 'pending'
	`, id, strings.TrimSpace(status), nextCheckAt, strings.TrimSpace(lastError))
	if err != nil {
		return fmt.Errorf("update openai video task compensation: %w", err)
	}
	return nil
}
