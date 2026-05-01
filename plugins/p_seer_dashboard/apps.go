package p_seer_dashboard

import (
	"log"
	"net/url"

	"github.com/UniquityVentures/lago/lago"
)

func registerDashboardPlugin() {
	u, err := url.Parse(AppUrl)
	if err != nil {
		log.Panic(err)
	}
	if err := lago.RegistryPlugin.Register("p_seer_dashboard", lago.Plugin{
		Type:        lago.PluginTypeAddon,
		Icon:        "map",
		URL:         u,
		VerboseName: "Seer dashboard map",
	}); err != nil {
		log.Panic(err)
	}
}
