package p_seer_intel

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
)

const AppUrl = "/seer-intel/"

func init() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}

	p := lago.Plugin{
		Type:        lago.PluginTypeApp,
		Icon:        "eye",
		URL:         u,
		VerboseName: "Intel",
	}
	err = lago.RegistryPlugin.Register("p_seer_intel", p)
	if err != nil {
		log.Panic(err)
	}
	p_seer_workerregistry.RegisterPluginInCategory("Intel", p)
}
