package p_seer_reddit

import (
	"github.com/UniquityVentures/lamu/lamu"
)

func registerRoutes() {
	registerPluginRoute("seer_reddit.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_reddit.RedditSourceListView"),
	})

	registerPluginRoute("seer_reddit.RedditSourceCreateRoute", lamu.Route{
		Path:    AppUrl + "sources/create/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditSourceCreateView"),
	})

	registerPluginRoute("seer_reddit.RedditSourceUnsetSelectRoute", lamu.Route{
		Path:    AppUrl + "sources/unset/select/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditSourceUnsetSelectView"),
	})

	registerPluginRoute("seer_reddit.RedditPostListBySourceRoute", lamu.Route{
		Path:    AppUrl + "sources/{source_id}/posts/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditPostListBySourceView"),
	})

	registerPluginRoute("seer_reddit.RedditPostListBySourceBulkAddIntelRoute", lamu.Route{
		Path:    AppUrl + "sources/{source_id}/posts/bulk-add-intel/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditPostListBySourceBulkAddIntelView"),
	})

	registerPluginRoute("seer_reddit.RedditSourceFetchPostsRoute", lamu.Route{
		Path:    AppUrl + "sources/{source_id}/fetch-posts/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditSourceFetchPostsView"),
	})

	registerPluginRoute("seer_reddit.RedditSourceDetailRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditSourceDetailView"),
	})

	registerPluginRoute("seer_reddit.RedditSourceUpdateRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditSourceUpdateView"),
	})

	registerPluginRoute("seer_reddit.RedditSourceDeleteRoute", lamu.Route{
		Path:    AppUrl + "sources/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditSourceDeleteView"),
	})

	registerPluginRoute("seer_reddit.RedditPostListRoute", lamu.Route{
		Path:    AppUrl + "posts/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditPostListView"),
	})

	registerPluginRoute("seer_reddit.RedditPostListBulkAddIntelRoute", lamu.Route{
		Path:    AppUrl + "posts/bulk-add-intel/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditPostListBulkAddIntelView"),
	})

	registerPluginRoute("seer_reddit.RedditPostDetailRoute", lamu.Route{
		Path:    AppUrl + "posts/{id}/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditPostDetailView"),
	})

	registerPluginRoute("seer_reddit.RedditPostDeleteRoute", lamu.Route{
		Path:    AppUrl + "posts/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditPostSoftDeleteView"),
	})

	registerPluginRoute("seer_reddit.RedditPostAddIntelRoute", lamu.Route{
		Path:    AppUrl + "posts/{id}/add-intel/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditPostAddIntelView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerListRoute", lamu.Route{
		Path:    AppUrl + "workers/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerListView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerCreateRoute", lamu.Route{
		Path:    AppUrl + "workers/create/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerCreateView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerSelectRoute", lamu.Route{
		Path:    AppUrl + "workers/select/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerSelectView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerWorkerPoolStartRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/start/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerWorkerPoolStartView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerWorkerPoolStopRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/worker-pool/stop/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerWorkerPoolStopView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerDetailRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerDetailView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerUpdateRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/edit/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerUpdateView"),
	})

	registerPluginRoute("seer_reddit.RedditRunnerDeleteRoute", lamu.Route{
		Path:    AppUrl + "workers/{id}/delete/",
		Handler: lamu.NewDynamicView("seer_reddit.RedditRunnerDeleteView"),
	})
}

func init() {
	registerRoutes()
}
