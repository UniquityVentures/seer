package p_seer_node_fleet

import (
	"fmt"

	"github.com/UniquityVentures/seer/plugins/p_seer_node_fleet/messages"
	"golang.org/x/net/websocket"
	"google.golang.org/protobuf/proto"
)

var ScraperCodec = websocket.Codec{
	Marshal:   ScraperMarshal,
	Unmarshal: ScraperUnmarshal,
}

func ScraperMarshal(v any) ([]byte, byte, error) {
	command, ok := v.(*messages.Command)
	if !ok {
		return nil, 0, fmt.Errorf("Couldn't marshal %T, only support *messages.Command)", v)
	}
	data, err := proto.Marshal(command)
	return data, websocket.BinaryFrame, err
}

func ScraperUnmarshal(data []byte, payloadType byte, v any) error {
	if payloadType != websocket.BinaryFrame {
		return fmt.Errorf("Only Binary frames are supported")
	}
	respContainer, ok := v.(*messages.Response)
	if !ok {
		return fmt.Errorf("Couldn't unmarshal to %T, only support *messages.Response)", v)
	}
	return proto.Unmarshal(data, respContainer)
}
