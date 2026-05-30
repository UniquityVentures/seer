package p_seer_aisstream

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

const AppUrl = "/seer-aisstream/"

func GetPlugin() registry.Pair[string, lamu.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lamu.Plugin{
		Type:        lamu.PluginTypeApp,
		Icon:        "radio",
		URL:         u,
		VerboseName: "AISstream",
		Pages:       lamu.PluginStages(pluginPages),
		Views:       lamu.PluginStages(pluginViews),
		Routes:      lamu.PluginStages(pluginRoutes),
		Configs:     lamu.PluginStages(pluginConfigs),
		DBInitHooks: lamu.PluginStages(pluginDBInitHooks),
	}
	p_seer_workerregistry.RegisterPluginInCategory("Live Maps", p)
	return registry.Pair[string, lamu.Plugin]{Key: "p_seer_aisstream", Value: p}
}
