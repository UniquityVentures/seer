package p_seer_dashboard

import (
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/views"
	"github.com/UniquityVentures/seer/plugins/p_seer_intel"
)

func init() {
	lago.RegistryView.Patch("dashboard.AppsView", func(v *views.View) *views.View {
		return v.WithLayer("seer_dashboard.intel_latest", views.LayerList[p_seer_intel.Intel]{
			Key:      getters.Static("seerDashboardIntelLatest"),
			PageSize: getters.Static(uint(20)),
			QueryPatchers: views.QueryPatchers[p_seer_intel.Intel]{
				{Key: "seer_dashboard.intel_order", Value: views.QueryPatcherOrderBy[p_seer_intel.Intel]{Order: "datetime DESC, id DESC"}},
			},
		})
	})
}
