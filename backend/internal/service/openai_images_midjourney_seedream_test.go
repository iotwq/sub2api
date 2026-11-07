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
	for _, model := range []string{"Midjourney-V7", "Midjourney V7（满血）", "Midjourney V7 (满血)", "midjourney-D", "seedream-5.0-pro", "seedream-5.0-pro-x", "seedream-5.0-pro（X）", "seedream-5.0-pro (X)", "seedream-5.0-pro（满血）", "seedream-5.0-pro(G)", "seedream-5.0-lite(G)", "seedream-5.0-lite-x", "seedream-5.0-Flash（满血）", "seedream-5.0-flash", "nano-banana2-pro（满血）", "nano-banana2（满血）", "nano-banana-2.1（as）", "nano-banana-2（as）", "nano-banana-pro（as）", "gpt-image-2-D", "gpt-image-2.5-D", "grok-imagine-image", "grok-imagine-image-2.0", "grok-imagine-image-quality"} {
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
	for _, model := range []string{"midjourney-chat", "Midjourney-V70", "midjourney-D-chat", "seedream-5.0-pro-chat", "seedream-5.0-lite-chat", "seedream-5.0-lite-video", "seedream-5.0-flash-chat", "seedream-video", "nano-banana2-chat", "nano-banana2-pro-video", "nano-banana-2.1-chat", "nano-banana-2-video", "nano-banana-pro-chat", "gpt-6-astra"} {
		require.Error(t, validateCompatibleImagesModel(model))
	}
}

func TestCompatibleImagesScreenshotModelMapping(t *testing.T) {
	for _, models := range [][2]string{
		{"Midjourney-V7", "Midjourney V7（满血）"},
		{"Midjourney-V7", "midjourney-D"},
		{"midjourney-D", "midjourney-D"},
		{"seedream-5.0-pro", "seedream-5.0-pro（X）"},
		{"seedream-5.0-pro", "seedream-5.0-pro（满血）"},
		{"seedream-5.0-pro", "seedream-5.0-pro-x"},
		{"seedream-5.0-pro", "seedream-5.0-pro(G)"},
		{"seedream-5.0-lite", "seedream-5.0-lite(G)"},
		{"seedream-5.0-pro-x", "seedream-5.0-pro(G)"},
		{"seedream-5.0-lite-x", "seedream-5.0-lite(G)"},
		{"seedream-5.0-pro(G)", "seedream-5.0-pro(G)"},
		{"seedream-5.0-lite(G)", "seedream-5.0-lite(G)"},
		{"seedream-5.0-flash", "seedream-5.0-Flash（满血）"},
		{"nano-banana2-pro", "nano-banana2-pro（满血）"},
		{"nano-banana2", "nano-banana2（满血）"},
		{"nano-banana-2.1（as）", "nano-banana-2.1（as）"},
		{"nano-banana-2（as）", "nano-banana-2（as）"},
		{"nano-banana-pro（as）", "nano-banana-pro（as）"},
		{"nano-banana-2.1", "nano-banana-2.1（as）"},
		{"nano-banana2", "nano-banana-2（as）"},
		{"nano-banana2-pro", "nano-banana-pro（as）"},
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
		{"gpt-image-2-D", "gpt-image-2-D"},
		{"gpt-image-2.5-D", "gpt-image-2.5-D"},
		{"gpt-image-2", "gpt-image-2-D"},
		{"gpt-image-2", "gpt-image-2.5-D"},
		{"grok-imagine-image", "grok-imagine-image"},
		{"grok-imagine-image-2.0", "grok-imagine-image-2.0"},
		{"grok-imagine-image-quality", "grok-imagine-image-quality"},
		{"grok-imagine-image", "grok-imagine-image-2.0"},
		{"grok-imagine-image-2.0", "grok-imagine-image-quality"},
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

func TestCompatibleImagesNanoBananaASChannelMapping(t *testing.T) {
	for _, model := range []string{"nano-banana-2.1（as）", "nano-banana-2（as）", "nano-banana-pro（as）"} {
		t.Run(model, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"draw","size":"2K","aspect_ratio":"16:9","response_format":"url"}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, openAIImagesGenerationsEndpoint, bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			response := `{"data":[{"url":"https://cdn.example/image.png"}]}`
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
				Body: io.NopCloser(strings.NewReader(response)),
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			require.Equal(t, OpenAIImagesCapabilityAPIKey, parsed.RequiredCapabilityForModel(model))
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "image-key", "base_url": "https://compatible.example/v1"}}
			result, err := svc.ForwardImages(c.Request.Context(), c, account, body, parsed, model)
			require.NoError(t, err)
			require.Equal(t, model, result.UpstreamModel)
			require.Equal(t, 1, result.ImageCount)
			require.Equal(t, "https://compatible.example/v1/images/generations", upstream.lastReq.URL.String())
			require.Equal(t, "Bearer image-key", upstream.lastReq.Header.Get("Authorization"))
			require.JSONEq(t, fmt.Sprintf(`{"model":%q,"prompt":"draw","size":"2K","aspect_ratio":"16:9","response_format":"url"}`, model), string(upstream.lastBody))
			require.Equal(t, http.StatusOK, rec.Code)
			require.JSONEq(t, response, rec.Body.String())
		})
	}
}

func TestCompatibleImagesNativeRejection(t *testing.T) {
	for _, model := range []string{"gpt-image-2-D", "gpt-image-2.5-D", "nano-banana-2.1（as）", "nano-banana-2（as）", "nano-banana-pro（as）", "seedream-5.0-pro(G)", "seedream-5.0-lite(G)", "seedream-5.0-lite-x", "grok-imagine-image", "grok-imagine-image-2.0", "grok-imagine-image-quality"} {
		t.Run(model, func(t *testing.T) {
			body := map[string]any{"model": model, "input": "draw"}
			require.False(t, normalizeOpenAIResponsesImageOnlyModel(body))
			require.Equal(t, model, body["model"])
			for _, typ := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
				for _, mapping := range []string{"direct", "channel", "account"} {
					t.Run(typ+"/"+mapping, func(t *testing.T) {
						requestModel, channelModel := model, ""
						account := &Account{Platform: PlatformOpenAI, Type: typ, Credentials: map[string]any{"access_token": "unused"}}
						if mapping == "channel" {
							requestModel, channelModel = "gpt-image-2", model
						} else if mapping == "account" {
							requestModel = "gpt-image-2"
							account.Credentials["model_mapping"] = map[string]any{requestModel: model}
						}
						c, _ := gin.CreateTestContext(httptest.NewRecorder())
						c.Request = httptest.NewRequest(http.MethodPost, openAIImagesGenerationsEndpoint, nil)
						upstream := &httpUpstreamRecorder{}
						svc := &OpenAIGatewayService{httpUpstream: upstream}
						_, err := svc.ForwardImages(c.Request.Context(), c, account, nil, &OpenAIImagesRequest{Model: requestModel}, channelModel)
						require.ErrorContains(t, err, "images endpoint requires an image model")
						require.Empty(t, upstream.requests)
					})
				}
			}
		})
	}
}
