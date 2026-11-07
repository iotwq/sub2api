package admin

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSettingsModelPlazaRechargeRatio(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{})
	require.Equal(t, float64(1), h.settingService.GetModelPlazaRuntime(context.Background()).RechargeRatio)
	for _, value := range []float64{6, 2.5} {
		rec := doUpdateSettings(t, h, map[string]any{"model_plaza_recharge_ratio": value}, nil)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Equal(t, value, gjson.Get(rec.Body.String(), "data.model_plaza_recharge_ratio").Float())
		require.Equal(t, value, h.settingService.GetModelPlazaRuntime(context.Background()).RechargeRatio)
		rec = httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
		h.GetSettings(c)
		require.Equal(t, value, gjson.Get(rec.Body.String(), "data.model_plaza_recharge_ratio").Float())
	}
	rec := doUpdateSettings(t, h, map[string]any{"site_name": "unchanged ratio"}, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "2.5", repo.values[service.SettingKeyModelPlazaRechargeRatio])
	for _, value := range []any{0, -1, "invalid"} {
		rec = doUpdateSettings(t, h, map[string]any{"model_plaza_recharge_ratio": value}, nil)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		require.Equal(t, "2.5", repo.values[service.SettingKeyModelPlazaRechargeRatio])
	}
}

func TestSettingsModelPlazaRechargeRatioInvalidStoredValue(t *testing.T) {
	for _, raw := range []string{"", "0", "-1", "NaN", "+Inf", "invalid"} {
		h, _ := newStepUpSwitchTestHandler(t, map[string]string{service.SettingKeyModelPlazaRechargeRatio: raw})
		settings, err := h.settingService.GetAllSettings(context.Background())
		require.NoError(t, err)
		require.Equal(t, float64(1), settings.ModelPlazaRechargeRatio)
		require.Equal(t, float64(1), h.settingService.GetModelPlazaRuntime(context.Background()).RechargeRatio)
	}
}
