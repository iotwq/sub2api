package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuthSnapshotMergedGroupFieldsRoundtrip(t *testing.T) {
	groupID := int64(1)
	key := &APIKey{
		ID: 1, UserID: 2, GroupID: &groupID, Status: StatusActive,
		User: &User{ID: 2, Status: StatusActive},
		Group: &Group{
			ID: groupID, Status: StatusActive, Platform: PlatformOpenAI,
			RequirePrivacySet: true, GPT61SolSubscriptionOptimizationEnabled: true,
		},
	}
	svc := &APIKeyService{}
	payload, err := json.Marshal(&APIKeyAuthCacheEntry{Snapshot: svc.snapshotFromAPIKey(context.Background(), key)})
	require.NoError(t, err)
	var entry APIKeyAuthCacheEntry
	require.NoError(t, json.Unmarshal(payload, &entry))
	restored, used, err := svc.applyAuthCacheEntry("merged-key", &entry)
	require.NoError(t, err)
	require.True(t, used)
	require.True(t, restored.Group.RequirePrivacySet)
	require.True(t, restored.Group.GPT61SolSubscriptionOptimizationEnabled)

	for _, version := range []int{25, 26, 27} {
		entry.Snapshot.Version = version
		restored, used, err = svc.applyAuthCacheEntry("merged-key", &entry)
		require.NoError(t, err)
		require.False(t, used, "snapshot v%d predates the combined group fields", version)
		require.Nil(t, restored)
	}
}
