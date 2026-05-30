package p_seer_gdelt

import (
	"context"
	"net/http"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	"gorm.io/gorm"
)

var gdeltEventListPatchers = views.QueryPatchers[Event]{
	{Key: "seer_gdelt.event_list.order", Value: views.QueryPatcherOrderBy[Event]{Order: "id DESC"}},
}

// gdeltWorkerPoolStateLayer sets [workerPoolIsRunning] after [views.LayerDetail] for [GDELTWorker].
type gdeltWorkerPoolStateLayer struct{}

func (gdeltWorkerPoolStateLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		wk, ok := ctx.Value("gdeltWorker").(GDELTWorker)
		running := ok && wk.ID != 0 && GDELTWorkerPoolIsRunning(wk.ID)
		ctx = context.WithValue(ctx, "workerPoolIsRunning", running)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	gdeltSourcePatchers := views.QueryPatchers[GDELTSource]{
		{Key: "seer_gdelt.source.order", Value: views.QueryPatcherOrderBy[GDELTSource]{Order: "id DESC"}},
	}
	gdeltSourceDetailPatchers := views.QueryPatchers[GDELTSource]{
		{Key: "seer_gdelt.source.preload_worker", Value: views.QueryPatcherPreload[GDELTSource]{Fields: []string{"GDELTWorker"}}},
	}
	gdeltSourceUnsetPatchers := views.QueryPatchers[GDELTSource]{
		{Key: "seer_gdelt.source.unset_worker", Value: gdeltSourceUnsetWorkerPatcher{}},
		{Key: "seer_gdelt.source.order", Value: views.QueryPatcherOrderBy[GDELTSource]{Order: "id DESC"}},
	}

	registerPluginView("seer_gdelt.GDELTSourceListView",
		lamu.GetPageView("seer_gdelt.GDELTSourceTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.list", views.LayerList[GDELTSource]{
				Key:           getters.Static("gdeltSources"),
				QueryPatchers: gdeltSourcePatchers,
			}))

	registerPluginView("seer_gdelt.GDELTSourceUnsetSelectView",
		lamu.GetPageView("seer_gdelt.GDELTSourceUnsetSelectionTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.unset_select_list", views.LayerList[GDELTSource]{
				Key:           getters.Static("gdeltSources"),
				QueryPatchers: gdeltSourceUnsetPatchers,
			}))

	registerPluginView("seer_gdelt.GDELTSourceDetailView",
		lamu.GetPageView("seer_gdelt.GDELTSourceDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.detail", views.LayerDetail[GDELTSource]{
				Key:           getters.Static("gdeltSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: gdeltSourceDetailPatchers,
			}))

	registerPluginView("seer_gdelt.GDELTSourceCreateView",
		lamu.GetPageView("seer_gdelt.GDELTSourceCreateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.create", views.LayerCreate[GDELTSource]{
				SuccessURL: lamu.RoutePath("seer_gdelt.GDELTSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.source.form_validate", Value: gdeltSourceFormValidate{}},
				},
			}))

	registerPluginView("seer_gdelt.GDELTSourceUpdateView",
		lamu.GetPageView("seer_gdelt.GDELTSourceUpdateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.detail_for_update", views.LayerDetail[GDELTSource]{
				Key:           getters.Static("gdeltSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: gdeltSourceDetailPatchers,
			}).
			WithLayer("seer_gdelt.gdelt_source.update", views.LayerUpdate[GDELTSource]{
				Key: getters.Static("gdeltSource"),
				SuccessURL: lamu.RoutePath("seer_gdelt.GDELTSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltSource.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.source.form_validate", Value: gdeltSourceFormValidate{}},
				},
			}))

	registerPluginView("seer_gdelt.GDELTSourceDeleteView",
		lamu.GetPageView("seer_gdelt.GDELTSourceDeleteForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.delete_detail", views.LayerDetail[GDELTSource]{
				Key:           getters.Static("gdeltSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: gdeltSourceDetailPatchers,
			}).
			WithLayer("seer_gdelt.gdelt_source.delete", views.LayerDelete[GDELTSource]{
				Key:        getters.Static("gdeltSource"),
				SuccessURL: lamu.RoutePath("seer_gdelt.DefaultRoute", nil),
			}))

	gdeltWorkerPatchers := views.QueryPatchers[GDELTWorker]{
		{Key: "seer_gdelt.worker.order", Value: views.QueryPatcherOrderBy[GDELTWorker]{Order: "id DESC"}},
	}

	registerPluginView("seer_gdelt.GDELTWorkerListView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.list", views.LayerList[GDELTWorker]{
				Key:           getters.Static("gdeltWorkers"),
				QueryPatchers: gdeltWorkerPatchers,
			}))

	registerPluginView("seer_gdelt.GDELTWorkerSelectView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerSelectionTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.select_list", views.LayerList[GDELTWorker]{
				Key:           getters.Static("gdeltWorkers"),
				QueryPatchers: gdeltWorkerPatchers,
			}))

	registerPluginView("seer_gdelt.GDELTWorkerDetailView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.detail", views.LayerDetail[GDELTWorker]{
				Key:          getters.Static("gdeltWorker"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.gdelt_worker.worker_pool_state", gdeltWorkerPoolStateLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.run_logs", p_seer_workerregistry.RunnerRunLogsLayer{
				RunnerContextKey: "gdeltWorker",
				Kind:             p_seer_workerregistry.WorkerRunnerKindGDELT,
			}))

	registerPluginView("seer_gdelt.GDELTWorkerCreateView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerCreateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.create", views.LayerCreate[GDELTWorker]{
				SuccessURL: lamu.RoutePath("seer_gdelt.GDELTWorkerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.worker.validate", Value: gdeltWorkerValidate{}},
				},
			}))

	registerPluginView("seer_gdelt.GDELTWorkerUpdateView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerUpdateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.detail_for_update", views.LayerDetail[GDELTWorker]{
				Key:          getters.Static("gdeltWorker"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.gdelt_worker.enrich_source_ids", gdeltWorkerEnrichSourceIDsLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.update", views.LayerUpdate[GDELTWorker]{
				Key: getters.Static("gdeltWorker"),
				SuccessURL: lamu.RoutePath("seer_gdelt.GDELTWorkerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltWorker.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.worker.validate", Value: gdeltWorkerValidate{}},
				},
			}))

	registerPluginView("seer_gdelt.GDELTWorkerDeleteView",
		lamu.GetPageView("seer_gdelt.GDELTWorkerDeleteForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.delete_detail", views.LayerDetail[GDELTWorker]{
				Key:          getters.Static("gdeltWorker"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.gdelt_worker.delete", views.LayerDelete[GDELTWorker]{
				Key:        getters.Static("gdeltWorker"),
				SuccessURL: lamu.RoutePath("seer_gdelt.GDELTWorkerListRoute", nil),
			}))

	registerGDELTWorkerPoolViews()

	registerPluginView("seer_gdelt.MapView",
		lamu.GetPageView("seer_gdelt.MapPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.map", gdeltMapLayer{}))

	registerPluginView("seer_gdelt.SearchView",
		lamu.GetPageView("seer_gdelt.SearchPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.search", gdeltSearchLayer{}))

	registerPluginView("seer_gdelt.EventListView",
		lamu.GetPageView("seer_gdelt.EventTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.list", views.LayerList[Event]{
				Key:           getters.Static("gdeltEvents"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: gdeltEventListPatchers,
			}))

	registerPluginView("seer_gdelt.EventCreateView",
		lamu.GetPageView("seer_gdelt.EventCreateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.create", views.LayerCreate[Event]{
				SuccessURL: lamu.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	registerPluginView("seer_gdelt.EventDetailView",
		lamu.GetPageView("seer_gdelt.EventDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}))

	registerPluginView("seer_gdelt.EventUpdateView",
		lamu.GetPageView("seer_gdelt.EventUpdateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail_for_update", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.event.update", views.LayerUpdate[Event]{
				Key: getters.Static("gdeltEvent"),
				SuccessURL: lamu.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
				}),
			}))

	registerPluginView("seer_gdelt.EventDeleteView",
		lamu.GetPageView("seer_gdelt.EventDeleteForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail_for_delete", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.event.delete", views.LayerDelete[Event]{
				Key:        getters.Static("gdeltEvent"),
				SuccessURL: lamu.RoutePath("seer_gdelt.EventListRoute", nil),
			}))
}

type gdeltSourceUnsetWorkerPatcher struct{}

func (gdeltSourceUnsetWorkerPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[GDELTSource]) gorm.ChainInterface[GDELTSource] {
	return q.Where("gdelt_worker_id IS NULL")
}
