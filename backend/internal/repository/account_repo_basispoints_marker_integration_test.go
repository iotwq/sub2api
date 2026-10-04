//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBasispoints403MarkerPersistsAcrossEditsAndClearsOnOptIn(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	repo := newAccountRepositoryWithSQL(tx.Client(), tx, nil)
	account := mustCreateAccount(t, tx.Client(), newBasispointsAutoDisableAccount())
	changed, err := repo.DisableBasispointsOn403(ctx, account)
	require.NoError(t, err)
	require.True(t, changed)
	current, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	stamp, ok := current.Extra[service.Basispoints403DisabledAtKey].(string)
	require.True(t, ok)
	_, err = time.Parse(time.RFC3339, stamp)
	require.NoError(t, err)
	current.Extra[service.Basispoints403DisabledAtKey] = "forged"
	require.NoError(t, repo.Update(ctx, current))
	current, err = repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, stamp, current.Extra[service.Basispoints403DisabledAtKey])
	current.Extra["openai_oauth_responses_endpoint"] = "basispoints"
	require.NoError(t, repo.Update(ctx, current))
	current, err = repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.NotContains(t, current.Extra, service.Basispoints403DisabledAtKey)
}
