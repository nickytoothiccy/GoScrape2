package hermesenv

import (
	"os"
	"path/filepath"
	"strings"

	"stealthfetch/internal/envutil"
	"stealthfetch/internal/models"
)

// ResolveHome returns the active Hermes home directory.
func ResolveHome() string {
	if home := strings.TrimSpace(os.Getenv("HERMES_HOME")); home != "" {
		return home
	}
	baseHome, err := os.UserHomeDir()
	if err != nil {
		return ".hermes"
	}
	return filepath.Join(baseHome, ".hermes")
}

// LoadEnv hydrates process env from HERMES_HOME/.env without overriding existing env vars.
func LoadEnv() error {
	return envutil.LoadDotEnv(filepath.Join(ResolveHome(), ".env"))
}

// DefaultConfig returns a scraper config seeded from Hermes-owned env/config.
func DefaultConfig() *models.Config {
	cfg := models.DefaultConfig()

	if model := strings.TrimSpace(firstSet("HERMES_GOSCRAPE_MODEL", "OPENAI_MODEL")); model != "" {
		cfg.LLMModel = model
	} else if model := readHermesModelDefault(); model != "" {
		cfg.LLMModel = model
	}

	cfg.LLMAPIKey = strings.TrimSpace(firstSet("HERMES_GOSCRAPE_API_KEY", "OPENAI_API_KEY"))
	cfg.LLMBaseURL = strings.TrimSpace(firstSet("HERMES_GOSCRAPE_BASE_URL", "OPENAI_BASE_URL"))

	return cfg
}

func firstSet(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func readHermesModelDefault() string {
	data, err := os.ReadFile(filepath.Join(ResolveHome(), "config.yaml"))
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	inModel := false
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inModel = trimmed == "model:"
			continue
		}
		if !inModel {
			continue
		}
		if strings.HasPrefix(trimmed, "default:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "default:"))
			return strings.Trim(value, `"'`)
		}
	}
	return ""
}
