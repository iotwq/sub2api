package service

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const openAIBasispointsCacheCreationAsInputExtraKey = "openai_basispoints_cache_creation_as_input"

// Both the selected mode and the actual endpoint must match: native Codex
// fallbacks and WebSocket requests keep their existing accounting.
func openAIBasispointsCacheCreationAsInput(account *Account, upstreamEndpoint string) bool {
	if upstreamEndpoint != "/basispoints/api/responses" || openAIOAuthResponsesEndpointMode(account) != openAIOAuthResponsesEndpointBasis {
		return false
	}
	enabled, _ := account.Extra[openAIBasispointsCacheCreationAsInputExtraKey].(bool)
	return enabled
}

// Adapted from ranxi2001/sub2api 49bb7b0. Total input already includes cache
// creation. Clear only write counters; preserve total input, reads and output.
// Call after collecting the original upstream usage for accounting.
func normalizeOpenAIBasispointsUsage(payload []byte) ([]byte, error) {
	for _, path := range []string{"usage", "response.usage"} {
		usage := gjson.GetBytes(payload, path)
		if !usage.IsObject() {
			continue
		}
		normalized := []byte(usage.Raw)
		changed := false
		for _, field := range []string{
			"input_tokens_details.cache_write_tokens", "prompt_tokens_details.cache_write_tokens",
			"input_tokens_details.cache_creation_tokens", "prompt_tokens_details.cache_creation_tokens",
			"cache_write_tokens", "cache_creation_input_tokens", "cache_write_input_tokens", "cache_creation_tokens",
			"cache_creation.ephemeral_5m_input_tokens", "cache_creation.ephemeral_1h_input_tokens",
		} {
			if value := usage.Get(field); !value.Exists() || value.Raw == "0" {
				continue
			}
			var err error
			normalized, err = sjson.SetBytes(normalized, field, 0)
			if err != nil {
				return nil, err
			}
			changed = true
		}
		if changed {
			var err error
			payload, err = sjson.SetRawBytes(payload, path, normalized)
			if err != nil {
				return nil, err
			}
		}
	}
	return payload, nil
}
