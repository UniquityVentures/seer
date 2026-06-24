package p_seer_aisstream

import (
	"github.com/UniquityVentures/lamu/lamu"
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
}

func init() {
	registerRoutes()
}
