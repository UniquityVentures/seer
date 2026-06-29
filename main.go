package main

import (
	"log/slog"

	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_dashboard"
	"github.com/UniquityVentures/lamu/plugins/p_filesystem"
	"github.com/UniquityVentures/lamu/plugins/p_google_genai"
	"github.com/UniquityVentures/lamu/plugins/p_llm_assistant"
	"github.com/UniquityVentures/lamu/plugins/p_pwa"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_aisstream"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_assistant"
	"github.com/UniquityVentures/seer/plugins/p_seer_dashboard"
	"github.com/UniquityVentures/seer/plugins/p_seer_deepsearch"
	"github.com/UniquityVentures/seer/plugins/p_seer_gdelt"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet"
	"github.com/UniquityVentures/seer/plugins/p_seer_opensky"
	"github.com/UniquityVentures/seer/plugins/p_seer_reddit"
	"github.com/UniquityVentures/seer/plugins/p_seer_websites"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

func main() {
	plugins := []registry.Pair[string, lamu.Plugin]{
		p_dashboard.GetPlugin(),
		p_filesystem.GetPlugin(),
		p_google_genai.GetPlugin(),
		p_users.GetPlugin(),
		p_pwa.GetPlugin(),
		p_seer_workerregistry.GetPlugin(),
		p_seer_intel.GetPlugin(),
		p_seer_dashboard.GetPlugin(),
		p_seer_reddit.GetPlugin(),
		p_seer_websites.GetPlugin(),
		p_seer_gdelt.GetPlugin(),
		p_seer_opensky.GetPlugin(),
		p_seer_aisstream.GetPlugin(),
		p_seer_deepsearch.GetPlugin(),
		p_llm_assistant.GetPlugin(),
		p_seer_node_fleet.GetPlugin(),
	}
	p_seer_workerregistry.BuildAllRegistries()

	config, err := lamu.LoadConfigFromFile("seer.toml", plugins)
	if err != nil {
		panic(err)
	}
	if err := lamu.Start(config, plugins); err != nil {
		slog.Error(err.Error())
	}
}
