package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/service/basispoints"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const basispointsImageTestPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRusAAAAASUVORK5CYII="

type basispointsImageIsolationUpstream struct{ basispointsTestUpstream }

func (u *basispointsImageIsolationUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.send(req)
}

func TestBasispointsImageRelayLeavesDefaultCodexAndAPIKeyUnchanged(t *testing.T) {
	for _, kind := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("DATA_DIR", dir)
			repo := &excelBPSImageSettingsRepo{values: map[string]string{
				SettingKeyExcelBPSImageRelayEnabled: "true", SettingKeyExcelBPSImageBaseURL: "https://relay.example",
			}}
			called := false
			upstream := &basispointsImageIsolationUpstream{basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
				called = true
				require.NotEqual(t, "bps.openai.com", req.URL.Host)
				require.Contains(t, string(mustReadRequestBody(t, req)), "data:image/png;base64,"+basispointsImageTestPNG)
				require.Empty(t, req.Header.Get("x-basispoints-auth-mode"))
				return basispointsTestCompletedResponse(t, []any{}), nil
			}}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream, settingService: NewSettingService(repo, &config.Config{})}
			account := &Account{ID: 42, Platform: PlatformOpenAI, Type: kind,
				Credentials: map[string]any{"access_token": "test-token", "api_key": "test-key", "chatgpt_account_id": "acct-test"}}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			body := []byte(fmt.Sprintf(`{"model":"gpt-6-sol","input":[{"role":"user","content":[{"type":"input_image","image_url":"data:image/png;base64,%s"}]}]}`, basispointsImageTestPNG))
			_, err := svc.Forward(context.Background(), c, account, body)
			require.NoError(t, err)
			require.True(t, called)
			require.Nil(t, svc.excelBPSImages)
			require.NoDirExists(t, filepath.Join(dir, "bps-images"))
		})
	}
}

func TestBasispointsImageForwardConvertsOnlyWhenEnabled(t *testing.T) {
	for _, kind := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
		for _, enabled := range []bool{false, true} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/enabled=%t/stream=%t", kind, enabled, stream), func(t *testing.T) {
					dir := t.TempDir()
					t.Setenv("DATA_DIR", dir)
					repo := &excelBPSImageSettingsRepo{values: map[string]string{
						SettingKeyExcelBPSImageRelayEnabled: fmt.Sprint(enabled),
						SettingKeyExcelBPSImageBaseURL:      "https://relay.example",
					}}
					svc := &OpenAIGatewayService{cfg: &config.Config{}, settingService: NewSettingService(repo, &config.Config{})}
					t.Cleanup(func() { require.NoError(t, svc.CloseExcelBPSImages()) })
					sent := 0
					svc.httpUpstream = &basispointsTestUpstream{send: func(req *http.Request) (*http.Response, error) {
						sent++
						body := mustReadRequestBody(t, req)
						require.NotContains(t, string(body), "data:image")
						require.Equal(t, "xhigh", gjson.GetBytes(body, "reasoning_effort").String())
						imageURL := gjson.GetBytes(body, `input.#(role=="user").content.0.image_url`).String()
						require.True(t, strings.HasPrefix(imageURL, "https://relay.example/api/bps-images/"), string(body))
						rec := httptest.NewRecorder()
						c, _ := gin.CreateTestContext(rec)
						c.Request = httptest.NewRequest(http.MethodGet, imageURL, nil)
						svc.ServeExcelBPSImage(c)
						require.Equal(t, http.StatusOK, rec.Code)
						data, err := base64.StdEncoding.DecodeString(basispointsImageTestPNG)
						require.NoError(t, err)
						require.Equal(t, data, rec.Body.Bytes())
						return basispointsTestCompletedResponse(t, []any{}), nil
					}}
					body := []byte(fmt.Sprintf(`{"model":"gpt-6-sol","stream":%t,"reasoning":{"effort":"max"},"input":[{"role":"user","content":[{"type":"input_image","image_url":"data:image/png;base64,%s"}]}]}`, stream, basispointsImageTestPNG))
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
					account := &Account{ID: 42, Platform: PlatformOpenAI, Type: kind,
						Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "acct-test"},
						Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
					result, err := svc.Forward(context.Background(), c, account, body)
					if !enabled {
						require.Error(t, err)
						require.Equal(t, http.StatusBadRequest, rec.Code)
						require.Zero(t, sent)
						require.NoDirExists(t, filepath.Join(dir, "bps-images"))
						return
					}
					require.NoError(t, err)
					require.Equal(t, http.StatusOK, rec.Code)
					require.Equal(t, 1, sent)
					require.Equal(t, 12, result.Usage.InputTokens)
					require.Equal(t, 4, result.Usage.OutputTokens)
					// The forwarding defer must release the complete request budget.
					release, err := svc.excelBPSImages.AdmitRequest(context.Background(), 64<<20)
					require.NoError(t, err)
					release()
				})
			}
		}
	}
}

func TestBasispointsImageRelayShutdownAndAnonymousRead(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DATA_DIR", dir)
	repo := &excelBPSImageSettingsRepo{values: map[string]string{
		SettingKeyExcelBPSImageRelayEnabled: "true", SettingKeyExcelBPSImageBaseURL: "https://relay.example",
	}}
	svc := &OpenAIGatewayService{settingService: NewSettingService(repo, &config.Config{})}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/bps-images/"+strings.Repeat("a", 43), nil)
	svc.ServeExcelBPSImage(c)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.NoDirExists(t, filepath.Join(dir, "bps-images"))
	_, err := svc.excelBPSImageRelay(context.Background())
	require.NoError(t, err)
	require.NoError(t, svc.CloseExcelBPSImages())
	entries, err := os.ReadDir(filepath.Join(dir, "bps-images"))
	require.NoError(t, err)
	require.Empty(t, entries)
	_, err = svc.excelBPSImageRelay(context.Background())
	require.ErrorIs(t, err, basispoints.ErrImageRelayStorage)
}

func TestBasispointsImageAdmissionFailsBeforeUpstreamAndReleasesOnInvalidImage(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	repo := &excelBPSImageSettingsRepo{values: map[string]string{
		SettingKeyExcelBPSImageRelayEnabled: "true", SettingKeyExcelBPSImageBaseURL: "https://relay.example",
		SettingKeyExcelBPSImageBudgetMiB: "512", SettingKeyExcelBPSImageMaxRequests: "32",
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, settingService: NewSettingService(repo, &config.Config{})}
	t.Cleanup(func() { require.NoError(t, svc.CloseExcelBPSImages()) })
	svc.excelBPSImageAdmission.SetAdmissionLimits(64, 512, 32)
	release, err := svc.excelBPSImageAdmission.AdmitRequest(context.Background(), 64<<20)
	require.NoError(t, err)
	defer release()
	account := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Extra: map[string]any{openAIOAuthResponsesEndpointExtraKey: "basispoints"}}
	call := func(body string, status int) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
		_, err := svc.Forward(context.Background(), c, account, []byte(body))
		require.Error(t, err)
		require.Equal(t, status, rec.Code, rec.Body.String())
	}
	call(`{"model":"gpt-6-sol","input":"hi"}`, http.StatusServiceUnavailable)
	release()
	call(`{"model":"gpt-6-sol","input":[{"role":"user","content":[{"type":"input_image","image_url":"data:image/png;base64,invalid"}]}]}`, http.StatusBadRequest)
	release, err = svc.excelBPSImageAdmission.AdmitRequest(context.Background(), 64<<20)
	require.NoError(t, err)
	release()
}
