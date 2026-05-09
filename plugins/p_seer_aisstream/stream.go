package p_seer_aisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const maxBufferedAISPackets = 25000

func startAISStreamWorkerIfConfigured(db *gorm.DB) {
	if db == nil || Config == nil {
		return
	}
	if !Config.Enabled {
		slog.Info("p_seer_aisstream: stream worker disabled")
		return
	}
	if Config.APIKey == "" {
		slog.Error("p_seer_aisstream: stream worker not started: apiKey required in [Plugins.p_seer_aisstream]")
		return
	}
	go runAISStreamWorker(context.Background(), db, Config)
}

func runAISStreamWorker(ctx context.Context, db *gorm.DB, cfg *AISStreamConfig) {
	backoff := 2 * time.Second
	for {
		err := runAISStreamSession(ctx, db, cfg, func() {
			backoff = 2 * time.Second
		})
		if err == nil {
			return
		}
		slog.Error("p_seer_aisstream: stream session ended", "error", err, "backoff", backoff.String())
		select {
		case <-ctx.Done():
			slog.Info("p_seer_aisstream: stream worker stopped")
			return
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > time.Minute {
			backoff = time.Minute
		}
	}
}

func runAISStreamSession(ctx context.Context, db *gorm.DB, cfg *AISStreamConfig, onConnected func()) error {
	if cfg == nil {
		return nil
	}
	dialer := websocket.Dialer{
		Proxy:            websocket.DefaultDialer.Proxy,
		HandshakeTimeout: 20 * time.Second,
	}
	ws, resp, err := dialer.DialContext(ctx, cfg.StreamURL, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("dial %s: status %s: %w", cfg.StreamURL, resp.Status, err)
		}
		return fmt.Errorf("dial %s: %w", cfg.StreamURL, err)
	}
	defer ws.Close()

	sub := aisstream.SubscriptionMessage{
		APIKey:        cfg.APIKey,
		BoundingBoxes: [][][]float64{{{-90.0, -180.0}, {90.0, 180.0}}},
	}
	b, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	if onConnected != nil {
		onConnected()
	}
	slog.Info("p_seer_aisstream: stream worker connected", "url", cfg.StreamURL)

	sessionCtx, cancelSession := context.WithCancel(ctx)
	defer cancelSession()

	var packets []aisstream.AisStreamMessage
	packetsLock := sync.Mutex{}
	var droppedPackets uint64
	refreshInterval := Config.MapRefreshEvery()
	if refreshInterval == 0 {
		return fmt.Errorf("refreshInterval set to 0")
	}
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-sessionCtx.Done():
				return
			case <-ticker.C:
			}
			packetsLock.Lock()
			localPackets := slices.Clone(packets)
			packets = packets[0:0]
			packetsLock.Unlock()
			if err := ingestAISStreamPackets(ctx, db, localPackets); err != nil {
				slog.Error("Error while ingesting aisstream packets", "error", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		var packet aisstream.AisStreamMessage
		if err := ws.ReadJSON(&packet); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("p_seer_aisstream: read message", "error", err)
			return fmt.Errorf("read message: %w", err)
		}
		packetsLock.Lock()
		if len(packets) >= maxBufferedAISPackets {
			dropped := atomic.AddUint64(&droppedPackets, 1)
			if dropped == 1 || dropped%100 == 0 {
				slog.Warn("p_seer_aisstream: dropping packet because buffer is full", "buffer_limit", maxBufferedAISPackets, "dropped", dropped)
			}
			packetsLock.Unlock()
			continue
		}
		packets = append(packets, packet)
		packetsLock.Unlock()
	}
}
