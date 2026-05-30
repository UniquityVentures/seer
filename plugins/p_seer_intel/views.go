package p_seer_intel

import (
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
	"github.com/UniquityVentures/lamu/views"
)

func init() {
	intelListPatchers := views.QueryPatchers[Intel]{
		{Key: "seer_intel.intel.order", Value: views.QueryPatcherOrderBy[Intel]{Order: "datetime DESC, id DESC"}},
	}

	registerPluginView("seer_intel.ListView",
		lamu.GetPageView("seer_intel.IntelTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_intel.intel.list", views.LayerList[Intel]{
				Key:           getters.Static("intels"),
				QueryPatchers: intelListPatchers,
			}))

	registerPluginView("seer_intel.DetailView",
		lamu.GetPageView("seer_intel.IntelDetail").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_intel.intel.detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_intel.intel.source_href", intelSourceDetailHrefLayer{}))
}
