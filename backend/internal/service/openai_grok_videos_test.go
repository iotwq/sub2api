package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
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

func TestOpenAIGrokVideoCreateForwardsAndBindsRequestID(t *testing.T) {
	for _, tc := range []struct {
		model         string
		upstreamModel string
		baseURL       string
	}{
		{"grok-imagine-video", "grok-imagine-video", "https://video.test"},
		{"grok-imagine-video-1.5", "grok-imagine-video-1.5", "https://video.test/v1"},
		{"grok-imagine-video", "grok-imagine-video-1.5", "https://video.test/v1"},
		{"grok-imagine-video-1.5", "grok-imagine-video", "https://video.test"},
	} {
		t.Run(tc.model+"_to_"+tc.upstreamModel, func(t *testing.T) {
			body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"waves","duration":8,"resolution":"720p","aspect_ratio":"16:9","image":{"url":"https://example.com/frame.jpg"}}`, tc.model))
			parsed, err := ParseOpenAIVideoCreateRequest(body)
			require.NoError(t, err)
			require.Equal(t, tc.model, parsed.Model)
			require.Equal(t, "waves", parsed.Prompt)
			require.Equal(t, body, parsed.BillingBody)

			account := newOpenAIVideoTestAccount()
			account.Credentials["base_url"] = tc.baseURL
			account.Credentials["model_mapping"] = map[string]any{tc.model: tc.upstreamModel}
			require.True(t, AccountSupportsOpenAIVideoEndpoint(account, tc.model))
			upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"request_id":"grok-task-123"}`)}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))

			result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", parsed.Model)

			require.NoError(t, err)
			require.Equal(t, "https://video.test/v1/videos", upstream.lastReq.URL.String())
			require.Equal(t, http.MethodPost, upstream.lastReq.Method)
			require.Equal(t, "Bearer video-key", upstream.lastReq.Header.Get("Authorization"))
			require.Equal(t, "application/json", upstream.lastReq.Header.Get("Content-Type"))
			require.JSONEq(t, strings.Replace(string(body), tc.model, tc.upstreamModel, 1), string(upstream.lastBody))
			require.Equal(t, tc.model, result.Model)
			require.Equal(t, tc.upstreamModel, result.UpstreamModel)
			require.Equal(t, "/v1/videos", result.UpstreamEndpoint)
			require.Equal(t, "grok-task-123", result.ResponseID)
			require.Equal(t, "grok-task-123", result.BillingTaskID)
			require.Equal(t, "video", result.MediaType)
			require.Equal(t, 1, result.VideoCount)
			require.Equal(t, VideoBillingResolution720P, result.VideoResolution)
			require.Equal(t, 8, result.VideoDurationSeconds)
			require.JSONEq(t, `{"request_id":"grok-task-123"}`, recorder.Body.String())

			groupID := int64(125)
			repo := newOpenAIVideoTaskBindingRepoStub()
			svc.openAIVideoTaskBindingRepo = repo
			require.NoError(t, svc.BindOpenAIVideoTaskAccount(context.Background(), &groupID, 42, result.ResponseID, account.ID))
			binding := repo.bindings["125:42:grok-task-123"]
			require.Equal(t, result.ResponseID, binding.TaskID)
			require.Equal(t, account.ID, binding.AccountID)
		})
	}
}

func TestOpenAIGrokVideoRejectsNonAPIKeyAccounts(t *testing.T) {
	for _, model := range []string{"grok-imagine-video", "grok-imagine-video-1.5"} {
		for _, accountType := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
			t.Run(model+"_"+accountType, func(t *testing.T) {
				account := newOpenAIVideoTestAccount()
				account.Type = accountType
				require.False(t, AccountSupportsOpenAIVideoEndpoint(account, model))
				svc := &OpenAIGatewayService{}
				_, err := svc.ForwardOpenAIVideo(context.Background(), nil, account, OpenAIVideoEndpointCreate, "", "", nil, "application/json", model)
				require.ErrorContains(t, err, "requires an OpenAI APIKey account")
			})
		}
	}
	for _, model := range []string{"grok-imagine-video-unknown", "svggrok-imagine-video", "grok-imagine-image"} {
		require.False(t, IsOpenAIVideoModel(model))
	}
}

func TestOpenAIGrokVideoStatusPreservesResultAndRecognizesDone(t *testing.T) {
	account := newOpenAIVideoTestAccount()
	body := `{"status":"done","model":"grok-imagine-video-1.5","video":{"url":"https://example.com/output.mp4","duration":8}}`
	upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, body)}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/grok-task-123", nil)

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointStatus, "grok-task-123", "grok-task-123", nil, "", "")

	require.NoError(t, err)
	require.Equal(t, "https://video.test/v1/videos/grok-task-123", upstream.lastReq.URL.String())
	require.Equal(t, http.MethodGet, upstream.lastReq.Method)
	require.Empty(t, upstream.lastBody)
	require.Equal(t, "grok-task-123", result.ResponseID)
	require.True(t, OpenAIVideoStatusSucceeded(result.ResponseBody))
	require.False(t, OpenAIVideoStatusFailed(result.ResponseBody))
	require.Zero(t, result.VideoCount)
	require.JSONEq(t, body, recorder.Body.String())
}

func TestOpenAIGrokVideoRequestIDCompatibility(t *testing.T) {
	for _, body := range []string{
		`{"request_id":"grok-task-123"}`,
		`{"data":{"request_id":"grok-task-123"}}`,
		`{"video":{"request_id":"grok-task-123"}}`,
		`{"task_id":"grok-task-123","request_id":"trace-id"}`,
	} {
		require.Equal(t, "grok-task-123", extractOpenAIVideoTaskID([]byte(body)))
	}
}

func TestOpenAIGrokVideoBillingUsesConfiguredModeAndDefaults(t *testing.T) {
	groupID := int64(125)
	for _, model := range []string{"grok-imagine-video", "grok-imagine-video-1.5"} {
		defaultPrice := defaultGrokImagineVideoPrice720P
		if model == "grok-imagine-video-1.5" {
			defaultPrice = defaultGrokImagineVideo15Price720P
		}
		for _, tc := range []struct {
			name      string
			mode      BillingMode
			wantTotal float64
		}{
			{"per_request", BillingModePerRequest, 3},
			{"per_second", BillingModeVideo, 24},
			{"default", "", defaultPrice * 8},
		} {
			t.Run(model+"_"+tc.name, func(t *testing.T) {
				svc := &OpenAIGatewayService{cfg: &config.Config{}, billingService: NewBillingService(&config.Config{}, nil)}
				if tc.mode != "" {
					svc.resolver = newOpenAIPerRequestChannelPricingResolverForTest(t, groupID, model, 3)
					cache := svc.resolver.channelService.cache.Load().(*channelCache)
					cache.pricingByGroupModel[channelModelKey{groupID: groupID, model: model}].BillingMode = tc.mode
				}
				cost, err := svc.EstimateOpenAIVideoCreateCost(context.Background(), &APIKey{
					GroupID: &groupID,
					Group:   &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 0.5},
				}, &User{ID: 42}, model, []byte(fmt.Sprintf(`{"model":%q,"duration":8,"resolution":"720p"}`, model)))

				require.NoError(t, err)
				require.InDelta(t, tc.wantTotal, cost.TotalCost, 1e-12)
				require.InDelta(t, tc.wantTotal*0.5, cost.ActualCost, 1e-12)
				if tc.mode != "" {
					require.Equal(t, string(tc.mode), cost.BillingMode)
				}
			})
		}
	}
}

func TestOpenAIGrokVideoJSONImageMayOmitPrompt(t *testing.T) {
	for _, model := range []string{"grok-imagine-video", "grok-imagine-video-1.5"} {
		for _, imageURL := range []string{"https://example.com/frame.png", "data:image/png;base64,aW1hZ2U=", "aW1hZ2U="} {
			body := []byte(fmt.Sprintf(`{"model":%q,"image":{"url":%q},"seconds":15,"size":"1280x720","generate_audio":false}`, model, imageURL))
			parsed, err := ParseOpenAIVideoCreateRequest(body)
			require.NoError(t, err)
			require.Empty(t, parsed.Prompt)
			require.Equal(t, body, parsed.Body)
			resolution, seconds := OpenAIVideoBillingParameters(model, parsed.BillingBody)
			require.Equal(t, "720p", resolution)
			require.Equal(t, 15, seconds)
		}
		for _, image := range []string{`null`, `{}`, `{"url":""}`, `{"data":"aW1hZ2U=","mime_type":"image/png"}`} {
			_, err := ParseOpenAIVideoCreateRequest([]byte(fmt.Sprintf(`{"model":%q,"prompt":"waves","image":%s}`, model, image)))
			require.ErrorContains(t, err, "image.url")
		}
		_, err := ParseOpenAIVideoCreateRequest([]byte(fmt.Sprintf(`{"model":%q}`, model)))
		require.Error(t, err)
	}
}

func TestOpenAIGrokVideoMultipartPreservesInputReferenceAndBillingAliases(t *testing.T) {
	for _, model := range []string{"grok-imagine-video", "grok-imagine-video-1.5"} {
		t.Run(model, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			for _, field := range [][2]string{{"model", model}, {"seconds", "15"}, {"size", "1280x720"}, {"generate_audio", "false"}} {
				require.NoError(t, writer.WriteField(field[0], field[1]))
			}
			file, err := writer.CreateFormFile("input_reference", "frame.png")
			require.NoError(t, err)
			_, err = file.Write([]byte("\x89PNG\r\n\x1a\nimage-bytes"))
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			contentType := writer.FormDataContentType()
			parsed, err := ParseOpenAIVideoOperationRequestWithContentType(OpenAIVideoEndpointCreate, body.Bytes(), contentType)
			require.NoError(t, err)
			require.Empty(t, parsed.Prompt)

			account := newOpenAIVideoTestAccount()
			account.Credentials["model_mapping"] = map[string]any{model: "grok-imagine-video-1.5"}
			upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"id":"grok-multipart-task","status":"queued","seconds":"15"}`)}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body.Bytes()))
			result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", parsed.Body, contentType, model)
			require.NoError(t, err)
			require.Equal(t, "https://video.test/v1/videos", upstream.lastReq.URL.String())
			require.Equal(t, "grok-multipart-task", result.ResponseID)
			require.Equal(t, 15, result.VideoDurationSeconds)
			require.Equal(t, "720p", result.VideoResolution)
			require.Equal(t, contentType, upstream.lastReq.Header.Get("Content-Type"))
			_, params, err := mime.ParseMediaType(contentType)
			require.NoError(t, err)
			form, err := multipart.NewReader(bytes.NewReader(upstream.lastBody), params["boundary"]).ReadForm(1 << 20)
			require.NoError(t, err)
			t.Cleanup(func() { _ = form.RemoveAll() })
			require.Equal(t, []string{"grok-imagine-video-1.5"}, form.Value["model"])
			require.Equal(t, []string{"false"}, form.Value["generate_audio"])
			require.Equal(t, []string{"15"}, form.Value["seconds"])
			require.Equal(t, []string{"1280x720"}, form.Value["size"])
			require.Len(t, form.File["input_reference"], 1)
			require.Equal(t, "frame.png", form.File["input_reference"][0].Filename)
			image, err := form.File["input_reference"][0].Open()
			require.NoError(t, err)
			defer image.Close()
			imageBytes, err := io.ReadAll(image)
			require.NoError(t, err)
			require.Equal(t, []byte("\x89PNG\r\n\x1a\nimage-bytes"), imageBytes)
		})
	}
}

func TestOpenAIGrokVideoSizeAliasBilling(t *testing.T) {
	for _, tc := range []struct {
		body, resolution string
		seconds          int
	}{
		{`{}`, "480p", 8},
		{`{"size":"848x480","duration":1}`, "480p", 1},
		{`{"size":"544x544"}`, "480p", 8},
		{`{"size":"1280x720","seconds":"15"}`, "720p", 15},
		{`{"size":"944x960"}`, "720p", 8},
		{`{"size":"1920x1088"}`, "1080p", 8},
		{`{"size":"1440x1440"}`, "1080p", 8},
		{`{"resolution":"1080p","size":"1280x720","duration":5,"seconds":8}`, "1080p", 5},
	} {
		resolution, seconds := OpenAIVideoBillingParameters("grok-imagine-video-1.5", []byte(tc.body))
		require.Equal(t, tc.resolution, resolution, tc.body)
		require.Equal(t, tc.seconds, seconds, tc.body)
	}
	_, err := ParseOpenAIVideoCreateRequest([]byte(`{"model":"grok-imagine-video-1.5","prompt":"waves","size":"unknown"}`))
	require.ErrorContains(t, err, "resolution")
	svc := &OpenAIGatewayService{cfg: &config.Config{}, billingService: NewBillingService(&config.Config{}, nil)}
	groupID := int64(125)
	cost, err := svc.EstimateOpenAIVideoCreateCost(context.Background(), &APIKey{
		GroupID: &groupID,
		Group:   &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 0.5},
	}, &User{ID: 42}, "grok-imagine-video-1.5", []byte(`{"size":"1280x720","seconds":15}`))
	require.NoError(t, err)
	require.InDelta(t, defaultGrokImagineVideo15Price720P*15, cost.TotalCost, 1e-12)
	require.InDelta(t, defaultGrokImagineVideo15Price720P*15*0.5, cost.ActualCost, 1e-12)
}

func TestOpenAIGrokVideoGenerationsEndpointPreserved(t *testing.T) {
	for _, model := range []string{"grok-imagine-video", "grok-imagine-video-1.5"} {
		for _, path := range []string{"/v1/videos/generations", "/videos/generations", "/videos"} {
			body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"waves"}`, model))
			upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"request_id":"grok-task-123"}`)}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
			result, err := svc.ForwardOpenAIVideo(context.Background(), c, newOpenAIVideoTestAccount(), OpenAIVideoEndpointCreate, "", "", body, "application/json", model)
			require.NoError(t, err)
			wantPath := "/v1" + strings.TrimPrefix(path, "/v1")
			require.Equal(t, "https://video.test"+wantPath, upstream.lastReq.URL.String())
			require.Equal(t, wantPath, result.UpstreamEndpoint)
		}
	}
}

func TestOpenAIGrokVideoMultipartRequiresPromptOrImageFile(t *testing.T) {
	for _, tc := range []struct {
		prompt, field, content string
		valid                  bool
	}{
		{"waves", "", "", true},
		{"", "", "", false},
		{"", "image", "png-bytes", false},
		{"", "input_reference", "", false},
	} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		require.NoError(t, writer.WriteField("model", "grok-imagine-video-1.5"))
		require.NoError(t, writer.WriteField("prompt", tc.prompt))
		if tc.field != "" {
			file, err := writer.CreateFormFile(tc.field, "frame.png")
			require.NoError(t, err)
			_, err = io.WriteString(file, tc.content)
			require.NoError(t, err)
		}
		require.NoError(t, writer.Close())
		_, err := ParseOpenAIVideoOperationRequestWithContentType(OpenAIVideoEndpointCreate, body.Bytes(), writer.FormDataContentType())
		if tc.valid {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, "prompt or an input_reference")
		}
	}
}

func TestOpenAIGrokVideoStandardCreateQueryAndDownload(t *testing.T) {
	for _, model := range []string{"grok-imagine-video", "grok-imagine-video-1.5"} {
		t.Run(model, func(t *testing.T) {
			account := newOpenAIVideoTestAccount()
			upstream := &httpUpstreamRecorder{}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			for _, step := range []struct {
				endpoint                            OpenAIVideoEndpoint
				method, path, response, contentType string
			}{
				{OpenAIVideoEndpointCreate, http.MethodPost, "/v1/videos", `{"id":"video-123","object":"video","status":"queued","seconds":"8"}`, "application/json"},
				{OpenAIVideoEndpointStatus, http.MethodGet, "/v1/videos/video-123", `{"id":"video-123","status":"in_progress","progress":50,"seconds":"8"}`, "application/json"},
				{OpenAIVideoEndpointStatus, http.MethodGet, "/v1/videos/video-123", `{"id":"video-123","status":"completed","progress":100,"seconds":"8"}`, "application/json"},
				{OpenAIVideoEndpointContent, http.MethodGet, "/v1/videos/video-123/content", "mp4-bytes", "video/mp4"},
			} {
				body := []byte(nil)
				taskID := "video-123"
				if step.endpoint == OpenAIVideoEndpointCreate {
					body = []byte(fmt.Sprintf(`{"model":%q,"prompt":"waves","duration":8}`, model))
					taskID = ""
				}
				upstream.resp = &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{step.contentType}}, Body: io.NopCloser(strings.NewReader(step.response))}
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(step.method, step.path, bytes.NewReader(body))
				result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, step.endpoint, taskID, taskID, body, "application/json", model)
				require.NoError(t, err)
				require.Equal(t, "https://video.test"+step.path, upstream.lastReq.URL.String())
				require.Equal(t, step.method, upstream.lastReq.Method)
				require.Equal(t, "Bearer video-key", upstream.lastReq.Header.Get("Authorization"))
				require.Equal(t, "video-123", result.ResponseID)
				require.Equal(t, step.response, recorder.Body.String())
				if step.endpoint == OpenAIVideoEndpointStatus {
					require.False(t, gjson.GetBytes(result.ResponseBody, "url").Exists())
					require.Equal(t, strings.Contains(step.response, "completed"), OpenAIVideoStatusSucceeded(result.ResponseBody))
				}
				if step.endpoint != OpenAIVideoEndpointCreate {
					require.Zero(t, result.VideoCount)
				}
			}
		})
	}
}
