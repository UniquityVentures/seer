package p_seer_opensky

import (
	"net/http"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
	"gorm.io/gorm"
)

var stateListQueryPatchers = views.QueryPatchers[OpenSkyState]{
	{Key: "seer_opensky.state_list.order", Value: views.QueryPatcherOrderBy[OpenSkyState]{Order: "id DESC"}},
}

var transitionListQueryPatchers = views.QueryPatchers[p_seer_intel.Intel]{
	{Key: "seer_opensky.transition_list.filter_opensky", Value: transitionFilterOpenskyPatcher{}},
	{Key: "seer_opensky.transition_list.order", Value: views.QueryPatcherOrderBy[p_seer_intel.Intel]{Order: "id DESC"}},
}

type transitionFilterOpenskyPatcher struct{}

func (transitionFilterOpenskyPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[p_seer_intel.Intel]) gorm.ChainInterface[p_seer_intel.Intel] {
	return q.Where("kind = ?", "opensky")
}


func init() {
	registerPluginView("seer_opensky.StateListView",
		lamu.GetPageView("seer_opensky.StateTablePage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.list", views.LayerList[OpenSkyState]{
				Key:           getters.Static("openskyStates"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: stateListQueryPatchers,
			}))

	registerPluginView("seer_opensky.StateCreateView",
		lamu.GetPageView("seer_opensky.StateCreateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.create", views.LayerCreate[OpenSkyState]{
				SuccessURL: lamu.RoutePath("seer_opensky.StateDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: openSkyFormPatchers,
			}))

	registerPluginView("seer_opensky.StateDetailView",
		lamu.GetPageView("seer_opensky.StateDetailPage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.detail", views.LayerDetail[OpenSkyState]{
				Key:          getters.Static("openskyState"),
				PathParamKey: getters.Static("id"),
			}))

	registerPluginView("seer_opensky.StateUpdateView",
		lamu.GetPageView("seer_opensky.StateUpdateForm").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.detail_for_update", views.LayerDetail[OpenSkyState]{
				Key:          getters.Static("openskyState"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_opensky.state.update", views.LayerUpdate[OpenSkyState]{
				Key: getters.Static("openskyState"),
				SuccessURL: lamu.RoutePath("seer_opensky.StateDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("openskyState.ID")),
				}),
				FormPatchers: openSkyFormPatchers,
			}))

	registerPluginView("seer_opensky.StateDeleteView",
		lamu.GetPageView("seer_opensky.StateDeleteFormModal").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.state.detail_for_delete", views.LayerDetail[OpenSkyState]{
				Key:          getters.Static("openskyState"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_opensky.state.delete", views.LayerDelete[OpenSkyState]{
				Key:        getters.Static("openskyState"),
				SuccessURL: lamu.RoutePath("seer_opensky.StateListRoute", nil),
			}))

	registerPluginView("seer_opensky.TransitionListView",
		lamu.GetPageView("seer_opensky.TransitionTablePage").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_opensky.transition.list", views.LayerList[p_seer_intel.Intel]{
				Key:           getters.Static("openskyTransitions"),
				PageSize:      getters.Static(uint(25)),
				QueryPatchers: transitionListQueryPatchers,
			}))
}

