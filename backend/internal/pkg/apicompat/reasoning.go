package apicompat

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

// NormalizeOpenAIReasoningEffort keeps the accepted OpenAI reasoning effort
// values consistent across chat-completions and responses code paths.
func NormalizeOpenAIReasoningEffort(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}

	value = strings.NewReplacer("-", "", "_", "", " ", "").Replace(value)

	switch value {
	case "none", "minimal":
		return ""
	case "low", "medium", "high", "max":
		return value
	case "xhigh", "extrahigh":
		return "xhigh"
	default:
		return ""
	}
}

// EffectiveReasoningEffort returns the normalized reasoning effort regardless
// of whether the caller used the legacy flat field or the nested reasoning
// object.
func (req *ChatCompletionsRequest) EffectiveReasoningEffort() string {
	if req == nil {
		return ""
	}
	if openai.IsGPT6SolOrLunaModelSpelling(req.Model) {
		raw := strings.TrimSpace(req.ReasoningEffort)
		if raw == "" && req.Reasoning != nil {
			raw = strings.TrimSpace(req.Reasoning.Effort)
		}
		if strings.EqualFold(raw, "none") {
			return "none"
		}
	}

	if effort := NormalizeOpenAIReasoningEffort(req.ReasoningEffort); effort != "" {
		return effort
	}
	if req.Reasoning != nil {
		return NormalizeOpenAIReasoningEffort(req.Reasoning.Effort)
	}
	return ""
}
