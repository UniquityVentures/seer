package p_seer_deepsearch

import (
	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

var deepSearchHomeFormName = getters.Static("seer_deepsearch.HomeSearchForm")

func registerDeepSearchMenuPages() {
	registerPluginPage("seer_deepsearch.DeepSearchMenu", &components.SidebarMenu{
		Title: getters.Static("Deep search"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lamu.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("New search"),
				Url:   lamu.RoutePath("seer_deepsearch.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("History"),
				Url:   lamu.RoutePath("seer_deepsearch.HistoryRoute", nil),
			},
		},
	})
}

func registerDeepSearchSearchPages() {
	registerPluginPage("seer_deepsearch.DeepSearchHome", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_deepsearch.DeepSearchMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      deepSearchHomeFormName,
				ActionURL: lamu.RoutePath("seer_deepsearch.StartRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[DeepSearch]{
						Getter:   getters.Static(DeepSearch{}),
						Title:    "Deep search",
						Subtitle: "Enter a research question. The app expands queries, searches the web (Google Programmable Search), scrapes pages, ingests Intel, then writes a markdown report. Requires [Plugins.p_seer_deepsearch] apiKey+cx and [Plugins.p_google_genai] for LLM calls.",
						Classes:  "@container max-w-2xl mx-auto",
						ChildrenInput: []components.PageInterface{
							&components.InputText{
								Page:     components.Page{Key: "seer_deepsearch.HomeQueryInput"},
								Label:    "Question",
								Name:     "Query",
								Required: true,
								Getter:   getters.Key[string]("$in.Query"),
								Classes:  "w-full",
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Run deep search"},
						},
						Attr: deepSearchHomeFormAttr(),
					},
				},
			},
		},
	})

	registerPluginPage("seer_deepsearch.StartBlank", &components.ContainerColumn{
		Page: components.Page{Key: "seer_deepsearch.StartBlankRoot"},
	})
}
