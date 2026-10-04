package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *usageBillingRepository) SavePendingImageSettlement(ctx context.Context, pending *service.PendingImageSettlement) error {
	// Persist billing fields only, never API keys, credentials, or user objects.
	snapshot := *pending
	usage := *pending.Usage
	usage.User, usage.APIKey, usage.Account, usage.Group, usage.Subscription = nil, nil, nil, nil, nil
	snapshot.Usage = &usage
	body, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	var fingerprint string
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO pending_image_settlements (request_id, api_key_id, request_fingerprint, snapshot, next_check_at)
		VALUES ($1, $2, $3, $4, NOW() + INTERVAL '2 minutes')
		ON CONFLICT (request_id, api_key_id) DO UPDATE
		SET request_id = EXCLUDED.request_id
		RETURNING request_fingerprint
	`, pending.Command.RequestID, pending.Command.APIKeyID, pending.Command.RequestFingerprint, body).Scan(&fingerprint)
	if err == nil && fingerprint != pending.Command.RequestFingerprint {
		return service.ErrUsageBillingRequestConflict
	}
	return err
}

func (r *usageBillingRepository) ClaimPendingImageSettlements(ctx context.Context, now time.Time) ([]service.PendingImageSettlement, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH due AS (
			SELECT request_id, api_key_id FROM pending_image_settlements
			WHERE next_check_at <= $1 ORDER BY next_check_at
			LIMIT 10 FOR UPDATE SKIP LOCKED
		)
		UPDATE pending_image_settlements p
		SET next_check_at = $1 + INTERVAL '5 minutes', attempts = attempts + 1
		FROM due WHERE p.request_id = due.request_id AND p.api_key_id = due.api_key_id
		RETURNING p.snapshot, p.attempts
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []service.PendingImageSettlement
	for rows.Next() {
		var raw []byte
		var item service.PendingImageSettlement
		var attempts int
		if err := rows.Scan(&raw, &attempts); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		item.Attempts = attempts
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *usageBillingRepository) FinishPendingImageSettlement(ctx context.Context, pending *service.PendingImageSettlement, cause error) error {
	cmd := pending.Command
	if cause == nil {
		_, err := r.db.ExecContext(ctx, `DELETE FROM pending_image_settlements
			WHERE request_id = $1 AND api_key_id = $2 AND request_fingerprint = $3`,
			cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint)
		return err
	}
	var next sql.NullTime
	if !errors.Is(cause, service.ErrMediaBalanceHoldInconsistent) && !errors.Is(cause, service.ErrUsageBillingRequestConflict) {
		delay := time.Duration(min(pending.Attempts+1, 20)) * 30 * time.Second
		next = sql.NullTime{Time: time.Now().Add(delay), Valid: true}
	}
	_, err := r.db.ExecContext(ctx, `UPDATE pending_image_settlements
		SET next_check_at = $4, last_error = $5
		WHERE request_id = $1 AND api_key_id = $2 AND request_fingerprint = $3`,
		cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint, next, cause.Error())
	return err
}
