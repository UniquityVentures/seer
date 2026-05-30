package p_seer_websites

import (
	"time"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

func websiteListColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "URL",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter:  pageURLStringFromKey("$row.URL"),
					Classes: "break-all max-w-prose",
				},
			},
		},
		{
			Label: "Saved",
			Children: []components.PageInterface{
				&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.CreatedAt")},
			},
		},
	}
}

func registerWebsiteTablePages() {
	registerPluginPage("seer_websites.WebsiteTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Website]{
				Page:    components.Page{Key: "seer_websites.WebsiteTableBody"},
				UID:     "seer-websites-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Website]]("websites"),
				Actions: []components.PageInterface{
					&components.ButtonPost{
						Page:    components.Page{Key: "seer_websites.WebsiteTableAddAllIntelBtn"},
						Label:   "Add all to Intel",
						URL:     lamu.RoutePath("seer_websites.WebsiteAddAllIntelRoute", nil),
						Icon:    "document-plus",
						Classes: "btn-outline btn-primary btn-sm shrink-0",
					},
					&components.TableButtonCreate{Link: lamu.RoutePath("seer_websites.WebsiteAddRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lamu.RoutePath("seer_websites.WebsiteDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: websiteListColumns(),
			},
		},
	})
}
