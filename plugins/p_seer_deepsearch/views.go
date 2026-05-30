package p_seer_deepsearch

import (
	"net/http"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
	"gorm.io/gorm"
)

func init() {
	deepSearchListPatchers := views.QueryPatchers[DeepSearch]{
		{Key: "seer_deepsearch.list.not_deleted", Value: deepSearchActiveOnlyPatcher{}},
		{Key: "seer_deepsearch.list.order", Value: views.QueryPatcherOrderBy[DeepSearch]{Order: "id DESC"}},
	}

	deepSearchDetailPatchers := views.QueryPatchers[DeepSearch]{
		{Key: "seer_deepsearch.detail_preload_logs", Value: views.QueryPatcherPreload[DeepSearch]{
			Fields: []string{"Logs"},
			PreloadBuilder: func(_ views.View, _ *http.Request, pb gorm.PreloadBuilder) error {
				pb.Order(`"created_at" DESC`)
				return nil
			},
		}},
	}

	registerPluginView("seer_deepsearch.HomeView",
		lamu.GetPageView("seer_deepsearch.DeepSearchHome").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}))

	registerPluginView("seer_deepsearch.HistoryView",
		lamu.GetPageView("seer_deepsearch.HistoryTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.deepsearch.list", views.LayerList[DeepSearch]{
				Key:           getters.Static("deepSearches"),
				QueryPatchers: deepSearchListPatchers,
			}))

	registerPluginView("seer_deepsearch.StartView",
		lamu.GetPageView("seer_deepsearch.StartBlank").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.start_method", deepSearchStartRejectGetLayer{}).
			WithLayer("seer_deepsearch.start_post", deepSearchStartPostLayer{}))

	registerPluginView("seer_deepsearch.DetailView",
		lamu.GetPageView("seer_deepsearch.DeepSearchDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.deepsearch.detail", views.LayerDetail[DeepSearch]{
				Key:           getters.Static("deepSearch"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: deepSearchDetailPatchers,
			}))

	registerPluginView("seer_deepsearch.StopView",
		lamu.GetPageView("seer_deepsearch.StartBlank").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.stop_method", deepSearchStartRejectGetLayer{}).
			WithLayer("seer_deepsearch.stop_post", deepSearchStopPostLayer{}))

	registerPluginView("seer_deepsearch.RestartView",
		lamu.GetPageView("seer_deepsearch.StartBlank").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_deepsearch.restart_method", deepSearchStartRejectGetLayer{}).
			WithLayer("seer_deepsearch.restart_post", deepSearchRestartPostLayer{}))
}
