package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	BasispointsAutoMoveOn403Key    = "openai_basispoints_auto_move_on_403"
	Basispoints403TargetGroupIDKey = "openai_basispoints_403_target_group_id"
)

type AccountBasispointsGroupRepository interface {
	MoveBasispointsOn403(context.Context, *Account) (bool, error)
}

// Basispoints403GroupTarget returns an explicitly configured destination. Zero
// means leave all groups; a missing or invalid destination never removes groups.
func (a *Account) Basispoints403GroupTarget() (int64, bool) {
	if !a.UsesBasispointsResponses() || a.ParentAccountID != nil || a.IsOpenAIPersonalAccessToken() || a.Extra[BasispointsAutoMoveOn403Key] != true {
		return 0, false
	}
	return basispoints403GroupID(a.Extra[Basispoints403TargetGroupIDKey])
}

func basispoints403GroupID(value any) (int64, bool) {
	var id int64
	switch value := value.(type) {
	case int:
		id = int64(value)
	case int64:
		id = value
	case float64:
		if math.IsNaN(value) || value < 0 || value >= math.MaxInt64 || math.Trunc(value) != value {
			return 0, false
		}
		id = int64(value)
	case json.Number:
		var err error
		id, err = value.Int64()
		if err != nil {
			return 0, false
		}
	default:
		return 0, false
	}
	return id, id >= 0
}

func validateBasispoints403GroupExtra(extra map[string]any) error {
	for _, key := range []string{BasispointsIgnoreImagesKey, BasispointsIgnoreEncryptedContentKey} {
		if value, exists := extra[key]; exists {
			if _, ok := value.(bool); !ok {
				return infraerrors.BadRequest("OPENAI_BASISPOINTS_INVALID", key+" must be a boolean")
			}
		}
	}
	invalid := func(message string) error {
		return infraerrors.BadRequest("OPENAI_BASISPOINTS_INVALID", message)
	}
	if raw, exists := extra[BasispointsAutoMoveOn403Key]; exists {
		if _, ok := raw.(bool); !ok {
			return invalid(BasispointsAutoMoveOn403Key + " must be a boolean")
		}
	}
	if raw, exists := extra[Basispoints403TargetGroupIDKey]; exists && raw != nil {
		if _, ok := basispoints403GroupID(raw); !ok {
			return invalid(Basispoints403TargetGroupIDKey + " must be a nonnegative integer")
		}
	}
	if extra[BasispointsAutoMoveOn403Key] == true {
		if _, ok := basispoints403GroupID(extra[Basispoints403TargetGroupIDKey]); !ok {
			return invalid("BPS 403 group action requires an explicit target group, or 0 to leave all groups")
		}
	}
	return nil
}

func (s *adminServiceImpl) validateBasispoints403GroupSettings(ctx context.Context, account *Account) error {
	if err := validateBasispoints403GroupExtra(account.Extra); err != nil {
		return err
	}
	if account.Extra[BasispointsAutoMoveOn403Key] != true {
		return nil
	}
	if account.Platform != PlatformOpenAI || !account.IsOpenAIOAuthLike() || account.ParentAccountID != nil || account.IsOpenAIAgentIdentity() || account.IsOpenAIPersonalAccessToken() {
		return infraerrors.BadRequest("OPENAI_BASISPOINTS_INVALID", "BPS 403 group actions require an OpenAI ChatGPT OAuth account")
	}
	target, _ := basispoints403GroupID(account.Extra[Basispoints403TargetGroupIDKey])
	if target == 0 {
		return nil
	}
	if s.groupRepo == nil {
		return errors.New("group repository not configured")
	}
	group, err := s.groupRepo.GetByIDLite(ctx, target)
	if err != nil {
		return err
	}
	if group == nil || (group.Platform != PlatformOpenAI && group.Platform != PlatformComposite) {
		return infraerrors.BadRequest("OPENAI_BASISPOINTS_INVALID", "BPS 403 destination must be an OpenAI or composite group")
	}
	return s.ValidateAccountGroupBindings(ctx, []int64{target})
}
