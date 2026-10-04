package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeBasispoints403Marker(t *testing.T) {
	const at = "2026-09-26T15:04:05Z"
	stored := map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointChatGPT, Basispoints403DisabledAtKey: at}
	for _, tc := range []struct {
		name    string
		extra   map[string]any
		current map[string]any
		want    map[string]any
	}{
		{
			name:    "ordinary edit keeps stored marker",
			extra:   map[string]any{"openai_passthrough": true},
			current: stored,
			want:    map[string]any{"openai_passthrough": true, Basispoints403DisabledAtKey: at},
		},
		{
			name:    "explicitly disabled protocol keeps marker",
			extra:   map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointChatGPT},
			current: stored,
			want:    map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointChatGPT, Basispoints403DisabledAtKey: at},
		},
		{
			name:    "edit cannot replace marker",
			extra:   map[string]any{Basispoints403DisabledAtKey: "2000-01-01T00:00:00Z"},
			current: stored,
			want:    map[string]any{Basispoints403DisabledAtKey: at},
		},
		{
			name:    "edit cannot add marker",
			extra:   map[string]any{Basispoints403DisabledAtKey: at},
			current: map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointChatGPT},
			want:    map[string]any{},
		},
		{
			name:    "re-enabling protocol clears marker",
			extra:   map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis, Basispoints403DisabledAtKey: at},
			current: stored,
			want:    map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
		},
		{
			name:    "non-boolean switch does not clear marker",
			extra:   map[string]any{"openai_basispoints": "true"},
			current: stored,
			want:    map[string]any{"openai_basispoints": "true", Basispoints403DisabledAtKey: at},
		},
		{
			name:    "nil edit keeps marker",
			current: stored,
			want:    map[string]any{Basispoints403DisabledAtKey: at},
		},
		{
			name: "nil maps",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, MergeBasispoints403Marker(tc.extra, tc.current))
		})
	}
}

func TestMergeBasispoints403MoveMarker(t *testing.T) {
	const at = "2026-09-27T01:02:03Z"
	stored := map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis, Basispoints403MovedAtKey: at, Basispoints403MovedGroupIDKey: float64(7)}
	for _, tc := range []struct {
		name    string
		extra   map[string]any
		current map[string]any
		want    map[string]any
	}{
		{
			name:    "ordinary edit keeps group action",
			extra:   map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
			current: stored,
			want:    map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis, Basispoints403MovedAtKey: at, Basispoints403MovedGroupIDKey: float64(7)},
		},
		{
			name:    "edit cannot rewrite group action",
			extra:   map[string]any{Basispoints403MovedAtKey: "2000-01-01T00:00:00Z", Basispoints403MovedGroupIDKey: float64(9)},
			current: stored,
			want:    map[string]any{Basispoints403MovedAtKey: at, Basispoints403MovedGroupIDKey: float64(7)},
		},
		{
			name:    "edit cannot add group action",
			extra:   map[string]any{Basispoints403MovedAtKey: at, Basispoints403MovedGroupIDKey: float64(0)},
			current: map[string]any{},
			want:    map[string]any{},
		},
		{
			name:    "re-enabling clears only the shutdown record",
			extra:   map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
			current: map[string]any{Basispoints403DisabledAtKey: at, Basispoints403MovedAtKey: at, Basispoints403MovedGroupIDKey: float64(0)},
			want:    map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis, Basispoints403MovedAtKey: at, Basispoints403MovedGroupIDKey: float64(0)},
		},
		{
			name:    "nil edit keeps group action",
			current: stored,
			want:    map[string]any{Basispoints403MovedAtKey: at, Basispoints403MovedGroupIDKey: float64(7)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, MergeBasispoints403Marker(tc.extra, tc.current))
		})
	}
}
