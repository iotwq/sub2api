package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCompatibleImagesMidjourneySeedreamAdmission(t *testing.T) {
	for _, model := range []string{"Midjourney-V7", "Midjourney V7（满血）", "Midjourney V7 (满血)", "seedream-5.0-pro", "seedream-5.0-pro-x", "seedream-5.0-pro（X）", "seedream-5.0-pro (X)", "seedream-5.0-pro（满血）", "nano-banana2-pro（满血）", "nano-banana2（满血）"} {
		t.Run(model, func(t *testing.T) {
			body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"draw"}`, model))
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, openAIImagesGenerationsEndpoint, bytes.NewReader(body))
			parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			require.Equal(t, model, parsed.Model)
			require.Equal(t, OpenAIImagesCapabilityAPIKey, parsed.RequiredCapability)
			require.Equal(t, OpenAIImagesCapabilityAPIKey, (&OpenAIImagesRequest{RequiredCapability: OpenAIImagesCapabilityNative}).RequiredCapabilityForModel(model))
			require.True(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}).SupportsOpenAIImageCapability(parsed.RequiredCapability))
			for _, typ := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
				require.False(t, (&Account{Platform: PlatformOpenAI, Type: typ}).SupportsOpenAIImageCapability(parsed.RequiredCapability))
			}
			require.Error(t, validateOpenAIImagesModel(model), "do not enable native Codex image conversion")
		})
	}
	for _, model := range []string{"midjourney-chat", "Midjourney-V70", "seedream-5.0-pro-chat", "seedream-video", "nano-banana2-chat", "nano-banana2-pro-video", "gpt-6-astra"} {
		require.Error(t, validateCompatibleImagesModel(model))
	}
}

func TestCompatibleImagesScreenshotModelMapping(t *testing.T) {
	for _, models := range [][2]string{
		{"Midjourney-V7", "Midjourney V7（满血）"},
		{"seedream-5.0-pro", "seedream-5.0-pro（X）"},
		{"seedream-5.0-pro", "seedream-5.0-pro（满血）"},
		{"seedream-5.0-pro", "seedream-5.0-pro-x"},
		{"nano-banana2-pro", "nano-banana2-pro（满血）"},
		{"nano-banana2", "nano-banana2（满血）"},
		{"Midjourney V7（满血）", "Midjourney V7（满血）"},
		{"Midjourney-V7", "Midjourney-V7"},
		{"seedream-5.0-pro（X）", "seedream-5.0-pro（X）"},
		{"seedream-5.0-pro", "seedream-5.0-pro"},
		{"seedream-5.0-pro-x", "seedream-5.0-pro-x"},
		{"seedream-5.0-pro（满血）", "seedream-5.0-pro（满血）"},
		{"nano-banana2-pro（满血）", "nano-banana2-pro（满血）"},
		{"nano-banana2（满血）", "nano-banana2（满血）"},
		{"gpt-image-2", "gpt-image-2（满血）"},
		{"gpt-image-2.5-flare", "gpt-image-2.5-flare（满血）"},
		{"gpt-image-2.5-sunburst", "gpt-image-2.5-sunburst（满血）"},
	} {
		for _, edit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s->%s/edit=%t", models[0], models[1], edit), func(t *testing.T) {
				endpoint, contentType := openAIImagesGenerationsEndpoint, "application/json"
				body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"draw","custom_field":"preserved"}`, models[0]))
				if edit {
					endpoint = openAIImagesEditsEndpoint
					var buf bytes.Buffer
					writer := multipart.NewWriter(&buf)
					require.NoError(t, writer.WriteField("model", models[0]))
					require.NoError(t, writer.WriteField("prompt", "draw"))
					require.NoError(t, writer.WriteField("custom_field", "preserved"))
					part, err := writer.CreateFormFile("image", "input.png")
					require.NoError(t, err)
					_, err = part.Write([]byte("original-image-bytes"))
					require.NoError(t, err)
					require.NoError(t, writer.Close())
					body, contentType = buf.Bytes(), writer.FormDataContentType()
				}
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
				c.Request.Header.Set("Content-Type", contentType)
				upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"aW1hZ2U=","url":"https://cdn.example/image.png"}]}`))}}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
				parsed, err := svc.ParseOpenAIImagesRequest(c, body)
				require.NoError(t, err)
				account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
					"api_key": "image-key", "base_url": "https://compatible.example/v1",
				}, Extra: map[string]any{OpenAIImageURLToB64JSONExtraKey: true}}
				if models[0] != models[1] {
					account.Credentials["model_mapping"] = map[string]any{models[0]: models[1]}
				}
				result, err := svc.ForwardImages(c.Request.Context(), c, account, body, parsed, "")
				require.NoError(t, err)
				require.Equal(t, models[0], result.Model)
				require.Equal(t, models[1], result.UpstreamModel)
				require.Equal(t, 1, result.ImageCount)
				require.Equal(t, "https://compatible.example"+endpoint, upstream.lastReq.URL.String())
				require.Equal(t, "Bearer image-key", upstream.lastReq.Header.Get("Authorization"))
				if edit {
					r := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(upstream.lastBody))
					r.Header.Set("Content-Type", upstream.lastReq.Header.Get("Content-Type"))
					require.NoError(t, r.ParseMultipartForm(1<<20))
					defer r.MultipartForm.RemoveAll()
					require.Equal(t, models[1], r.FormValue("model"))
					require.Equal(t, "preserved", r.FormValue("custom_field"))
					file, _, err := r.FormFile("image")
					require.NoError(t, err)
					defer file.Close()
					data, err := io.ReadAll(file)
					require.NoError(t, err)
					require.Equal(t, "original-image-bytes", string(data))
				} else {
					require.Equal(t, models[1], gjson.GetBytes(upstream.lastBody, "model").String())
					require.Equal(t, "preserved", gjson.GetBytes(upstream.lastBody, "custom_field").String())
				}
				require.Equal(t, 200, rec.Code)
				require.Equal(t, "aW1hZ2U=", gjson.Get(rec.Body.String(), "data.0.b64_json").String())
				require.False(t, gjson.Get(rec.Body.String(), "data.0.url").Exists())
			})
		}
	}
}
