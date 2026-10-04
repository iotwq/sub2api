//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMediaBillingReview_UnexplainedHoldDoesNotDebitOrRefund(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := &usageBillingRepository{db: integrationDB}
	id := uuid.NewString()
	user := mustCreateUser(t, client, &service.User{Email: id + "@example.com", PasswordHash: "hash", Balance: 10})
	key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + id, Name: "review"})
	account := mustCreateAccount(t, client, &service.Account{Name: id, Type: service.AccountTypeAPIKey})
	hold := &service.BatchImageBalanceHoldCommand{RequestID: service.BatchImageHoldRequestID(id), BatchID: id,
		APIKeyID: key.ID, UserID: user.ID, HoldAmount: 2, ActualAmount: 2, CompletedMedia: true}
	_, err := repo.ReserveBatchImageBalance(ctx, hold)
	require.NoError(t, err)
	// Reproduce missing aggregate frozen funds with no matching release evidence.
	_, err = integrationDB.ExecContext(ctx, `UPDATE users SET frozen_balance = 0 WHERE id=$1`, user.ID)
	require.NoError(t, err)
	capture := *hold
	capture.RequestID, capture.RequestFingerprint = service.BatchImageCaptureRequestID(id), ""
	cmd := &service.UsageBillingCommand{RequestID: id, APIKeyID: key.ID, UserID: user.ID, AccountID: account.ID, CapturedBalanceHold: &capture}
	cmd.Normalize()
	usage := &service.UsageLog{RequestID: id, UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID,
		Model: "gpt-image-2", ImageCount: 1, TotalCost: 2, ActualCost: 2, CreatedAt: time.Now()}
	pending := &service.PendingImageSettlement{Command: cmd, Usage: usage}
	require.NoError(t, repo.SavePendingImageSettlement(ctx, pending))
	_, err = repo.ApplyWithUsageLog(ctx, cmd, usage)
	require.ErrorIs(t, err, service.ErrMediaBalanceHoldInconsistent)
	require.NoError(t, repo.FinishPendingImageSettlement(ctx, pending, err))
	var balance, frozen float64
	var review bool
	var count int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT balance,frozen_balance FROM users WHERE id=$1`, user.ID).Scan(&balance, &frozen))
	require.Equal(t, 8.0, balance)
	require.Zero(t, frozen)
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT count(*) FROM usage_logs WHERE request_id=$1`, id).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT next_check_at IS NULL FROM pending_image_settlements WHERE request_id=$1`, id).Scan(&review))
	require.True(t, review)
}

func TestVideoBillingReview_PreservesUniqueOwnerAndPausesRefunds(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	id := uuid.NewString()
	user := mustCreateUser(t, client, &service.User{Email: id + "@example.com", PasswordHash: "hash", Balance: 10})
	key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + id, Name: "video-review"})
	account := mustCreateAccount(t, client, &service.Account{Name: id, Type: service.AccountTypeAPIKey})
	repo := &openAIVideoTaskBindingRepository{db: integrationDB}
	now := time.Now()
	binding := service.OpenAIVideoTaskBinding{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, TaskID: id,
		RecoveryBaseline: []byte(`[]`), RecoverySignature: []byte(`{}`), RecoveryBilling: []byte(`{}`),
		RecoveryNextCheckAt: &now, RecoveryExpiresAt: &now, ExpiresAt: now.Add(time.Hour)}
	require.NoError(t, repo.PrepareOpenAIVideoRecovery(ctx, binding))
	stored, err := repo.GetOpenAIVideoRecovery(ctx, 0, user.ID, id)
	require.NoError(t, err)
	require.NoError(t, repo.UpdateOpenAIVideoRecovery(ctx, stored.ID, service.OpenAIVideoRecoveryBillingReview, "upstream-"+id, nil, "missing hold"))
	// Actual-ID route and local recovery route share the original billing ID.
	require.NoError(t, repo.UpsertOpenAIVideoTaskRoute(ctx, service.OpenAIVideoTaskBinding{
		UserID: user.ID, AccountID: account.ID, TaskID: "upstream-" + id, BillingTaskID: id, ExpiresAt: now.Add(time.Hour)}))
	require.NoError(t, repo.ArmOpenAIVideoTaskCompensation(ctx, 0, user.ID, "upstream-"+id, key, now.Add(-time.Minute)))
	refunds, err := repo.ClaimOpenAIVideoTaskCompensations(ctx, now, now.Add(time.Minute), 10)
	require.NoError(t, err)
	require.Empty(t, refunds, "unsettled tasks must not loop on a missing original charge")
	recoveries, err := repo.ClaimOpenAIVideoRecoveries(ctx, now, now.Add(time.Minute), 10)
	require.NoError(t, err)
	require.Empty(t, recoveries, "billing review requires reconciliation before another attempt")
	binding.TaskID = id + "-different-request"
	require.NoError(t, repo.PrepareOpenAIVideoRecovery(ctx, binding))
	other, err := repo.GetOpenAIVideoRecovery(ctx, 0, user.ID, binding.TaskID)
	require.NoError(t, err)
	require.Error(t, repo.UpdateOpenAIVideoRecovery(ctx, other.ID, service.OpenAIVideoRecoveryIdentified, "upstream-"+id, &now, ""), "a billing review task retains its exclusive upstream binding")
	// Once accounting is reconciled, normal terminal-failure compensation resumes.
	require.NoError(t, repo.UpdateOpenAIVideoRecovery(ctx, stored.ID, service.OpenAIVideoRecoveryMatched, "upstream-"+id, nil, ""))
	refunds, err = repo.ClaimOpenAIVideoTaskCompensations(ctx, now, now.Add(time.Minute), 10)
	require.NoError(t, err)
	require.Len(t, refunds, 1)
	require.Equal(t, "upstream-"+id, refunds[0].TaskID)
}
