package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPublicAssetUploadAndServe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := t.TempDir()
	cfg := &config.Config{}
	cfg.Pricing.DataDir = dataDir
	cfg.Server.FrontendURL = "https://admin.example.test/password"
	h := NewPublicAssetHandler(nil, cfg)
	groupID := int64(1)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("kind", "image"))
	part, err := writer.CreateFormFile("file", "../reference.png")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-png"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "http://sub2api:8080/pg/assets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "api.example.test")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{GroupID: &groupID})
	h.Upload(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			URL         string `json:"url"`
			Filename    string `json:"filename"`
			ContentType string `json:"content_type"`
			Size        int64  `json:"size"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, "image/png", response.Data.ContentType)
	require.Equal(t, int64(len("fake-png")), response.Data.Size)
	require.NotContains(t, response.Data.Filename, "..")
	require.True(t, strings.HasPrefix(response.Data.URL, "https://api.example.test/pg/assets/"))
	require.NotContains(t, response.Data.URL, "/password/")

	serveReq := httptest.NewRequest(http.MethodGet, response.Data.URL, nil)
	serveRec := httptest.NewRecorder()
	serveCtx, _ := gin.CreateTestContext(serveRec)
	serveCtx.Request = serveReq
	serveCtx.Params = gin.Params{{Key: "asset_id", Value: filepath.Base(filepath.Dir(strings.TrimPrefix(response.Data.URL, "https://api.example.test/pg/assets/")))}, {Key: "filename", Value: response.Data.Filename}}
	h.Serve(serveCtx)
	require.Equal(t, http.StatusOK, serveRec.Code)
	require.Equal(t, "image/png", serveRec.Header().Get("Content-Type"))
	require.Equal(t, "fake-png", serveRec.Body.String())

	entries, err := os.ReadDir(filepath.Join(dataDir, "pg", "assets"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}
