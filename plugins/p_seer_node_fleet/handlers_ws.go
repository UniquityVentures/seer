package p_seer_node_fleet

import (
	"log/slog"

	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet/messages"
	"golang.org/x/net/websocket"
)

func fleetWebSocketConn(ws *websocket.Conn) {
	defer ws.Close()

	getIDCmd := &messages.Command{
		Id: 1,
		CommandType: &messages.Command_GetId_{
			GetId_: &messages.GetID{},
		},
	}
	if err := ScraperCodec.Send(ws, getIDCmd); err != nil {
		slog.Warn("p_seer_node_fleet: get id send failed", "error", err)
		return
	}

	var getIDResp messages.Response
	if err := ScraperCodec.Receive(ws, &getIDResp); err != nil {
		slog.Warn("p_seer_node_fleet: get id receive failed", "error", err)
		return
	}
	if getIDResp.GetError() != nil {
		slog.Warn("p_seer_node_fleet: get id returned error")
		return
	}
	nodeID := getIDResp.GetOk().GetId().GetId()
	if nodeID == 0 {
		slog.Warn("p_seer_node_fleet: get id returned zero id")
		return
	}

	version := "unknown"
	var versionInfo *messages.VersionResponse
	getVersionCmd := &messages.Command{
		Id: 2,
		CommandType: &messages.Command_GetVersion{
			GetVersion: &messages.GetVersion{},
		},
	}
	if err := ScraperCodec.Send(ws, getVersionCmd); err != nil {
		slog.Warn("p_seer_node_fleet: get version send failed", "node_id", nodeID, "error", err)
	} else {
		var getVersionResp messages.Response
		if err := ScraperCodec.Receive(ws, &getVersionResp); err != nil {
			slog.Warn("p_seer_node_fleet: get version receive failed", "node_id", nodeID, "error", err)
		} else if getVersionResp.GetError() != nil {
			slog.Warn("p_seer_node_fleet: get version returned error", "node_id", nodeID)
		} else if v := getVersionResp.GetOk().GetVersion(); v != nil {
			copied := *v
			versionInfo = &copied
			version = formatVersionResponse(versionInfo)
		}
	}

	commandCh := make(chan *messages.Command)
	responseCh := make(chan *messages.Response)
	channels := nodeChannels{Key: commandCh, Value: responseCh}
	registerNodeConnection(nodeID, channels, versionInfo)
	defer func() {
		close(commandCh)
		close(responseCh)
		unregisterNodeConnection(nodeID, channels)
	}()

	slog.Info("p_seer_node_fleet: node connected", "node_id", nodeID, "version", version)

	for cmd := range commandCh {
		if err := ScraperCodec.Send(ws, cmd); err != nil {
			slog.Warn("p_seer_node_fleet: command send failed", "node_id", nodeID, "error", err)
			return
		}

		var resp messages.Response
		if err := ScraperCodec.Receive(ws, &resp); err != nil {
			slog.Warn("p_seer_node_fleet: response receive failed", "node_id", nodeID, "error", err)
			return
		}
		respCopy := resp
		responseCh <- &respCopy
	}
}
