package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCompatibleImagesSeedreamGParameters(t *testing.T) {
	for _, tt := range []struct {
		model      string
		resolution string
		references int
	}{
		{"seedream-5.0-pro(G)", "1K", 0},
		{"seedream-5.0-pro(G)", "2K", 10},
		{"seedream-5.0-lite(G)", "2K", 0},
		{"seedream-5.0-lite(G)", "3K", 14},
		{"seedream-5.0-lite(G)", "4K", 0},
	} {
		for _, channelMapping := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%s/channel=%t", tt.model, tt.resolution, channelMapping), func(t *testing.T) {
				requestModel, channelModel := tt.model, ""
				if channelMapping {
					requestModel, channelModel = strings.TrimSuffix(tt.model, "(G)")+"-x", tt.model
				}
				endpoint, format := openAIImagesGenerationsEndpoint, "url"
				payload := map[string]any{
					"model": requestModel, "prompt": strings.Repeat("图", 32000),
					"resolution": tt.resolution, "aspect_ratio": "16:9", "size": "auto", "n": 10,
				}
				if tt.references > 0 {
					endpoint, format = openAIImagesEditsEndpoint, "b64_json"
					images := make([]map[string]string, tt.references)
					for i := range images {
						images[i] = map[string]string{"image_url": fmt.Sprintf("https://input.example/%d.png", i)}
					}
					payload["images"] = images
				}
				payload["response_format"] = format
				body, err := json.Marshal(payload)
				require.NoError(t, err)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
				c.Request.Header.Set("Content-Type", "application/json")
				response := `{"data":[{"url":"https://output.example/1.png"},{"url":"https://output.example/2.png"}]}`
				if format == "b64_json" {
					response = `{"data":[{"b64_json":"aW1hZ2Ux"},{"b64_json":"aW1hZ2Uy"}]}`
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
					Body: io.NopCloser(strings.NewReader(response)),
				}}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
				parsed, err := svc.ParseOpenAIImagesRequest(c, body)
				require.NoError(t, err)
				require.Equal(t, tt.references, len(parsed.InputImageURLs))
				require.Equal(t, OpenAIImagesCapabilityAPIKey, parsed.RequiredCapabilityForModel(tt.model))
				account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
					Credentials: map[string]any{"api_key": "image-key", "base_url": "https://compatible.example/v1"}}
				result, err := svc.ForwardImages(c.Request.Context(), c, account, body, parsed, channelModel)
				require.NoError(t, err)
				require.Equal(t, tt.model, result.UpstreamModel)
				require.Equal(t, 2, result.ImageCount, "use returned image count instead of requested n=10")
				require.Equal(t, "https://compatible.example"+endpoint, upstream.lastReq.URL.String())
				require.Equal(t, "Bearer image-key", upstream.lastReq.Header.Get("Authorization"))
				payload["model"] = tt.model
				wantBody, err := json.Marshal(payload)
				require.NoError(t, err)
				require.JSONEq(t, string(wantBody), string(upstream.lastBody))
				require.False(t, gjson.GetBytes(upstream.lastBody, "quality").Exists())
				require.Equal(t, http.StatusOK, rec.Code)
				require.JSONEq(t, response, rec.Body.String())
			})
		}
	}
}
