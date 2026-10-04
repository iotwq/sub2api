package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// CodexSignalExtraKey contains observations only, never the opaque state itself.
const CodexSignalExtraKey = "codex_signal"

// CodexSignalStatus is a header-length heuristic, not a model capability verdict.
type CodexSignalStatus struct {
	Length     int        `json:"length"`
	ObservedAt time.Time  `json:"observed_at"`
	Last312At  *time.Time `json:"last_312_at,omitempty"`
}

func CodexSignalFromAccount(account *Account) *CodexSignalStatus {
	if account == nil || !account.IsOpenAIOAuthLike() || account.IsShadow() {
		return nil
	}
	raw, err := json.Marshal(account.Extra[CodexSignalExtraKey])
	if err != nil {
		return nil
	}
	var signal CodexSignalStatus
	if json.Unmarshal(raw, &signal) != nil || signal.ObservedAt.IsZero() || signal.Length <= 0 {
		return nil
	}
	return &signal
}

func (s *OpenAIGatewayService) observeCodexSignal(account *Account, headers http.Header) {
	if s == nil || s.deferredService == nil || account == nil || account.ID <= 0 || !account.IsOpenAIOAuthLike() || account.IsShadow() {
		return
	}
	state := strings.TrimSpace(headers.Get(openAICodexTurnStateHeader))
	// Do not treat missing/malformed headers as recovery or record echoed client state.
	if len(state) < 7 || len(state) > 4096 || !strings.HasPrefix(state, "gAAAAA") {
		return
	}
	for _, c := range state {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '=') {
			return
		}
	}
	now := time.Now().UTC()
	signal := CodexSignalStatus{Length: len(state), ObservedAt: now}
	if signal.Length == 312 {
		signal.Last312At = &now
	}
	s.deferredService.scheduleCodexSignal(account.ID, signal)
}

// MergeCodexSignals preserves both the newest observation and the most recent
// 312 sighting, including when concurrent requests finish out of order.
func MergeCodexSignals(a, b CodexSignalStatus) CodexSignalStatus {
	latest := a
	if b.ObservedAt.After(a.ObservedAt) {
		latest = b
	}
	if a.Last312At != nil && (latest.Last312At == nil || a.Last312At.After(*latest.Last312At)) {
		latest.Last312At = a.Last312At
	}
	if b.Last312At != nil && (latest.Last312At == nil || b.Last312At.After(*latest.Last312At)) {
		latest.Last312At = b.Last312At
	}
	return latest
}

// CodexSignalRepository is implemented by the account repository with an atomic,
// timestamp-ordered merge. No schema change or scheduler invalidation is needed.
type CodexSignalRepository interface {
	RecordCodexSignal(context.Context, int64, CodexSignalStatus) error
}

func (s *DeferredService) scheduleCodexSignal(id int64, signal CodexSignalStatus) {
	s.codexSignalMu.Lock()
	defer s.codexSignalMu.Unlock()
	if s.codexSignals == nil {
		s.codexSignals = make(map[int64]CodexSignalStatus)
	}
	s.codexSignals[id] = MergeCodexSignals(s.codexSignals[id], signal)
}

func (s *DeferredService) flushCodexSignals() {
	repo, ok := s.accountRepo.(CodexSignalRepository)
	if !ok {
		return
	}
	s.codexSignalMu.Lock()
	updates := s.codexSignals
	s.codexSignals = nil
	s.codexSignalMu.Unlock()
	if len(updates) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for id, signal := range updates {
		if err := repo.RecordCodexSignal(ctx, id, signal); err != nil {
			slog.Warn("codex signal persistence failed", "account_id", id, "error", err)
			s.scheduleCodexSignal(id, signal)
		}
	}
}
