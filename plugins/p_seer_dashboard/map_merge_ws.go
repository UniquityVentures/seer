package p_seer_dashboard

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/seer/plugins/p_seer_aisstream"
	"github.com/UniquityVentures/seer/plugins/p_seer_gdelt"
	"github.com/UniquityVentures/seer/plugins/p_seer_opensky"
	"github.com/fxamacker/cbor/v2"
)

const (
	dashboardMapViewportMarginDeg = 0.25
	dashboardMapMaxFrameBytes     = 1 << 20
)

type dashboardMapViewportMessage struct {
	Type   string          `json:"type" cbor:"type"`
	Bounds *viewportBounds `json:"bounds" cbor:"bounds"`
	Zoom   float64         `json:"zoom" cbor:"zoom"`
}

type viewportBounds struct {
	West  float64 `json:"west" cbor:"west"`
	South float64 `json:"south" cbor:"south"`
	East  float64 `json:"east" cbor:"east"`
	North float64 `json:"north" cbor:"north"`
}

type dashboardMapDataHandler struct{}

func (h dashboardMapDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") ||
		!headerContainsToken(r.Header.Get("Connection"), "upgrade") ||
		key == "" {
		http.Error(w, "bad websocket request", http.StatusBadRequest)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket unsupported", http.StatusInternalServerError)
		return
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		slog.Error("p_seer_dashboard: map websocket hijack failed", "error", err)
		return
	}
	ws := &dashboardMapWSConn{conn: conn, reader: rw.Reader}
	defer ws.close()

	accept := dashboardWebSocketAccept(key)
	if _, err := fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", accept); err != nil {
		slog.Warn("p_seer_dashboard: map websocket handshake failed", "error", err)
		return
	}
	if err := rw.Flush(); err != nil {
		slog.Warn("p_seer_dashboard: map websocket handshake flush failed", "error", err)
		return
	}

	ctx := r.Context()
	if _, err := getters.DBFromContext(ctx); err != nil {
		slog.Error("p_seer_dashboard: map websocket: db from context", "error", err)
		return
	}

	var writeMu sync.Mutex
	if err := sendDashboardMergedMapPoints(ctx, ws, &writeMu, nil); err != nil {
		if !errors.Is(err, ctx.Err()) {
			slog.Warn("p_seer_dashboard: map websocket: initial send failed", "error", err)
		}
		return
	}

	for {
		opcode, payload, err := ws.readFrame()
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, ctx.Err()) {
				slog.Debug("p_seer_dashboard: map websocket receive closed", "error", err)
			}
			return
		}
		if opcode != 0x1 && opcode != 0x2 {
			continue
		}
		var msg dashboardMapViewportMessage
		if err := cbor.Unmarshal(payload, &msg); err != nil {
			slog.Debug("p_seer_dashboard: map websocket ignored malformed message", "error", err)
			continue
		}
		if msg.Type != "mapDisplayViewport" {
			continue
		}
		if err := sendDashboardMergedMapPoints(ctx, ws, &writeMu, msg.Bounds); err != nil {
			if !errors.Is(err, ctx.Err()) {
				slog.Warn("p_seer_dashboard: map websocket: viewport send failed", "error", err)
			}
			return
		}
	}
}

func sendDashboardMergedMapPoints(ctx context.Context, ws *dashboardMapWSConn, writeMu *sync.Mutex, raw *viewportBounds) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}

	var osVp *p_seer_opensky.OpenSkyViewportBounds
	var aiVp *p_seer_aisstream.AisstreamViewportBounds
	var gdWest, gdSouth, gdEast, gdNorth float64
	gdSkipBBox := true

	if raw != nil && raw.South <= raw.North {
		gdSkipBBox = false
		gdWest = raw.West - dashboardMapViewportMarginDeg
		gdSouth = raw.South - dashboardMapViewportMarginDeg
		gdEast = raw.East + dashboardMapViewportMarginDeg
		gdNorth = raw.North + dashboardMapViewportMarginDeg
		osVp = &p_seer_opensky.OpenSkyViewportBounds{
			West:  gdWest,
			South: gdSouth,
			East:  gdEast,
			North: gdNorth,
		}
		if !osVp.IsValid() {
			osVp = nil
		}
		aiVp = &p_seer_aisstream.AisstreamViewportBounds{
			West:  gdWest,
			South: gdSouth,
			East:  gdEast,
			North: gdNorth,
		}
		if !aiVp.IsValid() {
			aiVp = nil
		}
	}

	var merged []any

	osPts, err := p_seer_opensky.MapDisplayPointsForBounds(ctx, db, osVp, "opensky")
	if err != nil {
		slog.Warn("p_seer_dashboard: opensky map points", "error", err)
	} else {
		for i := range osPts {
			merged = append(merged, osPts[i])
		}
	}

	aiPts, err := p_seer_aisstream.MapDisplayPointsForBounds(ctx, db, aiVp, "aisstream")
	if err != nil {
		slog.Warn("p_seer_dashboard: aisstream map points", "error", err)
	} else {
		for i := range aiPts {
			merged = append(merged, aiPts[i])
		}
	}

	var gdPts []p_seer_gdelt.MapDisplayPointWire
	if gdSkipBBox {
		gdPts, err = p_seer_gdelt.MapDisplayPointsForBounds(ctx, nil, 0, 1, 0, -1, "gdelt")
	} else {
		gdPts, err = p_seer_gdelt.MapDisplayPointsForBounds(ctx, nil, gdWest, gdSouth, gdEast, gdNorth, "gdelt")
	}
	if err != nil {
		slog.Warn("p_seer_dashboard: gdelt map points", "error", err)
	} else {
		for i := range gdPts {
			merged = append(merged, gdPts[i])
		}
	}

	b, err := cbor.Marshal(merged)
	if err != nil {
		return err
	}
	if len(b) > dashboardMapMaxFrameBytes {
		return fmt.Errorf("dashboard map payload exceeds 1 MiB: bytes=%d", len(b))
	}
	writeMu.Lock()
	defer writeMu.Unlock()
	return ws.writeBinary(b)
}

func headerContainsToken(header, token string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func dashboardWebSocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(sum[:])
}

type dashboardMapWSConn struct {
	conn   net.Conn
	reader *bufio.Reader
	write  sync.Mutex
}

func (c *dashboardMapWSConn) close() error {
	return c.conn.Close()
}

func (c *dashboardMapWSConn) readFrame() (byte, []byte, error) {
	for {
		header := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, header); err != nil {
			return 0, nil, err
		}
		fin := header[0]&0x80 != 0
		opcode := header[0] & 0x0f
		masked := header[1]&0x80 != 0
		length := uint64(header[1] & 0x7f)

		if !fin {
			return 0, nil, fmt.Errorf("fragmented websocket frames unsupported")
		}
		if !masked {
			return 0, nil, fmt.Errorf("client websocket frame not masked")
		}
		switch length {
		case 126:
			var b [2]byte
			if _, err := io.ReadFull(c.reader, b[:]); err != nil {
				return 0, nil, err
			}
			length = uint64(binary.BigEndian.Uint16(b[:]))
		case 127:
			var b [8]byte
			if _, err := io.ReadFull(c.reader, b[:]); err != nil {
				return 0, nil, err
			}
			length = binary.BigEndian.Uint64(b[:])
		}
		if opcode >= 0x8 && length > 125 {
			return 0, nil, fmt.Errorf("websocket control frame too large")
		}
		if length > 1<<20 {
			return 0, nil, fmt.Errorf("websocket frame exceeds 1 MiB")
		}

		var mask [4]byte
		if _, err := io.ReadFull(c.reader, mask[:]); err != nil {
			return 0, nil, err
		}
		payload := make([]byte, int(length))
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return 0, nil, err
		}
		for i := range payload {
			payload[i] ^= mask[i%4]
		}

		switch opcode {
		case 0x1:
			return opcode, payload, nil
		case 0x2:
			return opcode, payload, nil
		case 0x8:
			return 0, nil, io.EOF
		case 0x9:
			if err := c.writeFrame(0xA, payload); err != nil {
				return 0, nil, err
			}
		case 0xA:
			continue
		case 0x0:
			return 0, nil, fmt.Errorf("websocket continuation frame unsupported")
		default:
			return 0, nil, fmt.Errorf("unsupported websocket opcode %d", opcode)
		}
	}
}

func (c *dashboardMapWSConn) writeBinary(b []byte) error {
	return c.writeFrame(0x2, b)
}

func (c *dashboardMapWSConn) writeFrame(opcode byte, payload []byte) error {
	c.write.Lock()
	defer c.write.Unlock()

	if len(payload) > dashboardMapMaxFrameBytes {
		return fmt.Errorf("websocket payload exceeds 1 MiB")
	}
	header := []byte{0x80 | opcode}
	switch {
	case len(payload) < 126:
		header = append(header, byte(len(payload)))
	case len(payload) <= 65535:
		header = append(header, 126, byte(len(payload)>>8), byte(len(payload)))
	default:
		header = append(header, 127)
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(len(payload)))
		header = append(header, b[:]...)
	}
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err := c.conn.Write(payload)
	return err
}
