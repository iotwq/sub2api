package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/tidwall/gjson"
)

const (
	miniMaxH3VideoResolution768P = "768P"
	miniMaxH3VideoResolution2K   = "2K"

	miniMaxH3MaxReferenceImages = 9
	miniMaxH3MaxReferenceVideos = 3
	miniMaxH3MaxReferenceAudios = 3
	miniMaxH3MaxVideoAssetBytes = 50 << 20
	miniMaxH3MinMediaSeconds    = 2.0
	miniMaxH3MaxMediaSeconds    = 15.0
	miniMaxH3DurationTolerance  = 0.15
)

type miniMaxH3VideoBillingMedia struct {
	InputVideoSeconds float64
}

type OpenAIVideoBillingInputError struct {
	Message string
}

func (e *OpenAIVideoBillingInputError) Error() string {
	if e == nil {
		return "invalid video billing input"
	}
	return e.Message
}

func newOpenAIVideoBillingInputError(message string) error {
	return &OpenAIVideoBillingInputError{Message: strings.TrimSpace(message)}
}

func normalizeMiniMaxH3VideoResolution(value string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case miniMaxH3VideoResolution768P:
		return miniMaxH3VideoResolution768P, true
	case miniMaxH3VideoResolution2K:
		return miniMaxH3VideoResolution2K, true
	default:
		return "", false
	}
}

func (s *OpenAIGatewayService) analyzeMiniMaxH3VideoBillingMedia(body []byte) (miniMaxH3VideoBillingMedia, error) {
	var result miniMaxH3VideoBillingMedia
	content := gjson.GetBytes(body, "content")
	if !content.Exists() {
		return result, nil
	}
	if !content.IsArray() {
		return result, fmt.Errorf("MiniMax-H3 content must be an array")
	}

	imageCount := 0
	videoURLs := make([]string, 0, miniMaxH3MaxReferenceVideos)
	audioCount := 0
	for _, item := range content.Array() {
		switch strings.ToLower(strings.TrimSpace(item.Get("type").String())) {
		case "image_url":
			imageCount++
		case "video_url":
			videoURL := strings.TrimSpace(item.Get("video_url.url").String())
			if videoURL == "" {
				return result, fmt.Errorf("MiniMax-H3 reference video URL is required")
			}
			videoURLs = append(videoURLs, videoURL)
		case "audio_url":
			audioCount++
		}
	}

	if imageCount > miniMaxH3MaxReferenceImages {
		return result, fmt.Errorf("MiniMax-H3 supports at most %d reference images", miniMaxH3MaxReferenceImages)
	}
	if len(videoURLs) > miniMaxH3MaxReferenceVideos {
		return result, fmt.Errorf("MiniMax-H3 supports at most %d reference videos", miniMaxH3MaxReferenceVideos)
	}
	if audioCount > miniMaxH3MaxReferenceAudios {
		return result, fmt.Errorf("MiniMax-H3 supports at most %d reference audios", miniMaxH3MaxReferenceAudios)
	}

	for _, rawURL := range videoURLs {
		path, err := s.miniMaxH3LocalVideoAssetPath(rawURL)
		if err != nil {
			return result, err
		}
		duration, err := readMiniMaxH3VideoDurationSeconds(path)
		if err != nil {
			return result, fmt.Errorf("cannot read MiniMax-H3 reference video duration: %w", err)
		}
		if duration < miniMaxH3MinMediaSeconds {
			return result, fmt.Errorf("MiniMax-H3 reference video must be at least %.0f seconds (got %.2f)", miniMaxH3MinMediaSeconds, duration)
		}
		if duration > miniMaxH3MaxMediaSeconds+miniMaxH3DurationTolerance {
			return result, fmt.Errorf("MiniMax-H3 reference video must not exceed %.0f seconds (got %.2f)", miniMaxH3MaxMediaSeconds, duration)
		}
		result.InputVideoSeconds += duration
	}
	if result.InputVideoSeconds > miniMaxH3MaxMediaSeconds+miniMaxH3DurationTolerance {
		return result, fmt.Errorf("MiniMax-H3 reference videos must not exceed %.0f seconds in total (got %.2f)", miniMaxH3MaxMediaSeconds, result.InputVideoSeconds)
	}
	return result, nil
}

func (s *OpenAIGatewayService) miniMaxH3LocalVideoAssetPath(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("MiniMax-H3 paid reference videos must use an absolute /pg/assets URL from this service")
	}
	if s == nil || s.cfg == nil || strings.TrimSpace(s.cfg.Server.FrontendURL) == "" {
		return "", fmt.Errorf("server.frontend_url must be configured before billing MiniMax-H3 reference videos")
	}
	frontendURL, err := url.Parse(strings.TrimSpace(s.cfg.Server.FrontendURL))
	if err != nil || frontendURL.Host == "" || !strings.EqualFold(parsed.Host, frontendURL.Host) {
		return "", fmt.Errorf("MiniMax-H3 paid reference videos must be uploaded through this service's /pg/assets endpoint")
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) == 5 && parts[0] == "v1" {
		parts = parts[1:]
	}
	if len(parts) != 4 || parts[0] != "pg" || parts[1] != "assets" {
		return "", fmt.Errorf("MiniMax-H3 paid reference videos must be uploaded through this service's /pg/assets endpoint")
	}
	assetID, filename := parts[2], parts[3]
	if !isMiniMaxH3AssetComponent(assetID) || !isMiniMaxH3AssetComponent(filename) {
		return "", fmt.Errorf("invalid MiniMax-H3 reference video asset URL")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".mp4" && ext != ".mov" {
		return "", fmt.Errorf("MiniMax-H3 reference videos must be MP4 or MOV files")
	}

	dataDir := strings.TrimSpace(s.cfg.Pricing.DataDir)
	if dataDir == "" {
		dataDir = "./data"
	}
	path := filepath.Join(dataDir, "pg", "assets", assetID, filename)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("MiniMax-H3 reference video asset was not found; upload it again through /pg/assets")
	}
	if info.Size() <= 0 || info.Size() > miniMaxH3MaxVideoAssetBytes {
		return "", fmt.Errorf("MiniMax-H3 reference video must be non-empty and no larger than 50 MiB")
	}
	return path, nil
}

func isMiniMaxH3AssetComponent(value string) bool {
	return value != "" && value != "." && value != ".." && filepath.Base(value) == value && !strings.ContainsAny(value, `/\\`)
}

func readMiniMaxH3VideoDurationSeconds(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	decoded, err := mp4.DecodeFile(file, mp4.WithDecodeMode(mp4.DecModeLazyMdat))
	if err != nil {
		return 0, err
	}
	moov := decoded.Moov
	if moov == nil && decoded.Init != nil {
		moov = decoded.Init.Moov
	}
	if moov == nil || moov.Mvhd == nil || moov.Mvhd.Timescale == 0 || moov.Mvhd.Duration == 0 {
		return 0, fmt.Errorf("video does not contain a usable movie duration")
	}
	return float64(moov.Mvhd.Duration) / float64(moov.Mvhd.Timescale), nil
}
