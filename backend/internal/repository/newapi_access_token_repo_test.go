package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestNewAPIBalanceAccessTokenRepositoryRotateAndFind(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &newAPIBalanceAccessTokenRepository{db: db}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE users
		SET newapi_access_token_hash = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`)).WithArgs("hash", int64(31)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Rotate(context.Background(), 31, "hash"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, role, email, concurrency
		FROM users
		WHERE newapi_access_token_hash = $1
		  AND status = 'active'
		  AND deleted_at IS NULL
	`)).WithArgs("hash").WillReturnRows(sqlmock.NewRows([]string{"id", "role", "email", "concurrency"}).AddRow(31, "user", "user@example.com", 4))
	subject, err := repo.FindSubjectByHash(context.Background(), "hash")
	require.NoError(t, err)
	require.Equal(t, int64(31), subject.UserID)
	require.Equal(t, "user", subject.Role)
	require.Equal(t, "user@example.com", subject.Email)
	require.Equal(t, 4, subject.Concurrency)
	require.NoError(t, mock.ExpectationsWereMet())
}
