package p_seer_node_fleet

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"sync"
	"time"

	"github.com/UniquityVentures/lamu/registry"
	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet/messages"
)

type nodeChannels = registry.Pair[chan *messages.Command, chan *messages.Response]

var (
	nodeConnectionsMu sync.RWMutex
	nodeConnections   = map[uint64]*nodeConnection{}
)

// ConnectedNode is one live scraper websocket registered on this server.
type ConnectedNode struct {
	ID      uint64
	Version *messages.VersionResponse
}

type nodeConnection struct {
	channels nodeChannels
	version  *messages.VersionResponse
	done     chan struct{}
	once     sync.Once
}

func (c *nodeConnection) Close() {
	c.once.Do(func() {
		close(c.done)
	})
}

func registerNodeConnection(id uint64, conn *nodeConnection) {
	nodeConnectionsMu.Lock()
	defer nodeConnectionsMu.Unlock()
	if existing, ok := nodeConnections[id]; ok {
		existing.Close()
	}
	nodeConnections[id] = conn
}

func unregisterNodeConnection(id uint64, conn *nodeConnection) {
	nodeConnectionsMu.Lock()
	defer nodeConnectionsMu.Unlock()
	current, ok := nodeConnections[id]
	if !ok || current != conn {
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
// Spinlocks when no node connections available
func DispatchCommand(cmd *messages.Command) (*messages.Response, error) {
	for {
		nodeConnectionsMu.RLock()
		hasNodes := len(nodeConnections) > 0
		nodeConnectionsMu.RUnlock()
		if hasNodes {
			break
		}
		time.Sleep(time.Second)
	}

	nodeConnectionsMu.RLock()
	ids := make([]uint64, 0, len(nodeConnections))
	for id := range nodeConnections {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		nodeConnectionsMu.RUnlock()
		return nil, fmt.Errorf("fleet node disconnected before command send")
	}
	nodeID := ids[rand.IntN(len(ids))]
	conn := nodeConnections[nodeID]
	nodeConnectionsMu.RUnlock()

	select {
	case conn.channels.Key <- cmd:
	case <-conn.done:
		return nil, fmt.Errorf("fleet node %d disconnected before command send", nodeID)
	}

	select {
	case resp := <-conn.channels.Value:
		return resp, nil
	case <-conn.done:
		return nil, fmt.Errorf("fleet node %d disconnected before response", nodeID)
	}
}
