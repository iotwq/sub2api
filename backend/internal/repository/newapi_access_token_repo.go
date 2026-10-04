package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type newAPIBalanceAccessTokenRepository struct {
	db *sql.DB
}

func NewNewAPIBalanceAccessTokenRepository(db *sql.DB) service.NewAPIBalanceAccessTokenRepository {
	return &newAPIBalanceAccessTokenRepository{db: db}
}

func (r *newAPIBalanceAccessTokenRepository) Rotate(ctx context.Context, userID int64, tokenHash string) error {
	if r == nil || r.db == nil {
		return errors.New("NewAPI balance access token repository is unavailable")
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET newapi_access_token_hash = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, tokenHash, userID)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated != 1 {
		return service.ErrUserNotFound
	}
	return nil
}

func (r *newAPIBalanceAccessTokenRepository) FindSubjectByHash(ctx context.Context, tokenHash string) (*service.NewAPIBalanceAccessSubject, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("NewAPI balance access token repository is unavailable")
	}
	var subject service.NewAPIBalanceAccessSubject
	err := r.db.QueryRowContext(ctx, `
		SELECT id, role, email, concurrency
		FROM users
		WHERE newapi_access_token_hash = $1
		  AND status = 'active'
		  AND deleted_at IS NULL
	`, tokenHash).Scan(&subject.UserID, &subject.Role, &subject.Email, &subject.Concurrency)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrNewAPIBalanceAccessTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	return &subject, nil
}
