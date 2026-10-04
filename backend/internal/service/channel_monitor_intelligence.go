package service

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/tidwall/gjson"
)

// Community candy question: https://github.com/haowang02/codex-candy-eval
// Shapes can be selected by touch; adaptive selection guarantees a pair in 21.
// See docs/channel-monitor-intelligence.md for the exhaustive lower-bound proof.
const monitorCandyPrompt = `不使用任何外部工具回答以下问题：

在一个黑色的袋子里放有三种口味的糖果，每种糖果有两种不同的形状（圆形和五角星形，不同的形状靠手感可以分辨）。现已知不同口味的糖和不同形状的数量统计如下表。参赛者需要在活动前决定摸出的糖果数目，那么，最少取出多少个糖果才能保证手中同时拥有不同形状的苹果味和桃子味的糖？（同时手中有圆形苹果味匹配五角星桃子味糖果，或者有圆形桃子味匹配五角星苹果味糖果都满足要求）

        苹果味  桃子味  西瓜味
圆形       7      9      8
五角星形   7      6      4

操作规则：你可以在取出前凭手感选择圆形或五角星形，但不能在取出前辨别口味。触摸但未取出的糖果不计入数量，取出的糖果不能放回。你可以事先安排两种形状各取多少颗，求无论取到什么口味，都能保证满足上述配对要求的最少总数。

可以简短解释。最后一行必须用 <answer>整数</answer> 给出唯一的最终答案。`

const monitorIntelligenceMaxTokens = 8192

var monitorCandyAnswer = regexp.MustCompile(`(?s)<answer>\s*([0-9]+)\s*</answer>\s*$`)
var monitorCandyBareAnswer = regexp.MustCompile(`^\s*([0-9]+)\s*(?:个|颗)?\s*[。.!！]?\s*$`)

func validateMonitorIntelligence(enabled bool, checkMode, bodyMode string) error {
	if enabled && (defaultCheckMode(checkMode) == MonitorCheckModeQuota || defaultBodyMode(bodyMode) == MonitorBodyOverrideModeReplace) {
		return infraerrors.BadRequest("MONITOR_INTELLIGENCE_INCOMPATIBLE", "intelligence check requires probe or quota_probe mode and an off/merge request body")
	}
	return nil
}

func monitorIntelligenceEnabled(opts *CheckOptions) bool {
	return opts != nil && opts.IntelligenceEnabled
}

func judgeMonitorCandy(text, rawBody string) *domain.MonitorIntelligenceResult {
	result := &domain.MonitorIntelligenceResult{Status: "inconclusive"}
	if !gjson.Valid(rawBody) || monitorIntelligenceResponseIncomplete(rawBody) {
		result.Reason = "response_incomplete"
		return result
	}
	text = strings.TrimSpace(text)
	if text == "" {
		result.Reason = "empty_response"
		return result
	}
	answer := monitorCandyAnswer.FindStringSubmatch(text)
	if len(answer) == 0 {
		answer = monitorCandyBareAnswer.FindStringSubmatch(strings.Trim(text, "*` \n"))
	}
	if len(answer) == 0 || strings.Count(text, "<answer>") > 1 {
		result.Reason = "missing_final_answer"
		return result
	}
	result.Answer = answer[1]
	result.Status = "failed"
	result.Reason = "answer_mismatch"
	if answer[1] == "21" {
		result.Status = "passed"
		result.Reason = ""
	}
	return result
}

func monitorIntelligenceResponseIncomplete(body string) bool {
	for _, path := range []string{"choices.0.finish_reason", "stop_reason", "candidates.0.finishReason", "status"} {
		switch strings.ToLower(gjson.Get(body, path).String()) {
		case "length", "max_tokens", "max_output_tokens", "incomplete", "failed", "content_filter", "safety", "recitation", "tool_calls", "tool_use":
			return true
		}
	}
	return gjson.Get(body, "error").Type != gjson.Null || gjson.Get(body, "promptFeedback.blockReason").String() != ""
}

// Apply after body merging so a legacy 50-token template cannot truncate the
// candy test and OpenAI reasoning effort is always low. Do not mutate the snapshot.
func prepareMonitorIntelligenceBody(body []byte, provider, apiMode string) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	switch {
	case provider == MonitorProviderGemini:
		config, _ := payload["generationConfig"].(map[string]any)
		if config == nil {
			config = map[string]any{}
		}
		config["maxOutputTokens"] = monitorIntelligenceMaxTokens
		payload["generationConfig"] = config
	case provider == MonitorProviderOpenAI && defaultAPIMode(apiMode) == MonitorAPIModeResponses:
		payload["max_output_tokens"] = monitorIntelligenceMaxTokens
		payload["instructions"] = "Solve the user's reasoning question. Give one final integer inside <answer></answer>."
		reasoning, _ := payload["reasoning"].(map[string]any)
		if reasoning == nil {
			reasoning = map[string]any{}
		}
		reasoning["effort"] = "low"
		payload["reasoning"] = reasoning
		delete(payload, "reasoning_effort")
	default:
		if _, ok := payload["max_completion_tokens"]; ok && provider != MonitorProviderAnthropic {
			delete(payload, "max_tokens")
			payload["max_completion_tokens"] = monitorIntelligenceMaxTokens
		} else {
			payload["max_tokens"] = monitorIntelligenceMaxTokens
		}
		if provider == MonitorProviderOpenAI {
			payload["reasoning_effort"] = "low"
			delete(payload, "reasoning")
		}
	}
	return json.Marshal(payload)
}

func extractGeminiMonitorText(respBytes []byte) string {
	var text strings.Builder
	gjson.GetBytes(respBytes, "candidates.0.content.parts").ForEach(func(_, part gjson.Result) bool {
		if !part.Get("thought").Bool() {
			text.WriteString(part.Get("text").String())
		}
		return true
	})
	return text.String()
}
