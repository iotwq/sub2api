package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBasispointsHostedToolsNeverSilentlySwitchToCodex(t *testing.T) {
	for _, tool := range []map[string]any{
		{"type": "image_generation"},
		{"type": "web_search", "external_web_access": true},
		{"type": "web_search", "search_context_size": "high"},
	} {
		for _, choice := range []any{nil, "auto", "none", "required", map[string]any{"type": tool["type"]}} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("%v/choice=%v/stream=%t", tool, choice, stream), func(t *testing.T) {
					completed, err := json.Marshal(map[string]any{"type": "response.completed", "response": map[string]any{"id": "resp_hosted", "status": "completed", "output": []any{}}})
					require.NoError(t, err)
					delta, err := json.Marshal(map[string]any{"type": "response.output_text.delta", "delta": "hello"})
					require.NoError(t, err)
					u := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("data: " + string(delta) + "\n\ndata: " + string(completed) + "\n\n"))}}
					source := map[string]any{"model": "gpt-6-astra", "input": "hello", "stream": stream, "tools": []any{tool, map[string]any{"type": "function", "name": "lookup_client", "parameters": map[string]any{"type": "object"}}}}
					if choice != nil {
						source["tool_choice"] = choice
					}
					body, err := json.Marshal(source)
					require.NoError(t, err)
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
					_, err = openAIClientToolsTestService(u).Forward(context.Background(), c, basispointsAccountForTest(), body)
					value, _ := choice.(string)
					if choice != nil && value != "auto" && value != "none" {
						require.Error(t, err)
						require.Equal(t, http.StatusBadRequest, rec.Code)
						require.Contains(t, rec.Body.String(), "basispoints supports tool_choice auto or none only")
						require.Empty(t, u.requests)
						return
					}
					require.NoError(t, err)
					require.Equal(t, http.StatusOK, rec.Code)
					require.Len(t, u.requests, 1)
					require.Equal(t, basispointsResponsesURL, u.lastReq.URL.String())
					if value == "none" {
						require.NotContains(t, string(u.lastBody), "Hosted tools unavailable through Basispoints")
						require.NotContains(t, string(u.lastBody), "lookup_client")
					} else {
						require.Contains(t, string(u.lastBody), "Hosted tools unavailable through Basispoints: "+fmt.Sprint(tool["type"]))
						require.Contains(t, string(u.lastBody), "Do not claim to have used them")
						require.Contains(t, string(u.lastBody), "lookup_client")
					}
				})
			}
		}
	}
}

func TestBasispointsHostedToolProbeDoesNotSwitchRoute(t *testing.T) {
	u := &httpUpstreamRecorder{}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Set(bpsAccountProbeRequiredContextKey, true)
	body, err := json.Marshal(map[string]any{"model": "gpt-6-astra", "input": "hello", "tools": []any{map[string]any{"type": "image_generation"}}})
	require.NoError(t, err)
	_, err = openAIClientToolsTestService(u).Forward(context.Background(), c, basispointsAccountForTest(), body)
	require.EqualError(t, err, "bps probe path is unavailable")
	require.Empty(t, u.requests)
}
