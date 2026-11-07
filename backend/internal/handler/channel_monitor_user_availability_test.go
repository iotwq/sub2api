package handler

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorUserListAvailabilityWindows(t *testing.T) {
	availability := 95.5
	item := userMonitorViewToItem(&service.UserMonitorView{Availability7d: 99, Availability15d: &availability}, false)
	data, err := json.Marshal(item)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, 99.0, decoded["availability_7d"])
	require.Equal(t, 95.5, decoded["availability_15d"])
	require.Contains(t, decoded, "availability_30d")
	require.Nil(t, decoded["availability_30d"])
}
