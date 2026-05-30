package p_seer_intel

import (
	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

func init() {
	registerMenuPages()
	registerTablePages()
	registerDetailPages()
}

func registerMenuPages() {
	registerPluginPage("seer_intel.IntelMenu", &components.SidebarMenu{
		Title: getters.Static("Intel"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lamu.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Intel"),
				Url:   lamu.RoutePath("seer_intel.DefaultRoute", nil),
			},
		},
	})

	registerPluginPage("seer_intel.IntelDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Intel: %s", getters.Any(getters.Key[string]("intel.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("All Intel"),
			Url:   lamu.RoutePath("seer_intel.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lamu.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("intel.ID")),
				}),
			},
		},
	})
}
