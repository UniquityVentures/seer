package p_seer_opensky

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

// AppUrl is the base path for the OpenSky plugin UI.
const AppUrl = "/seer-opensky/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "paper-airplane",
		URL:         u,
		VerboseName: "OpenSky",
	}
	err = lago.RegistryPlugin.Register("p_seer_opensky", p)
	if err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Live Maps", p)
}
