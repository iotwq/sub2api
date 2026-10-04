package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScheduledAccountTestPanicDoesNotEscapeWorker(t *testing.T) {
	// The missing test service deliberately panics, as a faulty background
	// account test would. Neither this plan nor a later plan may kill the worker.
	runner := &ScheduledTestRunnerService{}
	for _, id := range []int64{1, 2} {
		require.NotPanics(t, func() {
			runner.runOnePlan(context.Background(), &ScheduledTestPlan{ID: id, AccountID: 300})
		})
	}
}
