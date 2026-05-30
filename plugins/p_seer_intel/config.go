package p_seer_intel

import (
	"strings"

)

// IntelConfig holds Intel-specific settings loaded from [Plugins.p_seer_intel].
type IntelConfig struct {
	GeocodingAPIKey string `toml:"geocodingApiKey"`
	TitleModel      string `toml:"titleModel"`
	SummaryModel    string `toml:"summaryModel"`
	EmbeddingModel  string `toml:"embeddingModel"`
}

var IntelConfigValue = &IntelConfig{}

func (c *IntelConfig) PostConfig() {
	if c == nil {
		return
	}
	c.GeocodingAPIKey = strings.TrimSpace(c.GeocodingAPIKey)
}

func init() {
	registerPluginConfig("p_seer_intel", IntelConfigValue)
}
