package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestParseOpenAIVideoCreateRequest(t *testing.T) {
	parsed, err := ParseOpenAIVideoCreateRequest([]byte(`{"model":"video-ds-2.0-fast","prompt":"waves","seconds":15}`))
	require.NoError(t, err)
	require.Equal(t, "video-ds-2.0-fast", parsed.Model)
	require.Equal(t, "waves", parsed.Prompt)

	parsed, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"MiniMax-H3","content":[{"type":"text","text":"cinematic waves"},{"type":"image_url","image_url":{"url":"https://example.com/frame.jpg"},"role":"first_frame"}],"resolution":"2K","duration":5,"ratio":"adaptive"}`))
	require.NoError(t, err)
	require.Equal(t, "MiniMax-H3", parsed.Model)
	require.Equal(t, "cinematic waves", parsed.Prompt)

	parsed, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"sora-2-landscape-8s","prompt":"waves","image_url":"https://example.com/reference.jpg"}`))
	require.NoError(t, err)
	require.Equal(t, "sora-2-landscape-8s", parsed.Model)
	require.Equal(t, "waves", parsed.Prompt)

	parsed, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"as-sd2.0-fast","prompt":"waves","seconds":"15"}`))
	require.NoError(t, err)
	require.Equal(t, "as-sd2.0-fast", parsed.Model)
	require.Equal(t, "waves", parsed.Prompt)

	parsed, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"firefly-video-v2-fast","prompt":"waves","duration":5,"resolution":"720p"}`))
	require.NoError(t, err)
	require.Equal(t, "firefly-video-v2-fast", parsed.Model)
	require.Equal(t, "waves", parsed.Prompt)

	parsed, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"firefly-video-v2","duration":10,"shots":[{"prompt":"wide shot","duration":5},{"prompt":"close shot","duration":5}]}`))
	require.NoError(t, err)
	require.Equal(t, "wide shot\nclose shot", parsed.Prompt)

	_, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"video-v1-15s","prompt":"waves"}`))
	require.ErrorContains(t, err, "firefly-video-v2")

	_, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"gpt-image-2","prompt":"waves"}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "sora-2-landscape-8s")
}

func TestParseOpenAIVideoOperationRequestValidatesMiniMaxAdvancedInputs(t *testing.T) {
	parsed, err := ParseOpenAIVideoOperationRequest(OpenAIVideoEndpointRegenerate, []byte(`{"model":"MiniMax-H3","source_task_id":"task-source","resolution":"2K"}`))
	require.NoError(t, err)
	require.Equal(t, "task-source", parsed.SourceTaskID)

	_, err = ParseOpenAIVideoOperationRequest(OpenAIVideoEndpointRegenerate, []byte(`{"model":"MiniMax-H3","resolution":"2K"}`))
	require.ErrorContains(t, err, "source_task_id")

	_, err = ParseOpenAIVideoCreateRequest([]byte(`{"model":"MiniMax-H3","prompt":"waves","images":["https://example.com/frame.jpg"],"resolution":"2K","duration":5,"ratio":"adaptive"}`))
	require.ErrorContains(t, err, "native content[]")
}

func TestAccountSupportsMiniMaxVideoOnlyWhenExplicitlyEnabled(t *testing.T) {
	account := newOpenAIVideoTestAccount()
	require.False(t, AccountUsesMiniMaxV2Video(account))
	require.False(t, AccountSupportsOpenAIVideoEndpoint(account, "MiniMax-H3"))

	account.Credentials["openai_capabilities"] = []any{"minimax_video"}
	require.True(t, AccountUsesMiniMaxV2Video(account))
	require.True(t, AccountSupportsOpenAIVideoEndpoint(account, "MiniMax-H3"))

	account.Type = AccountTypeOAuth
	require.False(t, AccountUsesMiniMaxV2Video(account))
}

func TestOpenAIVideoStatusFailed(t *testing.T) {
	require.True(t, OpenAIVideoStatusFailed([]byte(`{"status":"FAILED"}`)))
	require.True(t, OpenAIVideoStatusFailed([]byte(`{"data":{"status":"cancelled"}}`)))
	require.True(t, OpenAIVideoStatusFailed([]byte(`{"task":{"status":"error"}}`)))
	require.True(t, OpenAIVideoStatusFailed([]byte(`{"error":{"message":"generation failed: generate error: An error occurred.","type":"server_error"}}`)))
	require.True(t, OpenAIVideoStatusFailed([]byte(`{"error":{"message":"generation failed: upstream task timed out","type":"upstream_error"}}`)))
	require.True(t, OpenAIVideoStatusFailed([]byte(`{"error":{"message":"generation failed: upstream task timed out"}}`)))
	require.True(t, OpenAIVideoStatusFailed([]byte(`generation failed: upstream task timed out`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`{"error":{"message":"upstream temporarily unavailable","type":"server_error"}}`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`{"error":{"message":"generation failed: invalid prompt","type":"invalid_request_error"}}`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`502 Bad Gateway`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`503 Service Unavailable`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`504 Gateway Timeout`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`{"status":"SUCCESS"}`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`{"status":"completed"}`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`{"status":"queued"}`)))
	require.False(t, OpenAIVideoStatusFailed([]byte(`not-json`)))
}

func TestOpenAIVideoStatusSucceeded(t *testing.T) {
	require.True(t, OpenAIVideoStatusSucceeded([]byte(`{"status":"completed"}`)))
	require.True(t, OpenAIVideoStatusSucceeded([]byte(`{"data":{"status":"SUCCESS"}}`)))
	require.True(t, OpenAIVideoStatusSucceeded([]byte(`{"task":{"status":"succeeded"}}`)))
	require.True(t, OpenAIVideoStatusSucceeded([]byte(`{"result":{"status":"finished"}}`)))
	require.False(t, OpenAIVideoStatusSucceeded([]byte(`{"status":"failed"}`)))
	require.False(t, OpenAIVideoStatusSucceeded([]byte(`{"status":"processing"}`)))
	require.False(t, OpenAIVideoStatusSucceeded([]byte(`not-json`)))
}

func TestEstimateOpenAIVideoCreateCostUsesPerRequestPricing(t *testing.T) {
	groupID := int64(125)
	price := 30.0
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		billingService: NewBillingService(&config.Config{}, nil),
		resolver:       newOpenAIPerRequestChannelPricingResolverForTest(t, groupID, "video-ds-2.0-fast", price),
	}

	cost, err := svc.EstimateOpenAIVideoCreateCost(context.Background(), &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID:             groupID,
			RateMultiplier: 1,
		},
	}, &User{ID: 42}, "video-ds-2.0-fast", []byte(`{"model":"video-ds-2.0-fast","seconds":"15"}`))

	require.NoError(t, err)
	require.NotNil(t, cost)
	require.Equal(t, string(BillingModePerRequest), cost.BillingMode)
	require.InDelta(t, price, cost.ActualCost, 1e-12)
}

func TestEstimateOpenAIVideoCreateCostRejectsMissingPricing(t *testing.T) {
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		billingService: NewBillingService(&config.Config{}, nil),
	}

	cost, err := svc.EstimateOpenAIVideoCreateCost(context.Background(), &APIKey{
		Group: &Group{RateMultiplier: 1},
	}, &User{ID: 42}, "video-ds-2.0-fast", []byte(`{"model":"video-ds-2.0-fast","seconds":"15"}`))

	require.ErrorIs(t, err, ErrModelPricingUnavailable)
	require.Nil(t, cost)
}

func TestOpenAIVideoBillingParametersUsesFireflyDefaultsAndExplicitValues(t *testing.T) {
	resolution, duration := OpenAIVideoBillingParameters("firefly-video-v2-fast", []byte(`{"model":"firefly-video-v2-fast","prompt":"waves"}`))
	require.Equal(t, VideoBillingResolution720P, resolution)
	require.Equal(t, 5, duration)

	resolution, duration = OpenAIVideoBillingParameters("firefly-video-v2", []byte(`{"model":"firefly-video-v2","resolution":"1080p","duration":"15"}`))
	require.Equal(t, VideoBillingResolution1080P, resolution)
	require.Equal(t, 15, duration)
}

func TestForwardOpenAIVideoCreateUsesVideosEndpointAndExtractsTaskID(t *testing.T) {

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"video-ds-2.0-fast","prompt":"waves","seconds":15}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	account := newOpenAIVideoTestAccount()
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Request-Id": []string{"video-create-req"},
		},
		Body: io.NopCloser(strings.NewReader(`{"task_id":"task-123","status":"queued"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", "video-ds-2.0-fast")

	require.NoError(t, err)
	require.Equal(t, "https://video.test/v1/videos", upstream.lastReq.URL.String())
	require.Equal(t, http.MethodPost, upstream.lastReq.Method)
	require.Equal(t, "Bearer video-key", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "application/json", upstream.lastReq.Header.Get("Content-Type"))
	require.JSONEq(t, `{"model":"video-ds-2.0-fast","prompt":"waves","seconds":15}`, string(upstream.lastBody))
	require.Equal(t, VideoBillingResolution720P, result.VideoResolution)
	require.Equal(t, 15, result.VideoDurationSeconds)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"task_id":"task-123","status":"queued"}`, recorder.Body.String())
	require.Equal(t, "task-123", result.ResponseID)
	require.Equal(t, "video-create-req", result.RequestID)
}

func TestForwardOpenAIVideoMiniMaxCreateUsesV2EndpointAndPreservesContent(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"MiniMax-H3","content":[{"type":"text","text":"cinematic waves"},{"type":"image_url","image_url":{"url":"https://example.com/frame.jpg"},"role":"first_frame"}],"resolution":"2K","duration":5,"ratio":"adaptive"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	account := newMiniMaxVideoTestAccount()
	upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"task_id":"minimax-task-1"}`)}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", "MiniMax-H3")

	require.NoError(t, err)
	require.Equal(t, "https://metaso.test/api/minimax/v2/video_generation", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer video-key", upstream.lastReq.Header.Get("Authorization"))
	require.JSONEq(t, string(body), string(upstream.lastBody))
	require.Equal(t, "minimax-task-1", result.ResponseID)
	require.Equal(t, "MiniMax-H3", result.UpstreamModel)
	require.Equal(t, "/api/minimax/v2/video_generation", result.UpstreamEndpoint)
	require.Equal(t, "2K", result.VideoResolution)
	require.Equal(t, 5, result.VideoDurationSeconds)
}

func TestForwardOpenAIVideoMiniMaxPromptCompatibilityBuildsTextContent(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"MiniMax-H3","prompt":"cinematic waves","resolution":"768P","seconds":"4","aspect_ratio":"16:9"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))

	upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"task_id":"minimax-task-2"}`)}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	_, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), OpenAIVideoEndpointCreate, "", "", body, "application/json", "MiniMax-H3")

	require.NoError(t, err)
	require.JSONEq(t, `{"model":"MiniMax-H3","content":[{"type":"text","text":"cinematic waves"}],"resolution":"768P","duration":4,"ratio":"16:9"}`, string(upstream.lastBody))
}

func TestForwardOpenAIVideoMiniMaxCreate504ReturnsRecoverableTaskWithoutFailover(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"MiniMax-H3","content":[{"type":"text","text":"cinematic waves"}],"resolution":"2K","duration":8,"ratio":"16:9"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusGatewayTimeout, `{"error":{"message":"upstream timed out"}}`)}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), OpenAIVideoEndpointCreate, "", "task_recovery_local", body, "application/json", "MiniMax-H3")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.SubmissionUncertain)
	require.Equal(t, "task_recovery_local", result.ResponseID)
	require.Equal(t, "task_recovery_local", result.BillingTaskID)
	require.Equal(t, "2k", strings.ToLower(result.VideoResolution))
	require.Equal(t, 8, result.VideoDurationSeconds)
	require.Empty(t, recorder.Body.String(), "handler must persist recovery before writing the normal task response")
}

func TestForwardOpenAIVideoMiniMaxCreateWrapped504ReturnsRecoverableTaskWithoutFailover(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"MiniMax-H3","content":[{"type":"text","text":"cinematic waves"}],"resolution":"2K","duration":8,"ratio":"16:9"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	upstreamBody := `{"error":{"http_code":"504","message":"Upstream request timed out","type":"gateway_timeout"},"request_id":"655e3798-ba5b-46a1-9393-c9df0733e815","type":"error"}`
	upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, upstreamBody)}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), OpenAIVideoEndpointCreate, "", "task_recovery_local", body, "application/json", "MiniMax-H3")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.SubmissionUncertain)
	require.Equal(t, "task_recovery_local", result.ResponseID)
	require.Equal(t, "task_recovery_local", result.BillingTaskID)
	require.Equal(t, "655e3798-ba5b-46a1-9393-c9df0733e815", result.RequestID)
	require.Empty(t, recorder.Body.String(), "handler must return the local recovery task instead of the wrapped error")
}

func TestIsMiniMaxWrappedGatewayTimeoutStrict(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "string 504", body: `{"error":{"http_code":"504","message":"timed out"}}`, want: true},
		{name: "numeric 504", body: `{"error":{"http_code":504,"message":"timed out"}}`, want: true},
		{name: "successful task", body: `{"task_id":"2087727370396942336","status":"queued"}`},
		{name: "different wrapped error", body: `{"error":{"http_code":"400","message":"invalid request"}}`},
		{name: "top level status is not accepted", body: `{"http_code":"504","message":"timed out"}`},
		{name: "invalid json", body: `gateway timeout`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isMiniMaxWrappedGatewayTimeout([]byte(tt.body)))
		})
	}
}

func TestForwardOpenAIVideoMiniMaxRecoveredStatusUsesUpstreamIDAndReturnsClientID(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task_recovery_local", nil)
	upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"task":{"id":"upstream-task-1","status":"succeeded","content":{"url":"https://cdn.example.com/video.mp4"}}}`)}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), OpenAIVideoEndpointStatus, "upstream-task-1", "task_recovery_local", nil, "", "MiniMax-H3")

	require.NoError(t, err)
	require.Equal(t, "/api/minimax/v2/query/video_generation/upstream-task-1", upstream.lastReq.URL.Path)
	require.Equal(t, "task_recovery_local", result.ResponseID)
	require.Equal(t, "task_recovery_local", gjson.GetBytes(recorder.Body.Bytes(), "task.id").String())
	require.Equal(t, "/v1/videos/task_recovery_local/content", gjson.GetBytes(recorder.Body.Bytes(), "task.content.url").String())
}

func TestForwardOpenAIVideoMiniMaxAdvancedCreateEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		endpoint OpenAIVideoEndpoint
		body     string
		path     string
	}{
		{name: "context ir", endpoint: OpenAIVideoEndpointContext, body: `{"model":"MiniMax-H3","content":[{"type":"text","text":"describe the scene"}],"duration":5,"ratio":"16:9"}`, path: "/api/minimax/v2/h3_context_ir"},
		{name: "regeneration", endpoint: OpenAIVideoEndpointRegenerate, body: `{"model":"MiniMax-H3","source_task_id":"source-1","resolution":"2K"}`, path: "/api/minimax/v2/video_regeneration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", strings.NewReader(tt.body))
			upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"task_id":"advanced-task"}`)}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

			result, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), tt.endpoint, "", "", []byte(tt.body), "application/json", "MiniMax-H3")

			require.NoError(t, err)
			require.Equal(t, tt.path, upstream.lastReq.URL.Path)
			require.Equal(t, "advanced-task", result.ResponseID)
		})
	}
}

func TestForwardOpenAIVideoMiniMaxStatusAndListRewriteContentURLs(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   OpenAIVideoEndpoint
		taskID     string
		requestURL string
		response   string
		wantPath   string
		wantBody   string
	}{
		{
			name:       "status",
			endpoint:   OpenAIVideoEndpointStatus,
			taskID:     "task-1",
			requestURL: "/v1/videos/task-1",
			response:   `{"task":{"id":"task-1","status":"succeeded","content":{"url":"https://cdn.example.com/video.mp4"}}}`,
			wantPath:   "/api/minimax/v2/query/video_generation/task-1",
			wantBody:   `{"task":{"id":"task-1","status":"succeeded","content":{"url":"/v1/videos/task-1/content"}}}`,
		},
		{
			name:       "list",
			endpoint:   OpenAIVideoEndpointList,
			requestURL: "/v1/videos?page_num=1&page_size=20",
			response:   `{"items":[{"id":"task-2","status":"succeeded","content":{"url":"https://cdn.example.com/video.mp4"}}],"total":1}`,
			wantPath:   "/api/minimax/v2/query/video_generation",
			wantBody:   `{"items":[{"id":"task-2","status":"succeeded","content":{"url":"/v1/videos/task-2/content"}}],"total":1}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, tt.requestURL, nil)
			upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, tt.response)}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

			result, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), tt.endpoint, tt.taskID, tt.taskID, nil, "", "MiniMax-H3")

			require.NoError(t, err)
			require.Equal(t, tt.wantPath, upstream.lastReq.URL.Path)
			require.JSONEq(t, tt.wantBody, recorder.Body.String())
			require.JSONEq(t, tt.response, string(result.ResponseBody))
		})
	}
}

func TestForwardOpenAIVideoMiniMaxDeleteMarksPreviouslyFailedTask(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodDelete, "/v1/videos/task-failed", nil)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		jsonVideoResponse(http.StatusOK, `{"task":{"id":"task-failed","status":"failed"}}`),
		jsonVideoResponse(http.StatusOK, `{"task_id":"task-failed","action":"deleted","status":"deleted"}`),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), OpenAIVideoEndpointDelete, "task-failed", "task-failed", nil, "", "")

	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, http.MethodGet, upstream.requests[0].Method)
	require.Equal(t, "/api/minimax/v2/query/video_generation/task-failed", upstream.requests[0].URL.Path)
	require.Equal(t, http.MethodDelete, upstream.requests[1].Method)
	require.Equal(t, "/api/minimax/v2/video_generation/task-failed", upstream.requests[1].URL.Path)
	require.True(t, result.TaskTerminalFailure)
}

func TestForwardOpenAIVideoMiniMaxContentQueriesStatusAndFetchesUnsignedURL(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task-1/content", nil)
	c.Request.Header.Set("Range", "bytes=0-3")
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		jsonVideoResponse(http.StatusOK, `{"task":{"id":"task-1","status":"succeeded","content":{"url":"https://cdn.example.com/video.mp4"}}}`),
		{
			StatusCode: http.StatusPartialContent,
			Header: http.Header{
				"Content-Type":  []string{"video/mp4"},
				"Content-Range": []string{"bytes 0-3/8"},
			},
			Body: io.NopCloser(strings.NewReader("mp4!")),
		},
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, newMiniMaxVideoTestAccount(), OpenAIVideoEndpointContent, "task-1", "task-1", nil, "", "")

	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "Bearer video-key", upstream.requests[0].Header.Get("Authorization"))
	require.Empty(t, upstream.requests[1].Header.Get("Authorization"))
	require.Equal(t, "bytes=0-3", upstream.requests[1].Header.Get("Range"))
	require.Equal(t, "https://cdn.example.com/video.mp4", upstream.requests[1].URL.String())
	require.Equal(t, "mp4!", recorder.Body.String())
	require.Equal(t, "task-1", result.ResponseID)
}

func TestForwardOpenAIVideoCreateAllowsSora2Models(t *testing.T) {

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"sora-2-landscape-8s","prompt":"waves","image_url":"https://example.com/reference.jpg","size":"1920x1080"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	account := newOpenAIVideoTestAccount()
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"task_id":"sora-task-123","status":"queued"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", "sora-2-landscape-8s")

	require.NoError(t, err)
	require.Equal(t, "https://video.test/v1/videos", upstream.lastReq.URL.String())
	require.JSONEq(t, `{"model":"sora-2-landscape-8s","prompt":"waves","image_url":"https://example.com/reference.jpg","size":"1920x1080"}`, string(upstream.lastBody))
	require.Equal(t, "sora-task-123", result.ResponseID)
	require.JSONEq(t, `{"task_id":"sora-task-123","status":"queued"}`, recorder.Body.String())
}

func TestForwardOpenAIVideoCreateAllowsASSD20FastUpstreamMapping(t *testing.T) {

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"video-ds-2.0-fast","prompt":"waves","seconds":"15"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	account := newOpenAIVideoTestAccount()
	account.Credentials["model_mapping"] = map[string]any{
		"video-ds-2.0-fast": "as-sd2.0-fast",
	}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"task_id":"as-task-123","status":"queued"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", "video-ds-2.0-fast")

	require.NoError(t, err)
	require.Equal(t, "https://video.test/v1/videos", upstream.lastReq.URL.String())
	require.JSONEq(t, `{"model":"as-sd2.0-fast","prompt":"waves","seconds":"15"}`, string(upstream.lastBody))
	require.Equal(t, "as-task-123", result.ResponseID)
	require.JSONEq(t, `{"task_id":"as-task-123","status":"queued"}`, recorder.Body.String())
}

func TestForwardOpenAIVideoCreateAllowsFireflyV2JSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"firefly-video-v2-fast","prompt":"waves","duration":5,"resolution":"720p"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	account := newOpenAIVideoTestAccount()
	account.Credentials["base_url"] = "https://ycyapi.cn"
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"task_id":"firefly-task-123","status":"queued","seconds":"5"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", "firefly-video-v2-fast")

	require.NoError(t, err)
	require.Equal(t, "https://ycyapi.cn/v1/videos", upstream.lastReq.URL.String())
	require.Equal(t, http.MethodPost, upstream.lastReq.Method)
	require.JSONEq(t, string(body), string(upstream.lastBody))
	require.Equal(t, "firefly-task-123", result.ResponseID)
	require.Equal(t, "firefly-video-v2-fast", result.UpstreamModel)
	require.Equal(t, VideoBillingResolution720P, result.VideoResolution)
	require.Equal(t, 5, result.VideoDurationSeconds)
}

func TestForwardOpenAIVideoCreatePreservesFireflyMultipartFilesAndMapsModel(t *testing.T) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	require.NoError(t, writer.WriteField("model", "firefly-video-v2-fast"))
	require.NoError(t, writer.WriteField("prompt", "preserve the subject"))
	require.NoError(t, writer.WriteField("duration", "15"))
	require.NoError(t, writer.WriteField("resolution", "1080p"))
	for _, file := range []struct {
		field    string
		filename string
		body     string
	}{
		{field: "images", filename: "one.png", body: "png-one"},
		{field: "images", filename: "two.jpg", body: "jpeg-two"},
		{field: "videos", filename: "motion.mp4", body: "mp4-data"},
		{field: "audios", filename: "sound.wav", body: "wav-data"},
	} {
		part, err := writer.CreateFormFile(file.field, file.filename)
		require.NoError(t, err)
		_, err = io.WriteString(part, file.body)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	contentType := writer.FormDataContentType()

	parsed, err := ParseOpenAIVideoOperationRequestWithContentType(OpenAIVideoEndpointCreate, requestBody.Bytes(), contentType)
	require.NoError(t, err)
	require.Equal(t, "firefly-video-v2-fast", parsed.Model)
	require.JSONEq(t, `{"model":"firefly-video-v2-fast","duration":"15","resolution":"1080p"}`, string(parsed.BillingBody))

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(requestBody.Bytes()))
	c.Request.Header.Set("Content-Type", contentType)

	account := newOpenAIVideoTestAccount()
	account.Credentials["base_url"] = "https://ycyapi.cn"
	account.Credentials["model_mapping"] = map[string]any{
		"firefly-video-v2-fast": "firefly-video-v2",
	}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"task_id":"firefly-multipart-task","status":"queued"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", requestBody.Bytes(), contentType, "firefly-video-v2-fast")

	require.NoError(t, err)
	require.Equal(t, "https://ycyapi.cn/v1/videos", upstream.lastReq.URL.String())
	require.Equal(t, contentType, upstream.lastReq.Header.Get("Content-Type"))
	require.Equal(t, "firefly-multipart-task", result.ResponseID)
	require.Equal(t, VideoBillingResolution1080P, result.VideoResolution)
	require.Equal(t, 15, result.VideoDurationSeconds)

	mediaType, params, err := mime.ParseMediaType(upstream.lastReq.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/form-data", mediaType)
	multipartReader := multipart.NewReader(bytes.NewReader(upstream.lastBody), params["boundary"])
	fields := map[string][]string{}
	files := map[string][]string{}
	for {
		part, partErr := multipartReader.NextPart()
		if errors.Is(partErr, io.EOF) {
			break
		}
		require.NoError(t, partErr)
		value, readErr := io.ReadAll(part)
		require.NoError(t, readErr)
		if part.FileName() == "" {
			fields[part.FormName()] = append(fields[part.FormName()], string(value))
		} else {
			files[part.FormName()] = append(files[part.FormName()], part.FileName()+":"+string(value))
		}
	}
	require.Equal(t, []string{"firefly-video-v2"}, fields["model"])
	require.Equal(t, []string{"preserve the subject"}, fields["prompt"])
	require.Equal(t, []string{"one.png:png-one", "two.jpg:jpeg-two"}, files["images"])
	require.Equal(t, []string{"motion.mp4:mp4-data"}, files["videos"])
	require.Equal(t, []string{"sound.wav:wav-data"}, files["audios"])
}

func TestForwardOpenAIVideoStatusUsesTaskEndpointWithoutBody(t *testing.T) {

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task-123", nil)

	account := newOpenAIVideoTestAccount()
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"id":"task-123","status":"completed"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointStatus, "task-123", "task-123", nil, "", "")

	require.NoError(t, err)
	require.Equal(t, "https://video.test/v1/videos/task-123", upstream.lastReq.URL.String())
	require.Equal(t, http.MethodGet, upstream.lastReq.Method)
	require.Empty(t, upstream.lastBody)
	require.Equal(t, "task-123", result.ResponseID)
	require.JSONEq(t, `{"id":"task-123","status":"completed"}`, string(result.ResponseBody))
	require.JSONEq(t, `{"id":"task-123","status":"completed"}`, recorder.Body.String())
}

func TestForwardOpenAIVideoStatusPreservesTerminalErrorResponse(t *testing.T) {
	for _, statusCode := range []int{
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	} {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			taskID := fmt.Sprintf("task-failed-%d", statusCode)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/"+taskID, nil)

			account := newOpenAIVideoTestAccount()
			body := `{"error":{"message":"generation failed: generate error: An error occurred.","type":"server_error"}}`
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: statusCode,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

			result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointStatus, taskID, taskID, nil, "", "")

			require.Error(t, err)
			require.NotNil(t, result)
			require.Equal(t, taskID, result.ResponseID)
			require.JSONEq(t, body, string(result.ResponseBody))
			require.Equal(t, statusCode, recorder.Code)
			require.JSONEq(t, body, recorder.Body.String())
		})
	}
}

func TestForwardOpenAIVideoContentStreamsVideoAndForwardsRange(t *testing.T) {

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task-123/content", nil)
	c.Request.Header.Set("Range", "bytes=0-3")

	account := newOpenAIVideoTestAccount()
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusPartialContent,
		Header: http.Header{
			"Content-Type":  []string{"video/mp4"},
			"Content-Range": []string{"bytes 0-3/8"},
			"Accept-Ranges": []string{"bytes"},
		},
		Body: io.NopCloser(strings.NewReader("mp4!")),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointContent, "task-123", "task-123", nil, "", "")

	require.NoError(t, err)
	require.Equal(t, "https://video.test/v1/videos/task-123/content", upstream.lastReq.URL.String())
	require.Equal(t, http.MethodGet, upstream.lastReq.Method)
	require.Equal(t, "bytes=0-3", upstream.lastReq.Header.Get("Range"))
	require.Equal(t, http.StatusPartialContent, recorder.Code)
	require.Equal(t, "video/mp4", recorder.Header().Get("Content-Type"))
	require.Equal(t, "bytes 0-3/8", recorder.Header().Get("Content-Range"))
	require.Equal(t, "bytes", recorder.Header().Get("Accept-Ranges"))
	require.Equal(t, "mp4!", recorder.Body.String())
	require.Equal(t, "task-123", result.ResponseID)
}

func TestForwardOpenAIVideoContentHEADPreservesHeadersWithoutBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodHead, "/v1/videos/task-123/content", nil)

	account := newOpenAIVideoTestAccount()
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":   []string{"video/mp4"},
			"Content-Length": []string{"12345"},
			"Accept-Ranges":  []string{"bytes"},
		},
		Body: io.NopCloser(strings.NewReader("must-not-be-written")),
	}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

	result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointContent, "task-123", "task-123", nil, "", "")

	require.NoError(t, err)
	require.Equal(t, http.MethodHead, upstream.lastReq.Method)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "video/mp4", recorder.Header().Get("Content-Type"))
	require.Equal(t, "12345", recorder.Header().Get("Content-Length"))
	require.Equal(t, "bytes", recorder.Header().Get("Accept-Ranges"))
	require.Empty(t, recorder.Body.String())
	require.Equal(t, "task-123", result.ResponseID)
}

func TestBindOpenAIVideoTaskAccountPersistsAndCachesWithVideoTTL(t *testing.T) {
	ctx := context.Background()
	groupID := int64(7)
	cache := newOpenAIVideoGatewayCacheStub()
	repo := newOpenAIVideoTaskBindingRepoStub()
	svc := &OpenAIGatewayService{
		cache:                      cache,
		openAIVideoTaskBindingRepo: repo,
	}

	err := svc.BindOpenAIVideoTaskAccount(ctx, &groupID, 42, " task-123 ", 91)

	require.NoError(t, err)
	binding := repo.bindings["7:42:task-123"]
	require.Equal(t, int64(7), binding.GroupID)
	require.Equal(t, int64(42), binding.UserID)
	require.Equal(t, "task-123", binding.TaskID)
	require.Equal(t, int64(91), binding.AccountID)
	require.WithinDuration(t, time.Now().Add(openAIVideoTaskBindingTTL), binding.ExpiresAt, 2*time.Second)

	sessionHash := svc.openAISessionCacheKey(OpenAIVideoTaskSessionHash(42, "task-123"))
	cacheKey := openAIVideoCacheStubKey(groupID, sessionHash)
	require.Equal(t, int64(91), cache.bindings[cacheKey])
	require.Equal(t, openAIVideoTaskBindingTTL, cache.ttls[cacheKey])
}

func TestRestoreOpenAIVideoTaskStickySessionLoadsPersistentBinding(t *testing.T) {
	ctx := context.Background()
	groupID := int64(7)
	cache := newOpenAIVideoGatewayCacheStub()
	repo := newOpenAIVideoTaskBindingRepoStub()
	repo.bindings["7:42:task-123"] = OpenAIVideoTaskBinding{
		GroupID:   groupID,
		UserID:    42,
		TaskID:    "task-123",
		AccountID: 91,
		ExpiresAt: time.Now().Add(openAIVideoTaskBindingTTL),
	}
	svc := &OpenAIGatewayService{
		cache:                      cache,
		openAIVideoTaskBindingRepo: repo,
	}

	accountID, err := svc.RestoreOpenAIVideoTaskStickySession(ctx, &groupID, 42, "task-123")

	require.NoError(t, err)
	require.Equal(t, int64(91), accountID)
	sessionHash := svc.openAISessionCacheKey(OpenAIVideoTaskSessionHash(42, "task-123"))
	cacheKey := openAIVideoCacheStubKey(groupID, sessionHash)
	require.Equal(t, int64(91), cache.bindings[cacheKey])
	require.Equal(t, openAIVideoTaskBindingTTL, cache.ttls[cacheKey])
	require.Equal(t, 1, repo.getCalls)
}

func TestRestoreOpenAIVideoTaskStickySessionDoesNotCrossUsers(t *testing.T) {
	ctx := context.Background()
	groupID := int64(7)
	cache := newOpenAIVideoGatewayCacheStub()
	repo := newOpenAIVideoTaskBindingRepoStub()
	svc := &OpenAIGatewayService{
		cache:                      cache,
		openAIVideoTaskBindingRepo: repo,
	}
	require.NoError(t, svc.BindOpenAIVideoTaskAccount(ctx, &groupID, 42, "task-123", 91))

	accountID, err := svc.RestoreOpenAIVideoTaskStickySession(ctx, &groupID, 43, "task-123")

	require.NoError(t, err)
	require.Zero(t, accountID)
	user42SessionHash := svc.openAISessionCacheKey(OpenAIVideoTaskSessionHash(42, "task-123"))
	user43SessionHash := svc.openAISessionCacheKey(OpenAIVideoTaskSessionHash(43, "task-123"))
	require.Equal(t, int64(91), cache.bindings[openAIVideoCacheStubKey(groupID, user42SessionHash)])
	require.Zero(t, cache.bindings[openAIVideoCacheStubKey(groupID, user43SessionHash)])
	require.Equal(t, 1, repo.getCalls)
}

func newOpenAIVideoTestAccount() *Account {
	return &Account{
		ID:          91,
		Name:        "video",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "video-key",
			"base_url": "https://video.test/v1",
		},
	}
}

func newMiniMaxVideoTestAccount() *Account {
	account := newOpenAIVideoTestAccount()
	account.Credentials["base_url"] = "https://metaso.test/api/minimax"
	account.Credentials["openai_capabilities"] = []any{"minimax_video"}
	return account
}

func jsonVideoResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type openAIVideoGatewayCacheStub struct {
	bindings map[string]int64
	ttls     map[string]time.Duration
}

func newOpenAIVideoGatewayCacheStub() *openAIVideoGatewayCacheStub {
	return &openAIVideoGatewayCacheStub{
		bindings: make(map[string]int64),
		ttls:     make(map[string]time.Duration),
	}
}

func openAIVideoCacheStubKey(groupID int64, sessionHash string) string {
	return fmt.Sprintf("%d:%s", groupID, sessionHash)
}

func (c *openAIVideoGatewayCacheStub) GetSessionAccountID(_ context.Context, groupID int64, sessionHash string) (int64, error) {
	return c.bindings[openAIVideoCacheStubKey(groupID, sessionHash)], nil
}

func (c *openAIVideoGatewayCacheStub) SetSessionAccountID(_ context.Context, groupID int64, sessionHash string, accountID int64, ttl time.Duration) error {
	key := openAIVideoCacheStubKey(groupID, sessionHash)
	c.bindings[key] = accountID
	c.ttls[key] = ttl
	return nil
}

func (c *openAIVideoGatewayCacheStub) RefreshSessionTTL(_ context.Context, groupID int64, sessionHash string, ttl time.Duration) error {
	c.ttls[openAIVideoCacheStubKey(groupID, sessionHash)] = ttl
	return nil
}

func (c *openAIVideoGatewayCacheStub) DeleteSessionAccountID(_ context.Context, groupID int64, sessionHash string) error {
	key := openAIVideoCacheStubKey(groupID, sessionHash)
	delete(c.bindings, key)
	delete(c.ttls, key)
	return nil
}

func (c *openAIVideoGatewayCacheStub) SetGrokVideoPendingBilling(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

func (c *openAIVideoGatewayCacheStub) GetGrokVideoPendingBilling(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}

func (c *openAIVideoGatewayCacheStub) ClaimGrokVideoBilled(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

func (c *openAIVideoGatewayCacheStub) ReleaseGrokVideoBilled(_ context.Context, _ string) error {
	return nil
}

func (c *openAIVideoGatewayCacheStub) SetReasoningContent(_ context.Context, _ string, _ string, _ time.Duration) error {
	return nil
}

func (c *openAIVideoGatewayCacheStub) GetReasoningContent(_ context.Context, _ string) (string, error) {
	return "", ErrReasoningContentNotFound
}

type openAIVideoTaskBindingRepoStub struct {
	bindings    map[string]OpenAIVideoTaskBinding
	upsertCalls int
	getCalls    int
}

func newOpenAIVideoTaskBindingRepoStub() *openAIVideoTaskBindingRepoStub {
	return &openAIVideoTaskBindingRepoStub{bindings: make(map[string]OpenAIVideoTaskBinding)}
}

func (r *openAIVideoTaskBindingRepoStub) UpsertOpenAIVideoTaskBinding(_ context.Context, binding OpenAIVideoTaskBinding) error {
	r.upsertCalls++
	r.bindings[fmt.Sprintf("%d:%d:%s", binding.GroupID, binding.UserID, binding.TaskID)] = binding
	return nil
}

func (r *openAIVideoTaskBindingRepoStub) GetOpenAIVideoTaskBinding(_ context.Context, groupID, userID int64, taskID string) (*OpenAIVideoTaskBinding, error) {
	r.getCalls++
	binding, ok := r.bindings[fmt.Sprintf("%d:%d:%s", groupID, userID, taskID)]
	if !ok {
		return nil, nil
	}
	return &binding, nil
}

func (r *openAIVideoTaskBindingRepoStub) ArmOpenAIVideoTaskCompensation(_ context.Context, groupID, userID int64, taskID string, apiKey *APIKey, nextCheckAt time.Time) error {
	key := fmt.Sprintf("%d:%d:%s", groupID, userID, taskID)
	binding, ok := r.bindings[key]
	if !ok {
		return fmt.Errorf("binding not found")
	}
	binding.APIKeyID = apiKey.ID
	binding.APIKeyQuotaLimited = apiKey.Quota > 0
	binding.APIKeyRateLimited = apiKey.HasRateLimits()
	binding.CompensationStatus = openAIVideoCompensationPending
	binding.NextCheckAt = &nextCheckAt
	r.bindings[key] = binding
	return nil
}

func (r *openAIVideoTaskBindingRepoStub) ClaimOpenAIVideoTaskCompensations(_ context.Context, _, _ time.Time, _ int) ([]OpenAIVideoTaskBinding, error) {
	return nil, nil
}

func (r *openAIVideoTaskBindingRepoStub) UpdateOpenAIVideoTaskCompensation(_ context.Context, id int64, status string, nextCheckAt *time.Time, lastError string) error {
	for key, binding := range r.bindings {
		if binding.ID != id {
			continue
		}
		binding.CompensationStatus = status
		binding.NextCheckAt = nextCheckAt
		if lastError == "" {
			binding.LastCheckError = nil
		} else {
			binding.LastCheckError = &lastError
		}
		r.bindings[key] = binding
		break
	}
	return nil
}
