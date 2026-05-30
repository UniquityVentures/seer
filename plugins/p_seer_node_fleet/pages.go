package p_seer_node_fleet

import (
	"context"
	"fmt"
	"strconv"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet/messages"
)

func formatVersionResponse(v *messages.VersionResponse) string {
	if v == nil {
		return "unknown"
	}
	return fmt.Sprintf("%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
}

func init() {
	registerNodeFleetPages()
}

func registerNodeFleetPages() {
	registerPluginPage("seer_node_fleet.NodeFleetMenu", &components.SidebarMenu{
		Title: getters.Static("Node fleet"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lamu.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Connected scrapers"),
				Url:   lamu.RoutePath("seer_node_fleet.DefaultRoute", nil),
			},
		},
	})

	registerPluginPage("seer_node_fleet.ConnectedNodesTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_node_fleet.NodeFleetMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[ConnectedNode]{
				Page:    components.Page{Key: "seer_node_fleet.ConnectedNodesTableBody"},
				UID:     "seer-node-fleet-table",
				Title:   "Connected scrapers",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[ConnectedNode]](connectedNodesKey),
				Columns: []components.TableColumn{
					{
						Label: "Node ID",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Map(getters.Key[uint64]("$row.ID"), func(_ context.Context, id uint64) (string, error) {
									return strconv.FormatUint(id, 10), nil
								}),
							},
						},
					},
					{
						Label: "Version",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Map(getters.Key[*messages.VersionResponse]("$row.Version"), func(_ context.Context, v *messages.VersionResponse) (string, error) {
									return formatVersionResponse(v), nil
								}),
							},
						},
					},
				},
			},
		},
	})
}
