package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// SunoMusicPricePerRequest is the fallback price of one successful music task
// when no group/channel per-request price is configured.
const SunoMusicPricePerRequest = 0.08

const sunoMaxTextCharacters = 520000
const sunoV6MaxPromptUTF16Units = 3000

var sunoMusicModelIDs = map[string]struct{}{
	"suno":         {},
	"suno-v6-mini": {},
	"suno-v6":      {},
	"suno-v6-wild": {},
}

// IsSunoMusicModel reports whether model is one of the supported Suno music
// models. Keep this list explicit so arbitrary custom models cannot
// accidentally use music validation or billing.
func IsSunoMusicModel(model string) bool {
	_, ok := sunoMusicModelIDs[strings.ToLower(strings.TrimSpace(model))]
	return ok
}

func IsSunoV6MusicModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return model == "suno-v6-mini" || model == "suno-v6" || model == "suno-v6-wild"
}

// ValidateSunoChatRequest validates the intentionally small Suno Chat
// Completions contract. A nil error for other models keeps this helper safe to
// call from the shared OpenAI handler.
func ValidateSunoChatRequest(body []byte) error {
	var request map[string]json.RawMessage
	if err := json.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("invalid JSON request body")
	}
	model, _ := requestString(request, "model")
	if !IsSunoMusicModel(model) {
		return nil
	}

	var messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	rawMessages, ok := request["messages"]
	if !ok || json.Unmarshal(rawMessages, &messages) != nil || len(messages) < 1 || len(messages) > 32 {
		return fmt.Errorf("suno requires 1 to 32 text messages")
	}

	parts := make([]string, 0, len(messages))
	for _, message := range messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "system", "user", "assistant":
		default:
			return fmt.Errorf("suno message role must be system, user, or assistant")
		}
		text, err := sunoMessageText(message.Content)
		if err != nil {
			return err
		}
		parts = append(parts, strings.TrimSpace(text))
	}
	prompt := strings.TrimSpace(strings.Join(parts, "\n"))
	length := utf8.RuneCountInString(prompt)
	maxLength := sunoMaxTextCharacters
	lengthUnit := "characters"
	if IsSunoV6MusicModel(model) {
		length = len(utf16.Encode([]rune(prompt)))
		maxLength = sunoV6MaxPromptUTF16Units
		lengthUnit = "UTF-16 units"
	}
	if length < 1 || length > maxLength {
		return fmt.Errorf("suno prompt must be between 1 and %d %s", maxLength, lengthUnit)
	}

	if rawN, exists := request["n"]; exists {
		var n int
		if json.Unmarshal(rawN, &n) != nil || n != 1 {
			return fmt.Errorf("suno only supports n=1")
		}
	}
	for _, field := range []string{"style", "temperature", "duration", "resolution"} {
		if _, exists := request[field]; exists {
			return fmt.Errorf("suno does not support %s", field)
		}
	}
	if rawInstrumental, exists := request["instrumental"]; exists {
		if !IsSunoV6MusicModel(model) {
			return fmt.Errorf("suno does not support instrumental")
		}
		var instrumental bool
		if json.Unmarshal(rawInstrumental, &instrumental) != nil {
			return fmt.Errorf("suno instrumental must be boolean")
		}
	}
	for _, field := range []string{"lyrics", "style_text"} {
		if IsSunoV6MusicModel(model) {
			if _, exists := request[field]; exists {
				return fmt.Errorf("%s does not support %s", model, field)
			}
			continue
		}
		if err := validateSunoOptionalTextField(request, field); err != nil {
			return err
		}
	}
	if rawStreamOptions, exists := request["stream_options"]; exists {
		var options map[string]json.RawMessage
		if json.Unmarshal(rawStreamOptions, &options) != nil {
			return fmt.Errorf("stream_options must be an object")
		}
		if rawIncludeUsage, exists := options["include_usage"]; exists {
			var includeUsage bool
			if json.Unmarshal(rawIncludeUsage, &includeUsage) != nil {
				return fmt.Errorf("stream_options.include_usage must be boolean")
			}
		}
	}
	return nil
}

func validateSunoOptionalTextField(request map[string]json.RawMessage, field string) error {
	raw, exists := request[field]
	if !exists {
		return nil
	}
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return fmt.Errorf("suno %s must be a string", field)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("suno %s must be a string", field)
	}
	if utf8.RuneCountInString(value) > sunoMaxTextCharacters {
		return fmt.Errorf("suno %s must be at most %d characters", field, sunoMaxTextCharacters)
	}
	return nil
}

func requestString(request map[string]json.RawMessage, key string) (string, bool) {
	raw, ok := request[key]
	if !ok {
		return "", false
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return "", false
	}
	return value, true
}

func sunoMessageText(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", fmt.Errorf("suno messages must contain text content")
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if strings.TrimSpace(text) == "" {
			return "", fmt.Errorf("suno messages must contain text content")
		}
		return text, nil
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil || len(blocks) == 0 {
		return "", fmt.Errorf("suno only supports plain text messages")
	}
	var texts []string
	for _, block := range blocks {
		if !strings.EqualFold(strings.TrimSpace(block.Type), "text") || strings.TrimSpace(block.Text) == "" {
			return "", fmt.Errorf("suno only supports plain text messages")
		}
		texts = append(texts, block.Text)
	}
	return strings.Join(texts, "\n"), nil
}

func sunoMusicCost(rateMultiplier float64) *CostBreakdown {
	return sunoMusicCostAtPrice(SunoMusicPricePerRequest, rateMultiplier)
}

func sunoMusicCostAtPrice(unitPrice, rateMultiplier float64) *CostBreakdown {
	if rateMultiplier < 0 {
		rateMultiplier = 0
	}
	return &CostBreakdown{
		TotalCost:   unitPrice,
		ActualCost:  unitPrice * rateMultiplier,
		BillingMode: string(BillingModePerRequest),
	}
}

func isSunoResult(result *OpenAIForwardResult) bool {
	if result == nil {
		return false
	}
	return IsSunoMusicModel(result.Model) || IsSunoMusicModel(result.BillingModel) || IsSunoMusicModel(result.UpstreamModel)
}

// markSunoForwardResult adds the request-count marker used by usage metrics and
// fixed per-request billing without changing the upstream response body.
func markSunoForwardResult(result *OpenAIForwardResult) {
	if !isSunoResult(result) {
		return
	}
	result.RequestCount = 1
	result.MediaType = "audio"
}
