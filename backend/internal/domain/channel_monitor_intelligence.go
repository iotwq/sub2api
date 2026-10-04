package domain

// MonitorIntelligenceResult is independent of channel availability. A nil result
// means intelligence checking was disabled (including all historical probes).
type MonitorIntelligenceResult struct {
	Status string `json:"status"` // passed / failed / inconclusive
	Reason string `json:"reason,omitempty"`
	Answer string `json:"answer,omitempty"` // Parsed final number only; never raw reasoning.
}
