package p_seer_gdelt

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

const AppUrl = "/seer-gdelt/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "magnifying-glass-circle",
		URL:         u,
		VerboseName: "GDELT search",
	}
	err = lago.RegistryPlugin.Register("p_seer_gdelt", p)
	if err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Sources", p)
}
