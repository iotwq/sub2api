package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorIntelligenceUserDTO(t *testing.T) {
	result := &domain.MonitorIntelligenceResult{Status: "passed", Answer: "21"}
	item := userMonitorViewToItem(&service.UserMonitorView{
		IntelligenceEnabled: true, Intelligence: result,
		ExtraModels: []service.ExtraModelStatus{{Intelligence: result}},
		Timeline:    []service.UserMonitorTimelinePoint{{Intelligence: result}, {}},
	}, false)
	require.True(t, item.IntelligenceEnabled)
	require.Equal(t, result, item.Intelligence)
	require.Equal(t, result, item.ExtraModels[0].Intelligence)
	require.Equal(t, result, item.Timeline[0].Intelligence)
	require.Nil(t, item.Timeline[1].Intelligence)
	detail := userMonitorDetailToResponse(&service.UserMonitorDetail{IntelligenceEnabled: true, Models: []service.ModelDetail{{Intelligence: result}}})
	require.True(t, detail.IntelligenceEnabled)
	require.Equal(t, result, detail.Models[0].Intelligence)
}
