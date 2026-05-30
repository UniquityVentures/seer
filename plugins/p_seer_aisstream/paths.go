package p_seer_aisstream

import (
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
)

func registerRoutes() {
	registerPluginRoute("seer_aisstream.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_aisstream.MessageListView"),
	})
	registerPluginRoute("seer_aisstream.MessageListRoute", lamu.Route{
		Path:    AppUrl + "messages/",
		Handler: lamu.NewDynamicView("seer_aisstream.MessageListView"),
	})
	registerPluginRoute("seer_aisstream.MessageDetailRoute", lamu.Route{
		Path:    AppUrl + "messages/{id}/",
		Handler: lamu.NewDynamicView("seer_aisstream.MessageDetailView"),
	})
	registerPluginRoute("seer_aisstream.MapRouteUnderMessages", lamu.Route{
		Path:    AppUrl + "messages/map/",
		Handler: lamu.NewDynamicView("seer_aisstream.MapView"),
	})
	registerPluginRoute("seer_aisstream.MapRoute", lamu.Route{
		Path:    AppUrl + "map/",
		Handler: lamu.NewDynamicView("seer_aisstream.MapView"),
	})
	registerPluginRoute("seer_aisstream.MapDataRoute", lamu.Route{
		Path:    AppUrl + "map/data/",
		Handler: p_users.RequireAuth(aisStreamMapDataHandler{}),
	})
}

func init() {
	registerRoutes()
}
