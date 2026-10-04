package service

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// These upstreams use the ordinary asynchronous /v1/videos JSON protocol.
// Resolution is tied to the model; reference media remain upstream-owned fields.
type viraleeVideoSpec struct {
	resolution     string
	minSeconds     int
	maxSeconds     int
	defaultSeconds int // zero means duration is required
}

func lookupViraleeVideoSpec(model string) (viraleeVideoSpec, bool) {
	switch normalizeOpenAIVideoModel(model) {
	case "wan3.0x":
		return viraleeVideoSpec{VideoBillingResolution720P, 2, 30, 0}, true
	case "wan3.0x-480p":
		return viraleeVideoSpec{VideoBillingResolution480P, 2, 30, 0}, true
	case "wan3.0x-1080p":
		return viraleeVideoSpec{VideoBillingResolution1080P, 2, 30, 0}, true
	case "viraldance933", "viraldance933-fast", "dola-viraldance2.0":
		return viraleeVideoSpec{VideoBillingResolution720P, 4, 15, 0}, true
	case "dola-viraldance2.5":
		return viraleeVideoSpec{VideoBillingResolution720P, 4, 30, 0}, true
	case "viraldance2.5-30":
		return viraleeVideoSpec{VideoBillingResolution720P, 4, 30, 5}, true
	case "viraldance2.5-15":
		return viraleeVideoSpec{VideoBillingResolution720P, 4, 15, 15}, true
	case "viraldance2.5-480p-15":
		return viraleeVideoSpec{VideoBillingResolution480P, 4, 15, 15}, true
	default:
		return viraleeVideoSpec{}, false
	}
}

func validateViraleeVideoDuration(model string, body []byte) error {
	spec, ok := lookupViraleeVideoSpec(model)
	if !ok {
		return nil
	}
	value := gjson.GetBytes(body, "duration")
	if !value.Exists() && spec.defaultSeconds > 0 {
		return nil
	}
	// Usage records store whole seconds. Reject fractional values before creating
	// a paid task instead of forwarding them while silently truncating the charge.
	seconds := value.Float()
	if value.Type != gjson.Number || seconds < float64(spec.minSeconds) || seconds > float64(spec.maxSeconds) || seconds != float64(value.Int()) {
		return fmt.Errorf("%s requires an integer duration from %d to %d seconds", model, spec.minSeconds, spec.maxSeconds)
	}
	return nil
}
