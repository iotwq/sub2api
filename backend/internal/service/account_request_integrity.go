package service

import (
	"github.com/gin-gonic/gin"
	"log/slog"
)

// Integrity is opt-in and only protects the Codex OAuth route.
func (a *Account) RequestIntegrityMode() string {
	plan, err := AccountTrafficPlanFor(a)
	if err == nil && plan.Policy.Active {
		return plan.Policy.IntegrityMode
	}
	return "off"
}

func checkAccountRequestIntegrity(c *gin.Context, a *Account, original, forwarded []byte) error {
	mode := a.RequestIntegrityMode()
	if mode == "off" {
		return nil
	}
	err := validateMode1RequestIntegrityForAccount(a, original, forwarded)
	if err == nil {
		return nil
	}
	if c != nil {
		c.Set("request_integrity_difference", true)
	}
	// Error text is produced by the validator and contains field names only;
	// never record request bodies, credentials or conversation content here.
	slog.Warn("account_request_integrity_difference", "account_id", a.ID, "mode", mode, "error", err.Error())
	if mode == "enforce" {
		return err
	}
	return nil
}
