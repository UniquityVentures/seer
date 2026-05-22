package p_seer_node_fleet

import (
	"github.com/UniquityVentures/lago/lago"
	"golang.org/x/net/websocket"
)

const fleetWebSocketPath = "/fleet/websocket"

func init() {
	_ = lago.RegistryRoute.Register("seer_node_fleet.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("seer_node_fleet.HomeView"),
	})

	_ = lago.RegistryRoute.Register("seer_node_fleet.WebSocketRoute", lago.Route{
		Path: fleetWebSocketPath,
		Handler: websocket.Server{
			Handler: fleetWebSocketConn,
		},
	})
}
