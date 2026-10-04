package service

// Basispoints403DisabledAtKey records when an upstream HTTP 403 automatically
// turned Excel BPS off (RFC 3339, UTC). Only DisableBasispointsOn403 writes it; the
// admin account list shows it as a suspected Excel ban.
const Basispoints403DisabledAtKey = "openai_basispoints_403_disabled_at"

// Basispoints403MovedAtKey and Basispoints403MovedGroupIDKey record the last automatic
// group action after an upstream HTTP 403 (group 0 means all groups were left).
// Only MoveBasispointsOn403 writes them; the account list flags the suspected ban
// while the account's groups still match that action.
const (
	Basispoints403MovedAtKey      = "openai_basispoints_403_moved_at"
	Basispoints403MovedGroupIDKey = "openai_basispoints_403_moved_group_id"
)

var basispoints403MarkerKeys = []string{Basispoints403DisabledAtKey, Basispoints403MovedAtKey, Basispoints403MovedGroupIDKey}

// MergeBasispoints403Marker keeps the persisted 403 records across account edits
// and ignores values supplied by the edit. Turning Excel BPS back on
// acknowledges the automatic shutdown and clears its record; group-action
// records stay until the next automatic move. The repository applies this
// under the row lock.
func MergeBasispoints403Marker(extra, current map[string]any) map[string]any {
	for _, key := range basispoints403MarkerKeys {
		delete(extra, key)
	}
	enabled := extra[openAIOAuthResponsesEndpointExtraKey] == openAIOAuthResponsesEndpointBasis
	for _, key := range basispoints403MarkerKeys {
		value, ok := current[key]
		if !ok || (enabled && key == Basispoints403DisabledAtKey) {
			continue
		}
		if extra == nil {
			extra = make(map[string]any, len(basispoints403MarkerKeys))
		}
		extra[key] = value
	}
	return extra
}
