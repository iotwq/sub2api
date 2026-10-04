package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

const (
	openAIVideoRecoveryInactive      = "inactive"
	openAIVideoRecoverySubmitting    = "submitting"
	openAIVideoRecoveryPending       = "pending"
	openAIVideoRecoveryIdentified    = "identified"
	openAIVideoRecoveryMatched       = "matched"
	openAIVideoRecoveryAmbiguous     = "ambiguous"
	openAIVideoRecoveryFailed        = "failed"
	openAIVideoRecoveryCancelled     = "cancelled"
	openAIVideoRecoveryBillingReview = "billing_review"

	openAIVideoRecoveryInitialDelay = 30 * time.Second
	openAIVideoRecoveryRetryDelay   = 15 * time.Second
	openAIVideoRecoveryWindow       = 30 * time.Minute
	openAIVideoRecoveryLease        = 45 * time.Second
	openAIVideoRecoveryBatchSize    = 10
)

const (
	OpenAIVideoRecoveryInactive      = openAIVideoRecoveryInactive
	OpenAIVideoRecoveryPending       = openAIVideoRecoveryPending
	OpenAIVideoRecoveryIdentified    = openAIVideoRecoveryIdentified
	OpenAIVideoRecoveryMatched       = openAIVideoRecoveryMatched
	OpenAIVideoRecoveryAmbiguous     = openAIVideoRecoveryAmbiguous
	OpenAIVideoRecoveryFailed        = openAIVideoRecoveryFailed
	OpenAIVideoRecoveryCancelled     = openAIVideoRecoveryCancelled
	OpenAIVideoRecoveryBillingReview = openAIVideoRecoveryBillingReview
)

type MiniMaxVideoRecoverySignature struct {
	Model           string `json:"model"`
	Resolution      string `json:"resolution"`
	DurationSeconds int    `json:"duration_seconds"`
	Ratio           string `json:"ratio"`
	ImageCount      int    `json:"image_count"`
	VideoCount      int    `json:"video_count"`
	AudioCount      int    `json:"audio_count"`
}

type OpenAIVideoRecoveryBillingSnapshot struct {
	Cost                      *CostBreakdown          `json:"cost,omitempty"`
	Hold                      *OpenAIMediaBalanceHold `json:"hold,omitempty"`
	RequestModel              string                  `json:"request_model"`
	UpstreamModel             string                  `json:"upstream_model"`
	BillingModel              string                  `json:"billing_model"`
	VideoResolution           string                  `json:"video_resolution"`
	VideoDurationSeconds      int                     `json:"video_duration_seconds"`
	VideoInputDurationSeconds float64                 `json:"video_input_duration_seconds"`
	RequestPayloadHash        string                  `json:"request_payload_hash"`
	SubscriptionID            int64                   `json:"subscription_id,omitempty"`
	InboundEndpoint           string                  `json:"inbound_endpoint"`
	UpstreamEndpoint          string                  `json:"upstream_endpoint"`
	UserAgent                 string                  `json:"user_agent"`
	IPAddress                 string                  `json:"ip_address"`
	QuotaPlatform             string                  `json:"quota_platform"`
	PricingAt                 time.Time               `json:"pricing_at"`
	ChannelUsageFields
}

type MiniMaxVideoRecoveryTask struct {
	Binding *OpenAIVideoTaskBinding
}

type miniMaxVideoTaskSummary struct {
	ID              string
	Signature       MiniMaxVideoRecoverySignature
	ImageCountKnown bool
	VideoCountKnown bool
	AudioCountKnown bool
}

func BuildMiniMaxVideoRecoverySignature(model string, body []byte) MiniMaxVideoRecoverySignature {
	resolution, duration := OpenAIVideoBillingParameters(model, body)
	signature := MiniMaxVideoRecoverySignature{
		Model:           canonicalOpenAIVideoModel(model),
		Resolution:      normalizeMiniMaxRecoveryValue(resolution),
		DurationSeconds: duration,
		Ratio:           normalizeMiniMaxRecoveryValue(firstGJSONValue(body, "ratio", "aspect_ratio")),
	}
	content := gjson.GetBytes(body, "content")
	if content.IsArray() {
		content.ForEach(func(_, item gjson.Result) bool {
			switch strings.ToLower(strings.TrimSpace(item.Get("type").String())) {
			case "image_url", "input_image", "image":
				signature.ImageCount++
			case "video_url", "input_video", "video":
				signature.VideoCount++
			case "audio_url", "input_audio", "audio":
				signature.AudioCount++
			}
			return true
		})
	}
	return signature
}

func (s *OpenAIGatewayService) PrepareMiniMaxVideoRecovery(
	ctx context.Context,
	groupID *int64,
	userID int64,
	apiKey *APIKey,
	account *Account,
	requestBody []byte,
	billing OpenAIVideoRecoveryBillingSnapshot,
) (*MiniMaxVideoRecoveryTask, error) {
	repo, ok := s.openAIVideoTaskBindingRepo.(OpenAIVideoRecoveryRepository)
	if !ok || repo == nil {
		return nil, errors.New("MiniMax-H3 submission recovery is unavailable")
	}
	if apiKey == nil || account == nil {
		return nil, errors.New("MiniMax-H3 recovery identity is unavailable")
	}
	tasks, err := s.fetchMiniMaxVideoTaskList(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("read MiniMax-H3 recovery baseline: %w", err)
	}
	baselineIDs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		baselineIDs = append(baselineIDs, task.ID)
	}
	signature := BuildMiniMaxVideoRecoverySignature(billing.UpstreamModel, requestBody)
	baselineJSON, _ := json.Marshal(baselineIDs)
	signatureJSON, _ := json.Marshal(signature)
	billingJSON, err := json.Marshal(billing)
	if err != nil {
		return nil, fmt.Errorf("encode MiniMax-H3 recovery billing snapshot: %w", err)
	}
	now := time.Now()
	nextCheckAt := now.Add(openAIVideoRecoveryInitialDelay)
	recoveryExpiresAt := now.Add(openAIVideoRecoveryWindow)
	binding := OpenAIVideoTaskBinding{
		GroupID:             derefGroupID(groupID),
		UserID:              userID,
		TaskID:              "task_recovery_" + generateRequestID(),
		AccountID:           account.ID,
		APIKeyID:            apiKey.ID,
		APIKeyQuotaLimited:  apiKey.Quota > 0,
		APIKeyRateLimited:   apiKey.HasRateLimits(),
		BillingTaskID:       "",
		RecoveryStatus:      openAIVideoRecoverySubmitting,
		RecoveryBaseline:    baselineJSON,
		RecoverySignature:   signatureJSON,
		RecoveryBilling:     billingJSON,
		RecoveryNextCheckAt: &nextCheckAt,
		RecoveryExpiresAt:   &recoveryExpiresAt,
		ExpiresAt:           now.Add(openAIVideoTaskBindingTTL),
	}
	binding.BillingTaskID = binding.TaskID
	if err := repo.PrepareOpenAIVideoRecovery(ctx, binding); err != nil {
		return nil, err
	}
	persisted, err := repo.GetOpenAIVideoRecovery(ctx, binding.GroupID, binding.UserID, binding.TaskID)
	if err != nil {
		return nil, err
	}
	if persisted == nil {
		return nil, errors.New("MiniMax-H3 recovery record was not persisted")
	}
	return &MiniMaxVideoRecoveryTask{Binding: persisted}, nil
}

func (s *OpenAIGatewayService) GetMiniMaxVideoRecovery(ctx context.Context, groupID *int64, userID int64, taskID string) (*OpenAIVideoTaskBinding, error) {
	repo, ok := s.openAIVideoTaskBindingRepo.(OpenAIVideoRecoveryRepository)
	if !ok || repo == nil {
		return nil, nil
	}
	return repo.GetOpenAIVideoRecovery(ctx, derefGroupID(groupID), userID, taskID)
}

func (s *OpenAIGatewayService) MarkMiniMaxVideoRecovery(
	ctx context.Context,
	recovery *MiniMaxVideoRecoveryTask,
	status string,
	upstreamTaskID string,
	lastError string,
) error {
	if recovery == nil || recovery.Binding == nil {
		return nil
	}
	repo, ok := s.openAIVideoTaskBindingRepo.(OpenAIVideoRecoveryRepository)
	if !ok || repo == nil {
		return errors.New("MiniMax-H3 submission recovery is unavailable")
	}
	var nextCheckAt *time.Time
	if status == openAIVideoRecoverySubmitting || status == openAIVideoRecoveryPending || status == openAIVideoRecoveryIdentified {
		next := time.Now().Add(openAIVideoRecoveryRetryDelay)
		nextCheckAt = &next
	}
	if err := repo.UpdateOpenAIVideoRecovery(ctx, recovery.Binding.ID, status, upstreamTaskID, nextCheckAt, lastError); err != nil {
		return err
	}
	recovery.Binding.RecoveryStatus = status
	if strings.TrimSpace(upstreamTaskID) != "" {
		recovery.Binding.UpstreamTaskID = strings.TrimSpace(upstreamTaskID)
	}
	return nil
}

func (s *OpenAIGatewayService) BindOpenAIVideoTaskRoute(
	ctx context.Context,
	groupID *int64,
	userID int64,
	clientTaskID string,
	upstreamTaskID string,
	billingTaskID string,
	accountID int64,
) error {
	repo, ok := s.openAIVideoTaskBindingRepo.(OpenAIVideoRecoveryRepository)
	if !ok || repo == nil {
		return s.BindOpenAIVideoTaskAccount(ctx, groupID, userID, clientTaskID, accountID)
	}
	binding := OpenAIVideoTaskBinding{
		GroupID:        derefGroupID(groupID),
		UserID:         userID,
		TaskID:         strings.TrimSpace(clientTaskID),
		UpstreamTaskID: strings.TrimSpace(upstreamTaskID),
		BillingTaskID:  strings.TrimSpace(billingTaskID),
		AccountID:      accountID,
		ExpiresAt:      time.Now().Add(openAIVideoTaskBindingTTL),
	}
	if err := repo.UpsertOpenAIVideoTaskRoute(ctx, binding); err != nil {
		return err
	}
	return s.setStickySessionAccountID(ctx, groupID, OpenAIVideoTaskSessionHash(userID, clientTaskID), accountID, openAIVideoTaskBindingTTL)
}

func (s *OpenAIGatewayService) fetchMiniMaxVideoTaskList(ctx context.Context, account *Account) ([]miniMaxVideoTaskSummary, error) {
	if s == nil || s.httpUpstream == nil || account == nil || !AccountUsesMiniMaxV2Video(account) {
		return nil, errors.New("MiniMax-H3 task list is unavailable")
	}
	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}
	baseURL, err := s.validateUpstreamBaseURL(account.GetOpenAIBaseURL())
	if err != nil {
		return nil, err
	}
	targetURL := buildMiniMaxV2VideoEndpointURL(baseURL, OpenAIVideoEndpointList, "")
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	query := parsedURL.Query()
	query.Set("page_num", "1")
	query.Set("page_size", "100")
	parsedURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(WithHTTPUpstreamProfile(ctx, HTTPUpstreamProfileOpenAI), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	account.ApplyHeaderOverrides(req.Header)
	if userAgent := account.GetOpenAIUserAgent(); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("MiniMax-H3 task list returned HTTP %d: %s", resp.StatusCode, sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(body)))
	}
	return parseMiniMaxVideoTaskList(body)
}

func parseMiniMaxVideoTaskList(body []byte) ([]miniMaxVideoTaskSummary, error) {
	if !gjson.ValidBytes(body) {
		return nil, errors.New("MiniMax-H3 task list returned invalid JSON")
	}
	var items gjson.Result
	for _, path := range []string{"items", "data.items", "tasks", "data.tasks"} {
		candidate := gjson.GetBytes(body, path)
		if candidate.Exists() && candidate.IsArray() {
			items = candidate
			break
		}
	}
	if !items.Exists() {
		return nil, errors.New("MiniMax-H3 task list did not contain items")
	}
	tasks := make([]miniMaxVideoTaskSummary, 0, len(items.Array()))
	items.ForEach(func(_, item gjson.Result) bool {
		id := firstGJSONResultValue(item, "id", "task_id", "task.id")
		if id == "" {
			return true
		}
		imageCount, imageCountKnown := firstGJSONResultPositiveInt(item, "input_image_count", "image_count", "reference_image_count", "usage.input_image_count", "usage.image_count", "usage.reference_image_count")
		videoCount, videoCountKnown := firstGJSONResultPositiveInt(item, "input_video_count", "video_count", "reference_video_count", "usage.input_video_count", "usage.video_count", "usage.reference_video_count")
		audioCount, audioCountKnown := firstGJSONResultPositiveInt(item, "input_audio_count", "audio_count", "reference_audio_count", "usage.input_audio_count", "usage.audio_count", "usage.reference_audio_count")
		tasks = append(tasks, miniMaxVideoTaskSummary{
			ID:              id,
			ImageCountKnown: imageCountKnown,
			VideoCountKnown: videoCountKnown,
			AudioCountKnown: audioCountKnown,
			Signature: MiniMaxVideoRecoverySignature{
				Model:           firstGJSONResultValue(item, "model", "usage.model", "metadata.model"),
				Resolution:      normalizeMiniMaxRecoveryValue(firstGJSONResultValue(item, "resolution", "usage.resolution", "metadata.resolution")),
				DurationSeconds: firstGJSONResultInt(item, "duration", "duration_seconds", "usage.duration", "usage.duration_seconds", "metadata.duration"),
				Ratio:           normalizeMiniMaxRecoveryValue(firstGJSONResultValue(item, "ratio", "aspect_ratio", "usage.ratio", "metadata.ratio")),
				ImageCount:      imageCount,
				VideoCount:      videoCount,
				AudioCount:      audioCount,
			},
		})
		return true
	})
	return tasks, nil
}

func findMiniMaxVideoRecoveryCandidates(
	baselineIDs []string,
	want MiniMaxVideoRecoverySignature,
	tasks []miniMaxVideoTaskSummary,
) []miniMaxVideoTaskSummary {
	baseline := make(map[string]struct{}, len(baselineIDs))
	for _, id := range baselineIDs {
		baseline[strings.TrimSpace(id)] = struct{}{}
	}
	matches := make([]miniMaxVideoTaskSummary, 0, 1)
	for _, task := range tasks {
		if _, existed := baseline[task.ID]; existed || !miniMaxRecoverySignaturesMatch(want, task) {
			continue
		}
		matches = append(matches, task)
	}
	return matches
}

func miniMaxRecoverySignaturesMatch(want MiniMaxVideoRecoverySignature, got miniMaxVideoTaskSummary) bool {
	if !strings.EqualFold(strings.TrimSpace(want.Model), strings.TrimSpace(got.Signature.Model)) {
		return false
	}
	return want.Resolution != "" && want.Resolution == got.Signature.Resolution &&
		want.DurationSeconds > 0 && want.DurationSeconds == got.Signature.DurationSeconds &&
		want.Ratio != "" && want.Ratio == got.Signature.Ratio &&
		(!got.ImageCountKnown || want.ImageCount == got.Signature.ImageCount) &&
		(!got.VideoCountKnown || want.VideoCount == got.Signature.VideoCount) &&
		(!got.AudioCountKnown || want.AudioCount == got.Signature.AudioCount)
}

func firstGJSONValue(body []byte, paths ...string) string {
	for _, path := range paths {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ""
}

func firstGJSONResultValue(value gjson.Result, paths ...string) string {
	for _, path := range paths {
		if result := strings.TrimSpace(value.Get(path).String()); result != "" {
			return result
		}
	}
	return ""
}

func firstGJSONResultInt(value gjson.Result, paths ...string) int {
	for _, path := range paths {
		result := value.Get(path)
		if !result.Exists() {
			continue
		}
		if parsed, err := strconv.Atoi(strings.TrimSpace(result.String())); err == nil {
			return parsed
		}
	}
	return 0
}

func firstGJSONResultPositiveInt(value gjson.Result, paths ...string) (int, bool) {
	for _, path := range paths {
		result := value.Get(path)
		if !result.Exists() {
			continue
		}
		if parsed, err := strconv.Atoi(strings.TrimSpace(result.String())); err == nil && parsed > 0 {
			return parsed, true
		}
	}
	return 0, false
}

func normalizeMiniMaxRecoveryValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
