//go:build integration

package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountCodexSignalPersistenceConcurrentEdits(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := newAccountRepositoryWithSQL(client, integrationDB, nil)
	var recorder service.CodexSignalRepository = repo
	a := mustCreateAccount(t, client, &service.Account{
		Name: "codex-signal", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth,
		Extra: map[string]any{"custom": true, "codex_ticket_enabled": true},
	})
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(ctx, "DELETE FROM scheduler_outbox WHERE account_id = $1", a.ID)
		_, _ = integrationDB.ExecContext(ctx, "DELETE FROM accounts WHERE id = $1", a.ID)
	})
	start := time.Now().UTC()
	stale, err := repo.GetByID(ctx, a.ID)
	require.NoError(t, err)
	var wg sync.WaitGroup
	errs := make(chan error, 21)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ts := start.Add(time.Duration(i) * time.Second)
			signal := service.CodexSignalStatus{Length: 292, ObservedAt: ts}
			if i == 10 {
				signal.Length = 312
				signal.Last312At = &ts
			}
			errs <- recorder.RecordCodexSignal(ctx, a.ID, signal)
		}(i)
	}
	// An administrator saves a stale view while observations are in flight.
	wg.Add(1)
	go func() {
		defer wg.Done()
		stale.Extra = map[string]any{"custom": true, "edited": true, service.CodexSignalExtraKey: map[string]any{"length": 999}}
		errs <- repo.Update(ctx, stale)
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	got, err := repo.GetByID(ctx, a.ID)
	require.NoError(t, err)
	signal := service.CodexSignalFromAccount(got)
	require.NotNil(t, signal)
	require.Equal(t, 292, signal.Length)
	require.True(t, start.Add(19*time.Second).Equal(signal.ObservedAt))
	require.NotNil(t, signal.Last312At)
	require.True(t, start.Add(10*time.Second).Equal(*signal.Last312At))
	require.Equal(t, true, got.Extra["edited"])
	require.Equal(t, true, got.Extra["custom"])
	// A later save must not restore the old summary either.
	stale.Extra = map[string]any{"custom": true}
	require.NoError(t, repo.Update(ctx, stale))
	got, err = repo.GetByID(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, signal, service.CodexSignalFromAccount(got))
}
