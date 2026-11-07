package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestExcelOAuthNativeRoutesRejectBeforeSend(t *testing.T) {
	account := basispointsAccountForTest()
	account.Credentials["client_id"] = openai.ExcelClientID
	require.True(t, account.IsExcelOAuth())
	require.False(t, account.IsBasispointsAutoDisableOn403Enabled())
	account.Extra["openai_basispoints_auto_disable_on_403"] = true
	require.False(t, account.IsBasispointsAutoDisableOn403Enabled())
	require.False(t, account.SupportsOpenAIImageCapability(OpenAIImagesCapabilityBasic))
	for _, mode := range []string{"basispoints", "chatgpt_codex"} {
		account.Extra["openai_oauth_responses_endpoint"] = mode
		require.Equal(t, mode == "basispoints", account.IsModelSupported("gpt-6-astra"))
		upstream := &httpUpstreamRecorder{}
		svc := openAIClientToolsTestService(upstream)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
		body := []byte(`{"model":"gpt-6-astra","input":"test"}`)
		_, err := svc.buildUpstreamRequest(context.Background(), c, account, body, "synthetic-token", false, "", false)
		require.ErrorIs(t, err, errExcelOAuthRouteUnavailable)
		_, err = svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, body, "synthetic-token")
		require.ErrorIs(t, err, errExcelOAuthRouteUnavailable)
		_, err = svc.buildOpenAIResponsesWSURL(account)
		require.ErrorIs(t, err, errExcelOAuthRouteUnavailable)
		require.ErrorIs(t, validateOpenAIWSBearerToken(account, "synthetic-token"), errExcelOAuthRouteUnavailable)
		_, err = svc.Forward(withOpenAIImagesSelfBuiltRequest(context.Background()), c, account, body)
		require.ErrorIs(t, err, errExcelOAuthRouteUnavailable)
		if mode == "chatgpt_codex" {
			_, err = svc.FetchOpenAIModelsList(context.Background(), account)
			require.Error(t, err)
			_, err = svc.FetchCodexModelsManifest(context.Background(), account, "", "")
			require.Error(t, err)
		}
		require.Empty(t, upstream.requests)
	}
}

func TestExcelOAuthForwardsBPSAndPreservesClientOnRefresh(t *testing.T) {
	account := basispointsAccountForTest()
	account.Credentials["client_id"] = openai.ExcelClientID
	upstream := &httpUpstreamRecorder{resp: bpsCompletionResponse(http.StatusOK, basispointsRepairWire(t, "excel", "Run", `{"name":"shell","arguments":{"cmd":"test"}}`))}
	svc := openAIClientToolsTestService(upstream)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	result, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-6-astra","input":"test","tools":[{"type":"function","name":"shell"}]}`))
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "bps.openai.com", upstream.requests[0].URL.Host)
	oauth := NewOpenAIOAuthService(nil, &excelRefreshClient{})
	defer oauth.Stop()
	account.Credentials["refresh_token"] = "synthetic-refresh"
	token, err := oauth.RefreshAccountToken(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, openai.ExcelClientID, token.ClientID)
	require.Equal(t, openai.ExcelClientID, oauth.BuildAccountCredentials(token)["client_id"])
}

type excelRefreshClient struct{ openaiOAuthClientAuthURLStub }

func (*excelRefreshClient) RefreshTokenWithClientID(_ context.Context, refresh, _ string, client string) (*openai.TokenResponse, error) {
	if refresh != "synthetic-refresh" || client != openai.ExcelClientID {
		return nil, errExcelOAuthRouteUnavailable
	}
	return &openai.TokenResponse{AccessToken: "synthetic-new", RefreshToken: "synthetic-rotated", ExpiresIn: 3600}, nil
}
