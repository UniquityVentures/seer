package p_seer_node_fleet

import (
	"github.com/UniquityVentures/lamu/lamu"
	"golang.org/x/net/websocket"
)

const fleetWebSocketPath = "/fleet/websocket"

func init() {
	registerPluginRoute("seer_node_fleet.DefaultRoute", lamu.Route{
		Path:    AppUrl,
		Handler: lamu.NewDynamicView("seer_node_fleet.HomeView"),
	})

	registerPluginRoute("seer_node_fleet.WebSocketRoute", lamu.Route{
		Path: fleetWebSocketPath,
		Handler: websocket.Server{
			Handler: fleetWebSocketConn,
		},
	})
}
