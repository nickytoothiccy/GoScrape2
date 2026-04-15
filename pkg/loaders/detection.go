package loaders

import (
	"strings"

	"stealthfetch/internal/models"
)

var blockMarkers = []string{
	"cf-chl",
	"captcha",
	"access denied",
	"verify you are human",
	"enable javascript and cookies",
	"just a moment",
	"attention required",
	"request blocked",
}

// IsLikelyBlocked returns true when a fetch result looks like an anti-bot block.
func IsLikelyBlocked(result *models.FetchResult) bool {
	if result == nil {
		return true
	}
	if result.Error != nil {
		return true
	}
	if result.StatusCode == 403 || result.StatusCode == 429 || result.StatusCode == 503 {
		return true
	}
	body := strings.ToLower(result.HTML)
	for _, marker := range blockMarkers {
		if strings.Contains(body, marker) {
			return true
		}
	}
	return false
}
