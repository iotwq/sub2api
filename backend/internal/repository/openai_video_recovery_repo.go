package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *openAIVideoTaskBindingRepository) PrepareOpenAIVideoRecovery(ctx context.Context, binding service.OpenAIVideoTaskBinding) error {
	if r == nil || r.db == nil {
		return errors.New("openai video recovery repository is unavailable")
	}
	if binding.UserID <= 0 || binding.AccountID <= 0 || binding.APIKeyID <= 0 || strings.TrimSpace(binding.TaskID) == "" {
		return errors.New("openai video recovery identity is incomplete")
	}
	if binding.ExpiresAt.IsZero() || binding.RecoveryNextCheckAt == nil || binding.RecoveryExpiresAt == nil {
		return errors.New("openai video recovery timing is incomplete")
	}
	if !json.Valid(binding.RecoveryBaseline) || !json.Valid(binding.RecoverySignature) || !json.Valid(binding.RecoveryBilling) {
		return errors.New("openai video recovery snapshot is invalid")
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO openai_video_task_bindings (
			group_id, user_id, task_id, account_id, api_key_id,
			api_key_quota_limited, api_key_rate_limited,
			upstream_task_id, billing_task_id, recovery_status,
			recovery_baseline, recovery_signature, recovery_billing,
			recovery_next_check_at, recovery_expires_at, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, '', $3, 'submitting', $8, $9, $10, $11, $12, $13)
		ON CONFLICT (user_id, group_id, task_id) DO NOTHING
	`, binding.GroupID, binding.UserID, strings.TrimSpace(binding.TaskID), binding.AccountID, binding.APIKeyID,
		binding.APIKeyQuotaLimited, binding.APIKeyRateLimited,
		binding.RecoveryBaseline, binding.RecoverySignature, binding.RecoveryBilling,
		binding.RecoveryNextCheckAt, binding.RecoveryExpiresAt, binding.ExpiresAt)
	if err != nil {
		return fmt.Errorf("prepare openai video recovery: %w", err)
	}
	return nil
}

func (r *openAIVideoTaskBindingRepository) GetOpenAIVideoRecovery(ctx context.Context, groupID, userID int64, taskID string) (*service.OpenAIVideoTaskBinding, error) {
	if r == nil || r.db == nil || userID <= 0 || strings.TrimSpace(taskID) == "" {
		return nil, nil
	}
	row := r.db.QueryRowContext(ctx, openAIVideoRecoverySelect+`
		WHERE binding.group_id = $1
			AND binding.user_id = $2
			AND binding.task_id = $3
			AND binding.expires_at > NOW()
	`, groupID, userID, strings.TrimSpace(taskID))
	binding, err := scanOpenAIVideoRecovery(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get openai video recovery: %w", err)
	}
	return binding, nil
}

func (r *openAIVideoTaskBindingRepository) UpsertOpenAIVideoTaskRoute(ctx context.Context, binding service.OpenAIVideoTaskBinding) error {
	if r == nil || r.db == nil {
		return errors.New("openai video task route repository is unavailable")
	}
	taskID := strings.TrimSpace(binding.TaskID)
	upstreamTaskID := strings.TrimSpace(binding.UpstreamTaskID)
	billingTaskID := strings.TrimSpace(binding.BillingTaskID)
	if taskID == "" || binding.UserID <= 0 || binding.AccountID <= 0 || binding.ExpiresAt.IsZero() {
		return errors.New("openai video task route is incomplete")
	}
	if upstreamTaskID == "" {
		upstreamTaskID = taskID
	}
	if billingTaskID == "" {
		billingTaskID = taskID
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO openai_video_task_bindings (
			group_id, user_id, task_id, account_id, upstream_task_id, billing_task_id, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, group_id, task_id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			upstream_task_id = EXCLUDED.upstream_task_id,
			billing_task_id = EXCLUDED.billing_task_id,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
	`, binding.GroupID, binding.UserID, taskID, binding.AccountID, upstreamTaskID, billingTaskID, binding.ExpiresAt)
	if err != nil {
		return fmt.Errorf("upsert openai video task route: %w", err)
	}
	return nil
}

func (r *openAIVideoTaskBindingRepository) UpdateOpenAIVideoRecovery(
	ctx context.Context,
	id int64,
	status string,
	upstreamTaskID string,
	nextCheckAt *time.Time,
	lastError string,
) error {
	if r == nil || r.db == nil || id <= 0 {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE openai_video_task_bindings
		SET recovery_status = $2,
			upstream_task_id = CASE WHEN $3 = '' THEN upstream_task_id ELSE $3 END,
			recovery_next_check_at = $4,
			recovery_last_error = NULLIF($5, ''),
			recovery_checked_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`, id, strings.TrimSpace(status), strings.TrimSpace(upstreamTaskID), nextCheckAt, strings.TrimSpace(lastError))
	if err != nil {
		return fmt.Errorf("update openai video recovery: %w", err)
	}
	return nil
}

func (r *openAIVideoTaskBindingRepository) ClaimOpenAIVideoRecoveries(ctx context.Context, now, leaseUntil time.Time, limit int) ([]service.OpenAIVideoTaskBinding, error) {
	if r == nil || r.db == nil || limit <= 0 {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		WITH candidates AS (
				SELECT id
				FROM openai_video_task_bindings
				WHERE recovery_status IN ('submitting', 'pending', 'identified')
					AND (
						recovery_next_check_at <= $1
						OR (
							recovery_status = 'identified'
							AND upstream_task_id <> ''
							AND recovery_next_check_at IS NULL
						)
					)
			ORDER BY recovery_next_check_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $2
		), claimed AS (
			UPDATE openai_video_task_bindings AS binding
			SET recovery_next_check_at = $3,
				recovery_attempts = binding.recovery_attempts + 1,
				recovery_checked_at = $1,
				updated_at = NOW()
			FROM candidates
			WHERE binding.id = candidates.id
			RETURNING binding.id
		)
		`+openAIVideoRecoverySelect+`
		WHERE binding.id IN (SELECT id FROM claimed)
		ORDER BY binding.id
	`, now, limit, leaseUntil)
	if err != nil {
		return nil, fmt.Errorf("claim openai video recoveries: %w", err)
	}
	defer rows.Close()

	bindings := make([]service.OpenAIVideoTaskBinding, 0, limit)
	for rows.Next() {
		binding, scanErr := scanOpenAIVideoRecovery(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan openai video recovery: %w", scanErr)
		}
		bindings = append(bindings, *binding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate openai video recoveries: %w", err)
	}
	return bindings, nil
}

const openAIVideoRecoverySelect = `
	SELECT binding.id, binding.group_id, binding.user_id, binding.task_id,
		binding.account_id, COALESCE(binding.api_key_id, 0),
		binding.api_key_quota_limited, binding.api_key_rate_limited,
		binding.compensation_status, binding.next_check_at, binding.check_attempts,
		binding.last_check_error, binding.checked_at, binding.expires_at,
		binding.created_at, binding.updated_at,
		binding.upstream_task_id, binding.billing_task_id, binding.recovery_status,
		binding.recovery_baseline, binding.recovery_signature, binding.recovery_billing,
		binding.recovery_next_check_at, binding.recovery_expires_at,
		binding.recovery_attempts, binding.recovery_last_error, binding.recovery_checked_at
	FROM openai_video_task_bindings AS binding
`

type openAIVideoRecoveryScanner interface {
	Scan(dest ...any) error
}

func scanOpenAIVideoRecovery(scanner openAIVideoRecoveryScanner) (*service.OpenAIVideoTaskBinding, error) {
	binding := &service.OpenAIVideoTaskBinding{}
	var nextCheckAt, checkedAt, recoveryNextCheckAt, recoveryExpiresAt, recoveryCheckedAt sql.NullTime
	var lastCheckError, recoveryLastError sql.NullString
	err := scanner.Scan(
		&binding.ID,
		&binding.GroupID,
		&binding.UserID,
		&binding.TaskID,
		&binding.AccountID,
		&binding.APIKeyID,
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
		&binding.RecoveryStatus,
		&binding.RecoveryBaseline,
		&binding.RecoverySignature,
		&binding.RecoveryBilling,
		&recoveryNextCheckAt,
		&recoveryExpiresAt,
		&binding.RecoveryAttempts,
		&recoveryLastError,
		&recoveryCheckedAt,
	)
	if err != nil {
		return nil, err
	}
	if nextCheckAt.Valid {
		binding.NextCheckAt = &nextCheckAt.Time
	}
	if checkedAt.Valid {
		binding.CheckedAt = &checkedAt.Time
	}
	if lastCheckError.Valid {
		binding.LastCheckError = &lastCheckError.String
	}
	if recoveryNextCheckAt.Valid {
		binding.RecoveryNextCheckAt = &recoveryNextCheckAt.Time
	}
	if recoveryExpiresAt.Valid {
		binding.RecoveryExpiresAt = &recoveryExpiresAt.Time
	}
	if recoveryCheckedAt.Valid {
		binding.RecoveryCheckedAt = &recoveryCheckedAt.Time
	}
	if recoveryLastError.Valid {
		binding.RecoveryLastError = &recoveryLastError.String
	}
	return binding, nil
}
