package p_seer_gdelt

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

const AppUrl = "/seer-gdelt/"

func GetPlugin() registry.Pair[string, lamu.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lamu.Plugin{
		Type:        lamu.PluginTypeApp,
		Icon:        "magnifying-glass-circle",
		URL:         u,
		VerboseName: "GDELT search",
		Pages:       lamu.PluginStages(pluginPages),
		Views:       lamu.PluginStages(pluginViews),
		Routes:      lamu.PluginStages(pluginRoutes),
		Configs:     lamu.PluginStages(pluginConfigs),
		DBInitHooks: lamu.PluginStages(pluginDBInitHooks),
	}
	p_seer_workerregistry.RegisterPluginInCategory("Sources", p)
	return registry.Pair[string, lamu.Plugin]{Key: "p_seer_gdelt", Value: p}
}
