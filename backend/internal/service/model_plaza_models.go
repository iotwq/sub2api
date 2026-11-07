package service

import (
	"context"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
)

// accountModels supplements channels with the group's configured account models.
// Persistent availability keeps the catalog stable during temporary rate limits.
func (s *ModelPlazaService) accountModels(ctx context.Context, g *Group, channelModels []PlazaModel) ([]PlazaModel, error) {
	if s.accountRepo == nil {
		return nil, nil
	}
	platforms := []string{PlatformOpenAI, PlatformAnthropic, PlatformGrok, PlatformGemini}
	if g.Platform != PlatformComposite {
		switch g.Platform {
		case PlatformOpenAI, PlatformAnthropic, PlatformGrok, PlatformGemini:
			platforms = []string{g.Platform}
		default:
			return nil, nil
		}
	}
	queryPlatforms := append(append([]string(nil), platforms...), PlatformAntigravity)
	accounts, err := s.accountRepo.ListModelAvailabilityCandidates(ctx, &g.ID, queryPlatforms, false)
	if err != nil {
		return nil, err
	}
	models := make([]PlazaModel, 0)
	seen := make(map[string]bool)
	for _, platform := range platforms {
		defaults := plazaDefaultModels(platform)
		for i := range accounts {
			account := &accounts[i]
			if account.Platform != platform && !mixedListingAccountAllowed(platform, account) {
				continue
			}
			// Configured request names come first so their spelling is preserved.
			candidates := make([]string, 0)
			for name := range account.GetModelMapping() {
				candidates = append(candidates, name)
			}
			sort.Strings(candidates)
			candidates = append(candidates, defaults...)
			for _, model := range channelModels {
				if model.Platform == platform {
					candidates = append(candidates, model.Name)
				}
			}
			if g.ModelAllowlistEnabled() {
				candidates = append(candidates, g.ModelAllowlist.Models...)
			}
			for _, name := range candidates {
				name = strings.TrimSpace(name)
				if name == "" || strings.Contains(name, "*") || !g.ModelAllowlist.Allows(name) {
					continue
				}
				if account.Platform != platform && !mixedListingModelAllowed(platform, name) {
					continue
				}
				if !account.IsModelSupported(name) {
					continue
				}
				key := platform + ":" + strings.ToLower(name)
				if seen[key] {
					continue
				}
				seen[key] = true
				model := PlazaModel{Name: name, Platform: platform}
				if !account.IsOpenAIPassthroughEnabled() {
					mapped := strings.TrimSpace(account.GetMappedModel(name))
					if mapped != "" && !strings.EqualFold(mapped, name) {
						model.mappedModel = mapped
					}
				}
				models = append(models, model)
			}
		}
	}
	return models, nil
}

func plazaDefaultModels(platform string) []string {
	switch platform {
	case PlatformOpenAI:
		return openai.DefaultModelIDs()
	case PlatformAnthropic:
		return claude.DefaultModelIDs()
	case PlatformGrok:
		return xai.DefaultModelIDs()
	case PlatformGemini:
		models := make([]string, 0, len(geminicli.DefaultModels))
		for _, m := range geminicli.DefaultModels {
			models = append(models, m.ID)
		}
		return models
	default:
		return nil
	}
}
