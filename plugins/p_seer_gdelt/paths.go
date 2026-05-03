package p_seer_gdelt

import "github.com/UniquityVentures/lago/lago"

func registerRoutes() {
	_ = lago.RegistryRoute.Register("seer_gdelt.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_gdelt.GDELTSourceListView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.SearchRoute", lago.Route{
		Path:    AppUrl + "search/",
		Handler: lago.NewDynamicView("seer_gdelt.SearchView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTSourceCreateRoute", lago.Route{
		Path:    AppUrl + "sources/create/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTSourceCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTSourceUnsetSelectRoute", lago.Route{
		Path:    AppUrl + "sources/unset/select/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTSourceUnsetSelectView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTSourceDetailRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTSourceDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTSourceUpdateRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/edit/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTSourceUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTSourceDeleteRoute", lago.Route{
		Path:    AppUrl + "sources/{id}/delete/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTSourceDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerListRoute", lago.Route{
		Path:    AppUrl + "workers/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerListView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerCreateRoute", lago.Route{
		Path:    AppUrl + "workers/create/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerSelectRoute", lago.Route{
		Path:    AppUrl + "workers/select/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerSelectView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerPoolStartRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/start/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerPoolStartView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerPoolStopRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/stop/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerPoolStopView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerDetailRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerUpdateRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/edit/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.GDELTWorkerDeleteRoute", lago.Route{
		Path:    AppUrl + "workers/{id}/delete/",
		Handler: lago.NewDynamicView("seer_gdelt.GDELTWorkerDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventListRoute", lago.Route{
		Path:    AppUrl + "events/",
		Handler: lago.NewDynamicView("seer_gdelt.EventListView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventCreateRoute", lago.Route{
		Path:    AppUrl + "events/create/",
		Handler: lago.NewDynamicView("seer_gdelt.EventCreateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventDetailRoute", lago.Route{
		Path:    AppUrl + "events/{id}/",
		Handler: lago.NewDynamicView("seer_gdelt.EventDetailView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventUpdateRoute", lago.Route{
		Path:    AppUrl + "events/{id}/edit/",
		Handler: lago.NewDynamicView("seer_gdelt.EventUpdateView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.EventDeleteRoute", lago.Route{
		Path:    AppUrl + "events/{id}/delete/",
		Handler: lago.NewDynamicView("seer_gdelt.EventDeleteView"),
	})

	_ = lago.RegistryRoute.Register("seer_gdelt.MapRoute", lago.Route{
		Path:    AppUrl + "map/",
		Handler: lago.NewDynamicView("seer_gdelt.MapView"),
	})
}

func init() {
	registerRoutes()
}
