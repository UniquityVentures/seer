package p_seer_opensky

import (
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
)

func registerRoutes() {
	registerPluginRoute("seer_opensky.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_opensky.StateListView"),
	})
	registerPluginRoute("seer_opensky.StateListRoute", lamu.Route{
		Path:    AppUrl + "states/",
		Handler: lamu.NewDynamicView("seer_opensky.StateListView"),
	})
	registerPluginRoute("seer_opensky.MapRouteUnderStates", lamu.Route{
		Path:    AppUrl + "states/map/",
		Handler: lamu.NewDynamicView("seer_opensky.MapView"),
	})
	registerPluginRoute("seer_opensky.StateCreateRoute", lamu.Route{
		Path:    AppUrl + "states/create/",
		Handler: lamu.NewDynamicView("seer_opensky.StateCreateView"),
	})
	registerPluginRoute("seer_opensky.StateDetailRoute", lamu.Route{
		Path:    AppUrl + "states/{id}/",
		Handler: lamu.NewDynamicView("seer_opensky.StateDetailView"),
	})
	registerPluginRoute("seer_opensky.StateUpdateRoute", lamu.Route{
		Path:    AppUrl + "states/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_opensky.StateUpdateView"),
	})
	registerPluginRoute("seer_opensky.StateDeleteRoute", lamu.Route{
		Path:    AppUrl + "states/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_opensky.StateDeleteView"),
	})
	registerPluginRoute("seer_opensky.MapRoute", lamu.Route{
		Path:    AppUrl + "map/",
		Handler: lamu.NewDynamicView("seer_opensky.MapView"),
	})
	registerPluginRoute("seer_opensky.MapDataRoute", lamu.Route{
		Path:    AppUrl + "map/data/",
		Handler: p_users.RequireAuth(openSkyMapDataHandler{}),
	})
}

func init() {
	registerRoutes()
}
