package p_seer_node_fleet

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

// AppUrl is the HTTP prefix for this plugin (trailing slash).
const AppUrl = "/seer-node-fleet/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "server-stack",
		URL:         u,
		VerboseName: "Node fleet",
	}
	if err := lago.RegistryPlugin.Register("p_seer_node_fleet", p); err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Infrastructure", p)
}
