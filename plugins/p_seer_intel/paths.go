package p_seer_intel

import (
	"github.com/UniquityVentures/lamu/lamu"
)

func registerRoutes() {
	registerPluginRoute("seer_intel.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_intel.ListView"),
	})

	registerPluginRoute("seer_intel.DetailRoute", lamu.Route{
		Path:    AppUrl + "{id}/",
		Handler: lamu.NewDynamicView("seer_intel.DetailView"),
	})
}

func init() {
	registerRoutes()
}
