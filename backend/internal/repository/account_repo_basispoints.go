package repository

import (
	"context"
	"encoding/json"
	"errors"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

var _ service.AccountBasispointsRepository = (*accountRepository)(nil)

// DisableBasispointsOn403 changes only the protocol switch. A stale request cannot
// disable an account whose credentials or opt-in have since been changed.
func (r *accountRepository) DisableBasispointsOn403(ctx context.Context, account *service.Account) (bool, error) {
	if !account.IsBasispointsAutoDisableOn403Enabled() {
		return false, nil
	}
	if dbent.TxFromContext(ctx) != nil {
		return r.disableBasispointsOn403InTx(ctx, account)
	}
	tx, err := r.client.Tx(ctx)
	if errors.Is(err, dbent.ErrTxStarted) {
		return r.disableBasispointsOn403InTx(ctx, account)
	}
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	changed, err := r.disableBasispointsOn403InTx(dbent.NewTxContext(ctx, tx), account)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	if changed {
		r.syncSchedulerAccountSnapshot(ctx, account.ID)
	}
	return changed, nil
}

func (r *accountRepository) disableBasispointsOn403InTx(ctx context.Context, account *service.Account) (bool, error) {
	credentials, err := json.Marshal(account.Credentials)
	if err != nil {
		return false, err
	}
	client := clientFromContext(ctx, r.client)
	result, err := client.ExecContext(ctx, `
UPDATE accounts
SET extra = jsonb_set(extra, '{openai_oauth_responses_endpoint}', '"chatgpt_codex"'::jsonb) || jsonb_build_object('openai_basispoints_403_disabled_at', to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND parent_account_id IS NULL
  AND platform = 'openai' AND type IN ('oauth', 'setup-token')
  AND credentials = $2::jsonb
  AND lower(btrim(extra ->> 'openai_oauth_responses_endpoint')) = 'basispoints'
  AND extra -> 'openai_basispoints_auto_disable_on_403' = 'true'::jsonb`,
		account.ID, string(credentials))
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return false, err
	}
	if err := enqueueSchedulerOutbox(ctx, client, service.SchedulerOutboxEventAccountChanged, &account.ID, nil, nil); err != nil {
		return false, err
	}
	return true, nil
}
