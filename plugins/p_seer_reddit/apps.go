package p_seer_reddit

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

const AppUrl = "/seer-reddit/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "chat-bubble-left-right",
		URL:         u,
		VerboseName: "Reddit",
	}
	err = lago.RegistryPlugin.Register("p_seer_reddit", p)
	if err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Sources", p)
}
