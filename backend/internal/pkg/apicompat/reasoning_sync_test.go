package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatCompletionsToResponsesSyncedReasoning(t *testing.T) {
	for _, tc := range []struct {
		name, model, flat, nested, want string
		keepSampling                    bool
	}{
		{"sol flat none", "gpt-6-sol", "none", "", "none", true},
		{"luna nested none", "gpt-6-luna", "", "none", "none", true},
		{"flat wins", "gpt-6-sol", "max", "none", "max", false},
		{"nested max", "gpt-6-luna", "", "max", "max", false},
		{"claude max", "claude-fable-5-1", "max", "", "max", true},
		{"legacy none", "gpt-5.4", "none", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			temperature := 0.5
			req := &ChatCompletionsRequest{
				Model: tc.model, ReasoningEffort: tc.flat,
				Reasoning:   &ChatCompletionsReasoning{Effort: tc.nested},
				Temperature: &temperature,
				Messages:    []ChatMessage{{Role: "user", Content: json.RawMessage(`"hi"`)}},
			}
			got, err := ChatCompletionsToResponses(req)
			require.NoError(t, err)
			if tc.want == "" {
				require.Nil(t, got.Reasoning)
			} else {
				require.NotNil(t, got.Reasoning)
				require.Equal(t, tc.want, got.Reasoning.Effort)
			}
			require.Equal(t, tc.keepSampling, got.Temperature != nil)
		})
	}
}
