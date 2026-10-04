package repository

import (
	"context"
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// RecordCodexSignal serializes observations with account edits and other instances.
// It changes only the observation key; it does not suspend the account or alter routing.
func (r *accountRepository) RecordCodexSignal(ctx context.Context, id int64, signal service.CodexSignalStatus) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	client := tx.Client()
	rows, err := client.QueryContext(ctx, `SELECT extra->'codex_signal' FROM accounts
		WHERE id = $1 AND deleted_at IS NULL AND platform = 'openai'
		AND type IN ('oauth', 'setup-token') AND parent_account_id IS NULL FOR NO KEY UPDATE`, id)
	if err != nil {
		return err
	}
	if !rows.Next() {
		err = rows.Err()
		_ = rows.Close()
		return err
	}
	var raw []byte
	err = rows.Scan(&raw)
	_ = rows.Close()
	if err != nil {
		return err
	}
	var previous service.CodexSignalStatus
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &previous); err != nil {
			return err
		}
	}
	merged := service.MergeCodexSignals(previous, signal)
	payload, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	_, err = client.ExecContext(ctx, `UPDATE accounts SET extra =
		COALESCE(extra, '{}'::jsonb) || jsonb_build_object('codex_signal', $1::jsonb),
		updated_at = NOW() WHERE id = $2`, string(payload), id)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	r.syncSchedulerAccountSnapshot(ctx, id)
	return nil
}
