package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestViraleeVideoCreateAndResultPassthrough(t *testing.T) {
	cases := []struct {
		model, params, resolution, completed string
		seconds                              int
	}{
		{"dola-viraldance2.0", `"duration":4,"ratio":"16:9","image_urls":["https://media.test/a.png"],"video_urls":["https://media.test/a.mp4"],"audio_urls":["https://media.test/a.mp3"],"generate_audio":true`, "720p", `"status":"succeeded","video":{"url":"https://cdn.test/dola20.mp4"}`, 4},
		{"dola-viraldance2.5", `"duration":30,"ratio":"9:16","resolution":"720p","elements":[{"frontal_image_url":"https://media.test/front.png","reference_image_urls":["https://media.test/side.png"]}]`, "720p", `"status":"completed","result_url":"https://cdn.test/dola25.mp4"`, 30},
		{"wan3.0x", `"duration":30,"ratio":"16:9","image_urls":["https://media.test/a.png"],"video_urls":["https://media.test/a.mp4"],"audio_urls":["https://media.test/a.mp3"]`, "720p", `"status":"completed","metadata":{"url":"https://cdn.test/wan.mp4"}`, 30},
		{"wan3.0x-480p", `"duration":2,"ratio":"1:1","images":["https://media.test/a.png"]`, "480p", `"status":"succeeded","url":"https://cdn.test/480.mp4"`, 2},
		{"wan3.0x-1080p", `"duration":30,"ratio":"9:16","videos":["https://media.test/a.mp4"],"audios":["https://media.test/a.mp3"]`, "1080p", `"status":"completed","metadata":{"url":"https://cdn.test/1080.mp4"}`, 30},
		{"viraldance933", `"duration":15,"ratio":"4:3","image_url":"https://media.test/a.png","audio_url":"https://media.test/a.mp3","elements":[{"frontal_image_url":"https://media.test/front.png","reference_image_urls":["https://media.test/side.png"]}]`, "720p", `"status":"succeeded","video":{"url":"https://cdn.test/933.mp4"}`, 15},
		{"viraldance933-fast", `"duration":4,"ratio":"16:9","video_url":"https://media.test/a.mp4","generate_audio":true`, "720p", `"status":"completed","result_url":"https://cdn.test/fast.mp4"`, 4},
		{"viraldance2.5-30", `"duration":30,"aspect_ratio":"9:16","size":"720x1280","start_image_url":"https://media.test/start.png","end_image_url":"https://media.test/end.png","video_reference":[{"url":"https://media.test/a.mp4"}],"audio_reference":[{"url":"https://media.test/a.mp3"}],"async":true`, "720p", `"status":"completed","url":"https://cdn.test/30.mp4"`, 30},
		{"viraldance2.5-15", `"ratio":"16:9","audio_urls":["https://media.test/a.mp3"],"generate_audio":false`, "720p", `"status":"succeeded","video":{"url":"https://cdn.test/15.mp4"}`, 15},
		{"viraldance2.5-480p-15", `"ratio":"1:1","images":["https://media.test/a.png"]`, "480p", `"status":"completed","result_url":"https://cdn.test/15-480.mp4"`, 15},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"@Image1 @图片1 镜头推进",%s}`, tc.model, tc.params))
			parsed, err := ParseOpenAIVideoCreateRequest(body)
			require.NoError(t, err)
			require.Equal(t, body, parsed.Body)
			require.Equal(t, body, parsed.BillingBody)
			account := newOpenAIVideoTestAccount()
			require.True(t, AccountSupportsOpenAIVideoEndpoint(account, tc.model))
			for _, base := range []string{"https://video.test", "https://video.test/v1"} {
				account.Credentials["base_url"] = base
				for _, idField := range []string{"id", "task_id"} {
					response := fmt.Sprintf(`{%q:"task-123","status":"queued"}`, idField)
					upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, response)}
					svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
					recorder := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(recorder)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
					result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", tc.model)
					require.NoError(t, err)
					require.Equal(t, "https://video.test/v1/videos", upstream.lastReq.URL.String())
					require.Equal(t, "Bearer video-key", upstream.lastReq.Header.Get("Authorization"))
					require.JSONEq(t, string(body), string(upstream.lastBody))
					require.JSONEq(t, response, recorder.Body.String())
					require.Equal(t, "task-123", result.ResponseID)
					require.Equal(t, tc.resolution, result.VideoResolution)
					require.Equal(t, tc.seconds, result.VideoDurationSeconds)
				}
			}
			for _, state := range []string{
				`"status":"submitted"`, `"status":"queued"`, `"status":"processing"`, `"status":"in_progress"`, tc.completed,
				`"status":"failed","error":{"message":"generation_failed"}`,
			} {
				response := `{"id":"task-123",` + state + `}`
				upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, response)}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task-123", nil)
				result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointStatus, "task-123", "task-123", nil, "", "")
				require.NoError(t, err)
				require.Equal(t, "https://video.test/v1/videos/task-123", upstream.lastReq.URL.String())
				require.Equal(t, http.MethodGet, upstream.lastReq.Method)
				require.Empty(t, upstream.lastBody)
				require.JSONEq(t, response, recorder.Body.String(), "preserve completed URL and error fields")
				require.Zero(t, result.VideoCount, "polling must not charge for another video")
				require.Equal(t, state == tc.completed, OpenAIVideoStatusSucceeded(result.ResponseBody))
				require.Equal(t, state == `"status":"failed","error":{"message":"generation_failed"}`, result.TaskTerminalFailure)
			}
		})
	}
}

func TestDolaVideoMappedUpstreamModel(t *testing.T) {
	for _, model := range []string{"dola-viraldance2.0", "dola-viraldance2.5"} {
		t.Run(model, func(t *testing.T) {
			body := []byte(`{"model":"viraldance933","prompt":"waves","duration":8,"ratio":"16:9","image_urls":["https://media.test/a.png"]}`)
			account := newOpenAIVideoTestAccount()
			account.Credentials["model_mapping"] = map[string]any{"viraldance933": model}
			require.True(t, AccountSupportsOpenAIVideoEndpoint(account, "viraldance933"))
			upstream := &httpUpstreamRecorder{resp: jsonVideoResponse(http.StatusOK, `{"task_id":"task-dola","status":"queued"}`)}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
			result, err := svc.ForwardOpenAIVideo(context.Background(), c, account, OpenAIVideoEndpointCreate, "", "", body, "application/json", "viraldance933")
			require.NoError(t, err)
			require.JSONEq(t, fmt.Sprintf(`{"model":%q,"prompt":"waves","duration":8,"ratio":"16:9","image_urls":["https://media.test/a.png"]}`, model), string(upstream.lastBody))
			require.Equal(t, model, result.UpstreamModel)
			require.Equal(t, "viraldance933", result.Model)
		})
	}
}

func TestDolaVideoRejectsInvalidDurationBeforeBilling(t *testing.T) {
	for _, tc := range []struct {
		model      string
		maxSeconds int
	}{{"dola-viraldance2.0", 15}, {"dola-viraldance2.5", 30}} {
		for _, value := range []string{"", `,"duration":3`, fmt.Sprintf(`,"duration":%d`, tc.maxSeconds+1), `,"duration":4.5`, `,"duration":"8"`, `,"duration":null`, `,"seconds":8`} {
			t.Run(tc.model+value, func(t *testing.T) {
				body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"waves"%s}`, tc.model, value))
				wantError := fmt.Sprintf("integer duration from 4 to %d", tc.maxSeconds)
				_, err := ParseOpenAIVideoCreateRequest(body)
				require.ErrorContains(t, err, wantError)
				svc := &OpenAIGatewayService{}
				_, _, err = svc.EstimateOpenAIVideoCreateBilling(context.Background(), &APIKey{}, &User{ID: 1}, tc.model, body)
				require.ErrorContains(t, err, wantError)
			})
		}
	}
}

func TestDolaVideoAcceptsValidDurationAndPreservesBillingSeconds(t *testing.T) {
	for _, tc := range []struct {
		model   string
		seconds []int
	}{{"dola-viraldance2.0", []int{4, 15}}, {"dola-viraldance2.5", []int{4, 15, 16, 30}}} {
		for _, seconds := range tc.seconds {
			t.Run(fmt.Sprintf("%s/%d", tc.model, seconds), func(t *testing.T) {
				body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"waves","duration":%d}`, tc.model, seconds))
				parsed, err := ParseOpenAIVideoCreateRequest(body)
				require.NoError(t, err)
				require.Equal(t, body, parsed.Body)
				resolution, billedSeconds := OpenAIVideoBillingParameters(tc.model, parsed.BillingBody)
				require.Equal(t, "720p", resolution)
				require.Equal(t, seconds, billedSeconds)
			})
		}
	}
}

func newViraleeVideoPricingResolver(groupID int64, model string, mode BillingMode, price float64) *ModelPricingResolver {
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: model}] = &ChannelModelPricing{
		Platform: PlatformOpenAI, Models: []string{model}, BillingMode: mode, PerRequestPrice: &price,
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channels := &ChannelService{}
	channels.cache.Store(cache)
	return NewModelPricingResolver(channels, NewBillingService(&config.Config{}, nil))
}

func TestViraleeVideoBillingPreservesThirtySecondsAndDefaults(t *testing.T) {
	for _, tc := range []struct {
		model, body string
		mode        BillingMode
		want        float64
	}{
		{"wan3.0x", `{"duration":30}`, BillingModeVideo, 15},
		{"wan3.0x-480p", `{"duration":2}`, BillingModeVideo, 1},
		{"wan3.0x-1080p", `{"duration":30}`, BillingModeVideo, 15},
		{"viraldance2.5-30", `{"duration":30}`, BillingModeVideo, 15},
		{"viraldance2.5-30", `{}`, BillingModeVideo, 2.5},
		{"viraldance933", `{"duration":15}`, BillingModeVideo, 7.5},
		{"viraldance933-fast", `{"duration":4}`, BillingModeVideo, 2},
		{"dola-viraldance2.0", `{"duration":4}`, BillingModeVideo, 2},
		{"dola-viraldance2.5", `{"duration":15}`, BillingModeVideo, 7.5},
		{"dola-viraldance2.5", `{"duration":16}`, BillingModeVideo, 8},
		{"dola-viraldance2.5", `{"duration":30}`, BillingModeVideo, 15},
		{"dola-viraldance2.0", `{"duration":15}`, BillingModePerRequest, .5},
		{"dola-viraldance2.5", `{"duration":4}`, BillingModePerRequest, .5},
		{"dola-viraldance2.5", `{"duration":30}`, BillingModePerRequest, .5},
		{"viraldance2.5-15", `{}`, BillingModePerRequest, .5},
		{"viraldance2.5-15", `{"duration":4}`, BillingModePerRequest, .5},
		{"viraldance2.5-480p-15", `{"duration":15}`, BillingModePerRequest, .5},
	} {
		t.Run(tc.model+tc.body, func(t *testing.T) {
			const groupID = int64(729)
			svc := &OpenAIGatewayService{cfg: &config.Config{}, billingService: NewBillingService(&config.Config{}, nil), resolver: newViraleeVideoPricingResolver(groupID, tc.model, tc.mode, .5)}
			key := miniMaxH3BillingAPIKey(groupID, .8)
			cost, inputSeconds, err := svc.EstimateOpenAIVideoCreateBilling(context.Background(), key, &User{ID: 1}, tc.model, []byte(tc.body))
			require.NoError(t, err)
			require.Zero(t, inputSeconds)
			require.InDelta(t, tc.want, cost.TotalCost, 1e-8)
			require.InDelta(t, tc.want*.8, cost.ActualCost, 1e-8)
		})
	}
}

func TestViraleeVideoRejectsInvalidDurationBeforeBilling(t *testing.T) {
	for _, tc := range []struct{ model, value string }{
		{"wan3.0x", ``}, {"wan3.0x", `,"duration":1`}, {"wan3.0x", `,"duration":31`},
		{"viraldance933", ``}, {"viraldance933-fast", `,"duration":16`},
		{"viraldance2.5-30", `,"duration":0`}, {"viraldance2.5-30", `,"duration":4.5`},
		{"viraldance2.5-30", `,"duration":"30"`}, {"viraldance2.5-30", `,"duration":null`},
		{"viraldance2.5-15", `,"duration":16`}, {"viraldance2.5-480p-15", `,"duration":3`},
	} {
		t.Run(tc.model+tc.value, func(t *testing.T) {
			body := []byte(fmt.Sprintf(`{"model":%q,"prompt":"waves"%s}`, tc.model, tc.value))
			_, err := ParseOpenAIVideoCreateRequest(body)
			require.ErrorContains(t, err, "integer duration")
			svc := &OpenAIGatewayService{}
			_, _, err = svc.EstimateOpenAIVideoCreateBilling(context.Background(), &APIKey{}, &User{ID: 1}, tc.model, body)
			require.ErrorContains(t, err, "integer duration")
		})
	}
	for _, model := range []string{"wan3.0x-unknown", "viraldance933-unknown", "dola-viraldance2.0-unknown", "dola-viraldance2.5-fast"} {
		require.False(t, IsOpenAIVideoModel(model), "only documented models should be enabled")
	}
}

func TestViraleeVideoGroupPriceAndUsageRefundKeepModelDuration(t *testing.T) {
	for _, model := range []string{"wan3.0x", "wan3.0x-480p", "wan3.0x-1080p", "viraldance2.5-30", "dola-viraldance2.0", "dola-viraldance2.5"} {
		t.Run(model, func(t *testing.T) {
			seconds := 30
			if model == "dola-viraldance2.0" {
				seconds = 15
			}
			body := []byte(fmt.Sprintf(`{"duration":%d}`, seconds))
			actualCost := float64(seconds) * .5 * .8
			const groupID = int64(730)
			usageRepo := &openAIRecordUsageLogRepoStub{inserted: true, logByRequest: make(map[string]*UsageLog)}
			billingStub := &openAIRecordUsageBillingRepoStub{applied: map[string]bool{}}
			billingRepo := &openAIVideoRefundAtomicBillingRepoStub{openAIRecordUsageBillingRepoStub: billingStub, usageRepo: usageRepo}
			svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
			// This billing test replaces dependencies after construction; stop the
			// unrelated harvester before changing the account repository.
			key := miniMaxH3BillingAPIKey(groupID, .8)
			key.ID = 1
			user := &User{ID: 2}
			account := newOpenAIVideoTestAccount()
			svc.accountRepo = &openAIRecordUsageAccountRepoStub{account: account}
			resolution, _ := OpenAIVideoBillingParameters(model, body)
			price := .5
			key.Group.VideoPrice480P = &price
			key.Group.VideoPrice720P = &price
			key.Group.VideoPrice1080P = &price
			cost, _, err := svc.EstimateOpenAIVideoCreateBilling(context.Background(), key, user, model, body)
			require.NoError(t, err)
			require.InDelta(t, float64(seconds)*.5, cost.TotalCost, 1e-8)
			require.InDelta(t, actualCost, cost.ActualCost, 1e-8)

			taskID := "task-" + model
			requestID := OpenAIVideoUsageRequestID(taskID)
			err = svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
				Result: &OpenAIForwardResult{
					RequestID: requestID, UseResultRequestID: true, Model: model, UpstreamModel: model,
					MediaType: "video", VideoCount: 1, RequestCount: 1, VideoResolution: resolution, VideoDurationSeconds: seconds,
				},
				APIKey: key, User: user, Account: account,
				MediaBalanceHold:    &OpenAIMediaBalanceHold{ID: "hold", APIKeyID: key.ID, UserID: user.ID, Amount: cost.ActualCost},
				BillingCostSnapshot: cost,
			})
			require.NoError(t, err)
			require.NotNil(t, usageRepo.lastLog)
			require.Equal(t, seconds, *usageRepo.lastLog.VideoDurationSeconds)
			require.Equal(t, resolution, *usageRepo.lastLog.VideoResolution)
			require.InDelta(t, actualCost, billingStub.lastCmd.CapturedBalanceHold.ActualAmount, 1e-8)
			usageRepo.logByRequest[requestID] = usageRepo.lastLog
			for i := 0; i < 2; i++ {
				require.NoError(t, svc.RefundFailedOpenAIVideoTask(context.Background(), key, taskID, nil))
			}
			require.Equal(t, 2, usageRepo.calls, "one original usage and one refund despite retry")
			require.Equal(t, OpenAIVideoRefundRequestID(taskID), usageRepo.lastLog.RequestID)
			require.InDelta(t, -actualCost, usageRepo.lastLog.ActualCost, 1e-8)
			require.InDelta(t, -actualCost, billingStub.lastCmd.BalanceCost, 1e-8)
		})
	}
}
