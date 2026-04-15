package loaders

import (
	"strings"
	"time"

	"stealthfetch/internal/models"
)

// NewFetchLoader chooses the best fetch backend for a source and config.
func NewFetchLoader(source string, config *models.Config) Loader {
	if strings.HasPrefix(strings.TrimSpace(source), "<") {
		return NewLocalLoader()
	}
	if config == nil {
		config = models.DefaultConfig()
	}
	switch resolveFetchStrategy(config) {
	case "rod":
		return NewDefaultRodLoader(config.Verbose)
	case "auto":
		return NewEscalatingLoader(
			NewUTLSLoader("chrome", "", 30*time.Second),
			NewDefaultRodLoader(config.Verbose),
		)
	default:
		return NewUTLSLoader("chrome", "", 30*time.Second)
	}
}

func resolveFetchStrategy(config *models.Config) string {
	strategy := strings.ToLower(strings.TrimSpace(config.FetchStrategy))
	if strategy == "utls" || strategy == "rod" || strategy == "auto" {
		return strategy
	}
	if config.Headless {
		return "rod"
	}
	return "utls"
}
