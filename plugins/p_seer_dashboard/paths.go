package p_seer_dashboard

import (
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
)

func init() {
	registerDashboardMapRoutes()
	registerDashboardMarkerIconRoutes()
	registerSeerDashboardHomePagePatch()
}

func registerDashboardMapRoutes() {
	registerPluginRoute("seer_dashboard.MapDataRoute", lamu.Route{
		Path:    AppUrl + "map/data/",
		Handler: p_users.RequireAuth(dashboardMapDataHandler{}),
	})
}
