package p_seer_deepsearch

import (
	"time"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/registry"
)

func deepSearchHistoryColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "Question",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter:  getters.Key[string]("$row.Query"),
					Classes: "break-words max-w-md line-clamp-2",
				},
			},
		},
		{
			Label: "Status",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter: registry.PairValueFromKey(getters.Key[string]("$row.Status"), DeepSearchStatusChoices),
				},
			},
		},
		{
			Label: "Started",
			Children: []components.PageInterface{
				&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.CreatedAt")},
			},
		},
	}
}

func registerDeepSearchHistoryPages() {
	registerPluginPage("seer_deepsearch.HistoryTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_deepsearch.DeepSearchMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[DeepSearch]{
				Page:    components.Page{Key: "seer_deepsearch.HistoryTableBody"},
				UID:     "seer-deepsearch-history-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[DeepSearch]]("deepSearches"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{
						Page:  components.Page{Key: "seer_deepsearch.HistoryNewSearchBtn"},
						Link:  lamu.RoutePath("seer_deepsearch.DefaultRoute", nil),
						Label: "New search",
					},
				},
				RowAttr: getters.RowAttrNavigate(
					lamu.RoutePath("seer_deepsearch.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: deepSearchHistoryColumns(),
			},
		},
	})
}
