package listen

import (
	"context"
	"sync"
	"time"

	"mikrobill/internal/infrastructure/mikrotik"
	pkg_logger "mikrobill/pkg/logger"

	"go.uber.org/zap"
)

/* =========================
   EVENT MODEL
========================= */

type EventType string

const (
	EventCreate  EventType = "CREATE"
	EventUpdate  EventType = "UPDATE"
	EventDelete  EventType = "DELETE"
	EventEnable  EventType = "ENABLE"
	EventDisable EventType = "DISABLE"
)

type ListenEvent struct {
	Command string
	ID      string
	Type    EventType
	Data    map[string]string
	Time    time.Time
}

/* =========================
   LISTENER CONFIG
========================= */

type ListenerConfig struct {
	Command  interface{}                             // string atau []string
	EventCh  chan<- ListenEvent                      // channel untuk event
	State    map[string]map[string]map[string]string // state storage
	Mu       *sync.Mutex                             // mutex untuk state
	QueueLen int                                     // queue length untuk listener
}

/* =========================
   LISTENER
========================= */

// listenCommand menerima command sebagai string atau []string
func listenCommand(
	ctx context.Context,
	client *mikrotik.Client,
	cfg ListenerConfig,
) {
	// Konversi command ke []string
	var cmdArgs []string
	var cmdKey string

	switch v := cfg.Command.(type) {
	case string:
		cmdArgs = []string{v}
		cmdKey = v
	case []string:
		if len(v) == 0 {
			pkg_logger.Error("empty command array")
			return
		}
		cmdArgs = v
		cmdKey = v[0] // gunakan command pertama sebagai key
	default:
		pkg_logger.Error("invalid command type",
			zap.Any("type", v),
		)
		return
	}

	// Pastikan state untuk command ini ada
	cfg.Mu.Lock()
	if _, exists := cfg.State[cmdKey]; !exists {
		cfg.State[cmdKey] = make(map[string]map[string]string)
	}
	cfg.Mu.Unlock()

	// Mulai listening
	reply, err := client.ListenArgsQueueContext(
		ctx,
		cmdArgs,
		cfg.QueueLen,
	)
	if err != nil {
		pkg_logger.Error("listen failed",
			zap.String("cmd", cmdKey),
			zap.Error(err),
		)
		return
	}

	pkg_logger.Info("listening started",
		zap.String("cmd", cmdKey),
		zap.Strings("args", cmdArgs),
	)

	for r := range reply.Chan() {
		if r == nil || r.Map == nil {
			continue
		}

		id := r.Map[".id"]
		if id == "" {
			continue
		}

		dead := r.Map[".dead"] == "true"
		disabled := r.Map["disabled"]

		cfg.Mu.Lock()
		prev, exists := cfg.State[cmdKey][id]
		prevDisabled := ""
		if exists {
			prevDisabled = prev["disabled"]
		}

		var eventType EventType

		switch {
		case cmdKey == "/log/listen":
			eventType = EventCreate

		case dead:
			eventType = EventDelete
			delete(cfg.State[cmdKey], id)

		case !exists:
			eventType = EventCreate

		case prevDisabled == "false" && disabled == "true":
			eventType = EventDisable

		case prevDisabled == "true" && disabled == "false":
			eventType = EventEnable

		default:
			eventType = EventUpdate
		}

		if eventType != EventDelete {
			snapshot := make(map[string]string, len(r.Map))
			for k, v := range r.Map {
				snapshot[k] = v
			}
			cfg.State[cmdKey][id] = snapshot
		}

		cfg.Mu.Unlock()

		cfg.EventCh <- ListenEvent{
			Command: cmdKey,
			ID:      id,
			Type:    eventType,
			Data:    r.Map,
			Time:    time.Now(),
		}
	}

	pkg_logger.Info("listening stopped",
		zap.String("cmd", cmdKey),
	)
}

/* =========================
   HELPER FUNCTIONS
========================= */

// StartListener - helper function untuk memulai listener dengan lebih mudah
func StartListener(
	ctx context.Context,
	client *mikrotik.Client,
	command interface{},
	eventCh chan<- ListenEvent,
	state map[string]map[string]map[string]string,
	mu *sync.Mutex,
) {
	cfg := ListenerConfig{
		Command:  command,
		EventCh:  eventCh,
		State:    state,
		Mu:       mu,
		QueueLen: 100,
	}
	go listenCommand(ctx, client, cfg)
}

// StartListenerWithConfig - memulai listener dengan config lengkap
func StartListenerWithConfig(
	ctx context.Context,
	client *mikrotik.Client,
	cfg ListenerConfig,
) {
	// Set default queue length jika tidak diisi
	if cfg.QueueLen == 0 {
		cfg.QueueLen = 100
	}
	go listenCommand(ctx, client, cfg)
}

/* =========================
   CONTOH PENGGUNAAN
========================= */

// Contoh 1: Single command string
// StartListener(ctx, client, "/interface/listen", eventCh, state, &mu)
//
// Contoh 2: Command dengan arguments
// StartListener(ctx, client, []string{"/log/listen", "follow=yes"}, eventCh, state, &mu)
//
// Contoh 3: Command dengan filter
// StartListener(ctx, client, []string{"/interface/listen", "where=type=ether"}, eventCh, state, &mu)
//
// Contoh 4: Dengan config lengkap dan custom queue size
// cfg := ListenerConfig{
//     Command:  []string{"/queue/simple/listen"},
//     EventCh:  eventCh,
//     State:    state,
//     Mu:       &mu,
//     QueueLen: 200,
// }
// StartListenerWithConfig(ctx, client, cfg)
//
// Contoh 5: Multiple listeners
// commands := []interface{}{
//     "/interface/listen",
//     "/ip/address/listen",
//     []string{"/log/listen", "follow=yes"},
// }
// for _, cmd := range commands {
//     StartListener(ctx, client, cmd, eventCh, state, &mu)
// }
