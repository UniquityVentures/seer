package p_seer_websites

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/UniquityVentures/lago/lago"
)

// WebsiteRodConfig is retained for future node-side browser configuration.
// Website scraping is dispatched to the Seer node fleet ([p_seer_node_fleet.DispatchCommand]).
type WebsiteRodConfig struct {
	UserDataDir string `toml:"userDataDir"`
	ProfileDir  string `toml:"profileDir"`
}

// WebsiteRod is the package-level Rod/Chromium launch config (filled from registry after config load).
var WebsiteRod = &WebsiteRodConfig{}

func defaultChromiumUserDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Chromium")
	case "windows":
		return filepath.Join(home, "AppData", "Local", "Chromium", "User Data")
	default:
		return filepath.Join(home, ".config", "chromium")
	}
}

func (c *WebsiteRodConfig) PostConfig() {
	if c == nil {
		return
	}
	if strings.TrimSpace(c.UserDataDir) == "" {
		c.UserDataDir = defaultChromiumUserDataDir()
	}
}

func init() {
	if d := defaultChromiumUserDataDir(); d != "" {
		WebsiteRod.UserDataDir = d
	}
	lago.RegistryConfig.Register("p_seer_websites", WebsiteRod)
}
