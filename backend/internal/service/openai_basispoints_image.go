package service

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Wei-Shaw/sub2api/internal/service/basispoints"
	"github.com/gin-gonic/gin"
)

func (s *OpenAIGatewayService) excelBPSImageRelay(ctx context.Context) (*basispoints.ImageRelay, error) {
	settings, err := s.settingService.GetExcelBPSImageRelaySettings(ctx)
	if err != nil || !settings.Enabled {
		return nil, err
	}
	return s.excelBPSImageRelayForSettings(settings)
}

func (s *OpenAIGatewayService) excelBPSImageRelayForSettings(settings ExcelBPSImageRelaySettings) (*basispoints.ImageRelay, error) {
	if !settings.Enabled || settings.Mode == ExcelBPSImageModeNative {
		return nil, nil
	}
	var err error
	s.excelBPSImagesMu.Lock()
	defer s.excelBPSImagesMu.Unlock()
	if s.excelBPSImagesClosed {
		return nil, basispoints.ErrImageRelayStorage
	}
	if s.excelBPSImages == nil {
		dataDir := os.Getenv("DATA_DIR")
		if dataDir == "" {
			dataDir = "./data"
		}
		s.excelBPSImages, err = basispoints.NewImageRelay(settings.BaseURL, filepath.Join(dataDir, "bps-images"))
	}
	if err == nil {
		err = s.excelBPSImages.Configure(settings.BaseURL, settings.Limits)
		s.excelBPSImages.SetAdmissionLimits(settings.BodyLimitMiB, settings.BudgetMiB, settings.MaxRequests)
	}
	return s.excelBPSImages, err
}

func (s *OpenAIGatewayService) CloseExcelBPSImages() error {
	s.excelBPSImageAdmission.CloseAdmission()
	s.excelBPSImagesMu.Lock()
	s.excelBPSImagesClosed = true
	relay := s.excelBPSImages
	s.excelBPSImagesMu.Unlock()
	return relay.Close()
}

func (s *OpenAIGatewayService) ServeExcelBPSImage(c *gin.Context) {
	settings, err := s.settingService.GetExcelBPSImageRelaySettings(c.Request.Context())
	if err != nil {
		c.Header("Cache-Control", "private, no-store")
		c.Status(http.StatusServiceUnavailable)
		return
	}
	var relay *basispoints.ImageRelay
	if settings.Enabled && settings.Mode != ExcelBPSImageModeNative {
		s.excelBPSImagesMu.Lock()
		relay = s.excelBPSImages
		s.excelBPSImagesMu.Unlock()
	}
	// Anonymous reads never create storage. Only authenticated Basispoints
	// requests can populate it; a missing/disabled relay safely returns 404.
	relay.ServeHTTP(c.Writer, c.Request)
}
