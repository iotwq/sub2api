package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type miniMaxVideoRecoveryRepoStub struct {
	OpenAIVideoTaskBindingRepository
	OpenAIVideoRecoveryRepository
	binding *OpenAIVideoTaskBinding
}

func (s *miniMaxVideoRecoveryRepoStub) PrepareOpenAIVideoRecovery(_ context.Context, binding OpenAIVideoTaskBinding) error {
	copy := binding
	copy.ID = 1
	s.binding = &copy
	return nil
}

func (s *miniMaxVideoRecoveryRepoStub) GetOpenAIVideoRecovery(_ context.Context, groupID, userID int64, taskID string) (*OpenAIVideoTaskBinding, error) {
	if s.binding == nil || s.binding.GroupID != groupID || s.binding.UserID != userID || s.binding.TaskID != taskID {
		return nil, nil
	}
	copy := *s.binding
	return &copy, nil
}

func TestPrepareMiniMaxVideoRecoveryAllowsThirtyMinutesForDelayedSubmission(t *testing.T) {
	repo := &miniMaxVideoRecoveryRepoStub{}
	upstream := &openAIVideoCompensationHTTPStub{statusCode: http.StatusOK, body: `{"items":[]}`}
	svc := newOpenAIVideoCompensationService(repo, upstream, nil)
	account := newMiniMaxVideoTestAccount()
	apiKey := &APIKey{ID: 101}
	before := time.Now()

	recovery, err := svc.PrepareMiniMaxVideoRecovery(
		context.Background(),
		nil,
		42,
		apiKey,
		account,
		[]byte(`{"model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9"}`),
		OpenAIVideoRecoveryBillingSnapshot{UpstreamModel: "MiniMax-H3"},
	)
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, recovery)
	require.NotNil(t, recovery.Binding.RecoveryExpiresAt)
	require.False(t, recovery.Binding.RecoveryExpiresAt.Before(before.Add(30*time.Minute)))
	require.False(t, recovery.Binding.RecoveryExpiresAt.After(after.Add(30*time.Minute)))
}

func TestBuildMiniMaxVideoRecoverySignatureCountsReferenceMedia(t *testing.T) {
	body := []byte(`{
		"model":"MiniMax-H3",
		"resolution":"2K",
		"duration":8,
		"ratio":"16:9",
		"content":[
			{"type":"text","text":"scene"},
			{"type":"image_url","image_url":{"url":"https://example.com/a.jpg"}},
			{"type":"video_url","video_url":{"url":"https://example.com/a.mp4"}},
			{"type":"audio_url","audio_url":{"url":"https://example.com/a.mp3"}}
		]
	}`)

	signature := BuildMiniMaxVideoRecoverySignature("MiniMax-H3", body)

	require.Equal(t, "MiniMax-H3", signature.Model)
	require.Equal(t, "2k", signature.Resolution)
	require.Equal(t, 8, signature.DurationSeconds)
	require.Equal(t, "16:9", signature.Ratio)
	require.Equal(t, 1, signature.ImageCount)
	require.Equal(t, 1, signature.VideoCount)
	require.Equal(t, 1, signature.AudioCount)
}

func TestFindMiniMaxVideoRecoveryCandidatesTreatsUnavailableMediaCountsAsUnknown(t *testing.T) {
	want := BuildMiniMaxVideoRecoverySignature("MiniMax-H3", []byte(`{
		"model":"MiniMax-H3",
		"resolution":"2K",
		"duration":15,
		"ratio":"16:9",
		"content":[
			{"type":"text","text":"scene"},
			{"type":"image_url","image_url":{"url":"https://example.com/a.jpg"}},
			{"type":"image_url","image_url":{"url":"https://example.com/b.jpg"}},
			{"type":"audio_url","audio_url":{"url":"https://example.com/a.mp3"}}
		]
	}`))
	require.Equal(t, 2, want.ImageCount)
	require.Equal(t, 1, want.AudioCount)
	tasks, err := parseMiniMaxVideoTaskList([]byte(`{
		"items":[
			{"id":"match","model":"MiniMax-H3","resolution":"2K","duration":15,"ratio":"16:9","input_image_count":0}
		]
	}`))
	require.NoError(t, err)

	matches := findMiniMaxVideoRecoveryCandidates(nil, want, tasks)

	require.Len(t, matches, 1)
	require.Equal(t, "match", matches[0].ID)
}

func TestFindMiniMaxVideoRecoveryCandidatesRejectsKnownPositiveMediaCountMismatch(t *testing.T) {
	want := MiniMaxVideoRecoverySignature{
		Model: "MiniMax-H3", Resolution: "2k", DurationSeconds: 15, Ratio: "16:9", ImageCount: 2, AudioCount: 1,
	}
	tasks, err := parseMiniMaxVideoTaskList([]byte(`{
		"items":[
			{"id":"wrong-image-count","model":"MiniMax-H3","resolution":"2K","duration":15,"ratio":"16:9","input_image_count":1}
		]
	}`))
	require.NoError(t, err)

	require.Empty(t, findMiniMaxVideoRecoveryCandidates(nil, want, tasks))
}

func TestFindMiniMaxVideoRecoveryCandidatesRequiresUniqueNewExactMatch(t *testing.T) {
	want := MiniMaxVideoRecoverySignature{
		Model: "MiniMax-H3", Resolution: "2k", DurationSeconds: 8, Ratio: "16:9", ImageCount: 1,
	}
	tasks, err := parseMiniMaxVideoTaskList([]byte(`{
		"items":[
			{"id":"before","model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9","image_count":1},
			{"id":"wrong-duration","model":"MiniMax-H3","resolution":"2K","duration":5,"ratio":"16:9","image_count":1},
			{"id":"match","model":"MiniMax-H3","resolution":"2K","duration":8,"ratio":"16:9","image_count":1}
		]
	}`))
	require.NoError(t, err)

	matches := findMiniMaxVideoRecoveryCandidates([]string{"before"}, want, tasks)

	require.Len(t, matches, 1)
	require.Equal(t, "match", matches[0].ID)

	tasks = append(tasks, miniMaxVideoTaskSummary{ID: "second-match", Signature: want})
	require.Len(t, findMiniMaxVideoRecoveryCandidates([]string{"before"}, want, tasks), 2)
}
