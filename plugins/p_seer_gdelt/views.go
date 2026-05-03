package p_seer_gdelt

import (
	"context"
	"net/http"

	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/plugins/p_users"
	"github.com/UniquityVentures/lago/views"
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

	lago.RegistryView.Register("seer_gdelt.GDELTSourceListView",
		lago.GetPageView("seer_gdelt.GDELTSourceTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.list", views.LayerList[GDELTSource]{
				Key:           getters.Static("gdeltSources"),
				QueryPatchers: gdeltSourcePatchers,
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTSourceUnsetSelectView",
		lago.GetPageView("seer_gdelt.GDELTSourceUnsetSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.unset_select_list", views.LayerList[GDELTSource]{
				Key:           getters.Static("gdeltSources"),
				QueryPatchers: gdeltSourceUnsetPatchers,
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTSourceDetailView",
		lago.GetPageView("seer_gdelt.GDELTSourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.detail", views.LayerDetail[GDELTSource]{
				Key:           getters.Static("gdeltSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: gdeltSourceDetailPatchers,
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTSourceCreateView",
		lago.GetPageView("seer_gdelt.GDELTSourceCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.create", views.LayerCreate[GDELTSource]{
				SuccessURL: lago.RoutePath("seer_gdelt.GDELTSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.source.form_validate", Value: gdeltSourceFormValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTSourceUpdateView",
		lago.GetPageView("seer_gdelt.GDELTSourceUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.detail_for_update", views.LayerDetail[GDELTSource]{
				Key:           getters.Static("gdeltSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: gdeltSourceDetailPatchers,
			}).
			WithLayer("seer_gdelt.gdelt_source.update", views.LayerUpdate[GDELTSource]{
				Key: getters.Static("gdeltSource"),
				SuccessURL: lago.RoutePath("seer_gdelt.GDELTSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltSource.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.source.form_validate", Value: gdeltSourceFormValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTSourceDeleteView",
		lago.GetPageView("seer_gdelt.GDELTSourceDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_source.delete_detail", views.LayerDetail[GDELTSource]{
				Key:           getters.Static("gdeltSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: gdeltSourceDetailPatchers,
			}).
			WithLayer("seer_gdelt.gdelt_source.delete", views.LayerDelete[GDELTSource]{
				Key:        getters.Static("gdeltSource"),
				SuccessURL: lago.RoutePath("seer_gdelt.DefaultRoute", nil),
			}))

	gdeltWorkerPatchers := views.QueryPatchers[GDELTWorker]{
		{Key: "seer_gdelt.worker.order", Value: views.QueryPatcherOrderBy[GDELTWorker]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("seer_gdelt.GDELTWorkerListView",
		lago.GetPageView("seer_gdelt.GDELTWorkerTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.list", views.LayerList[GDELTWorker]{
				Key:           getters.Static("gdeltWorkers"),
				QueryPatchers: gdeltWorkerPatchers,
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTWorkerSelectView",
		lago.GetPageView("seer_gdelt.GDELTWorkerSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.select_list", views.LayerList[GDELTWorker]{
				Key:           getters.Static("gdeltWorkers"),
				QueryPatchers: gdeltWorkerPatchers,
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTWorkerDetailView",
		lago.GetPageView("seer_gdelt.GDELTWorkerDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.detail", views.LayerDetail[GDELTWorker]{
				Key:          getters.Static("gdeltWorker"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.gdelt_worker.worker_pool_state", gdeltWorkerPoolStateLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.run_logs", p_seer_workerregistry.RunnerRunLogsLayer{
				RunnerContextKey: "gdeltWorker",
				Kind:             p_seer_workerregistry.WorkerRunnerKindGDELT,
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTWorkerCreateView",
		lago.GetPageView("seer_gdelt.GDELTWorkerCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.create", views.LayerCreate[GDELTWorker]{
				SuccessURL: lago.RoutePath("seer_gdelt.GDELTWorkerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.worker.validate", Value: gdeltWorkerValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTWorkerUpdateView",
		lago.GetPageView("seer_gdelt.GDELTWorkerUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.detail_for_update", views.LayerDetail[GDELTWorker]{
				Key:          getters.Static("gdeltWorker"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.gdelt_worker.enrich_source_ids", gdeltWorkerEnrichSourceIDsLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.update", views.LayerUpdate[GDELTWorker]{
				Key: getters.Static("gdeltWorker"),
				SuccessURL: lago.RoutePath("seer_gdelt.GDELTWorkerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltWorker.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_gdelt.worker.validate", Value: gdeltWorkerValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_gdelt.GDELTWorkerDeleteView",
		lago.GetPageView("seer_gdelt.GDELTWorkerDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.gdelt_worker.delete_detail", views.LayerDetail[GDELTWorker]{
				Key:          getters.Static("gdeltWorker"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.gdelt_worker.delete", views.LayerDelete[GDELTWorker]{
				Key:        getters.Static("gdeltWorker"),
				SuccessURL: lago.RoutePath("seer_gdelt.GDELTWorkerListRoute", nil),
			}))

	registerGDELTWorkerPoolViews()

	lago.RegistryView.Register("seer_gdelt.MapView",
		lago.GetPageView("seer_gdelt.MapPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.map", gdeltMapLayer{}))

	lago.RegistryView.Register("seer_gdelt.SearchView",
		lago.GetPageView("seer_gdelt.SearchPage").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.search", gdeltSearchLayer{}))

	lago.RegistryView.Register("seer_gdelt.EventListView",
		lago.GetPageView("seer_gdelt.EventTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.list", views.LayerList[Event]{
				Key:           getters.Static("gdeltEvents"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: gdeltEventListPatchers,
			}))

	lago.RegistryView.Register("seer_gdelt.EventCreateView",
		lago.GetPageView("seer_gdelt.EventCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.create", views.LayerCreate[Event]{
				SuccessURL: lago.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("seer_gdelt.EventDetailView",
		lago.GetPageView("seer_gdelt.EventDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("seer_gdelt.EventUpdateView",
		lago.GetPageView("seer_gdelt.EventUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail_for_update", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.event.update", views.LayerUpdate[Event]{
				Key: getters.Static("gdeltEvent"),
				SuccessURL: lago.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
				}),
			}))

	lago.RegistryView.Register("seer_gdelt.EventDeleteView",
		lago.GetPageView("seer_gdelt.EventDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_gdelt.event.detail_for_delete", views.LayerDetail[Event]{
				Key:          getters.Static("gdeltEvent"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_gdelt.event.delete", views.LayerDelete[Event]{
				Key:        getters.Static("gdeltEvent"),
				SuccessURL: lago.RoutePath("seer_gdelt.EventListRoute", nil),
			}))
}

type gdeltSourceUnsetWorkerPatcher struct{}

func (gdeltSourceUnsetWorkerPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[GDELTSource]) gorm.ChainInterface[GDELTSource] {
	return q.Where("gdelt_worker_id IS NULL")
}
