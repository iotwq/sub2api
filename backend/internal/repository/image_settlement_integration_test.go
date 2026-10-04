//go:build integration

package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestImageSettlement_OverEstimateReleasedHoldAndRetry(t *testing.T) {
	for _, released := range []bool{false, true} {
		name := "active_hold"
		if released {
			name = "previously_released_hold"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			client := testEntClient(t)
			repo := &usageBillingRepository{db: integrationDB}
			uid := uuid.NewString()
			user := mustCreateUser(t, client, &service.User{Email: uid + "@example.com", PasswordHash: "hash", Balance: 3})
			key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + uid, Name: "image-settlement", Quota: 100})
			account := mustCreateAccount(t, client, &service.Account{Name: uid, Type: service.AccountTypeAPIKey})
			hold := service.BatchImageBalanceHoldCommand{
				RequestID: service.BatchImageHoldRequestID(uid), APIKeyID: key.ID, UserID: user.ID,
				BatchID: uid, HoldAmount: 1, ActualAmount: 1, CompletedMedia: true,
			}
			_, err := repo.ReserveBatchImageBalance(ctx, &hold)
			require.NoError(t, err)
			other := hold
			other.BatchID, other.RequestID, other.RequestFingerprint = uid+"-other", service.BatchImageHoldRequestID(uid+"-other"), ""
			_, err = repo.ReserveBatchImageBalance(ctx, &other)
			require.NoError(t, err)
			release := hold
			release.RequestID, release.RequestFingerprint, release.ActualAmount = service.BatchImageReleaseRequestID(uid), "", 0
			if released {
				_, err = repo.ReleaseBatchImageBalance(ctx, &release)
				require.NoError(t, err)
			}
			capture := hold
			capture.RequestID, capture.RequestFingerprint, capture.ActualAmount = service.BatchImageCaptureRequestID(uid), "", 4
			cmd := &service.UsageBillingCommand{RequestID: uid, APIKeyID: key.ID, UserID: user.ID, AccountID: account.ID,
				AccountType: account.Type, CapturedBalanceHold: &capture, APIKeyQuotaCost: 4, ImageCount: 1}
			cmd.Normalize()
			usage := &service.UsageLog{RequestID: uid, UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID,
				Model: "gpt-image-2", ImageCount: 1, TotalCost: 4, ActualCost: 4, CreatedAt: time.Now()}
			// Relation objects must never leak credentials into the retry snapshot.
			usage.APIKey = &service.APIKey{Key: "MUST_NOT_PERSIST"}
			pending := &service.PendingImageSettlement{Command: cmd, Usage: usage}
			require.NoError(t, repo.SavePendingImageSettlement(ctx, pending))
			var leaked bool
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT snapshot::text LIKE '%MUST_NOT_PERSIST%' FROM pending_image_settlements WHERE request_id=$1`, uid).Scan(&leaked))
			require.False(t, leaked)
			// Failed usage insertion rolls back capture; the durable original survives.
			invalidUsage := *usage
			invalidUsage.AccountID += 999999999
			_, err = repo.ApplyWithUsageLog(ctx, cmd, &invalidUsage)
			require.Error(t, err)
			require.NoError(t, repo.FinishPendingImageSettlement(ctx, pending, err))
			items, err := repo.ClaimPendingImageSettlements(ctx, time.Now().Add(time.Minute))
			require.NoError(t, err)
			require.Len(t, items, 1)
			require.Equal(t, 4.0, items[0].Command.CapturedBalanceHold.ActualAmount)
			// Restart/repeated workers replay the saved command, without double debit.
			var wg sync.WaitGroup
			errs := make(chan error, 6)
			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c, h, u := *items[0].Command, *items[0].Command.CapturedBalanceHold, *items[0].Usage
					c.CapturedBalanceHold = &h
					_, err := repo.ApplyWithUsageLog(ctx, &c, &u)
					errs <- err
				}()
			}
			wg.Wait()
			close(errs)
			for err := range errs {
				require.NoError(t, err)
			}
			require.NoError(t, repo.FinishPendingImageSettlement(ctx, &items[0], nil))
			// A late handler release must leave the unrelated reservation intact.
			_, err = repo.ReleaseBatchImageBalance(ctx, &release)
			require.NoError(t, err)
			var balance, frozen, quota, total, ledger float64
			var count, pendingCount int
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT balance, frozen_balance FROM users WHERE id=$1`, user.ID).Scan(&balance, &frozen))
			require.Equal(t, -2.0, balance)
			require.Equal(t, 1.0, frozen)
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT quota_used FROM api_keys WHERE id=$1`, key.ID).Scan(&quota))
			require.Equal(t, 4.0, quota)
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT count(*), sum(u.actual_cost), sum(b.delta_usd) FROM usage_logs u JOIN billing_usage_entries b ON b.usage_log_id=u.id WHERE u.request_id=$1`, uid).Scan(&count, &total, &ledger))
			require.Equal(t, 1, count)
			require.Equal(t, 4.0, total)
			require.Equal(t, total, ledger)
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT count(*) FROM pending_image_settlements WHERE request_id=$1`, uid).Scan(&pendingCount))
			require.Zero(t, pendingCount)
			// A different usage ID cannot charge/log against an already captured hold.
			reusedCmd, reusedUsage := *cmd, *usage
			reusedCmd.RequestID, reusedCmd.RequestFingerprint = uid+"-reused", ""
			reusedUsage.RequestID, reusedUsage.ID = reusedCmd.RequestID, 0
			_, err = repo.ApplyWithUsageLog(ctx, &reusedCmd, &reusedUsage)
			require.ErrorIs(t, err, service.ErrUsageBillingRequestConflict)
		})
	}
}
