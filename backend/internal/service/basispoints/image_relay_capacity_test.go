package basispoints

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageRelayCapacityUsesActualBodyAndConfiguredSlots(t *testing.T) {
	r, err := newTestImageRelay(t, "https://relay.example")
	require.NoError(t, err)
	ctx := context.Background()
	for _, slots := range []int{1, 128, 512} {
		r.SetAdmissionLimits(128, 1024, slots)
		var releases []func()
		for i := 0; i < slots; i++ {
			release, err := r.AdmitRequest(ctx, 200) // Small bodies do not reserve their maximum limit.
			require.NoError(t, err)
			releases = append(releases, release)
		}
		_, err := r.AdmitRequest(ctx, 1)
		require.ErrorIs(t, err, ErrImageRelayBusy)
		for _, release := range releases {
			release()
		}
		require.Zero(t, r.requestBytes)
	}
	r.SetAdmissionLimits(128, 1024, 128)
	release, err := r.AdmitRequest(ctx, 128<<20)
	require.NoError(t, err)
	_, err = r.AdmitRequest(ctx, 1)
	require.ErrorIs(t, err, ErrImageRelayBusy)
	// Reducing limits does not evict in-flight work or lose its reservation.
	r.SetAdmissionLimits(32, 512, 1)
	_, err = r.AdmitRequest(ctx, 33<<20)
	require.ErrorIs(t, err, ErrImageRelayRequestTooLarge)
	release()
	release, err = r.AdmitRequest(ctx, 32<<20)
	require.NoError(t, err)
	release()
}
