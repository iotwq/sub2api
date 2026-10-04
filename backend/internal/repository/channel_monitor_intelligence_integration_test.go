//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorIntelligenceRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := NewChannelMonitorRepository(integrationEntClient, integrationDB)
	m := &service.ChannelMonitor{
		Name: "candy-enabled", Provider: "openai", APIMode: "chat_completions",
		Endpoint: "https://api.example.com", APIKey: "encrypted", PrimaryModel: "main",
		ExtraModels: []string{"extra"}, Enabled: true, IntervalSeconds: 300,
		GroupName: "same-display-name", IntelligenceEnabled: true,
	}
	require.NoError(t, repo.Create(ctx, m))
	t.Cleanup(func() { _ = repo.Delete(ctx, m.ID) })
	other := *m
	other.ID = 0
	other.Name = "candy-disabled"
	other.IntelligenceEnabled = false
	require.NoError(t, repo.Create(ctx, &other))
	t.Cleanup(func() { _ = repo.Delete(ctx, other.ID) })
	loaded, err := repo.GetByID(ctx, m.ID)
	require.NoError(t, err)
	require.True(t, loaded.IntelligenceEnabled)
	loaded.IntelligenceEnabled = false
	require.NoError(t, repo.Update(ctx, loaded))
	loaded, err = repo.GetByID(ctx, m.ID)
	require.NoError(t, err)
	require.False(t, loaded.IntelligenceEnabled)
	loaded.IntelligenceEnabled = true
	require.NoError(t, repo.Update(ctx, loaded))
	otherLoaded, err := repo.GetByID(ctx, other.ID)
	require.NoError(t, err)
	require.False(t, otherLoaded.IntelligenceEnabled, "same display name must not link switches")

	pass := &domain.MonitorIntelligenceResult{Status: "passed", Answer: "21"}
	fail := &domain.MonitorIntelligenceResult{Status: "failed", Reason: "answer_mismatch", Answer: "29"}
	unknown := &domain.MonitorIntelligenceResult{Status: "inconclusive", Reason: "request_failed"}
	now := time.Now().UTC()
	require.NoError(t, repo.InsertHistoryBatch(ctx, []*service.ChannelMonitorHistoryRow{
		{MonitorID: m.ID, Model: "main", Status: "operational", CheckedAt: now.Add(-3 * time.Minute)},
		{MonitorID: m.ID, Model: "main", Status: "error", CheckedAt: now.Add(-2 * time.Minute), Intelligence: unknown},
		{MonitorID: m.ID, Model: "main", Status: "operational", CheckedAt: now.Add(-time.Minute), Intelligence: fail},
		{MonitorID: m.ID, Model: "main", Status: "operational", CheckedAt: now, Intelligence: pass},
		{MonitorID: m.ID, Model: "extra", Status: "operational", CheckedAt: now, Intelligence: fail},
	}))
	history, err := repo.ListHistory(ctx, m.ID, "main", 60)
	require.NoError(t, err)
	require.Len(t, history, 4)
	require.Equal(t, pass, history[0].Intelligence)
	require.Equal(t, fail, history[1].Intelligence)
	require.Equal(t, unknown, history[2].Intelligence)
	require.Nil(t, history[3].Intelligence)
	latest, err := repo.ListLatestPerModel(ctx, m.ID)
	require.NoError(t, err)
	require.Len(t, latest, 2)
	for _, row := range latest {
		if row.Model == "main" {
			require.Equal(t, pass, row.Intelligence)
		} else {
			require.Equal(t, fail, row.Intelligence)
		}
	}
	batch, err := repo.ListLatestForMonitorIDs(ctx, []int64{m.ID, other.ID})
	require.NoError(t, err)
	require.Len(t, batch[m.ID], 2)
	require.Empty(t, batch[other.ID])
	timeline, err := repo.ListRecentHistoryForMonitors(ctx, []int64{m.ID}, map[int64]string{m.ID: "main"}, 60)
	require.NoError(t, err)
	require.Len(t, timeline[m.ID], 4)
	require.Equal(t, pass, timeline[m.ID][0].Intelligence)
	require.Equal(t, fail, timeline[m.ID][1].Intelligence)
	require.Equal(t, unknown, timeline[m.ID][2].Intelligence)
	require.Nil(t, timeline[m.ID][3].Intelligence)
	stats, err := repo.ComputeAvailability(ctx, m.ID, 7)
	require.NoError(t, err)
	for _, row := range stats {
		if row.Model == "main" {
			require.Equal(t, 75.0, row.AvailabilityPct, "wrong candy answer must not count as a channel outage")
		}
	}
}
