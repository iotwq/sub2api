package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildStatusSummaryUsesLatestPrimaryStatus(t *testing.T) {
	latency := 123
	summary := buildStatusSummary(
		map[string]*ChannelMonitorLatest{
			"gpt-test": {Model: "gpt-test", Status: MonitorStatusError, LatencyMs: &latency},
		},
		map[string]*ChannelMonitorAvailability{
			"gpt-test": {Model: "gpt-test", AvailabilityPct: 99},
		},
		"gpt-test",
		nil,
	)

	if summary.PrimaryStatus != MonitorStatusError {
		t.Fatalf("PrimaryStatus = %q, want %q", summary.PrimaryStatus, MonitorStatusError)
	}
	if summary.PrimaryLatencyMs == nil || *summary.PrimaryLatencyMs != latency {
		t.Fatalf("PrimaryLatencyMs = %v, want %d", summary.PrimaryLatencyMs, latency)
	}
	if summary.Availability7d != 99 {
		t.Fatalf("Availability7d = %v, want 99", summary.Availability7d)
	}
}

type userAvailabilityRepo struct {
	ChannelMonitorRepository
	monitors     []*ChannelMonitor
	windows      []int
	latestCalls  int
	failedWindow int
}

func (r *userAvailabilityRepo) ListEnabled(context.Context) ([]*ChannelMonitor, error) {
	return r.monitors, nil
}

func (r *userAvailabilityRepo) ListLatestForMonitorIDs(context.Context, []int64) (map[int64][]*ChannelMonitorLatest, error) {
	r.latestCalls++
	return map[int64][]*ChannelMonitorLatest{1: {{Model: "primary", Status: MonitorStatusOperational}}}, nil
}

func (r *userAvailabilityRepo) ComputeAvailabilityForMonitors(_ context.Context, ids []int64, days int) (map[int64][]*ChannelMonitorAvailability, error) {
	r.windows = append(r.windows, days)
	if days == r.failedWindow {
		return nil, errors.New("unavailable")
	}
	rows := make(map[int64][]*ChannelMonitorAvailability)
	for _, id := range ids {
		rows[id] = []*ChannelMonitorAvailability{
			{Model: "extra", AvailabilityPct: 1},
			{Model: "primary", AvailabilityPct: float64(100 - days)},
		}
	}
	return rows, nil
}

func (r *userAvailabilityRepo) ListRecentHistoryForMonitors(context.Context, []int64, map[int64]string, int) (map[int64][]*ChannelMonitorHistoryEntry, error) {
	return nil, nil
}

func TestListUserViewBatchesAllAvailabilityWindows(t *testing.T) {
	repo := &userAvailabilityRepo{}
	for id := int64(1); id <= 20; id++ {
		repo.monitors = append(repo.monitors, &ChannelMonitor{ID: id, PrimaryModel: "primary", ExtraModels: []string{"extra"}})
	}
	svc := &ChannelMonitorService{repo: repo}
	views, err := svc.ListUserView(context.Background())
	require.NoError(t, err)
	require.Len(t, views, 20)
	require.Equal(t, []int{7, 15, 30}, repo.windows)
	require.Equal(t, 1, repo.latestCalls)
	for _, view := range views {
		require.Equal(t, 93.0, view.Availability7d)
		require.NotNil(t, view.Availability15d)
		require.NotNil(t, view.Availability30d)
		require.Equal(t, 85.0, *view.Availability15d)
		require.Equal(t, 70.0, *view.Availability30d)
	}
	require.Equal(t, MonitorStatusOperational, views[0].PrimaryStatus)
}

func TestListUserViewAvailabilityFailureAndNoHistory(t *testing.T) {
	repo := &userAvailabilityRepo{failedWindow: 15, monitors: []*ChannelMonitor{{ID: 1, PrimaryModel: "no-history"}}}
	svc := &ChannelMonitorService{repo: repo}
	views, err := svc.ListUserView(context.Background())
	require.NoError(t, err)
	require.Nil(t, views[0].Availability15d)
	require.NotNil(t, views[0].Availability30d)
	require.Zero(t, *views[0].Availability30d)
}

func TestListUserViewEmptySkipsAggregation(t *testing.T) {
	repo := &userAvailabilityRepo{}
	views, err := (&ChannelMonitorService{repo: repo}).ListUserView(context.Background())
	require.NoError(t, err)
	require.Empty(t, views)
	require.Empty(t, repo.windows)
	require.Zero(t, repo.latestCalls)
}
