package p_seer_gdelt

import "github.com/UniquityVentures/lamu/lamu"

func registerRoutes() {
	registerPluginRoute("seer_gdelt.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTSourceListView"),
	})

	registerPluginRoute("seer_gdelt.SearchRoute", lamu.Route{
		Path:    AppUrl + "search/",
		Handler: lamu.NewDynamicView("seer_gdelt.SearchView"),
	})

	registerPluginRoute("seer_gdelt.GDELTSourceCreateRoute", lamu.Route{
		Path:    AppUrl + "sources/create/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTSourceCreateView"),
	})

	registerPluginRoute("seer_gdelt.GDELTSourceUnsetSelectRoute", lamu.Route{
		Path:    AppUrl + "sources/unset/select/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTSourceUnsetSelectView"),
	})

	registerPluginRoute("seer_gdelt.GDELTSourceDetailRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTSourceDetailView"),
	})

	registerPluginRoute("seer_gdelt.GDELTSourceUpdateRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTSourceUpdateView"),
	})

	registerPluginRoute("seer_gdelt.GDELTSourceDeleteRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTSourceDeleteView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerListRoute", lamu.Route{
		Path:    AppUrl + "workers/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerListView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerCreateRoute", lamu.Route{
		Path:    AppUrl + "workers/create/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerCreateView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerSelectRoute", lamu.Route{
		Path:    AppUrl + "workers/select/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerSelectView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerPoolStartRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/start/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerPoolStartView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerPoolStopRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/stop/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerPoolStopView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerDetailRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerDetailView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerUpdateRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerUpdateView"),
	})

	registerPluginRoute("seer_gdelt.GDELTWorkerDeleteRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_gdelt.GDELTWorkerDeleteView"),
	})

	registerPluginRoute("seer_gdelt.EventListRoute", lamu.Route{
		Path:    AppUrl + "events/",
		Handler: lamu.NewDynamicView("seer_gdelt.EventListView"),
	})

	registerPluginRoute("seer_gdelt.EventCreateRoute", lamu.Route{
		Path:    AppUrl + "events/create/",
		Handler: lamu.NewDynamicView("seer_gdelt.EventCreateView"),
	})

	registerPluginRoute("seer_gdelt.EventDetailRoute", lamu.Route{
		Path:    AppUrl + "events/{id}/",
		Handler: lamu.NewDynamicView("seer_gdelt.EventDetailView"),
	})

	registerPluginRoute("seer_gdelt.EventUpdateRoute", lamu.Route{
		Path:    AppUrl + "events/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_gdelt.EventUpdateView"),
	})

	registerPluginRoute("seer_gdelt.EventDeleteRoute", lamu.Route{
		Path:    AppUrl + "events/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_gdelt.EventDeleteView"),
	})

	registerPluginRoute("seer_gdelt.MapRoute", lamu.Route{
		Path:    AppUrl + "map/",
		Handler: lamu.NewDynamicView("seer_gdelt.MapView"),
	})
}

func init() {
	registerRoutes()
}
