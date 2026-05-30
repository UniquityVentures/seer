package p_seer_intel

import (
	"time"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

func registerTablePages() {
	registerPluginPage("seer_intel.IntelTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_intel.IntelMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Intel]{
				Page:    components.Page{Key: "seer_intel.IntelTableBody"},
				UID:     "seer-intel-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Intel]]("intels"),
				RowAttr: getters.RowAttrNavigate(
					lamu.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Kind",
						Name:  "Kind",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Kind")},
						},
					},
					{
						Label: "Datetime",
						Name:  "Datetime",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime")},
						},
					},
				},
			},
		},
	})
}
