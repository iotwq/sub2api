package service

import "strings"

func optionalTrimmedStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

// coalesceRequestedReasoningEffort prefers the client-requested value and falls
// back to the effective/forwarded effort for historical or unmapped rows.
func coalesceRequestedReasoningEffort(requested, forwarded *string) *string {
	if trimmed := optionalStringValue(requested); trimmed != "" {
		return &trimmed
	}
	if trimmed := optionalStringValue(forwarded); trimmed != "" {
		return &trimmed
	}
	return nil
}

func forwardResultBillingModel(requestedModel, upstreamModel string) string {
	if trimmed := strings.TrimSpace(requestedModel); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(upstreamModel)
}

func optionalInt64Ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func OpenAIForwardResultHasBillableUsage(result *OpenAIForwardResult) bool {
	if result == nil {
		return false
	}
	usage := result.Usage
	return usage.InputTokens > 0 || usage.OutputTokens > 0 ||
		usage.CacheReadInputTokens > 0 || usage.CacheCreationInputTokens > 0 ||
		usage.ImageInputTokens > 0 || usage.ImageOutputTokens > 0 ||
		result.ImageCount > 0 || result.VideoCount > 0 ||
		result.WebSearchCalls > 0 || result.RequestCount > 0
}

func ForwardResultHasBillableUsage(result *ForwardResult) bool {
	if result == nil {
		return false
	}
	usage := result.Usage
	return usage.InputTokens > 0 || usage.OutputTokens > 0 ||
		usage.CacheReadInputTokens > 0 || usage.CacheCreationInputTokens > 0 ||
		usage.CacheCreation5mTokens > 0 || usage.CacheCreation1hTokens > 0 ||
		usage.ImageOutputTokens > 0 || result.ImageCount > 0
}
