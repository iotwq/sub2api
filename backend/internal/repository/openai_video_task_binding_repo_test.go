package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestClaimOpenAIVideoTaskCompensationsUsesSkipLockedLease(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &openAIVideoTaskBindingRepository{db: db}
	now := time.Now().UTC()
	leaseUntil := now.Add(2 * time.Minute)
	query := regexp.QuoteMeta("FOR UPDATE SKIP LOCKED")
	mock.ExpectQuery(query).
		WithArgs(now, 10, leaseUntil).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "group_id", "user_id", "task_id", "account_id", "api_key_id", "key",
			"api_key_quota_limited", "api_key_rate_limited", "compensation_status", "next_check_at",
			"check_attempts", "last_check_error", "checked_at", "expires_at", "created_at", "updated_at",
		}))

	bindings, err := repo.ClaimOpenAIVideoTaskCompensations(context.Background(), now, leaseUntil, 10)

	require.NoError(t, err)
	require.Empty(t, bindings)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimOpenAIVideoRecoveriesUsesSkipLockedLease(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &openAIVideoTaskBindingRepository{db: db}
	now := time.Now().UTC()
	leaseUntil := now.Add(45 * time.Second)
	mock.ExpectQuery(`recovery_status = 'identified'[\s\S]+upstream_task_id <> ''[\s\S]+recovery_next_check_at IS NULL[\s\S]+FOR UPDATE SKIP LOCKED`).
		WithArgs(now, 10, leaseUntil).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "group_id", "user_id", "task_id", "account_id", "api_key_id",
			"api_key_quota_limited", "api_key_rate_limited", "compensation_status", "next_check_at",
			"check_attempts", "last_check_error", "checked_at", "expires_at", "created_at", "updated_at",
			"upstream_task_id", "billing_task_id", "recovery_status", "recovery_baseline", "recovery_signature",
			"recovery_billing", "recovery_next_check_at", "recovery_expires_at", "recovery_attempts",
			"recovery_last_error", "recovery_checked_at",
		}))

	bindings, err := repo.ClaimOpenAIVideoRecoveries(context.Background(), now, leaseUntil, 10)

	require.NoError(t, err)
	require.Empty(t, bindings)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimOpenAIVideoRecoveriesScansPersistedRecoveryAfterRestart(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &openAIVideoTaskBindingRepository{db: db}
	now := time.Now().UTC()
	leaseUntil := now.Add(45 * time.Second)
	expiresAt := now.Add(time.Hour)
	recoveryExpiresAt := now.Add(3 * time.Minute)
	createdAt := now.Add(-time.Minute)
	mock.ExpectQuery(regexp.QuoteMeta("FOR UPDATE SKIP LOCKED")).
		WithArgs(now, 10, leaseUntil).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "group_id", "user_id", "task_id", "account_id", "api_key_id",
			"api_key_quota_limited", "api_key_rate_limited", "compensation_status", "next_check_at",
			"check_attempts", "last_check_error", "checked_at", "expires_at", "created_at", "updated_at",
			"upstream_task_id", "billing_task_id", "recovery_status", "recovery_baseline", "recovery_signature",
			"recovery_billing", "recovery_next_check_at", "recovery_expires_at", "recovery_attempts",
			"recovery_last_error", "recovery_checked_at",
		}).AddRow(
			201, 7, 42, "task_recovery_local", 91, 101,
			true, false, "", nil,
			0, nil, nil, expiresAt, createdAt, createdAt,
			"", "task_recovery_local", "pending", []byte(`["before"]`), []byte(`{"model":"MiniMax-H3"}`),
			[]byte(`{}`), leaseUntil, recoveryExpiresAt, 2,
			nil, now,
		))

	bindings, err := repo.ClaimOpenAIVideoRecoveries(context.Background(), now, leaseUntil, 10)

	require.NoError(t, err)
	require.Len(t, bindings, 1)
	require.Equal(t, "task_recovery_local", bindings[0].TaskID)
	require.Equal(t, "pending", bindings[0].RecoveryStatus)
	require.Equal(t, 2, bindings[0].RecoveryAttempts)
	require.Equal(t, leaseUntil, *bindings[0].RecoveryNextCheckAt)
	require.NoError(t, mock.ExpectationsWereMet())
}
