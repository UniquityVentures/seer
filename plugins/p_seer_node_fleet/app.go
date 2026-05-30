package p_seer_node_fleet

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

const AppUrl = "/seer-node-fleet/"

func GetPlugin() registry.Pair[string, lamu.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lamu.Plugin{
		Type:        lamu.PluginTypeApp,
		Icon:        "server-stack",
		URL:         u,
		VerboseName: "Node fleet",
		Pages:       lamu.PluginStages(pluginPages),
		Views:       lamu.PluginStages(pluginViews),
		Routes:      lamu.PluginStages(pluginRoutes),
	}
	p_seer_workerregistry.RegisterPluginInCategory("Infrastructure", p)
	return registry.Pair[string, lamu.Plugin]{Key: "p_seer_node_fleet", Value: p}
}
