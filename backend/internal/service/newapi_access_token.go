package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

const NewAPIBalanceAccessTokenPrefix = "sub_bal_"

var ErrNewAPIBalanceAccessTokenInvalid = errors.New("invalid NewAPI balance access token")

type NewAPIBalanceAccessSubject struct {
	UserID      int64
	Role        string
	Email       string
	Concurrency int
}

type NewAPIBalanceAccessTokenRepository interface {
	Rotate(ctx context.Context, userID int64, tokenHash string) error
	FindSubjectByHash(ctx context.Context, tokenHash string) (*NewAPIBalanceAccessSubject, error)
}

type NewAPIBalanceAccessTokenService struct {
	repo NewAPIBalanceAccessTokenRepository
}

func NewNewAPIBalanceAccessTokenService(repo NewAPIBalanceAccessTokenRepository) *NewAPIBalanceAccessTokenService {
	return &NewAPIBalanceAccessTokenService{repo: repo}
}

func (s *NewAPIBalanceAccessTokenService) Generate(ctx context.Context, userID int64) (string, error) {
	if s == nil || s.repo == nil || userID <= 0 {
		return "", ErrNewAPIBalanceAccessTokenInvalid
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	token := NewAPIBalanceAccessTokenPrefix + base64.RawURLEncoding.EncodeToString(random)
	if err := s.repo.Rotate(ctx, userID, hashNewAPIBalanceAccessToken(token)); err != nil {
		return "", err
	}
	return token, nil
}

func (s *NewAPIBalanceAccessTokenService) Authenticate(ctx context.Context, token string) (*NewAPIBalanceAccessSubject, error) {
	token = strings.TrimSpace(token)
	if s == nil || s.repo == nil || !strings.HasPrefix(token, NewAPIBalanceAccessTokenPrefix) {
		return nil, ErrNewAPIBalanceAccessTokenInvalid
	}
	subject, err := s.repo.FindSubjectByHash(ctx, hashNewAPIBalanceAccessToken(token))
	if err != nil || subject == nil || subject.UserID <= 0 {
		return nil, ErrNewAPIBalanceAccessTokenInvalid
	}
	return subject, nil
}

func hashNewAPIBalanceAccessToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
