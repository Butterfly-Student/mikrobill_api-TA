package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mikrotik_adapter "MikrOps/internal/adapter/outbound/mikrotik"
	"MikrOps/internal/model"
	contextutil "MikrOps/utils/context"

	"github.com/palantir/stacktrace"
	log "github.com/sirupsen/logrus"
)

// Independent state for direct monitors
var (
	activeDirectMonitors = make(map[string]*DirectMonitor)
	directMu             sync.Mutex
	directMonitorLocks   = make(map[string]*sync.Mutex)
	directLocksMu        sync.Mutex
)

// DirectMonitor represents a monitored interface session (Direct Access)
type DirectMonitor struct {
	InterfaceName string
	Cancel        context.CancelFunc
	Clients       int
	Observers     map[chan model.CustomerTrafficData]bool
	restartCount  int
}

// StreamTrafficByInterface monitors traffic directly on an interface
func (d *monitorDomain) StreamTrafficByInterface(ctx context.Context, interfaceName string) (<-chan model.CustomerTrafficData, error) {
	// 1. Get lock for this interface
	directLocksMu.Lock()
	if _, ok := directMonitorLocks[interfaceName]; !ok {
		directMonitorLocks[interfaceName] = &sync.Mutex{}
	}
	lock := directMonitorLocks[interfaceName]
	directLocksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	directMu.Lock()
	monitor, exists := activeDirectMonitors[interfaceName]
	if exists {
		// Already monitoring, increment client count
		monitor.Clients++
		directMu.Unlock()
		log.Printf("[DirectMonitor] Interface %s: Client count incremented to %d", interfaceName, monitor.Clients)
		return d.addDirectObserver(ctx, interfaceName)
	}
	directMu.Unlock()

	// 2. Start new monitor
	// Extract tenant info for context
	tenantID, _ := contextutil.GetTenantID(ctx)
	user, _ := contextutil.GetUser(ctx)
	isSuper := contextutil.IsSuperAdmin(ctx)

	monitorCtx := contextutil.WithTenantContext(context.Background(), tenantID, user, isSuper)
	monitorCtx, cancel := context.WithCancel(monitorCtx)

	monitor = &DirectMonitor{
		InterfaceName: interfaceName,
		Cancel:        cancel,
		Clients:       1,
		Observers:     make(map[chan model.CustomerTrafficData]bool),
		restartCount:  0,
	}

	directMu.Lock()
	activeDirectMonitors[interfaceName] = monitor
	directMu.Unlock()

	// Start background loop
	go d.runDirectMonitorLoop(monitorCtx, interfaceName)

	log.Printf("[DirectMonitor] Started monitoring for interface %s", interfaceName)

	return d.addDirectObserver(ctx, interfaceName)
}

func (d *monitorDomain) StopDirectMonitoring(interfaceName string) {
	directLocksMu.Lock()
	if _, ok := directMonitorLocks[interfaceName]; !ok {
		directLocksMu.Unlock()
		return
	}
	lock := directMonitorLocks[interfaceName]
	directLocksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	directMu.Lock()
	monitor, exists := activeDirectMonitors[interfaceName]
	if !exists {
		directMu.Unlock()
		return
	}

	monitor.Clients--
	log.Printf("[DirectMonitor] Interface %s: Client count decremented to %d", interfaceName, monitor.Clients)

	if monitor.Clients <= 0 {
		monitor.Cancel()
		delete(activeDirectMonitors, interfaceName)
		for ch := range monitor.Observers {
			close(ch)
		}
		log.Printf("[DirectMonitor] Stopped monitoring for interface %s", interfaceName)
	}
	directMu.Unlock()
}

func (d *monitorDomain) runDirectMonitorLoop(ctx context.Context, interfaceName string) {
	maxRestarts := 3
	restartDelay := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
			activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
			if err != nil || activeMikrotik == nil {
				time.Sleep(restartDelay)
				continue
			}

			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				time.Sleep(restartDelay)
				continue
			}

			// Check if still active
			directMu.Lock()
			monitor, exists := activeDirectMonitors[interfaceName]
			if !exists {
				client.Close()
				directMu.Unlock()
				return
			}
			if monitor.restartCount >= maxRestarts {
				monitor.Cancel()
				delete(activeDirectMonitors, interfaceName)
				for ch := range monitor.Observers {
					close(ch)
				}
				directMu.Unlock()
				client.Close()
				return
			}
			directMu.Unlock()

			concreteClient, ok := client.(*mikrotik_adapter.Client)
			if !ok {
				client.Close()
				return
			}

			trafficChan, err := mikrotik_adapter.MonitorTraffic(ctx, concreteClient, interfaceName)
			if err != nil {
				directMu.Lock()
				if mon, ok := activeDirectMonitors[interfaceName]; ok {
					mon.restartCount++
				}
				directMu.Unlock()
				client.Close()
				time.Sleep(restartDelay)
				continue
			}

			// Process stream
			streamClosed := d.processDirectTrafficStream(ctx, interfaceName, trafficChan)
			client.Close()

			if streamClosed {
				directMu.Lock()
				if mon, ok := activeDirectMonitors[interfaceName]; ok {
					mon.restartCount++
				}
				directMu.Unlock()
				time.Sleep(restartDelay)
			}
		}
	}
}

func (d *monitorDomain) processDirectTrafficStream(ctx context.Context, interfaceName string, trafficChan <-chan model.InterfaceTraffic) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case traffic, ok := <-trafficChan:
			if !ok {
				return true
			}
			if traffic.Name == "" {
				continue
			}

			data := model.CustomerTrafficData{
				CustomerID:         "direct:" + interfaceName, // Mock ID for direct access
				CustomerName:       interfaceName,
				Username:           "direct",
				ServiceType:        "direct",
				InterfaceName:      traffic.Name,
				RxBitsPerSecond:    traffic.RxBitsPerSecond,
				TxBitsPerSecond:    traffic.TxBitsPerSecond,
				RxPacketsPerSecond: traffic.RxPacketsPerSecond,
				TxPacketsPerSecond: traffic.TxPacketsPerSecond,
				DownloadSpeed:      formatSpeed(traffic.RxBitsPerSecond),
				UploadSpeed:        formatSpeed(traffic.TxBitsPerSecond),
				Timestamp:          time.Now(),
			}

			d.publishDirectTrafficData(data)
		}
	}
}

func (d *monitorDomain) addDirectObserver(ctx context.Context, interfaceName string) (<-chan model.CustomerTrafficData, error) {
	directMu.Lock()
	monitor, exists := activeDirectMonitors[interfaceName]
	if !exists {
		directMu.Unlock()
		return nil, fmt.Errorf("monitor not running for interface %s", interfaceName)
	}

	ch := make(chan model.CustomerTrafficData, 50)
	monitor.Observers[ch] = true
	directMu.Unlock()

	go func() {
		<-ctx.Done()
		d.StopDirectMonitoring(interfaceName)
	}()

	return ch, nil
}

func (d *monitorDomain) publishDirectTrafficData(data model.CustomerTrafficData) {
	// Optional Redis pub
	jsonData, _ := json.Marshal(data)
	_ = d.cachePort.PubSub().Publish("mikrotik:traffic:direct", string(jsonData))

	directMu.Lock()
	defer directMu.Unlock()

	if monitor, ok := activeDirectMonitors[data.InterfaceName]; ok {
		for ch := range monitor.Observers {
			select {
			case ch <- data:
			default:
			}
		}
	}
}

// PingHost pings a target IP directly
func (d *monitorDomain) PingHost(ctx context.Context, targetIP string) (map[string]interface{}, error) {
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
	}
	defer client.Close()

	reply, err := client.RunArgs("/ping", map[string]string{
		"address": targetIP,
		"count":   "3",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to ping target")
	}

	m := reply.Done.Map
	isReachable := false
	loss := m["packet-loss"]
	if loss != "100" && loss != "" {
		isReachable = true
	}

	return map[string]interface{}{
		"target":       targetIP,
		"sent":         m["sent"],
		"received":     m["received"],
		"packet_loss":  loss + "%",
		"avg_time":     m["avg-rtt"],
		"min_time":     m["min-rtt"],
		"max_time":     m["max-rtt"],
		"is_reachable": isReachable,
	}, nil
}

// StreamPingHost streams ping results for a target IP
func (d *monitorDomain) StreamPingHost(ctx context.Context, targetIP string) (<-chan model.PingResponse, error) {
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
	}

	concreteClient, ok := client.(*mikrotik_adapter.Client)
	if !ok {
		client.Close()
		return nil, fmt.Errorf("client is not of type *mikrotik_adapter.Client")
	}

	outChan, err := concreteClient.StreamPing(ctx, targetIP, "56", "1")
	if err != nil {
		client.Close()
		return nil, err
	}

	wrapperChan := make(chan model.PingResponse)
	go func() {
		defer client.Close()
		defer close(wrapperChan)
		for resp := range outChan {
			wrapperChan <- resp
		}
	}()

	return wrapperChan, nil
}
