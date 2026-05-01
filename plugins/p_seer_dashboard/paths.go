package p_seer_dashboard

import (
	"log"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/plugins/p_users"
)

func init() {
	registerDashboardMapRoutes()
	registerDashboardAppsPagePatch()
	registerDashboardPlugin()
}

func registerDashboardMapRoutes() {
	_ = lago.RegistryRoute.Register("seer_dashboard.MapDataRoute", lago.Route{
		Path:    AppUrl + "map/data/",
		Handler: p_users.RequireAuth(dashboardMapDataHandler{}),
	})
}

func registerDashboardAppsPagePatch() {
	lago.RegistryPage.Patch("dashboard.AppsPage", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellTopbarScaffold)
		if !ok {
			log.Panic("dashboard.AppsPage was not *components.ShellTopbarScaffold")
		}
		scaffold.ExtraHead = append(scaffold.ExtraHead, &components.MapDisplayLibreHead{
			Page: components.Page{Key: "seer_dashboard.MapLibreHead"},
		})
		components.ReplaceChild(scaffold, "dashboard.AppsPageLayout", func(layout *components.LayoutSimple) *components.LayoutSimple {
			if len(layout.Children) != 1 {
				log.Panic("dashboard.AppsPageLayout: expected exactly one child (AppsGrid)")
			}
			appsGrid := layout.Children[0]
			layout.Children = []components.PageInterface{
				&components.ContainerColumn{
					Page:    components.Page{Key: "seer_dashboard.DashboardMapColumn"},
					Classes: "gap-2 w-full max-w-5xl mx-auto",
				Children: []components.PageInterface{
					&SeerDashboardMap{
						Page:    components.Page{Key: "seer_dashboard.DashboardMap"},
						DataURL: lago.RoutePath("seer_dashboard.MapDataRoute", nil),
						Classes: "w-full h-[min(48vh,420px)] min-h-64 rounded-box border border-base-300 relative z-[1]",
					},
				},
				},
				appsGrid,
			}
			return layout
		})
		return scaffold
	})
}
