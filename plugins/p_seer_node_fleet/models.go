package p_seer_node_fleet

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/UniquityVentures/lago/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet/messages"
)

type nodeChannels = registry.Pair[chan *messages.Command, chan *messages.Response]

var (
	nodeConnectionsMu sync.RWMutex
	nodeConnections   = map[uint64]nodeConnection{}
)

// ConnectedNode is one live scraper websocket registered on this server.
type ConnectedNode struct {
	ID      uint64
	Version *messages.VersionResponse
}

type nodeConnection struct {
	channels nodeChannels
	version  *messages.VersionResponse
}

func registerNodeConnection(id uint64, channels nodeChannels, version *messages.VersionResponse) {
	nodeConnectionsMu.Lock()
	defer nodeConnectionsMu.Unlock()
	if existing, ok := nodeConnections[id]; ok {
		close(existing.channels.Key)
		close(existing.channels.Value)
	}
	nodeConnections[id] = nodeConnection{
		channels: channels,
		version:  version,
	}
}

func unregisterNodeConnection(id uint64, channels nodeChannels) {
	nodeConnectionsMu.Lock()
	defer nodeConnectionsMu.Unlock()
	current, ok := nodeConnections[id]
	if !ok || current.channels.Key != channels.Key {
		return
	}
	delete(nodeConnections, id)
}

// ConnectedNodes returns a stable snapshot of attached scrapers and their versions.
func ConnectedNodes() []ConnectedNode {
	nodeConnectionsMu.RLock()
	defer nodeConnectionsMu.RUnlock()
	out := make([]ConnectedNode, 0, len(nodeConnections))
	for id, conn := range nodeConnections {
		out = append(out, ConnectedNode{
			ID:      id,
			Version: conn.version,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

// DispatchCommand sends cmd to a randomly selected connected fleet node and returns its response.
func DispatchCommand(cmd *messages.Command) (*messages.Response, error) {
	nodeConnectionsMu.RLock()
	if len(nodeConnections) == 0 {
		nodeConnectionsMu.RUnlock()
		return nil, fmt.Errorf("no fleet nodes connected")
	}
	ids := make([]uint64, 0, len(nodeConnections))
	for id := range nodeConnections {
		ids = append(ids, id)
	}
	nodeID := ids[rand.IntN(len(ids))]
	channels := nodeConnections[nodeID].channels
	nodeConnectionsMu.RUnlock()

	channels.Key <- cmd
	resp, ok := <-channels.Value
	if !ok {
		return nil, fmt.Errorf("fleet node %d disconnected before response", nodeID)
	}
	return resp, nil
}
