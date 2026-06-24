package p_seer_workerregistry

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
)

const AppUrl = "/seer-workerregistry/"

func GetPlugin() registry.Pair[string, lamu.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lamu.Plugin{
		Type:        lamu.PluginTypeApp,
		Icon:        "cpu-chip",
		URL:         u,
		VerboseName: "Worker Registry",
		Pages:       lamu.PluginStages(pluginPages),
		Views:       lamu.PluginStages(pluginViews),
		Routes:      lamu.PluginStages(pluginRoutes),
		Configs:     lamu.PluginStages(pluginConfigs),
		DBInitHooks: lamu.PluginStages(pluginDBInitHooks),
	}
	RegisterPluginInCategory("Management", p)
	return registry.Pair[string, lamu.Plugin]{Key: "p_seer_workerregistry", Value: p}
}

