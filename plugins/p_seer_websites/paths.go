package p_seer_websites

import (
	"github.com/UniquityVentures/lamu/lamu"
)

func registerRoutes() {
	registerPluginRoute("seer_websites.WebsiteListRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_websites.WebsiteListView"),
	})

	registerPluginRoute("seer_websites.WebsiteAddRoute", lamu.Route{
		Path:    AppUrl + "add/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteAddView"),
	})

	// Saved scrape rows live under pages/ so /seer-websites/{id}/… does not collide with
	// /seer-websites/workers/{id}/… (Go 1.22+ mux cannot rank those patterns).
	registerPluginRoute("seer_websites.WebsiteDetailRoute", lamu.Route{
		Path:    AppUrl + "pages/{id}/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteDetailView"),
	})

	registerPluginRoute("seer_websites.WebsiteDeleteRoute", lamu.Route{
		Path:    AppUrl + "pages/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteSoftDeleteView"),
	})

	registerPluginRoute("seer_websites.WebsiteAddIntelRoute", lamu.Route{
		Path:    AppUrl + "pages/{id}/add-intel/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteAddIntelView"),
	})

	registerPluginRoute("seer_websites.WebsiteAddAllIntelRoute", lamu.Route{
		Path:    AppUrl + "add-all-intel/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteAddAllIntelView"),
	})

	registerPluginRoute("seer_websites.WebsiteSourceListRoute", lamu.Route{
		Path:    AppUrl + "sources/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteSourceListView"),
	})

	registerPluginRoute("seer_websites.WebsiteSourceCreateRoute", lamu.Route{
		Path:    AppUrl + "sources/create/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteSourceCreateView"),
	})

	registerPluginRoute("seer_websites.WebsiteSourceDetailRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteSourceDetailView"),
	})

	registerPluginRoute("seer_websites.WebsiteSourceUpdateRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteSourceUpdateView"),
	})

	registerPluginRoute("seer_websites.WebsiteSourceDeleteRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteSourceDeleteView"),
	})

	registerPluginRoute("seer_websites.WebsiteSourceFetchRoute", lamu.Route{
		Path:    AppUrl + "sources/{source_id}/fetch/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteSourceFetchView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerListRoute", lamu.Route{
		Path:    AppUrl + "workers/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerListView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerCreateRoute", lamu.Route{
		Path:    AppUrl + "workers/create/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerCreateView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerSelectRoute", lamu.Route{
		Path:    AppUrl + "workers/select/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerSelectView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerWorkerPoolStartRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/start/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerWorkerPoolStartView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerWorkerPoolStopRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/stop/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerWorkerPoolStopView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerDetailRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerDetailView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerUpdateRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerUpdateView"),
	})

	registerPluginRoute("seer_websites.WebsiteRunnerDeleteRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_websites.WebsiteRunnerDeleteView"),
	})
}

func init() {
	registerRoutes()
}
