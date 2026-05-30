package p_seer_websites

import (
	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

var websiteAddFormName = getters.Static("seer_websites.WebsiteAddForm")

func registerWebsiteFormPages() {
	registerPluginPage("seer_websites.WebsiteAddForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      websiteAddFormName,
				ActionURL: lamu.RoutePath("seer_websites.WebsiteAddRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Website]{
						Getter:   getters.Static(Website{}),
						Attr:     getters.FormBubbling(websiteAddFormName),
						Title:    "Add website",
						Subtitle: "Enter a public http(s) URL.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							&components.InputText{
								Page:     components.Page{Key: "seer_websites.WebsiteAddURLInput"},
								Label:    "Page URL",
								Name:     "URL",
								Required: true,
								Getter:   pageURLStringFromKey("$in.URL"),
								Classes:  "w-full",
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Scrape and save"},
						},
					},
				},
			},
		},
	})
}
