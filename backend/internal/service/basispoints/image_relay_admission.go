package basispoints

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrImageRelayBusy            = errors.New("basispoints image relay is busy; retry later")
	ErrImageRelayRequestTooLarge = errors.New("basispoints image relay request exceeds the configured body limit")
)

type imageRelayAdmissionLimits struct {
	bodyBytes   int64
	budgetBytes int64
	maxRequests int
}

// ImageAdmission bounds preparation memory for both public relay and native
// attachment modes. Its zero value is ready to use with the default limits.
type ImageAdmission struct {
	mu              sync.Mutex
	closed          bool
	requests        int
	requestBytes    int64
	admissionLimits imageRelayAdmissionLimits
}

func (r *ImageAdmission) CloseAdmission() {
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
}

// Settings are validated by the service before applying to this process-wide relay.
func (r *ImageAdmission) SetAdmissionLimits(bodyLimitMiB, budgetMiB, maxRequests int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.admissionLimits = imageRelayAdmissionLimits{
		bodyBytes:   int64(bodyLimitMiB) << 20,
		budgetBytes: max(int64(budgetMiB), int64(maxRequests)*8) << 20,
		maxRequests: maxRequests,
	}
}

// AdmitRequest bounds additional JSON preparation memory only after selecting
// Basispoints with relay enabled. The gateway's existing ingress limits remain
// unchanged for Codex, API Key, WS, and other routes. Hold until forwarding ends.
func (r *ImageAdmission) AdmitRequest(ctx context.Context, size int) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cost := max(int64(size), 1<<20) * 8
	r.mu.Lock()
	defer r.mu.Unlock()
	limits := r.admissionLimits
	if limits.maxRequests == 0 {
		limits = imageRelayAdmissionLimits{bodyBytes: 64 << 20, budgetBytes: 1024 << 20, maxRequests: 128}
	}
	if int64(size) > limits.bodyBytes {
		return nil, ErrImageRelayRequestTooLarge
	}
	if r.closed {
		return nil, ErrImageRelayStorage
	}
	if r.requests >= limits.maxRequests || r.requestBytes+cost > limits.budgetBytes {
		return nil, ErrImageRelayBusy
	}
	r.requests++
	r.requestBytes += cost
	var once sync.Once
	return func() {
		once.Do(func() {
			r.mu.Lock()
			defer r.mu.Unlock()
			r.requests--
			r.requestBytes -= cost
		})
	}, nil
}
