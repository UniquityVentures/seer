package p_seer_assistant

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

// AppUrl is the HTTP prefix for this plugin (trailing slash).
const AppUrl = "/seer-assistant/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "sparkles",
		URL:         u,
		VerboseName: "Assistant",
	}
	err = lago.RegistryPlugin.Register("p_seer_assistant", p)
	if err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Search", p)
}
