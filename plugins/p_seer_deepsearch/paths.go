package p_seer_deepsearch

import "github.com/UniquityVentures/lamu/lamu"

func registerRoutes() {
	registerPluginRoute("seer_deepsearch.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_deepsearch.HomeView"),
	})

	registerPluginRoute("seer_deepsearch.StartRoute", lamu.Route{
		Path:    AppUrl + "start/",
		Handler: lamu.NewDynamicView("seer_deepsearch.StartView"),
	})

	// Literal "history/" before "{id}/" so /seer-deepsearch/history/ is not captured as an id segment.
	registerPluginRoute("seer_deepsearch.HistoryRoute", lamu.Route{
		Path:    AppUrl + "history/",
		Handler: lamu.NewDynamicView("seer_deepsearch.HistoryView"),
	})

	// Literal "{id}/stop/" and "{id}/restart/" before "{id}/" so they are not swallowed as id text.
	registerPluginRoute("seer_deepsearch.StopRoute", lamu.Route{
		Path:    AppUrl + "{id}/stop/",
		Handler: lamu.NewDynamicView("seer_deepsearch.StopView"),
	})
	registerPluginRoute("seer_deepsearch.RestartRoute", lamu.Route{
		Path:    AppUrl + "{id}/restart/",
		Handler: lamu.NewDynamicView("seer_deepsearch.RestartView"),
	})

	registerPluginRoute("seer_deepsearch.DetailRoute", lamu.Route{
		Path:    AppUrl + "{id}/",
		Handler: lamu.NewDynamicView("seer_deepsearch.DetailView"),
	})
}

func init() {
	registerRoutes()
}
