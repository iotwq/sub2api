package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type OpenAIVideoEndpoint string

const (
	OpenAIVideoEndpointCreate     OpenAIVideoEndpoint = "create"
	OpenAIVideoEndpointStatus     OpenAIVideoEndpoint = "status"
	OpenAIVideoEndpointContent    OpenAIVideoEndpoint = "content"
	OpenAIVideoEndpointList       OpenAIVideoEndpoint = "list"
	OpenAIVideoEndpointDelete     OpenAIVideoEndpoint = "delete"
	OpenAIVideoEndpointContext    OpenAIVideoEndpoint = "context_ir"
	OpenAIVideoEndpointRegenerate OpenAIVideoEndpoint = "regenerate"
)

type OpenAIVideoCreateRequest struct {
	Model        string
	Prompt       string
	SourceTaskID string
	Body         []byte
	BillingBody  []byte
}

const openAIVideoTaskBindingTTL = 7 * 24 * time.Hour
const openAIVideoUsageRequestIDPrefix = "openai-video:"
const openAIVideoRefundRequestIDPrefix = "openai-video-refund:"
const openAIVideoModelFireflyV2 = "firefly-video-v2"
const openAIVideoModelFireflyV2Fast = "firefly-video-v2-fast"
const openAIVideoModelMiniMaxH3 = "minimax-h3"
const openAIVideoModelMiniMaxH3Canonical = "MiniMax-H3"
const supportedOpenAIVideoModelsText = "video-ds-2.0-fast, video-ds-2.0, as-sd2.0-fast, firefly-video-v2, firefly-video-v2-fast, MiniMax-H3, sora-2-landscape-8s, sora-2-landscape-12s, sora-2-portrait-8s, sora-2-portrait-12s, wan3.0x, wan3.0x-480p, wan3.0x-1080p, viraldance933, viraldance933-fast, viraldance2.5-30, viraldance2.5-15, viraldance2.5-480p-15, dola-viraldance2.0, or dola-viraldance2.5"

type OpenAIVideoTaskBinding struct {
	ID        int64
	GroupID   int64
	UserID    int64
	TaskID    string
	AccountID int64
	APIKeyID  int64
	// APIKeyKey is populated only while a compensation item is claimed; it is
	// read from api_keys and is never copied into the task-binding table.
	APIKeyKey           string
	APIKeyQuotaLimited  bool
	APIKeyRateLimited   bool
	CompensationStatus  string
	NextCheckAt         *time.Time
	CheckAttempts       int
	LastCheckError      *string
	CheckedAt           *time.Time
	ExpiresAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	UpstreamTaskID      string
	BillingTaskID       string
	RecoveryStatus      string
	RecoveryBaseline    []byte
	RecoverySignature   []byte
	RecoveryBilling     []byte
	RecoveryNextCheckAt *time.Time
	RecoveryExpiresAt   *time.Time
	RecoveryAttempts    int
	RecoveryLastError   *string
	RecoveryCheckedAt   *time.Time
}

type OpenAIVideoRecoveryRepository interface {
	PrepareOpenAIVideoRecovery(ctx context.Context, binding OpenAIVideoTaskBinding) error
	GetOpenAIVideoRecovery(ctx context.Context, groupID, userID int64, taskID string) (*OpenAIVideoTaskBinding, error)
	UpsertOpenAIVideoTaskRoute(ctx context.Context, binding OpenAIVideoTaskBinding) error
	UpdateOpenAIVideoRecovery(ctx context.Context, id int64, status, upstreamTaskID string, nextCheckAt *time.Time, lastError string) error
	ClaimOpenAIVideoRecoveries(ctx context.Context, now, leaseUntil time.Time, limit int) ([]OpenAIVideoTaskBinding, error)
}

type OpenAIVideoTaskBindingRepository interface {
	UpsertOpenAIVideoTaskBinding(ctx context.Context, binding OpenAIVideoTaskBinding) error
	GetOpenAIVideoTaskBinding(ctx context.Context, groupID, userID int64, taskID string) (*OpenAIVideoTaskBinding, error)
	ArmOpenAIVideoTaskCompensation(ctx context.Context, groupID, userID int64, taskID string, apiKey *APIKey, nextCheckAt time.Time) error
	ClaimOpenAIVideoTaskCompensations(ctx context.Context, now, leaseUntil time.Time, limit int) ([]OpenAIVideoTaskBinding, error)
	UpdateOpenAIVideoTaskCompensation(ctx context.Context, id int64, status string, nextCheckAt *time.Time, lastError string) error
}

func ParseOpenAIVideoCreateRequest(body []byte) (*OpenAIVideoCreateRequest, error) {
	return ParseOpenAIVideoOperationRequest(OpenAIVideoEndpointCreate, body)
}

func ParseOpenAIVideoOperationRequest(endpoint OpenAIVideoEndpoint, body []byte) (*OpenAIVideoCreateRequest, error) {
	return ParseOpenAIVideoOperationRequestWithContentType(endpoint, body, "application/json")
}

func ParseOpenAIVideoOperationRequestWithContentType(endpoint OpenAIVideoEndpoint, body []byte, contentType string) (*OpenAIVideoCreateRequest, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("request body is empty")
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/json"
	}
	mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Type")
	}
	if strings.EqualFold(mediaType, "multipart/form-data") {
		return parseOpenAIVideoMultipartRequest(endpoint, body, params["boundary"])
	}
	if mediaType != "" && !strings.EqualFold(mediaType, "application/json") {
		return nil, fmt.Errorf("video endpoint requires application/json or multipart/form-data")
	}
	if !gjson.ValidBytes(body) {
		return nil, fmt.Errorf("failed to parse request body")
	}
	model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if model == "" {
		return nil, fmt.Errorf("video endpoint requires model")
	}
	if !IsOpenAIVideoModel(model) {
		return nil, fmt.Errorf("video endpoint supports %s, got %q", supportedOpenAIVideoModelsText, model)
	}
	if err := validateViraleeVideoDuration(model, body); err != nil {
		return nil, err
	}
	if endpoint == OpenAIVideoEndpointContext || endpoint == OpenAIVideoEndpointRegenerate {
		if !IsMiniMaxH3VideoModel(model) {
			return nil, fmt.Errorf("video endpoint %s requires %s", endpoint, openAIVideoModelMiniMaxH3Canonical)
		}
	}
	prompt := strings.TrimSpace(gjson.GetBytes(body, "prompt").String())
	if IsFireflyV2VideoModel(model) && prompt == "" {
		prompt = extractOpenAIVideoShotsPrompt(gjson.GetBytes(body, "shots"))
	}
	if IsMiniMaxH3VideoModel(model) {
		if contentPrompt := extractMiniMaxVideoPrompt(body); contentPrompt != "" {
			prompt = contentPrompt
		}
		if !gjson.GetBytes(body, "content").Exists() && miniMaxVideoLegacyMediaFieldsPresent(body) {
			return nil, fmt.Errorf("MiniMax-H3 media inputs require native content[] entries with explicit roles")
		}
	}
	sourceTaskID := strings.TrimSpace(gjson.GetBytes(body, "source_task_id").String())
	if endpoint == OpenAIVideoEndpointRegenerate && sourceTaskID == "" && !miniMaxVideoHasBaseVideo(body) {
		return nil, fmt.Errorf("MiniMax-H3 regeneration requires source_task_id or a content[] base_video entry")
	}
	if endpoint != OpenAIVideoEndpointRegenerate && prompt == "" {
		if IsFireflyV2VideoModel(model) {
			return nil, fmt.Errorf("Firefly video endpoint requires prompt or shots")
		}
		return nil, fmt.Errorf("video endpoint requires prompt")
	}
	return &OpenAIVideoCreateRequest{
		Model:        model,
		Prompt:       prompt,
		SourceTaskID: sourceTaskID,
		Body:         body,
		BillingBody:  body,
	}, nil
}

func parseOpenAIVideoMultipartRequest(endpoint OpenAIVideoEndpoint, body []byte, boundary string) (*OpenAIVideoCreateRequest, error) {
	boundary = strings.TrimSpace(boundary)
	if boundary == "" {
		return nil, fmt.Errorf("multipart boundary is required")
	}
	fields := make(map[string][]string)
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse multipart request")
		}
		if part.FileName() != "" {
			_, _ = io.Copy(io.Discard, part)
			_ = part.Close()
			continue
		}
		value, err := io.ReadAll(part)
		_ = part.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read multipart field %q", part.FormName())
		}
		fields[part.FormName()] = append(fields[part.FormName()], string(value))
	}

	model := firstOpenAIVideoMultipartField(fields, "model")
	if model == "" {
		return nil, fmt.Errorf("video endpoint requires model")
	}
	if !IsOpenAIVideoModel(model) {
		return nil, fmt.Errorf("video endpoint supports %s, got %q", supportedOpenAIVideoModelsText, model)
	}
	if !IsFireflyV2VideoModel(model) {
		return nil, fmt.Errorf("multipart video requests require firefly-video-v2 or firefly-video-v2-fast")
	}
	if endpoint != OpenAIVideoEndpointCreate {
		return nil, fmt.Errorf("multipart video requests are only supported for task creation")
	}
	prompt := firstOpenAIVideoMultipartField(fields, "prompt")
	if prompt == "" {
		prompt = extractOpenAIVideoShotsPrompt(gjson.Parse(firstOpenAIVideoMultipartField(fields, "shots")))
	}
	if prompt == "" {
		return nil, fmt.Errorf("Firefly video endpoint requires prompt or shots")
	}
	billingBody, err := json.Marshal(map[string]string{
		"model":      model,
		"duration":   firstOpenAIVideoMultipartField(fields, "duration"),
		"resolution": firstOpenAIVideoMultipartField(fields, "resolution"),
	})
	if err != nil {
		return nil, fmt.Errorf("build video billing request: %w", err)
	}
	return &OpenAIVideoCreateRequest{
		Model:       model,
		Prompt:      prompt,
		Body:        body,
		BillingBody: billingBody,
	}, nil
}

func firstOpenAIVideoMultipartField(fields map[string][]string, name string) string {
	values := fields[name]
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func extractOpenAIVideoShotsPrompt(shots gjson.Result) string {
	if !shots.Exists() || !shots.IsArray() {
		return ""
	}
	prompts := make([]string, 0)
	shots.ForEach(func(_, shot gjson.Result) bool {
		if prompt := strings.TrimSpace(shot.Get("prompt").String()); prompt != "" {
			prompts = append(prompts, prompt)
		}
		return true
	})
	return strings.Join(prompts, "\n")
}

func (r *OpenAIVideoCreateRequest) ModerationBody() []byte {
	if r == nil || len(r.Body) == 0 {
		return nil
	}
	payload := []byte(`{}`)
	if r.Prompt != "" {
		payload, _ = sjson.SetBytes(payload, "prompt", r.Prompt)
	}
	if images := gjson.GetBytes(r.Body, "images"); images.Exists() && images.IsArray() {
		payload, _ = sjson.SetRawBytes(payload, "images", []byte(images.Raw))
	} else if content := gjson.GetBytes(r.Body, "content"); content.Exists() && content.IsArray() {
		payload, _ = sjson.SetRawBytes(payload, "images", []byte(content.Raw))
	}
	return payload
}

func IsOpenAIVideoModel(model string) bool {
	if _, ok := lookupViraleeVideoSpec(model); ok {
		return true
	}
	switch normalizeOpenAIVideoModel(model) {
	case "video-ds-2.0-fast",
		"video-ds-2.0",
		"as-sd2.0-fast",
		openAIVideoModelFireflyV2,
		openAIVideoModelFireflyV2Fast,
		openAIVideoModelMiniMaxH3,
		"sora-2-landscape-8s",
		"sora-2-landscape-12s",
		"sora-2-portrait-8s",
		"sora-2-portrait-12s":
		return true
	default:
		return false
	}
}

func normalizeOpenAIVideoModel(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

func NormalizeOpenAIVideoModel(model string) string {
	return normalizeOpenAIVideoModel(model)
}

func IsMiniMaxH3VideoModel(model string) bool {
	return normalizeOpenAIVideoModel(model) == openAIVideoModelMiniMaxH3
}

func IsFireflyV2VideoModel(model string) bool {
	normalized := normalizeOpenAIVideoModel(model)
	return normalized == openAIVideoModelFireflyV2 || normalized == openAIVideoModelFireflyV2Fast
}

func canonicalOpenAIVideoModel(model string) string {
	if IsMiniMaxH3VideoModel(model) {
		return openAIVideoModelMiniMaxH3Canonical
	}
	return normalizeOpenAIVideoModel(model)
}

func AccountUsesMiniMaxV2Video(account *Account) bool {
	return account != nil && account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityMiniMaxVideo)
}

func AccountSupportsOpenAIVideoEndpoint(account *Account, requestedModel string) bool {
	if account == nil || !account.IsOpenAIApiKey() {
		return false
	}
	upstreamModel := normalizeOpenAIVideoModel(account.GetMappedModel(requestedModel))
	if IsMiniMaxH3VideoModel(upstreamModel) {
		return AccountUsesMiniMaxV2Video(account)
	}
	return IsOpenAIVideoModel(upstreamModel)
}

func OpenAIVideoTaskSessionHash(userID int64, taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if userID <= 0 || taskID == "" {
		return ""
	}
	return "openai-video:" + DeriveSessionHashFromSeed(fmt.Sprintf("%d:%s", userID, taskID))
}

func OpenAIVideoUsageRequestID(taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return ""
	}
	requestID := openAIVideoUsageRequestIDPrefix + taskID
	if len(requestID) <= 64 {
		return requestID
	}
	sum := sha256.Sum256([]byte(taskID))
	return openAIVideoUsageRequestIDPrefix + hex.EncodeToString(sum[:])[:50]
}

func OpenAIVideoRefundRequestID(taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return ""
	}
	requestID := openAIVideoRefundRequestIDPrefix + taskID
	if len(requestID) <= 64 {
		return requestID
	}
	sum := sha256.Sum256([]byte(taskID))
	return openAIVideoRefundRequestIDPrefix + hex.EncodeToString(sum[:])[:44]
}

func OpenAIVideoBillingParameters(model string, body []byte) (string, int) {
	if spec, ok := lookupViraleeVideoSpec(model); ok {
		duration := int(gjson.GetBytes(body, "duration").Int())
		if duration <= 0 {
			duration = spec.defaultSeconds
		}
		return spec.resolution, duration
	}
	model = strings.ToLower(strings.TrimSpace(model))
	resolution := strings.TrimSpace(gjson.GetBytes(body, "resolution").String())
	durationSeconds := int(gjson.GetBytes(body, "duration").Int())
	if durationSeconds <= 0 {
		durationSeconds = int(gjson.GetBytes(body, "seconds").Int())
	}

	switch {
	case IsFireflyV2VideoModel(model):
		if resolution == "" {
			resolution = VideoBillingResolution720P
		}
		if durationSeconds <= 0 {
			durationSeconds = 5
		}
	case strings.HasPrefix(model, "video-ds-2.0"), strings.HasPrefix(model, "as-sd2.0"):
		resolution = VideoBillingResolution720P
	case strings.Contains(model, "landscape"), strings.Contains(model, "portrait"):
		resolution = VideoBillingResolution1080P
		if durationSeconds <= 0 {
			switch {
			case strings.Contains(model, "-12s"):
				durationSeconds = 12
			case strings.Contains(model, "-8s"):
				durationSeconds = 8
			}
		}
	}

	if strings.TrimSpace(resolution) == "" {
		resolution = VideoBillingResolution480P
	}
	return resolution, NormalizeVideoBillingDurationSecondsOrDefault(durationSeconds)
}

func (s *OpenAIGatewayService) BindOpenAIVideoTaskAccount(ctx context.Context, groupID *int64, userID int64, taskID string, accountID int64) error {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" || userID <= 0 || accountID <= 0 {
		return nil
	}

	expiresAt := time.Now().Add(openAIVideoTaskBindingTTL)
	groupKey := derefGroupID(groupID)
	var persistErr error
	if s != nil && s.openAIVideoTaskBindingRepo != nil {
		persistErr = s.openAIVideoTaskBindingRepo.UpsertOpenAIVideoTaskBinding(ctx, OpenAIVideoTaskBinding{
			GroupID:   groupKey,
			UserID:    userID,
			TaskID:    taskID,
			AccountID: accountID,
			ExpiresAt: expiresAt,
		})
	}

	var cacheErr error
	if s != nil {
		cacheErr = s.setStickySessionAccountID(ctx, groupID, OpenAIVideoTaskSessionHash(userID, taskID), accountID, openAIVideoTaskBindingTTL)
	}
	return errors.Join(persistErr, cacheErr)
}

func (s *OpenAIGatewayService) RestoreOpenAIVideoTaskStickySession(ctx context.Context, groupID *int64, userID int64, taskID string) (int64, error) {
	taskID = strings.TrimSpace(taskID)
	if s == nil || taskID == "" || userID <= 0 {
		return 0, nil
	}

	sessionHash := OpenAIVideoTaskSessionHash(userID, taskID)
	if sessionHash == "" {
		return 0, nil
	}
	if accountID, err := s.getStickySessionAccountID(ctx, groupID, sessionHash); err == nil && accountID > 0 {
		return accountID, nil
	}

	if s.openAIVideoTaskBindingRepo == nil {
		return 0, nil
	}
	binding, err := s.openAIVideoTaskBindingRepo.GetOpenAIVideoTaskBinding(ctx, derefGroupID(groupID), userID, taskID)
	if err != nil || binding == nil || binding.AccountID <= 0 {
		return 0, err
	}
	if cacheErr := s.setStickySessionAccountID(ctx, groupID, sessionHash, binding.AccountID, openAIVideoTaskBindingTTL); cacheErr != nil {
		return binding.AccountID, cacheErr
	}
	return binding.AccountID, nil
}

func (e OpenAIVideoEndpoint) httpMethod() string {
	if e == OpenAIVideoEndpointDelete {
		return http.MethodDelete
	}
	if e.requiresRequestBody() {
		return http.MethodPost
	}
	return http.MethodGet
}

func (e OpenAIVideoEndpoint) requiresRequestBody() bool {
	return e == OpenAIVideoEndpointCreate || e == OpenAIVideoEndpointContext || e == OpenAIVideoEndpointRegenerate
}

func (e OpenAIVideoEndpoint) createsTask() bool {
	return e.requiresRequestBody()
}

func (e OpenAIVideoEndpoint) RequiresRequestBody() bool {
	return e.requiresRequestBody()
}

func (e OpenAIVideoEndpoint) CreatesTask() bool {
	return e.createsTask()
}

func buildOpenAIVideoEndpointURL(base string, endpoint OpenAIVideoEndpoint, taskID string) string {
	return buildOpenAIVideoEndpointURLForUpstream(base, endpoint, taskID)
}

func buildOpenAIVideoEndpointURLForUpstream(base string, endpoint OpenAIVideoEndpoint, taskID string) string {
	switch endpoint {
	case OpenAIVideoEndpointCreate:
		return buildOpenAIEndpointURL(base, "/v1/videos")
	case OpenAIVideoEndpointStatus:
		return buildOpenAIEndpointURL(base, "/v1/videos/"+url.PathEscape(strings.TrimSpace(taskID)))
	case OpenAIVideoEndpointContent:
		return buildOpenAIEndpointURL(base, "/v1/videos/"+url.PathEscape(strings.TrimSpace(taskID))+"/content")
	default:
		return buildOpenAIEndpointURL(base, "/v1/videos")
	}
}

func OpenAIVideoUpstreamEndpointPath(endpoint OpenAIVideoEndpoint, upstreamModel string) string {
	return "/v1/videos"
}

func OpenAIVideoEndpointURLForBase(base string, endpoint OpenAIVideoEndpoint, taskID string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "https://api.openai.com"
	}
	return buildOpenAIVideoEndpointURL(base, endpoint, taskID)
}

func extractOpenAIVideoTaskID(body []byte) string {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ""
	}
	for _, path := range []string{"task_id", "id", "data.task_id", "data.id", "video.task_id", "video.id"} {
		if id := strings.TrimSpace(gjson.GetBytes(body, path).String()); id != "" {
			return id
		}
	}
	return ""
}

func OpenAIVideoStatusFailed(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	if gjson.ValidBytes(body) {
		for _, path := range []string{"status", "data.status", "video.status", "task.status", "result.status"} {
			status := strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, path).String()))
			switch status {
			case "failed", "failure", "error", "cancelled", "canceled":
				return true
			}
		}
		return openAIVideoTerminalErrorFailed(body)
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(string(body))), "generation failed")
}

func OpenAIVideoStatusSucceeded(body []byte) bool {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return false
	}
	for _, path := range []string{"status", "data.status", "video.status", "task.status", "result.status"} {
		status := strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, path).String()))
		switch status {
		case "completed", "succeeded", "success", "finished":
			return true
		}
	}
	return false
}

func openAIVideoTerminalErrorFailed(body []byte) bool {
	errorType := strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "error.type").String()))
	errorMessage := strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "error.message").String()))
	if errorMessage == "" {
		errorValue := gjson.GetBytes(body, "error")
		if errorValue.Type == gjson.String {
			errorMessage = strings.ToLower(strings.TrimSpace(errorValue.String()))
		} else {
			errorMessage = strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "message").String()))
		}
	}
	if !strings.Contains(errorMessage, "generation failed") {
		return false
	}
	return errorType == "" || errorType == "server_error" || errorType == "upstream_error"
}

func (s *OpenAIGatewayService) RefundFailedOpenAIVideoTask(ctx context.Context, apiKey *APIKey, taskID string, apiKeyService APIKeyQuotaUpdater) error {
	return s.RefundFailedOpenAIVideoTaskWithBillingID(ctx, apiKey, taskID, taskID, apiKeyService)
}

func (s *OpenAIGatewayService) RefundFailedOpenAIVideoTaskWithBillingID(ctx context.Context, apiKey *APIKey, taskID, billingTaskID string, apiKeyService APIKeyQuotaUpdater) error {
	taskID = strings.TrimSpace(taskID)
	billingTaskID = strings.TrimSpace(billingTaskID)
	if s == nil || taskID == "" || apiKey == nil || apiKey.ID <= 0 {
		return nil
	}
	if billingTaskID == "" {
		billingTaskID = taskID
	}
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		return nil
	}
	if s.usageLogRepo == nil || s.usageBillingRepo == nil {
		return nil
	}

	originalRequestIDs := openAIVideoUsageRequestIDCandidates(billingTaskID)
	refundRequestID := OpenAIVideoRefundRequestID(billingTaskID)
	if len(originalRequestIDs) == 0 || refundRequestID == "" {
		return nil
	}

	original, err := s.getOpenAIVideoUsageLogForRefund(ctx, originalRequestIDs, apiKey.ID)
	if err != nil {
		return err
	}
	if !isRefundableOpenAIVideoUsageLog(original) {
		return nil
	}
	account, err := s.accountRepo.GetByID(ctx, original.AccountID)
	if err != nil {
		return err
	}

	cmd := &UsageBillingCommand{
		RequestID:             refundRequestID,
		APIKeyID:              original.APIKeyID,
		UserID:                original.UserID,
		AccountID:             original.AccountID,
		AccountType:           account.Type,
		Model:                 original.Model,
		BillingType:           original.BillingType,
		MediaType:             "video",
		RequestPayloadHash:    original.RequestID,
		RestoreUsageCreatedAt: original.CreatedAt,
	}
	if original.BillingType == BillingTypeSubscription && original.SubscriptionID != nil {
		cmd.SubscriptionID = cloneInt64Ptr(original.SubscriptionID)
		cmd.SubscriptionCost = -original.ActualCost
	} else {
		cmd.BalanceCost = -original.ActualCost
	}
	if apiKey.Quota > 0 {
		cmd.APIKeyQuotaCost = -original.ActualCost
	}
	if apiKey.HasRateLimits() {
		cmd.APIKeyRateLimitCost = -original.ActualCost
	}
	if account.IsAPIKeyOrBedrock() && account.HasAnyQuotaLimit() {
		accountMultiplier := account.BillingRateMultiplier()
		if original.AccountRateMultiplier != nil {
			accountMultiplier = *original.AccountRateMultiplier
		}
		cmd.AccountQuotaCost = -(original.TotalCost * accountMultiplier)
	}
	cmd.Normalize()

	atomicRepo, ok := s.usageBillingRepo.(AtomicUsageBillingRepository)
	if !ok {
		return errors.New("atomic usage billing repository is required for video refunds")
	}
	refundLog := buildOpenAIVideoRefundUsageLog(original, refundRequestID)
	result, err := atomicRepo.ApplyWithUsageLog(ctx, cmd, refundLog)
	if err != nil {
		return err
	}
	if result == nil || !result.Applied {
		return nil
	}
	if original.BillingType == BillingTypeSubscription && original.GroupID != nil && s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateSubscription(ctx, original.UserID, *original.GroupID)
	} else {
		s.invalidateOpenAIMediaBalanceCache(ctx, original.UserID)
	}
	if apiKey.HasRateLimits() && s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateAPIKeyRateLimit(ctx, original.APIKeyID)
	}
	if invalidator, ok := apiKeyService.(apiKeyAuthCacheInvalidator); ok && apiKey.Key != "" {
		invalidator.InvalidateAuthCacheByKey(ctx, apiKey.Key)
	}
	return nil
}

func openAIVideoUsageRequestIDCandidates(billingTaskID string) []string {
	requestID := OpenAIVideoUsageRequestID(billingTaskID)
	if requestID == "" {
		return nil
	}
	return []string{requestID, "grok-video:" + requestID}
}

func (s *OpenAIGatewayService) getOpenAIVideoUsageLogForRefund(ctx context.Context, requestIDs []string, apiKeyID int64) (*UsageLog, error) {
	const attempts = 6
	for attempt := 0; attempt < attempts; attempt++ {
		log, err := s.findOpenAIVideoUsageLogForRefund(ctx, requestIDs, apiKeyID)
		if err == nil {
			return log, nil
		}
		if !errors.Is(err, ErrUsageLogNotFound) {
			return nil, err
		}
		if attempt == attempts-1 {
			return nil, fmt.Errorf("video usage log not found after retry: %w", err)
		}
		delay := time.Duration(1<<attempt) * 50 * time.Millisecond
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, ErrUsageLogNotFound
}

func (s *OpenAIGatewayService) findOpenAIVideoUsageLogForRefund(ctx context.Context, requestIDs []string, apiKeyID int64) (*UsageLog, error) {
	var found *UsageLog
	for _, requestID := range requestIDs {
		log, err := s.usageLogRepo.GetByRequestIDAndAPIKey(ctx, requestID, apiKeyID)
		if errors.Is(err, ErrUsageLogNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if found != nil {
			return nil, fmt.Errorf("multiple video usage logs found for refund: request_ids=%q api_key_id=%d", requestIDs, apiKeyID)
		}
		found = log
	}
	if found == nil {
		return nil, ErrUsageLogNotFound
	}
	return found, nil
}

func isRefundableOpenAIVideoUsageLog(log *UsageLog) bool {
	if log == nil || log.ActualCost <= 0 || (log.BillingType != BillingTypeBalance && log.BillingType != BillingTypeSubscription) {
		return false
	}
	if log.MediaType != nil && !strings.EqualFold(strings.TrimSpace(*log.MediaType), "video") {
		return false
	}
	if log.APIKeyID <= 0 || log.UserID <= 0 || log.AccountID <= 0 {
		return false
	}
	return true
}

func buildOpenAIVideoRefundUsageLog(original *UsageLog, refundRequestID string) *UsageLog {
	mediaType := "video"
	refund := &UsageLog{
		UserID:                    original.UserID,
		APIKeyID:                  original.APIKeyID,
		AccountID:                 original.AccountID,
		RequestID:                 refundRequestID,
		Model:                     original.Model,
		RequestedModel:            original.RequestedModel,
		UpstreamModel:             cloneStringPtr(original.UpstreamModel),
		GroupID:                   cloneInt64Ptr(original.GroupID),
		SubscriptionID:            cloneInt64Ptr(original.SubscriptionID),
		InputTokens:               0,
		OutputTokens:              0,
		CacheCreationTokens:       0,
		CacheReadTokens:           0,
		CacheCreation5mTokens:     0,
		CacheCreation1hTokens:     0,
		InputCost:                 0,
		OutputCost:                0,
		CacheCreationCost:         0,
		CacheReadCost:             0,
		TotalCost:                 -original.TotalCost,
		ActualCost:                -original.ActualCost,
		RateMultiplier:            original.RateMultiplier,
		AccountRateMultiplier:     cloneFloat64Ptr(original.AccountRateMultiplier),
		BillingType:               original.BillingType,
		RequestType:               RequestTypeSync,
		MediaType:                 &mediaType,
		BillingMode:               cloneStringPtr(original.BillingMode),
		BillingTier:               cloneStringPtr(original.BillingTier),
		VideoCount:                original.VideoCount,
		VideoResolution:           cloneStringPtr(original.VideoResolution),
		VideoDurationSeconds:      cloneIntPtr(original.VideoDurationSeconds),
		VideoInputDurationSeconds: original.VideoInputDurationSeconds,
		VideoOutputCost:           -original.VideoOutputCost,
		VideoInputCost:            -original.VideoInputCost,
		ChannelID:                 cloneInt64Ptr(original.ChannelID),
		ModelMappingChain:         cloneStringPtr(original.ModelMappingChain),
		ServiceTier:               cloneStringPtr(original.ServiceTier),
		ReasoningEffort:           cloneStringPtr(original.ReasoningEffort),
		InboundEndpoint:           optionalTrimmedStringPtr("/v1/videos/{task_id}"),
		UpstreamEndpoint:          cloneStringPtr(original.UpstreamEndpoint),
		CreatedAt:                 time.Now(),
	}
	if refund.RequestedModel == "" {
		refund.RequestedModel = original.Model
	}
	return refund
}

func cloneStringPtr(v *string) *string {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

func cloneFloat64Ptr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

func cloneIntPtr(v *int) *int {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

func isOpenAIVideoMultipartContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	return err == nil && strings.EqualFold(mediaType, "multipart/form-data")
}

func rewriteOpenAIVideoMultipartModel(body []byte, contentType string, upstreamModel string) ([]byte, error) {
	_, params, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	if err != nil || strings.TrimSpace(params["boundary"]) == "" {
		return nil, fmt.Errorf("rewrite video multipart model: invalid Content-Type")
	}
	boundary := params["boundary"]
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	var rewritten bytes.Buffer
	writer := multipart.NewWriter(&rewritten)
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("rewrite video multipart model: %w", err)
	}
	replaced := false
	for {
		part, err := reader.NextRawPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("rewrite video multipart model: %w", err)
		}
		destination, err := writer.CreatePart(part.Header)
		if err != nil {
			_ = part.Close()
			return nil, fmt.Errorf("rewrite video multipart model: %w", err)
		}
		if part.FormName() == "model" && part.FileName() == "" {
			_, err = io.WriteString(destination, upstreamModel)
			replaced = true
		} else {
			_, err = io.Copy(destination, part)
		}
		_ = part.Close()
		if err != nil {
			return nil, fmt.Errorf("rewrite video multipart model: %w", err)
		}
	}
	if !replaced {
		return nil, fmt.Errorf("rewrite video multipart model: model field is required")
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("rewrite video multipart model: %w", err)
	}
	return rewritten.Bytes(), nil
}

func (s *OpenAIGatewayService) ForwardOpenAIVideo(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	endpoint OpenAIVideoEndpoint,
	taskID string,
	clientTaskID string,
	body []byte,
	contentType string,
	requestModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()
	if account == nil {
		return nil, fmt.Errorf("openai account is required")
	}
	if account.Type != AccountTypeAPIKey || !account.IsOpenAI() {
		return nil, fmt.Errorf("video endpoint requires an OpenAI APIKey account")
	}

	upstreamModel := ""
	forwardBody := body
	if endpoint.createsTask() {
		if !IsOpenAIVideoModel(requestModel) {
			return nil, fmt.Errorf("unsupported video model %q", requestModel)
		}
		upstreamModel = canonicalOpenAIVideoModel(account.GetMappedModel(requestModel))
		if !IsOpenAIVideoModel(upstreamModel) {
			return nil, fmt.Errorf("video upstream model must be %s, got %q", supportedOpenAIVideoModelsText, upstreamModel)
		}
		if IsMiniMaxH3VideoModel(upstreamModel) {
			if !AccountUsesMiniMaxV2Video(account) {
				return nil, fmt.Errorf("video upstream model %q requires a MiniMax H3 video capable account", upstreamModel)
			}
			rewritten, err := prepareMiniMaxV2VideoBody(endpoint, body, upstreamModel)
			if err != nil {
				return nil, err
			}
			forwardBody = rewritten
		} else if isOpenAIVideoMultipartContentType(contentType) {
			if !IsFireflyV2VideoModel(upstreamModel) {
				return nil, fmt.Errorf("multipart video upstream model must be firefly-video-v2 or firefly-video-v2-fast, got %q", upstreamModel)
			}
			rewritten, err := rewriteOpenAIVideoMultipartModel(body, contentType, upstreamModel)
			if err != nil {
				return nil, err
			}
			forwardBody = rewritten
		} else {
			rewritten, err := sjson.SetBytes(body, "model", upstreamModel)
			if err != nil {
				return nil, fmt.Errorf("rewrite video request model: %w", err)
			}
			forwardBody = rewritten
		}
	} else if endpoint == OpenAIVideoEndpointList {
		upstreamModel = openAIVideoModelMiniMaxH3Canonical
	}
	miniMaxProtocol := (IsMiniMaxH3VideoModel(upstreamModel) || miniMaxVideoEndpointOnly(endpoint) || endpoint == OpenAIVideoEndpointStatus || endpoint == OpenAIVideoEndpointContent) && AccountUsesMiniMaxV2Video(account)
	if miniMaxVideoEndpointOnly(endpoint) && !miniMaxProtocol {
		return nil, fmt.Errorf("video endpoint %s requires a MiniMax H3 video capable account", endpoint)
	}
	if miniMaxProtocol && upstreamModel == "" {
		upstreamModel = openAIVideoModelMiniMaxH3Canonical
	}

	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}
	baseURL := account.GetOpenAIBaseURL()
	validatedURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	if miniMaxProtocol && endpoint == OpenAIVideoEndpointContent {
		result, contentErr := s.forwardMiniMaxVideoContent(ctx, c, account, token, validatedURL, taskID)
		if result != nil && strings.TrimSpace(clientTaskID) != "" {
			result.ResponseID = strings.TrimSpace(clientTaskID)
		}
		return result, contentErr
	}

	preDeleteTerminalFailure := false
	if miniMaxProtocol && endpoint == OpenAIVideoEndpointDelete {
		statusResp, statusBody, statusErr := s.fetchMiniMaxVideoStatus(ctx, c, account, token, validatedURL, taskID)
		if statusErr != nil {
			return nil, statusErr
		}
		if statusResp == nil {
			return nil, fmt.Errorf("MiniMax video status response is empty")
		}
		if statusResp.StatusCode >= http.StatusBadRequest {
			statusURL := buildMiniMaxV2VideoEndpointURL(validatedURL, OpenAIVideoEndpointStatus, taskID)
			return s.handleOpenAIVideoErrorResponse(ctx, statusResp, c, account, statusURL, upstreamModel, taskID, false)
		}
		preDeleteTerminalFailure = miniMaxVideoTaskFailed(statusBody)
	}

	targetURL := buildOpenAIVideoEndpointURLForUpstream(validatedURL, endpoint, taskID)
	if miniMaxProtocol {
		targetURL = buildMiniMaxV2VideoEndpointURL(validatedURL, endpoint, taskID)
	}
	if c != nil && c.Request != nil && c.Request.URL != nil && c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	var bodyReader io.Reader
	if endpoint.requiresRequestBody() {
		bodyReader = bytes.NewReader(forwardBody)
	}
	method := endpoint.httpMethod()
	if endpoint == OpenAIVideoEndpointContent && c != nil && c.Request != nil && c.Request.Method == http.MethodHead {
		method = http.MethodHead
	}
	upstreamReq, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, err
	}
	upstreamReq = upstreamReq.WithContext(WithHTTPUpstreamProfile(upstreamReq.Context(), HTTPUpstreamProfileOpenAI))
	upstreamReq.Header.Set("Authorization", "Bearer "+token)
	applyOpenAIVideoRequestHeaders(upstreamReq, c, endpoint, contentType)
	account.ApplyHeaderOverrides(upstreamReq.Header)
	if customUA := account.GetOpenAIUserAgent(); customUA != "" {
		upstreamReq.Header.Set("User-Agent", customUA)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	upstreamStart := time.Now()
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
			Kind:               "request_error",
			Message:            safeErr,
		})
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}

	var wrappedMiniMaxTimeoutBody []byte
	if resp.StatusCode == http.StatusOK && miniMaxProtocol && endpoint == OpenAIVideoEndpointCreate && strings.TrimSpace(clientTaskID) != "" {
		wrappedMiniMaxTimeoutBody, err = ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
		if err != nil {
			_ = resp.Body.Close()
			return nil, err
		}
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(wrappedMiniMaxTimeoutBody))
		if !isMiniMaxWrappedGatewayTimeout(wrappedMiniMaxTimeoutBody) {
			wrappedMiniMaxTimeoutBody = nil
		}
	}
	if (resp.StatusCode == http.StatusGatewayTimeout || wrappedMiniMaxTimeoutBody != nil) && miniMaxProtocol && endpoint == OpenAIVideoEndpointCreate && strings.TrimSpace(clientTaskID) != "" {
		errorBody := wrappedMiniMaxTimeoutBody
		if errorBody == nil {
			errorBody = s.readUpstreamErrorBody(resp)
		}
		_ = resp.Body.Close()
		upstreamMsg := sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(errorBody))
		if upstreamMsg == "" {
			upstreamMsg = "MiniMax-H3 submission confirmation timed out"
		}
		requestID := firstNonEmptyString(strings.TrimSpace(resp.Header.Get("x-request-id")), strings.TrimSpace(gjson.GetBytes(errorBody, "request_id").String()))
		setOpsUpstreamError(c, http.StatusGatewayTimeout, upstreamMsg, requestID)
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: http.StatusGatewayTimeout,
			UpstreamRequestID:  requestID,
			UpstreamURL:        safeUpstreamURL(upstreamReq.URL.String()),
			Kind:               "submission_recovery",
			Message:            upstreamMsg,
		})
		videoResolution, videoDurationSeconds := OpenAIVideoBillingParameters(requestModel, forwardBody)
		return &OpenAIForwardResult{
			RequestID:            requestID,
			ResponseID:           strings.TrimSpace(clientTaskID),
			BillingTaskID:        strings.TrimSpace(clientTaskID),
			Model:                requestModel,
			UpstreamModel:        upstreamModel,
			UpstreamEndpoint:     upstreamReq.URL.Path,
			ResponseHeaders:      resp.Header.Clone(),
			ResponseBody:         append([]byte(nil), errorBody...),
			MediaType:            "video",
			VideoCount:           1,
			VideoResolution:      videoResolution,
			VideoDurationSeconds: videoDurationSeconds,
			SubmissionUncertain:  true,
			Duration:             time.Since(startTime),
		}, nil
	}
	if resp.StatusCode >= 400 {
		allowFailover := endpoint.createsTask()
		responseTaskID := strings.TrimSpace(taskID)
		if endpoint != OpenAIVideoEndpointCreate && strings.TrimSpace(clientTaskID) != "" {
			responseTaskID = strings.TrimSpace(clientTaskID)
		}
		return s.handleOpenAIVideoErrorResponse(ctx, resp, c, account, upstreamReq.URL.String(), upstreamModel, responseTaskID, allowFailover)
	}
	defer func() { _ = resp.Body.Close() }()

	if endpoint == OpenAIVideoEndpointContent {
		if err := s.streamOpenAIVideoContentResponse(resp, c); err != nil {
			return nil, err
		}
		return &OpenAIForwardResult{
			RequestID:       resp.Header.Get("x-request-id"),
			ResponseID:      firstNonEmptyString(strings.TrimSpace(clientTaskID), strings.TrimSpace(taskID)),
			Model:           requestModel,
			UpstreamModel:   upstreamModel,
			ResponseHeaders: resp.Header.Clone(),
			Duration:        time.Since(startTime),
		}, nil
	}

	respBody, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	if err != nil {
		return nil, err
	}
	clientBody := respBody
	if miniMaxProtocol {
		clientBody = rewriteMiniMaxVideoContentURLs(respBody, endpoint, clientTaskID)
	}
	writeOpenAIVideoJSONResponse(c, resp, clientBody, s)
	responseID := strings.TrimSpace(taskID)
	if endpoint.createsTask() {
		responseID = extractOpenAIVideoTaskID(respBody)
	} else if strings.TrimSpace(clientTaskID) != "" {
		responseID = strings.TrimSpace(clientTaskID)
	}
	upstreamEndpoint := ""
	if parsedURL, parseErr := url.Parse(targetURL); parseErr == nil {
		upstreamEndpoint = parsedURL.Path
	}
	videoResolution := ""
	videoDurationSeconds := 0
	mediaType := ""
	videoCount := 0
	if endpoint.createsTask() {
		billingBody := forwardBody
		if parsed, parseErr := ParseOpenAIVideoOperationRequestWithContentType(endpoint, forwardBody, contentType); parseErr == nil {
			billingBody = parsed.BillingBody
		}
		videoResolution, videoDurationSeconds = OpenAIVideoBillingParameters(requestModel, billingBody)
	}
	if endpoint.createsTask() {
		mediaType = "video"
		videoCount = 1
	}
	return &OpenAIForwardResult{
		RequestID:            resp.Header.Get("x-request-id"),
		ResponseID:           responseID,
		BillingTaskID:        firstNonEmptyString(strings.TrimSpace(clientTaskID), responseID),
		Model:                requestModel,
		UpstreamModel:        upstreamModel,
		UpstreamEndpoint:     upstreamEndpoint,
		ResponseHeaders:      resp.Header.Clone(),
		ResponseBody:         append([]byte(nil), respBody...),
		TaskTerminalFailure:  preDeleteTerminalFailure || OpenAIVideoStatusFailed(respBody),
		MediaType:            mediaType,
		VideoCount:           videoCount,
		VideoResolution:      videoResolution,
		VideoDurationSeconds: videoDurationSeconds,
		Duration:             time.Since(startTime),
	}, nil
}

func isMiniMaxWrappedGatewayTimeout(body []byte) bool {
	if len(body) == 0 || !gjson.ValidBytes(body) || !gjson.GetBytes(body, "error").IsObject() {
		return false
	}
	return strings.TrimSpace(gjson.GetBytes(body, "error.http_code").String()) == strconv.Itoa(http.StatusGatewayTimeout)
}

func applyOpenAIVideoRequestHeaders(req *http.Request, c *gin.Context, endpoint OpenAIVideoEndpoint, contentType string) {
	if req == nil {
		return
	}
	if c != nil && c.Request != nil {
		for _, key := range []string{"Accept", "Accept-Language", "User-Agent"} {
			if value := strings.TrimSpace(c.Request.Header.Get(key)); value != "" {
				req.Header.Set(key, value)
			}
		}
		if endpoint == OpenAIVideoEndpointContent {
			for _, key := range []string{"Range", "If-Range"} {
				if value := strings.TrimSpace(c.Request.Header.Get(key)); value != "" {
					req.Header.Set(key, value)
				}
			}
		}
	}
	if endpoint.requiresRequestBody() {
		contentType = strings.TrimSpace(contentType)
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}
	if req.Header.Get("Accept") == "" {
		if endpoint == OpenAIVideoEndpointContent {
			req.Header.Set("Accept", "*/*")
		} else {
			req.Header.Set("Accept", "application/json")
		}
	}
}

func (s *OpenAIGatewayService) handleOpenAIVideoErrorResponse(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	account *Account,
	upstreamURL string,
	upstreamModel string,
	taskID string,
	allowFailover bool,
) (*OpenAIForwardResult, error) {
	body := s.readUpstreamErrorBody(resp)
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(body))
	if upstreamMsg == "" {
		upstreamMsg = fmt.Sprintf("video upstream returned status %d", resp.StatusCode)
	}
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
	setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, "")

	if allowFailover && s.shouldFailoverOpenAIUpstreamResponse(account, resp.StatusCode, upstreamMsg, body) {
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  resp.Header.Get("x-request-id"),
			UpstreamURL:        safeUpstreamURL(upstreamURL),
			Kind:               "failover",
			Message:            upstreamMsg,
		})
		s.handleFailoverSideEffects(ctx, resp, account, body, upstreamModel)
		return nil, &UpstreamFailoverError{
			StatusCode:             resp.StatusCode,
			ResponseBody:           body,
			RetryableOnSameAccount: account.IsPoolMode() && account.IsPoolModeRetryableStatus(resp.StatusCode),
		}
	}

	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:           account.Platform,
		AccountID:          account.ID,
		AccountName:        account.Name,
		UpstreamStatusCode: resp.StatusCode,
		UpstreamRequestID:  resp.Header.Get("x-request-id"),
		UpstreamURL:        safeUpstreamURL(upstreamURL),
		Kind:               "http_error",
		Message:            upstreamMsg,
	})
	writeOpenAIVideoJSONResponse(c, resp, body, s)
	return &OpenAIForwardResult{
		ResponseID:      strings.TrimSpace(taskID),
		UpstreamModel:   upstreamModel,
		ResponseHeaders: resp.Header.Clone(),
		ResponseBody:    append([]byte(nil), body...),
	}, fmt.Errorf("upstream error: %d %s", resp.StatusCode, upstreamMsg)
}

func writeOpenAIVideoJSONResponse(c *gin.Context, resp *http.Response, body []byte, s *OpenAIGatewayService) {
	if c == nil || resp == nil || c.Writer == nil || c.Writer.Written() {
		return
	}
	writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/json"
	}
	MarkResponseCommitted(c)
	c.Data(resp.StatusCode, contentType, body)
}

func (s *OpenAIGatewayService) streamOpenAIVideoContentResponse(resp *http.Response, c *gin.Context) error {
	if c == nil || resp == nil || c.Writer == nil {
		return fmt.Errorf("response writer is required")
	}
	writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	for _, key := range []string{"Accept-Ranges", "Content-Range", "Content-Disposition", "Content-Length"} {
		if value := strings.TrimSpace(resp.Header.Get(key)); value != "" {
			c.Writer.Header().Set(key, value)
		}
	}
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "video/mp4"
	}
	c.Writer.Header().Set("Content-Type", contentType)
	c.Writer.WriteHeader(resp.StatusCode)
	MarkResponseCommitted(c)
	if c.Request != nil && c.Request.Method == http.MethodHead {
		return nil
	}
	_, err := io.Copy(c.Writer, resp.Body)
	return err
}
