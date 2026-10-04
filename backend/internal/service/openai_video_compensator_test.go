package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type openAIVideoCompensationRepoStub struct {
	OpenAIVideoTaskBindingRepository
	OpenAIVideoRecoveryRepository
	bindings              map[string]OpenAIVideoTaskBinding
	claimed               []OpenAIVideoTaskBinding
	recoveryClaimed       []OpenAIVideoTaskBinding
	updatedStatus         string
	updatedNext           *time.Time
	updatedError          string
	updatedUpstreamTaskID string
	updatedStatuses       []string
	updateCalls           int
	claimCalls            int
	recoveryClaimCalls    int
	armCalls              int
	lastClaimLimit        int
	failMatchedUpdates    int
	persistRecoveryClaims bool
}

func (r *openAIVideoCompensationRepoStub) UpsertOpenAIVideoTaskBinding(_ context.Context, binding OpenAIVideoTaskBinding) error {
	if r.bindings == nil {
		r.bindings = make(map[string]OpenAIVideoTaskBinding)
	}
	r.bindings[openAIVideoCompensationBindingKey(binding.GroupID, binding.UserID, binding.TaskID)] = binding
	return nil
}

func (r *openAIVideoCompensationRepoStub) GetOpenAIVideoTaskBinding(_ context.Context, groupID, userID int64, taskID string) (*OpenAIVideoTaskBinding, error) {
	binding, ok := r.bindings[openAIVideoCompensationBindingKey(groupID, userID, taskID)]
	if !ok {
		return nil, nil
	}
	return &binding, nil
}

func (r *openAIVideoCompensationRepoStub) ArmOpenAIVideoTaskCompensation(_ context.Context, groupID, userID int64, taskID string, apiKey *APIKey, nextCheckAt time.Time) error {
	r.armCalls++
	key := openAIVideoCompensationBindingKey(groupID, userID, taskID)
	binding, ok := r.bindings[key]
	if !ok {
		return errors.New("binding not found")
	}
	binding.APIKeyID = apiKey.ID
	binding.CompensationStatus = openAIVideoCompensationPending
	binding.NextCheckAt = &nextCheckAt
	r.bindings[key] = binding
	return nil
}

func (r *openAIVideoCompensationRepoStub) ClaimOpenAIVideoTaskCompensations(_ context.Context, _, _ time.Time, limit int) ([]OpenAIVideoTaskBinding, error) {
	r.claimCalls++
	r.lastClaimLimit = limit
	return append([]OpenAIVideoTaskBinding(nil), r.claimed...), nil
}

func (r *openAIVideoCompensationRepoStub) UpdateOpenAIVideoTaskCompensation(_ context.Context, _ int64, status string, nextCheckAt *time.Time, lastError string) error {
	r.updateCalls++
	r.updatedStatus = status
	r.updatedNext = nextCheckAt
	r.updatedError = lastError
	return nil
}

func (r *openAIVideoCompensationRepoStub) ClaimOpenAIVideoRecoveries(_ context.Context, _, _ time.Time, limit int) ([]OpenAIVideoTaskBinding, error) {
	r.recoveryClaimCalls++
	r.lastClaimLimit = limit
	if r.persistRecoveryClaims {
		claimed := make([]OpenAIVideoTaskBinding, 0, len(r.recoveryClaimed))
		for _, binding := range r.recoveryClaimed {
			switch binding.RecoveryStatus {
			case openAIVideoRecoverySubmitting, openAIVideoRecoveryPending, openAIVideoRecoveryIdentified:
				claimed = append(claimed, binding)
			}
		}
		return claimed, nil
	}
	return append([]OpenAIVideoTaskBinding(nil), r.recoveryClaimed...), nil
}

func (r *openAIVideoCompensationRepoStub) UpdateOpenAIVideoRecovery(_ context.Context, id int64, status, upstreamTaskID string, nextCheckAt *time.Time, lastError string) error {
	r.updatedStatuses = append(r.updatedStatuses, status)
	if status == openAIVideoRecoveryMatched && r.failMatchedUpdates > 0 {
		r.failMatchedUpdates--
		return errors.New("matched update failed")
	}
	r.updatedStatus = status
	r.updatedUpstreamTaskID = upstreamTaskID
	r.updatedNext = nextCheckAt
	r.updatedError = lastError
	r.updateCalls++
	if r.persistRecoveryClaims {
		for i := range r.recoveryClaimed {
			if r.recoveryClaimed[i].ID != id {
				continue
			}
			r.recoveryClaimed[i].RecoveryStatus = status
			r.recoveryClaimed[i].UpstreamTaskID = upstreamTaskID
			r.recoveryClaimed[i].RecoveryNextCheckAt = nextCheckAt
			if lastError == "" {
				r.recoveryClaimed[i].RecoveryLastError = nil
			} else {
				r.recoveryClaimed[i].RecoveryLastError = &lastError
			}
		}
	}
	return nil
}

func openAIVideoCompensationBindingKey(groupID, userID int64, taskID string) string {
	return strings.Join([]string{strconv.FormatInt(groupID, 10), strconv.FormatInt(userID, 10), taskID}, ":")
}

type openAIVideoCompensationAccountRepoStub struct {
	AccountRepository
	account *Account
	err     error
}

func (r *openAIVideoCompensationAccountRepoStub) GetByID(_ context.Context, _ int64) (*Account, error) {
	return r.account, r.err
}

type openAIVideoCompensationHTTPStub struct {
	HTTPUpstream
	statusCode int
	body       string
	err        error
	calls      int
}

type openAIVideoRecoveryAPIKeyServiceStub struct {
	apiKey *APIKey
}

func (s *openAIVideoRecoveryAPIKeyServiceStub) GetByID(_ context.Context, _ int64) (*APIKey, error) {
	return s.apiKey, nil
}

func (s *openAIVideoRecoveryAPIKeyServiceStub) UpdateQuotaUsed(context.Context, int64, float64) error {
	return nil
}

func (s *openAIVideoRecoveryAPIKeyServiceStub) UpdateRateLimitUsage(context.Context, int64, float64) error {
	return nil
}

type openAIVideoRecoveryBillingRepoStub struct {
	UsageBillingRepository
	releaseCalls int
}

func (s *openAIVideoRecoveryBillingRepoStub) ReleaseBatchImageBalance(_ context.Context, _ *BatchImageBalanceHoldCommand) (*BatchImageBalanceHoldResult, error) {
	s.releaseCalls++
	return &BatchImageBalanceHoldResult{Applied: true}, nil
}

func (s *openAIVideoCompensationHTTPStub) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return &http.Response{
		StatusCode: s.statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

func newOpenAIVideoCompensationService(repo OpenAIVideoTaskBindingRepository, upstream HTTPUpstream, accountRepo AccountRepository) *OpenAIGatewayService {
	return &OpenAIGatewayService{
		cfg:                        &config.Config{},
		openAIVideoTaskBindingRepo: repo,
		httpUpstream:               upstream,
		accountRepo:                accountRepo,
	}
}

func newOpenAIVideoCompensationBinding() OpenAIVideoTaskBinding {
	return OpenAIVideoTaskBinding{
		ID:                 1,
		GroupID:            7,
		UserID:             42,
		TaskID:             "task-compensation",
		AccountID:          91,
		APIKeyID:           101,
		APIKeyKey:          "test-key",
		CompensationStatus: openAIVideoCompensationPending,
		CheckAttempts:      1,
		ExpiresAt:          time.Now().Add(time.Hour),
	}
}

func newMiniMaxVideoRecoveryBinding(t *testing.T) OpenAIVideoTaskBinding {
	t.Helper()
	now := time.Now()
	expiresAt := now.Add(time.Minute)
	signature, err := json.Marshal(MiniMaxVideoRecoverySignature{
		Model: "MiniMax-H3", Resolution: "2k", DurationSeconds: 8, Ratio: "16:9", ImageCount: 1,
	})
	require.NoError(t, err)
	billing, err := json.Marshal(OpenAIVideoRecoveryBillingSnapshot{})
	require.NoError(t, err)
	return OpenAIVideoTaskBinding{
		ID:                  201,
		GroupID:             7,
		UserID:              42,
		TaskID:              "task_recovery_local",
		AccountID:           91,
		APIKeyID:            101,
		BillingTaskID:       "task_recovery_local",
		RecoveryStatus:      openAIVideoRecoveryPending,
		RecoveryBaseline:    []byte(`["before"]`),
		RecoverySignature:   signature,
		RecoveryBilling:     billing,
		RecoveryExpiresAt:   &expiresAt,
		RecoveryNextCheckAt: &now,
		ExpiresAt:           now.Add(time.Hour),
	}
}

func newOpenAIVideoCompensationAccount() *Account {
	return &Account{
		ID:          91,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "upstream-key",
			"base_url": "https://api.openai.com/v1",
		},
	}
}

func TestOpenAIVideoCompensatorReschedulesPendingTask(t *testing.T) {
	repo := &openAIVideoCompensationRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{"status":"processing"}`}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newOpenAIVideoCompensationAccount()})

	svc.reconcileOpenAIVideoTask(newOpenAIVideoCompensationBinding())

	require.Equal(t, openAIVideoCompensationPending, repo.updatedStatus)
	require.NotNil(t, repo.updatedNext)
	require.True(t, repo.updatedNext.After(time.Now()))
	require.Empty(t, repo.updatedError)
}

func TestOpenAIVideoCompensatorCompletesSucceededTask(t *testing.T) {
	repo := &openAIVideoCompensationRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{"task":{"status":"succeeded"}}`}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newOpenAIVideoCompensationAccount()})

	svc.reconcileOpenAIVideoTask(newOpenAIVideoCompensationBinding())

	require.Equal(t, openAIVideoCompensationCompleted, repo.updatedStatus)
	require.Nil(t, repo.updatedNext)
}

func TestOpenAIVideoCompensatorRetriesTransientProbeError(t *testing.T) {
	repo := &openAIVideoCompensationRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusServiceUnavailable, body: `{"error":{"message":"temporarily unavailable"}}`}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newOpenAIVideoCompensationAccount()})

	svc.reconcileOpenAIVideoTask(newOpenAIVideoCompensationBinding())

	require.Equal(t, openAIVideoCompensationPending, repo.updatedStatus)
	require.NotNil(t, repo.updatedNext)
	require.Contains(t, repo.updatedError, "HTTP 503")
}

func TestOpenAIVideoCompensatorRefundsTerminalFailureOnce(t *testing.T) {
	binding := newOpenAIVideoCompensationBinding()
	mediaType := "video"
	originalRequestID := "grok-video:" + OpenAIVideoUsageRequestID(binding.TaskID)
	require.Equal(t, "grok-video:openai-video:"+binding.TaskID, originalRequestID)
	usageRepo := &openAIRecordUsageLogRepoStub{
		inserted: true,
		logByRequest: map[string]*UsageLog{
			originalRequestID: {
				UserID:      binding.UserID,
				APIKeyID:    binding.APIKeyID,
				AccountID:   binding.AccountID,
				RequestID:   originalRequestID,
				Model:       "firefly-video-v2",
				TotalCost:   19.80,
				ActualCost:  19.80,
				BillingType: BillingTypeBalance,
				MediaType:   &mediaType,
				CreatedAt:   time.Now(),
			},
		},
	}
	billingStub := &openAIRecordUsageBillingRepoStub{applied: map[string]bool{}}
	billingRepo := &openAIVideoRefundAtomicBillingRepoStub{openAIRecordUsageBillingRepoStub: billingStub, usageRepo: usageRepo}
	repo := &openAIVideoCompensationRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusInternalServerError, body: `{"error":{"message":"generation failed: upstream task failed","type":"server_error"}}`}
	accountRepo := &openAIVideoCompensationAccountRepoStub{account: newOpenAIVideoCompensationAccount()}
	svc := newOpenAIVideoCompensationService(repo, upstream, accountRepo)
	svc.usageLogRepo = usageRepo
	svc.usageBillingRepo = billingRepo

	svc.reconcileOpenAIVideoTask(binding)
	svc.reconcileOpenAIVideoTask(binding)

	require.Equal(t, openAIVideoCompensationRefunded, repo.updatedStatus)
	require.Equal(t, 2, billingStub.calls)
	require.Equal(t, 1, usageRepo.calls, "idempotent refund must write one reversal log")
	require.Equal(t, OpenAIVideoRefundRequestID(binding.TaskID), usageRepo.lastLog.RequestID)
	require.InDelta(t, -19.80, usageRepo.lastLog.ActualCost, 1e-12)
}

func TestMiniMaxVideoRecoveryWorkerResumesCandidateAfterTwentyTwoMinutesWithoutRebilling(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	createdAt := time.Now().Add(-22*time.Minute - 30*time.Second)
	expiresAt := createdAt.Add(openAIVideoRecoveryWindow)
	binding.CreatedAt = createdAt
	binding.RecoveryExpiresAt = &expiresAt
	repo := &openAIVideoCompensationRepoStub{
		bindings: map[string]OpenAIVideoTaskBinding{
			openAIVideoCompensationBindingKey(binding.GroupID, binding.UserID, binding.TaskID): binding,
		},
		recoveryClaimed: []OpenAIVideoTaskBinding{binding},
	}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{"items":[
		{"id":"before","model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9","image_count":1},
		{"id":"recovered-upstream-task","model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9","image_count":1}
	]}`}
	accountRepo := &openAIVideoCompensationAccountRepoStub{account: newMiniMaxVideoTestAccount()}
	usageRepo := &openAIRecordUsageLogRepoStub{logByRequest: map[string]*UsageLog{
		OpenAIVideoUsageRequestID(binding.BillingTaskID): {APIKeyID: binding.APIKeyID},
	}}
	svc := newOpenAIVideoCompensationService(repo, upstream, accountRepo)
	svc.usageLogRepo = usageRepo
	svc.openAIVideoCompensatorAPIKeyService = &openAIVideoRecoveryAPIKeyServiceStub{apiKey: &APIKey{ID: binding.APIKeyID}}

	svc.reconcileMiniMaxVideoRecoveries(time.Now())

	require.Equal(t, 1, repo.recoveryClaimCalls, "persisted recovery must be claimable after a service restart")
	require.Equal(t, openAIVideoRecoveryMatched, repo.updatedStatus)
	require.Equal(t, "recovered-upstream-task", repo.updatedUpstreamTaskID)
	require.Equal(t, 1, repo.armCalls)
	require.Equal(t, 1, upstream.calls)
	require.Zero(t, usageRepo.calls, "an existing usage record must prevent duplicate billing")
}

func TestMiniMaxVideoRecoveryWorkerMatchesDelayedCandidateWithUnknownMediaCountsOnce(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	createdAt := time.Now().Add(-22*time.Minute - 30*time.Second)
	expiresAt := createdAt.Add(openAIVideoRecoveryWindow)
	binding.CreatedAt = createdAt
	binding.RecoveryExpiresAt = &expiresAt
	signature, err := json.Marshal(MiniMaxVideoRecoverySignature{
		Model: "MiniMax-H3", Resolution: "2k", DurationSeconds: 15, Ratio: "16:9", ImageCount: 2, AudioCount: 1,
	})
	require.NoError(t, err)
	binding.RecoverySignature = signature
	repo := &openAIVideoCompensationRepoStub{
		bindings: map[string]OpenAIVideoTaskBinding{
			openAIVideoCompensationBindingKey(binding.GroupID, binding.UserID, binding.TaskID): binding,
		},
		recoveryClaimed:       []OpenAIVideoTaskBinding{binding},
		persistRecoveryClaims: true,
	}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{"items":[
		{"id":"before","model":"MiniMax-H3","resolution":"2K","duration":15,"ratio":"16:9","input_image_count":0}
	]}`}
	usageRepo := &openAIRecordUsageLogRepoStub{logByRequest: map[string]*UsageLog{
		OpenAIVideoUsageRequestID(binding.BillingTaskID): {APIKeyID: binding.APIKeyID},
	}}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newMiniMaxVideoTestAccount()})
	svc.usageLogRepo = usageRepo
	svc.openAIVideoCompensatorAPIKeyService = &openAIVideoRecoveryAPIKeyServiceStub{apiKey: &APIKey{ID: binding.APIKeyID}}

	svc.reconcileMiniMaxVideoRecoveries(time.Now())
	require.Equal(t, openAIVideoRecoveryPending, repo.updatedStatus)
	require.Zero(t, repo.armCalls)

	upstream.body = `{"items":[
		{"id":"before","model":"MiniMax-H3","resolution":"2K","duration":15,"ratio":"16:9","input_image_count":0},
		{"id":"recovered-upstream-task","model":"MiniMax-H3","resolution":"2K","duration":15,"ratio":"16:9","input_image_count":0}
	]}`
	svc.reconcileMiniMaxVideoRecoveries(time.Now())
	svc.reconcileMiniMaxVideoRecoveries(time.Now())

	require.Equal(t, []string{openAIVideoRecoveryPending, openAIVideoRecoveryIdentified, openAIVideoRecoveryMatched}, repo.updatedStatuses)
	require.Equal(t, openAIVideoRecoveryMatched, repo.updatedStatus)
	require.Equal(t, "recovered-upstream-task", repo.updatedUpstreamTaskID)
	require.Equal(t, 1, repo.armCalls, "a matched recovery must arm compensation once")
	require.Equal(t, 2, upstream.calls, "a matched recovery must not be claimed again")
	require.Zero(t, usageRepo.calls, "an existing usage record must prevent duplicate billing across worker runs")
}

func TestMiniMaxVideoRecoveryWorkerCompletesExpiredIdentifiedTaskWithoutTaskListLookup(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	expired := time.Now().Add(-time.Minute)
	binding.RecoveryStatus = openAIVideoRecoveryIdentified
	binding.RecoveryExpiresAt = &expired
	binding.UpstreamTaskID = "recovered-upstream-task"
	repo := &openAIVideoCompensationRepoStub{
		bindings: map[string]OpenAIVideoTaskBinding{
			openAIVideoCompensationBindingKey(binding.GroupID, binding.UserID, binding.TaskID): binding,
		},
	}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{}`}
	usageRepo := &openAIRecordUsageLogRepoStub{logByRequest: map[string]*UsageLog{
		OpenAIVideoUsageRequestID(binding.BillingTaskID): {APIKeyID: binding.APIKeyID},
	}}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newMiniMaxVideoTestAccount()})
	svc.usageLogRepo = usageRepo
	svc.openAIVideoCompensatorAPIKeyService = &openAIVideoRecoveryAPIKeyServiceStub{apiKey: &APIKey{ID: binding.APIKeyID}}

	svc.reconcileMiniMaxVideoRecovery(binding)

	require.Equal(t, openAIVideoRecoveryMatched, repo.updatedStatus)
	require.Equal(t, "recovered-upstream-task", repo.updatedUpstreamTaskID)
	require.Equal(t, 1, repo.armCalls)
	require.Zero(t, upstream.calls, "an identified task must not query the task list again")
	require.Zero(t, usageRepo.calls, "existing billing must remain idempotent")
}

func TestMiniMaxVideoRecoveryWorkerReschedulesMatchedUpdateFailure(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	binding.RecoveryStatus = openAIVideoRecoveryIdentified
	binding.UpstreamTaskID = "recovered-upstream-task"
	repo := &openAIVideoCompensationRepoStub{
		bindings: map[string]OpenAIVideoTaskBinding{
			openAIVideoCompensationBindingKey(binding.GroupID, binding.UserID, binding.TaskID): binding,
		},
		failMatchedUpdates: 1,
	}
	usageRepo := &openAIRecordUsageLogRepoStub{logByRequest: map[string]*UsageLog{
		OpenAIVideoUsageRequestID(binding.BillingTaskID): {APIKeyID: binding.APIKeyID},
	}}
	svc := newOpenAIVideoCompensationService(repo, &openAIVideoCompensationHTTPStub{}, &openAIVideoCompensationAccountRepoStub{account: newMiniMaxVideoTestAccount()})
	svc.usageLogRepo = usageRepo
	svc.openAIVideoCompensatorAPIKeyService = &openAIVideoRecoveryAPIKeyServiceStub{apiKey: &APIKey{ID: binding.APIKeyID}}

	svc.reconcileMiniMaxVideoRecovery(binding)

	require.Equal(t, []string{openAIVideoRecoveryMatched, openAIVideoRecoveryIdentified}, repo.updatedStatuses)
	require.Equal(t, openAIVideoRecoveryIdentified, repo.updatedStatus)
	require.Equal(t, "recovered-upstream-task", repo.updatedUpstreamTaskID)
	require.NotNil(t, repo.updatedNext, "matched update failure must remain claimable")
	require.Contains(t, repo.updatedError, "matched update failed")

	svc.reconcileMiniMaxVideoRecovery(binding)
	require.Equal(t, openAIVideoRecoveryMatched, repo.updatedStatus)
	require.Nil(t, repo.updatedNext)
}

func TestMiniMaxVideoRecoveryWorkerReschedulesWhenNoCandidateExists(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	hold := &OpenAIMediaBalanceHold{ID: "hold-delayed", APIKeyID: binding.APIKeyID, UserID: binding.UserID, Amount: 18}
	billing, err := json.Marshal(OpenAIVideoRecoveryBillingSnapshot{Hold: hold})
	require.NoError(t, err)
	binding.RecoveryBilling = billing
	repo := &openAIVideoCompensationRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{"items":[
		{"id":"before","model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9","image_count":1}
	]}`}
	billingRepo := &openAIVideoRecoveryBillingRepoStub{}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newMiniMaxVideoTestAccount()})
	svc.usageBillingRepo = billingRepo

	svc.reconcileMiniMaxVideoRecovery(binding)

	require.Equal(t, openAIVideoRecoveryPending, repo.updatedStatus)
	require.NotNil(t, repo.updatedNext)
	require.Empty(t, repo.updatedUpstreamTaskID)
	require.Zero(t, repo.armCalls)
	require.Zero(t, billingRepo.releaseCalls, "a hold must remain reserved before the final recovery deadline")
}

func TestMiniMaxVideoRecoveryWorkerReleasesHoldAfterFinalDeadline(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	expired := time.Now().Add(-time.Minute)
	binding.RecoveryExpiresAt = &expired
	hold := &OpenAIMediaBalanceHold{ID: "hold-expired", APIKeyID: binding.APIKeyID, UserID: binding.UserID, Amount: 18}
	billing, err := json.Marshal(OpenAIVideoRecoveryBillingSnapshot{Hold: hold})
	require.NoError(t, err)
	binding.RecoveryBilling = billing
	repo := &openAIVideoCompensationRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{}
	usageRepo := &openAIRecordUsageLogRepoStub{}
	billingRepo := &openAIVideoRecoveryBillingRepoStub{}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newMiniMaxVideoTestAccount()})
	svc.usageLogRepo = usageRepo
	svc.usageBillingRepo = billingRepo

	svc.reconcileMiniMaxVideoRecovery(binding)

	require.Equal(t, openAIVideoRecoveryFailed, repo.updatedStatus)
	require.Nil(t, repo.updatedNext)
	require.Contains(t, repo.updatedError, "recovery window expired")
	require.Equal(t, 1, billingRepo.releaseCalls)
	require.Zero(t, upstream.calls, "an expired unidentified recovery must not query the upstream task list")
}

func TestMiniMaxVideoRecoveryWorkerStopsAmbiguousMatchAndReleasesHold(t *testing.T) {
	binding := newMiniMaxVideoRecoveryBinding(t)
	hold := &OpenAIMediaBalanceHold{ID: "hold-1", APIKeyID: binding.APIKeyID, UserID: binding.UserID, Amount: 1.5}
	billing, err := json.Marshal(OpenAIVideoRecoveryBillingSnapshot{Hold: hold})
	require.NoError(t, err)
	binding.RecoveryBilling = billing
	repo := &openAIVideoCompensationRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{"items":[
		{"id":"candidate-1","model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9","image_count":1},
		{"id":"candidate-2","model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9","image_count":1}
	]}`}
	usageRepo := &openAIRecordUsageLogRepoStub{}
	billingRepo := &openAIVideoRecoveryBillingRepoStub{}
	svc := newOpenAIVideoCompensationService(repo, upstream, &openAIVideoCompensationAccountRepoStub{account: newMiniMaxVideoTestAccount()})
	svc.usageLogRepo = usageRepo
	svc.usageBillingRepo = billingRepo

	svc.reconcileMiniMaxVideoRecovery(binding)

	require.Equal(t, openAIVideoRecoveryAmbiguous, repo.updatedStatus)
	require.Contains(t, repo.updatedError, "multiple matching upstream tasks")
	require.Equal(t, 1, billingRepo.releaseCalls)
	require.Zero(t, repo.armCalls)
}
