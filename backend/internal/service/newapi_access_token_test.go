package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type newAPIBalanceAccessTokenRepoStub struct {
	rotatedUserID int64
	rotatedHash   string
	subject       *NewAPIBalanceAccessSubject
	findHash      string
	findErr       error
}

func (r *newAPIBalanceAccessTokenRepoStub) Rotate(_ context.Context, userID int64, tokenHash string) error {
	r.rotatedUserID = userID
	r.rotatedHash = tokenHash
	return nil
}

func (r *newAPIBalanceAccessTokenRepoStub) FindSubjectByHash(_ context.Context, tokenHash string) (*NewAPIBalanceAccessSubject, error) {
	r.findHash = tokenHash
	return r.subject, r.findErr
}

func TestNewAPIBalanceAccessTokenServiceGenerateAndAuthenticate(t *testing.T) {
	repo := &newAPIBalanceAccessTokenRepoStub{}
	svc := NewNewAPIBalanceAccessTokenService(repo)

	token, err := svc.Generate(context.Background(), 31)
	require.NoError(t, err)
	require.True(t, len(token) > len(NewAPIBalanceAccessTokenPrefix)+32)
	require.Contains(t, token, NewAPIBalanceAccessTokenPrefix)
	require.Equal(t, int64(31), repo.rotatedUserID)
	require.Len(t, repo.rotatedHash, 64)
	require.NotContains(t, repo.rotatedHash, token)

	want := &NewAPIBalanceAccessSubject{UserID: 31, Role: RoleUser, Email: "user@example.com", Concurrency: 4}
	repo.subject = want
	got, err := svc.Authenticate(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, want, got)
	require.Equal(t, repo.rotatedHash, repo.findHash)
}

func TestNewAPIBalanceAccessTokenServiceRejectsInvalidToken(t *testing.T) {
	repo := &newAPIBalanceAccessTokenRepoStub{}
	svc := NewNewAPIBalanceAccessTokenService(repo)

	_, err := svc.Authenticate(context.Background(), "not-a-balance-token")
	require.ErrorIs(t, err, ErrNewAPIBalanceAccessTokenInvalid)
	require.Empty(t, repo.findHash)

	repo.findErr = ErrNewAPIBalanceAccessTokenInvalid
	_, err = svc.Authenticate(context.Background(), NewAPIBalanceAccessTokenPrefix+"missing")
	require.ErrorIs(t, err, ErrNewAPIBalanceAccessTokenInvalid)
}
