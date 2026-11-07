package apicompat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGPT61TerraChatPreservesDisabledReasoning(t *testing.T) {
	for _, nested := range []bool{false, true} {
		name := "flat"
		if nested {
			name = "nested"
		}
		t.Run(name, func(t *testing.T) {
			req := &ChatCompletionsRequest{Model: "gpt-6.1-terra"}
			if nested {
				req.Reasoning = &ChatCompletionsReasoning{Effort: "none"}
			} else {
				req.ReasoningEffort = "none"
			}
			out, err := ChatCompletionsToResponses(req)
			require.NoError(t, err)
			require.NotNil(t, out.Reasoning, "omitting none would restore upstream default reasoning")
			require.Equal(t, "none", out.Reasoning.Effort)
		})
	}
}
