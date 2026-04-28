package p_seer_intel

import (
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/plugins/p_users"
	"github.com/UniquityVentures/lago/views"
)

func init() {
	intelListPatchers := views.QueryPatchers[Intel]{
		{Key: "seer_intel.intel.order", Value: views.QueryPatcherOrderBy[Intel]{Order: "datetime DESC, id DESC"}},
	}

	lago.RegistryView.Register("seer_intel.ListView",
		lago.GetPageView("seer_intel.IntelTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_intel.intel.list", views.LayerList[Intel]{
				Key:           getters.Static("intels"),
				QueryPatchers: intelListPatchers,
			}))

	lago.RegistryView.Register("seer_intel.DetailView",
		lago.GetPageView("seer_intel.IntelDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_intel.intel.detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_intel.intel.source_href", intelSourceDetailHrefLayer{}))
}
