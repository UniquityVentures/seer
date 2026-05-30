package p_seer_node_fleet

import (
	"github.com/UniquityVentures/lamu/lamu"
	"github.com/UniquityVentures/lamu/plugins/p_users"
)

func init() {
	registerPluginView("seer_node_fleet.HomeView",
		lamu.GetPageView("seer_node_fleet.ConnectedNodesTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_node_fleet.connected_nodes", connectedNodesLayer{}))
}
