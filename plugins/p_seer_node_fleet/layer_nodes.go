package p_seer_node_fleet

import (
	"context"
	"net/http"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/views"
)

const connectedNodesKey = "seer_node_fleet.connectedNodes"

type connectedNodesLayer struct{}

func (connectedNodesLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nodes := ConnectedNodes()
		list := components.ObjectList[ConnectedNode]{
			Items:    nodes,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(nodes)),
		}
		ctx := context.WithValue(r.Context(), connectedNodesKey, list)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
