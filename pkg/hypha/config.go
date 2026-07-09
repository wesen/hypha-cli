package hypha

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the on-disk config file (~/.config/hypha/config.json, mode 600).
// Override the path with HYPHA_CONFIG.
type Config struct {
	BaseURL string `json:"base_url"`
	PAT     string `json:"pat"`
	Handle  string `json:"handle"` // the caller's handle, for whoami
}

// ConfigPath returns the config file path: $HYPHA_CONFIG or
// ~/.config/hypha/config.json.
func ConfigPath() (string, error) {
	if p := os.Getenv("HYPHA_CONFIG"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("hypha: find home dir: %w", err)
	}
	return filepath.Join(home, ".config", "hypha", "config.json"), nil
}

// LoadConfig reads the config file. A missing file is not an error: it
// returns a zero Config so callers can fall back to flags/env. A malformed
// file is an error.
func LoadConfig() (Config, error) {
	p, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("hypha: read config %s: %w", p, err)
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, fmt.Errorf("hypha: parse config %s: %w", p, err)
	}
	// Env overrides config (env is more explicit for CI/scripts).
	if b := os.Getenv("HYPHA_BASE_URL"); b != "" {
		c.BaseURL = b
	}
	if p := os.Getenv("HYPHA_PAT"); p != "" {
		c.PAT = p
	}
	if h := os.Getenv("HYPHA_HANDLE"); h != "" {
		c.Handle = h
	}
	return c, nil
}
