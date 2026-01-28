package log

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	mikrotik_outbound_adapter "MikrOps/internal/adapter/outbound/mikrotik"
	"MikrOps/internal/model"
	contextutil "MikrOps/utils/context"
)

// Global state for log monitor
var (
	activeLogMonitor *LogMonitor
	logMu            sync.Mutex
	logLock          sync.Mutex
)

// LogMonitor represents a global log monitoring session
type LogMonitor struct {
	Cancel       context.CancelFunc
	Clients      int
	Observers    map[chan model.LogStreamData]bool
	restartCount int
}

func (d *LogDomain) StreamLogs(ctx context.Context) (<-chan model.LogStreamData, error) {
	logLock.Lock()
	defer logLock.Unlock()

	logMu.Lock()
	if activeLogMonitor != nil {
		// Already monitoring, just increment client count
		activeLogMonitor.Clients++
		logMu.Unlock()
		log.Printf("[LogMonitor] Client count incremented to %d", activeLogMonitor.Clients)
		return d.addObserver(ctx)
	}
	logMu.Unlock()

	// Extract tenant info for background loop
	tenantID, _ := contextutil.GetTenantID(ctx)
	user, _ := contextutil.GetUser(ctx)
	isSuper := contextutil.IsSuperAdmin(ctx)

	// Create monitor context - independent of request context
	monitorCtx := contextutil.WithTenantContext(context.Background(), tenantID, user, isSuper)
	monitorCtx, cancel := context.WithCancel(monitorCtx)

	activeLogMonitor = &LogMonitor{
		Cancel:       cancel,
		Clients:      1,
		Observers:    make(map[chan model.LogStreamData]bool),
		restartCount: 0,
	}

	logMu.Lock()
	logMu.Unlock()

	// Start background monitoring
	go d.runLogMonitorLoop(monitorCtx)

	log.Printf("[LogMonitor] Started log monitoring")

	return d.addObserver(ctx)
}

func (d *LogDomain) StopLogStream() {
	logLock.Lock()
	defer logLock.Unlock()

	logMu.Lock()
	if activeLogMonitor == nil {
		logMu.Unlock()
		return
	}

	activeLogMonitor.Clients--
	log.Printf("[LogMonitor] Client count decremented to %d", activeLogMonitor.Clients)

	if activeLogMonitor.Clients <= 0 {
		// No more clients, stop monitoring
		activeLogMonitor.Cancel()

		// Close all observers
		for ch := range activeLogMonitor.Observers {
			close(ch)
		}

		activeLogMonitor = nil
		log.Printf("[LogMonitor] Stopped log monitoring")
	}
	logMu.Unlock()
}

func (d *LogDomain) runLogMonitorLoop(ctx context.Context) {
	maxRestarts := 3
	restartDelay := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Printf("[LogMonitor] Monitor context cancelled")
			return
		default:
			// Check active mikrotik
			activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
			if err != nil || activeMikrotik == nil {
				log.Errorf("[LogMonitor] Failed to get active mikrotik: %v", err)
				time.Sleep(restartDelay)
				continue
			}

			// Create client
			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				log.Errorf("[LogMonitor] Failed to create mikrotik client: %v", err)
				time.Sleep(restartDelay)
				continue
			}

			// Check restart count
			logMu.Lock()
			if activeLogMonitor == nil {
				client.Close()
				logMu.Unlock()
				return
			}

			if activeLogMonitor.restartCount >= maxRestarts {
				log.Printf("[LogMonitor] Max restart attempts (%d) reached, stopping monitor", maxRestarts)
				activeLogMonitor.Cancel()

				// Close all observers
				for ch := range activeLogMonitor.Observers {
					close(ch)
				}

				activeLogMonitor = nil
				logMu.Unlock()
				client.Close()
				return
			}
			logMu.Unlock()

			// Start log stream
			concreteClient, ok := client.(*mikrotik_outbound_adapter.Client)
			if !ok {
				log.Errorf("[LogMonitor] Client is not of type *mikrotik_outbound_adapter.Client")
				client.Close()
				return
			}

			logChan, err := mikrotik_outbound_adapter.MonitorLogs(ctx, concreteClient)
			if err != nil {
				log.Printf("[LogMonitor] Failed to start log monitoring: %v", err)

				logMu.Lock()
				if activeLogMonitor != nil {
					activeLogMonitor.restartCount++
				}
				logMu.Unlock()

				client.Close()
				time.Sleep(restartDelay)
				continue
			}

			log.Printf("[LogMonitor] Log stream active")

			// Process log stream
			streamClosed := d.processLogStream(ctx, logChan)

			client.Close()

			if streamClosed {
				log.Printf("[LogMonitor] Stream closed, attempting restart...")

				logMu.Lock()
				if activeLogMonitor != nil {
					activeLogMonitor.restartCount++
				}
				logMu.Unlock()

				time.Sleep(restartDelay)
			}
		}
	}
}

func (d *LogDomain) processLogStream(ctx context.Context, logChan <-chan model.MikrotikLog) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case logEntry, ok := <-logChan:
			if !ok {
				log.Printf("[LogMonitor] Log channel closed")
				return true // Stream closed
			}

			// Validate log data
			if logEntry.Message == "" {
				continue
			}

			data := model.LogStreamData{
				Timestamp: time.Now(),
				Log:       logEntry,
			}

			d.publishLogData(data)
		}
	}
}

func (d *LogDomain) addObserver(ctx context.Context) (<-chan model.LogStreamData, error) {
	logMu.Lock()
	if activeLogMonitor == nil {
		logMu.Unlock()
		return nil, fmt.Errorf("log monitor not running")
	}

	// Create buffered channel
	ch := make(chan model.LogStreamData, 50)
	activeLogMonitor.Observers[ch] = true
	logMu.Unlock()

	// Cleanup routine
	go func() {
		<-ctx.Done()
		d.StopLogStream()
	}()

	return ch, nil
}

func (d *LogDomain) publishLogData(data model.LogStreamData) {
	// 1. Publish to Redis
	jsonData, _ := json.Marshal(data)
	_ = d.cachePort.PubSub().Publish("mikrotik:logs:stream", string(jsonData))

	// 2. Broadcast to in-memory observers
	logMu.Lock()
	defer logMu.Unlock()

	if activeLogMonitor != nil {
		for ch := range activeLogMonitor.Observers {
			select {
			case ch <- data:
			default:
				// Skip if channel full
			}
		}
	}
}
