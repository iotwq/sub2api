package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const publicAssetTTL = 24 * time.Hour

type publicAssetKindSpec struct {
	Extensions map[string]string
	MaxBytes   int64
}

var publicAssetKinds = map[string]publicAssetKindSpec{
	"image": {
		Extensions: map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".webp": "image/webp"},
		MaxBytes:   32 << 20,
	},
	"video": {
		Extensions: map[string]string{".mp4": "video/mp4", ".mov": "video/quicktime", ".webm": "video/webm"},
		MaxBytes:   100 << 20,
	},
	"audio": {
		Extensions: map[string]string{".mp3": "audio/mpeg", ".m4a": "audio/mp4", ".wav": "audio/wav", ".aac": "audio/aac", ".ogg": "audio/ogg"},
		MaxBytes:   15 << 20,
	},
}

type PublicAssetHandler struct {
	assetsDir       string
	cleanupInterval time.Duration
	config          *config.Config
}

func NewPublicAssetHandler(_ *service.SettingService, cfg *config.Config) *PublicAssetHandler {
	dataDir := ""
	if cfg != nil {
		dataDir = cfg.Pricing.DataDir
	}
	assetsDir := filepath.Join(dataDir, "pg", "assets")
	_ = os.MkdirAll(assetsDir, 0755)
	cleanupInterval := time.Duration(0)
	if cfg != nil && cfg.Server.PublicAssetCleanupIntervalSeconds > 0 {
		cleanupInterval = time.Duration(cfg.Server.PublicAssetCleanupIntervalSeconds) * time.Second
	}
	return &PublicAssetHandler{assetsDir: assetsDir, cleanupInterval: cleanupInterval, config: cfg}
}

// StartCleanup starts the optional background cleanup loop for expired assets.
// The context owns the goroutine lifetime and should be cancelled during shutdown.
func (h *PublicAssetHandler) StartCleanup(ctx context.Context) {
	if h == nil || ctx == nil || h.cleanupInterval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(h.cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				_ = h.cleanupExpiredAssets(now)
			}
		}
	}()
}

func (h *PublicAssetHandler) cleanupExpiredAssets(now time.Time) error {
	entries, err := os.ReadDir(h.assetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !isSafeAssetComponent(entry.Name()) {
			continue
		}
		assetDir := filepath.Join(h.assetsDir, entry.Name())
		files, err := os.ReadDir(assetDir)
		if err != nil {
			continue
		}
		hasRegularFile := false
		allExpired := true
		for _, file := range files {
			if file.IsDir() {
				allExpired = false
				continue
			}
			info, err := file.Info()
			if err != nil {
				allExpired = false
				continue
			}
			hasRegularFile = true
			if !assetExpired(info.ModTime(), now) {
				allExpired = false
			}
		}
		if !hasRegularFile {
			if info, err := entry.Info(); err != nil || !assetExpired(info.ModTime(), now) {
				allExpired = false
			}
		}
		if allExpired {
			_ = os.RemoveAll(assetDir)
		}
	}
	return nil
}

// Upload accepts a temporary image, video, or audio asset for a video request.
// POST /pg/assets (multipart fields: kind, file)
func (h *PublicAssetHandler) Upload(c *gin.Context) {
	if _, ok := middleware2.GetAPIKeyFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"type": "authentication_error", "message": "Invalid API key"}})
		return
	}
	kind := strings.ToLower(strings.TrimSpace(c.PostForm("kind")))
	spec, ok := publicAssetKinds[kind]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"type": "invalid_request_error", "message": "kind must be image, video, or audio"}})
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"type": "invalid_request_error", "message": "file is required"}})
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	contentType, ok := spec.Extensions[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"type": "invalid_request_error", "message": "file type is not supported for this kind"}})
		return
	}
	if fileHeader.Size > spec.MaxBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"type": "invalid_request_error", "message": fmt.Sprintf("file exceeds %d MiB limit", spec.MaxBytes/(1<<20))}})
		return
	}

	assetID, err := randomAssetID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"type": "server_error", "message": "failed to allocate asset"}})
		return
	}
	assetDir := filepath.Join(h.assetsDir, assetID)
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"type": "server_error", "message": "failed to create asset directory"}})
		return
	}
	filename := safeAssetFilename(fileHeader.Filename, ext)
	path := filepath.Join(assetDir, filename)
	if err := savePublicAsset(fileHeader, path, spec.MaxBytes); err != nil {
		_ = os.RemoveAll(assetDir)
		status := http.StatusInternalServerError
		message := "failed to save asset"
		if err == errPublicAssetTooLarge {
			status = http.StatusRequestEntityTooLarge
			message = fmt.Sprintf("file exceeds %d MiB limit", spec.MaxBytes/(1<<20))
		}
		c.JSON(status, gin.H{"error": gin.H{"type": "invalid_request_error", "message": message}})
		return
	}
	size := fileHeader.Size
	if info, statErr := os.Stat(path); statErr == nil {
		size = info.Size()
	}

	urlBase := h.publicBaseURL(c)
	assetURL := strings.TrimRight(urlBase, "/") + "/pg/assets/" + assetID + "/" + filename
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"kind":         kind,
			"url":          assetURL,
			"filename":     filename,
			"content_type": contentType,
			"size":         size,
		},
	})
}

// Serve returns a temporary asset without authentication so an upstream can fetch it.
// GET /pg/assets/:asset_id/:filename
func (h *PublicAssetHandler) Serve(c *gin.Context) {
	assetID := c.Param("asset_id")
	filename := c.Param("filename")
	if !isSafeAssetComponent(assetID) || !isSafeAssetComponent(filename) {
		c.Status(http.StatusNotFound)
		return
	}
	path := filepath.Join(h.assetsDir, assetID, filename)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || (err == nil && assetExpired(info.ModTime(), time.Now())) {
		if err == nil && assetExpired(info.ModTime(), time.Now()) {
			_ = os.RemoveAll(filepath.Dir(path))
		}
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Content-Disposition", "inline; filename=\""+filename+"\"")
	c.File(path)
}

func (h *PublicAssetHandler) publicBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := strings.TrimSpace(c.Request.Host)
	if h.isTrustedProxy(c) {
		if forwardedProto := strings.ToLower(strings.TrimSpace(strings.Split(c.GetHeader("X-Forwarded-Proto"), ",")[0])); forwardedProto == "https" || forwardedProto == "http" {
			scheme = forwardedProto
		}
		if forwardedHost := validForwardedHost(c.GetHeader("X-Forwarded-Host")); forwardedHost != "" {
			host = forwardedHost
		}
	}
	return scheme + "://" + host
}

func (h *PublicAssetHandler) isTrustedProxy(c *gin.Context) bool {
	if h == nil || h.config == nil || len(h.config.Server.TrustedProxies) == 0 {
		return false
	}
	peer := strings.TrimSpace(c.Request.RemoteAddr)
	if host, _, err := net.SplitHostPort(peer); err == nil {
		peer = host
	}
	peerIP := net.ParseIP(peer)
	if peerIP == nil {
		return false
	}
	for _, raw := range h.config.Server.TrustedProxies {
		raw = strings.TrimSpace(raw)
		if ip := net.ParseIP(raw); ip != nil && ip.Equal(peerIP) {
			return true
		}
		if _, network, err := net.ParseCIDR(raw); err == nil && network.Contains(peerIP) {
			return true
		}
	}
	return false
}

func validForwardedHost(raw string) string {
	host := strings.TrimSpace(raw)
	if host == "" || strings.ContainsAny(host, ",\r\n\t /\\") {
		return ""
	}
	u, err := url.Parse("http://" + host)
	if err != nil || u.Host != host || u.Hostname() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return ""
	}
	return host
}

func assetExpired(modTime, now time.Time) bool {
	return now.Sub(modTime) > publicAssetTTL
}

func randomAssetID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func safeAssetFilename(raw, ext string) string {
	name := filepath.Base(strings.TrimSpace(raw))
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "asset" + ext
	}
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || strings.ContainsRune(`\\/:*?\"<>|`, r) {
			return '-'
		}
		return r
	}, name)
	if filepath.Ext(name) == "" {
		name += ext
	}
	return name
}

func isSafeAssetComponent(value string) bool {
	return value != "" && value != "." && value != ".." && filepath.Base(value) == value && !strings.ContainsAny(value, `/\\`)
}

var errPublicAssetTooLarge = fmt.Errorf("public asset too large")

func savePublicAsset(fileHeader *multipart.FileHeader, path string, maxBytes int64) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	limited := io.LimitReader(file, maxBytes+1)
	n, err := io.Copy(out, limited)
	if err != nil {
		return err
	}
	if n > maxBytes {
		return errPublicAssetTooLarge
	}
	return nil
}
