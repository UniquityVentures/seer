package p_seer_dashboard

import (
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/plugins/p_users"
)

func init() {
	registerDashboardMapRoutes()
	registerDashboardMarkerIconRoutes()
	registerSeerDashboardHomePagePatch()
	registerDashboardPlugin()
}

func registerDashboardMapRoutes() {
	_ = lago.RegistryRoute.Register("seer_dashboard.MapDataRoute", lago.Route{
		Path:    AppUrl + "map/data/",
		Handler: p_users.RequireAuth(dashboardMapDataHandler{}),
	})
}

