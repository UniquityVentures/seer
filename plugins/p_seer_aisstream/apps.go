package p_seer_aisstream

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

const AppUrl = "/seer-aisstream/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "radio",
		URL:         u,
		VerboseName: "AISstream",
	}
	err = lago.RegistryPlugin.Register("p_seer_aisstream", p)
	if err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Live Maps", p)
}
