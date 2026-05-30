package p_seer_dashboard

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
)

func GetPlugin() registry.Pair[string, lamu.Plugin] {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	return registry.Pair[string, lamu.Plugin]{
		Key: "p_seer_dashboard",
		Value: lamu.Plugin{
			Type:        lamu.PluginTypeAddon,
			Icon:        "map",
			URL:         u,
			VerboseName: "Seer dashboard map",
			Pages:       lamu.PluginStages(pluginPages),
			Views:       lamu.PluginStages(pluginViews),
			Routes:      lamu.PluginStages(pluginRoutes),
		},
	}
}
