package dto

import "github.com/Wei-Shaw/sub2api/internal/domain"

// ChannelMonitorExtraModelStatus 渠道监控附加模型最近一次状态。
// 同时被 admin handler（List 响应）与 user handler（List 响应）复用，
// 字段必须保持一致以保证前端拿到统一结构。
type ChannelMonitorExtraModelStatus struct {
	Intelligence *domain.MonitorIntelligenceResult `json:"intelligence,omitempty"`
	Model        string                            `json:"model"`
	Status       string                            `json:"status"`
	LatencyMs    *int                              `json:"latency_ms"`
}
