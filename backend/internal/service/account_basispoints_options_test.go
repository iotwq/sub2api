package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasispointsCompatibilityOptionsRequireOAuthOptIn(t *testing.T) {
	for _, accountType := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
		for _, mode := range []string{"", openAIOAuthResponsesEndpointChatGPT, openAIOAuthResponsesEndpointBasis} {
			a := basispointsAccountForTest()
			a.Type = accountType
			a.Extra[openAIOAuthResponsesEndpointExtraKey] = mode
			for _, key := range []string{BasispointsIgnoreImagesKey, BasispointsIgnoreEncryptedContentKey} {
				a.Extra[key] = true
			}
			want := accountType != AccountTypeAPIKey && mode == openAIOAuthResponsesEndpointBasis
			require.Equal(t, want, a.IsBasispointsIgnoreImagesEnabled())
			require.Equal(t, want, a.IsBasispointsIgnoreEncryptedContentEnabled())
		}
	}
	for _, key := range []string{BasispointsIgnoreImagesKey, BasispointsIgnoreEncryptedContentKey} {
		require.Error(t, validateBasispoints403GroupExtra(map[string]any{key: "true"}))
		require.NoError(t, validateBasispoints403GroupExtra(map[string]any{key: true}))
	}
}
