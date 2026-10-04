package basispoints

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageRelayAdmissionLimitsAndCancellation(t *testing.T) {
	r, err := newTestImageRelay(t, "https://relay.example")
	require.NoError(t, err)
	_, err = r.AdmitRequest(context.Background(), (64<<20)+1)
	require.ErrorIs(t, err, ErrImageRelayRequestTooLarge)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = r.AdmitRequest(ctx, 1)
	require.ErrorIs(t, err, context.Canceled)
	release, err := r.AdmitRequest(context.Background(), 64<<20)
	require.NoError(t, err)
	releaseSecond, err := r.AdmitRequest(context.Background(), 64<<20)
	require.NoError(t, err)
	_, err = r.AdmitRequest(context.Background(), 1)
	require.ErrorIs(t, err, ErrImageRelayBusy)
	release()
	release() // Idempotent release cannot undercount active requests.
	releaseSecond()
	var releases []func()
	for i := 0; i < 128; i++ {
		done, err := r.AdmitRequest(context.Background(), 1)
		require.NoError(t, err)
		releases = append(releases, done)
	}
	_, err = r.AdmitRequest(context.Background(), 1)
	require.ErrorIs(t, err, ErrImageRelayBusy)
	var wg sync.WaitGroup
	for _, done := range releases {
		wg.Add(1)
		go func() { defer wg.Done(); done() }()
	}
	wg.Wait()
	require.Zero(t, r.requests)
	require.Zero(t, r.requestBytes)
	require.NoError(t, r.Close())
	_, err = r.AdmitRequest(context.Background(), 1)
	require.ErrorIs(t, err, ErrImageRelayStorage)
}
