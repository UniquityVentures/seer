package p_seer_node_fleet

import (
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/lago/plugins/p_users"
)

func init() {
	lago.RegistryView.Register("seer_node_fleet.HomeView",
		lago.GetPageView("seer_node_fleet.ConnectedNodesTable").
			WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_node_fleet.connected_nodes", connectedNodesLayer{}))
}
