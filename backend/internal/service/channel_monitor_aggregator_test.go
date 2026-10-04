package service

import "testing"

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
