package admin

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorIntelligenceDTO(t *testing.T) {
	var create channelMonitorCreateRequest
	require.NoError(t, json.Unmarshal([]byte(`{"intelligence_enabled":true}`), &create))
	require.True(t, create.IntelligenceEnabled)
	var update channelMonitorUpdateRequest
	require.NoError(t, json.Unmarshal([]byte(`{"intelligence_enabled":false}`), &update))
	require.NotNil(t, update.IntelligenceEnabled)
	require.False(t, *update.IntelligenceEnabled)
	result := &domain.MonitorIntelligenceResult{Status: "failed", Reason: "answer_mismatch"}
	response := buildListItemResponse(&service.ChannelMonitor{IntelligenceEnabled: true}, service.MonitorStatusSummary{
		Intelligence: result, ExtraModels: []service.ExtraModelStatus{{Model: "extra", Intelligence: result}},
	})
	require.True(t, response.IntelligenceEnabled)
	require.Equal(t, result, response.Intelligence)
	require.Equal(t, result, response.ExtraModelsStatus[0].Intelligence)
	require.Equal(t, result, checkResultToResponse(&service.CheckResult{Intelligence: result}).Intelligence)
	require.Equal(t, result, historyEntryToResponse(&service.ChannelMonitorHistoryEntry{Intelligence: result}).Intelligence)
}
