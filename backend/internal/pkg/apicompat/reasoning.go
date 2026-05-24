package apicompat

import "strings"

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
	case "low", "medium", "high":
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

	if effort := NormalizeOpenAIReasoningEffort(req.ReasoningEffort); effort != "" {
		return effort
	}
	if req.Reasoning != nil {
		return NormalizeOpenAIReasoningEffort(req.Reasoning.Effort)
	}
	return ""
}
