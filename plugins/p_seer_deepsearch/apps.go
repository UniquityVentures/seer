package p_seer_deepsearch

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

// AppUrl is the HTTP prefix for this plugin (trailing slash).
const AppUrl = "/seer-deepsearch/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "magnifying-glass-circle",
		URL:         u,
		VerboseName: "Deep search",
	}
	err = lago.RegistryPlugin.Register("p_seer_deepsearch", p)
	if err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Search", p)
}
