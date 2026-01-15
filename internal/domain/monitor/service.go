package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	log "github.com/sirupsen/logrus"

	mikrotik_adapter "prabogo/internal/adapter/outbound/mikrotik"
	"prabogo/internal/model"
	contextutil "prabogo/utils/context"
)

// Global state for active monitors
var (
	activeMonitors = make(map[string]*CustomerMonitor)
	mu             sync.Mutex
	monitorLocks   = make(map[string]*sync.Mutex)
	locksMu        sync.Mutex
)

// CustomerMonitor represents a monitored customer session
type CustomerMonitor struct {
	CustomerID    string
	InterfaceName string
	Cancel        context.CancelFunc
	Clients       int
	Observers     map[chan model.CustomerTrafficData]bool
	restartCount  int // Track restart attempts
}

func (d *monitorDomain) StreamTraffic(ctx context.Context, customerID string) (<-chan model.CustomerTrafficData, error) {
	// 1. Get lock for this customer to prevent race conditions
	locksMu.Lock()
	if _, ok := monitorLocks[customerID]; !ok {
		monitorLocks[customerID] = &sync.Mutex{}
	}
	lock := monitorLocks[customerID]
	locksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	mu.Lock()
	custMonitor, exists := activeMonitors[customerID]
	if exists {
		// Already monitoring, just increment client count
		custMonitor.Clients++
		mu.Unlock()
		log.Printf("[OnDemand] Customer %s: Client count incremented to %d", customerID, custMonitor.Clients)

		// Subscribe to existing monitor
		return d.addObserver(ctx, customerID)
	}
	mu.Unlock()

	// 2. Not monitoring yet, need to start.
	// Get customer details first
	id, err := uuid.Parse(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	customer, err := d.databasePort.Customer().GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "customer not found: %s", customerID)
	}

	// Use the interface from the customer record directly
	if customer.Interface == nil || *customer.Interface == "" {
		return nil, fmt.Errorf("customer has no interface assigned")
	}
	interfaceName := *customer.Interface

	// Extract tenant info for background loop
	tenantID, _ := contextutil.GetTenantID(ctx)
	user, _ := contextutil.GetUser(ctx)
	isSuper := contextutil.IsSuperAdmin(ctx)

	// Create monitor context - independent of the request context but with same tenant info
	monitorCtx := contextutil.WithTenantContext(context.Background(), tenantID, user, isSuper)
	monitorCtx, cancel := context.WithCancel(monitorCtx)

	custMonitor = &CustomerMonitor{
		CustomerID:    customerID,
		InterfaceName: interfaceName,
		Cancel:        cancel,
		Clients:       1,
		Observers:     make(map[chan model.CustomerTrafficData]bool),
		restartCount:  0,
	}

	mu.Lock()
	activeMonitors[customerID] = custMonitor
	mu.Unlock()

	// Start the actual background monitoring for this customer
	go d.runMonitorLoop(monitorCtx, customer.ID.String(), customer.Name, customer.Username, customer.ServiceType, interfaceName)

	log.Printf("[OnDemand] Started monitoring for customer %s (%s) on interface %s",
		customer.Name, customer.Username, interfaceName)

	return d.addObserver(ctx, customerID)
}

func (d *monitorDomain) StopMonitoring(customerID string) {
	locksMu.Lock()
	if _, ok := monitorLocks[customerID]; !ok {
		locksMu.Unlock()
		return
	}
	lock := monitorLocks[customerID]
	locksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	mu.Lock()
	custMonitor, exists := activeMonitors[customerID]
	if !exists {
		mu.Unlock()
		return
	}

	custMonitor.Clients--
	log.Printf("[OnDemand] Customer %s: Client count decremented to %d", customerID, custMonitor.Clients)

	if custMonitor.Clients <= 0 {
		// No more clients, stop monitoring
		custMonitor.Cancel()
		delete(activeMonitors, customerID)

		// Close all observers
		for ch := range custMonitor.Observers {
			close(ch)
		}

		log.Printf("[OnDemand] Stopped monitoring for customer %s", customerID)
	}
	mu.Unlock()
}

// runMonitorLoop runs the actual MikroTik monitoring command with auto-restart
func (d *monitorDomain) runMonitorLoop(ctx context.Context, customerID, name, username string, serviceType model.ServiceType, interfaceName string) {
	maxRestarts := 3
	restartDelay := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Printf("[OnDemand] Monitor context cancelled for %s", name)
			return
		default:
			// check active mikrotik first
			activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
			if err != nil || activeMikrotik == nil {
				log.Errorf("failed to get active mikrotik for monitor: %v", err)
				time.Sleep(restartDelay)
				continue
			}

			// Create dedicated client
			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				log.Errorf("failed to create mikrotik client for monitor: %v", err)
				time.Sleep(restartDelay)
				continue
			}
			// Important: We manage client lifecycle here.
			// Ideally StartMonitoring should accept a client, but since we restart, we create new ones.

			// Check restart count
			mu.Lock()
			custMonitor, exists := activeMonitors[customerID]
			if !exists {
				client.Close()
				mu.Unlock()
				return
			}

			if custMonitor.restartCount >= maxRestarts {
				log.Printf("[OnDemand] Max restart attempts (%d) reached for %s, stopping monitor",
					maxRestarts, name)
				custMonitor.Cancel()
				delete(activeMonitors, customerID)

				// Notify observers about disconnection
				for ch := range custMonitor.Observers {
					select {
					case ch <- model.CustomerTrafficData{
						CustomerID:   customerID,
						CustomerName: name,
						Timestamp:    time.Now(),
						// Indicate finished?
					}:
					default:
					}
					close(ch)
				}
				mu.Unlock()
				client.Close()
				return
			}
			mu.Unlock()

			// Start monitoring stream
			concreteClient, ok := client.(*mikrotik_adapter.Client)
			if !ok {
				log.Errorf("client is not of type *mikrotik_adapter.Client")
				client.Close()
				return
			}

			trafficChan, err := mikrotik_adapter.MonitorTraffic(ctx, concreteClient, interfaceName)
			if err != nil {
				log.Printf("[OnDemand] Failed to start monitor for %s on %s: %v",
					name, interfaceName, err)

				mu.Lock()
				if mon, ok := activeMonitors[customerID]; ok {
					mon.restartCount++
				}
				mu.Unlock()

				client.Close()
				time.Sleep(restartDelay)
				continue
			}

			log.Printf("[OnDemand] Monitor stream active for %s on %s", name, interfaceName)

			// Process traffic data
			// This blocks until stream closes or context done
			streamClosed := d.processTrafficStream(ctx, customerID, name, username, serviceType, trafficChan)

			client.Close()

			if streamClosed {
				log.Printf("[OnDemand] Stream closed for %s, attempting restart...", name)

				mu.Lock()
				if mon, ok := activeMonitors[customerID]; ok {
					mon.restartCount++
				}
				mu.Unlock()

				time.Sleep(restartDelay)
			}
		}
	}
}

// processTrafficStream processes traffic data from the stream
func (d *monitorDomain) processTrafficStream(
	ctx context.Context,
	customerID, name, username string, serviceType model.ServiceType,
	trafficChan <-chan model.InterfaceTraffic,
) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case traffic, ok := <-trafficChan:
			if !ok {
				log.Printf("[OnDemand] Traffic channel closed for %s", name)
				return true // Stream closed
			}

			// Validate traffic data
			if traffic.Name == "" {
				// Sometimes initial data is empty
				continue
			}

			data := model.CustomerTrafficData{
				CustomerID:         customerID,
				CustomerName:       name,
				Username:           username,
				ServiceType:        serviceType,
				InterfaceName:      traffic.Name,
				RxBitsPerSecond:    traffic.RxBitsPerSecond,
				TxBitsPerSecond:    traffic.TxBitsPerSecond,
				RxPacketsPerSecond: traffic.RxPacketsPerSecond,
				TxPacketsPerSecond: traffic.TxPacketsPerSecond,
				DownloadSpeed:      formatSpeed(traffic.RxBitsPerSecond),
				UploadSpeed:        formatSpeed(traffic.TxBitsPerSecond),
				Timestamp:          time.Now(),
			}

			d.publishTrafficData(data)
		}
	}
}

// addObserver creates a channel and adds it to the monitor's observers
func (d *monitorDomain) addObserver(ctx context.Context, customerID string) (<-chan model.CustomerTrafficData, error) {
	mu.Lock()
	custMonitor, exists := activeMonitors[customerID]
	if !exists {
		mu.Unlock()
		return nil, fmt.Errorf("monitor not running for customer %s", customerID)
	}

	// Create buffered channel to prevent blocking
	ch := make(chan model.CustomerTrafficData, 50)
	custMonitor.Observers[ch] = true
	mu.Unlock()

	// Cleanup routine: remove observer when context is done
	go func() {
		<-ctx.Done()
		d.StopMonitoring(customerID) // Reuse StopMonitoring logic to decrement client count!
	}()

	return ch, nil
}

func (d *monitorDomain) publishTrafficData(data model.CustomerTrafficData) {
	// 1. Publish to Redis (optional, for history/other consumers)
	jsonData, _ := json.Marshal(data)
	// Ignore error on publish for now as it's fire-and-forget
	_ = d.cachePort.PubSub().Publish("mikrotik:traffic:customers", string(jsonData))

	// 2. Broadcast to in-memory observers (active websockets)
	mu.Lock()
	defer mu.Unlock()

	if custMonitor, ok := activeMonitors[data.CustomerID]; ok {
		for ch := range custMonitor.Observers {
			select {
			case ch <- data:
			default:
				// Skip if channel full to prevent blocking
			}
		}
	}
}

func formatSpeed(bps string) string {
	if bps == "" || bps == "0" {
		return "0 bps"
	}
	return bps + " bps"
}

func (d *monitorDomain) PingCustomer(ctx context.Context, customerID string) (map[string]interface{}, error) { // Parse/Validate Customer ID
	id, err := uuid.Parse(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	customer, err := d.databasePort.Customer().GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "customer not found: %s", customerID)
	}

	target, err := d.getCustomerIPAddress(&customer.Customer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to resolve customer ip")
	}

	// Get active mikrotik
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

	// Run Ping
	// /ping address=1.2.3.4 count=3
	reply, err := client.RunArgs("/ping", map[string]string{
		"address": target,
		"count":   "3",
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to ping customer")
	}

	// Parse reply
	m := reply.Done.Map

	// Default to unreachable if packet-loss is 100 or missing
	isReachable := false
	loss := m["packet-loss"]
	if loss != "100" && loss != "" {
		isReachable = true
	}

	stats := map[string]interface{}{
		"target":       target,
		"sent":         m["sent"],
		"received":     m["received"],
		"packet_loss":  loss + "%",
		"avg_time":     m["avg-rtt"],
		"min_time":     m["min-rtt"],
		"max_time":     m["max-rtt"],
		"is_reachable": isReachable,
	}

	return stats, nil
}

func (d *monitorDomain) StreamPing(ctx context.Context, customerID string) (<-chan model.PingResponse, error) {
	id, err := uuid.Parse(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	customer, err := d.databasePort.Customer().GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "customer not found: %s", customerID)
	}

	target, err := d.getCustomerIPAddress(&customer.Customer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to resolve customer ip")
	}

	// Get active mikrotik
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
	// Client lifecycle: StreamPing runs in background?
	// The client helper `StreamPing` uses `ListenArgsContext`.
	// We need to keep client open while streaming.
	// But `StreamPing` helper returns channel and runs goroutine.
	// The helper should close the client when done?
	// `mikrotik_adapter.StreamPing` doesn't seem to take ownership of closing client?
	// Let's check `ping.go`. It has `go func() ...` but no `defer client.Close()`.
	// Ideally `StreamPing` helper should probably manage it, or we manage it here.
	// If the helper uses `ListenArgsContext` which calls `Listen`, the connection is used.
	// If we close client here, `Listen` might fail?
	// Actually `StreamPing` in `ping.go` uses `ListenArgsContext`.
	// The `mikrotik_adapter` should probably handle `defer client.Close()` inside the goroutine if it created it?
	// Or we pass client and we are responsible.
	// `client.go` `NewClient` returns a client.

	concreteClient, ok := client.(*mikrotik_adapter.Client)
	if !ok {
		client.Close()
		return nil, fmt.Errorf("client is not of type *mikrotik_adapter.Client")
	}

	// We'll trust that we can't close client here immediately because StreamPing needs it.
	// We need to wrap the output channel to close the client when the channel closes.

	outChan, err := concreteClient.StreamPing(ctx, target, "56", "1")
	if err != nil {
		client.Close()
		return nil, err
	}

	// Create a wrapper channel to manage client closure
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

// getCustomerIPAddress extracts IP address based on service type
func (d *monitorDomain) getCustomerIPAddress(customer *model.Customer) (string, error) {
	// Adapting user logic to our model
	// User logic:
	/*
		switch customer.ServiceType {
		case "pppoe":
			if customer.AssignedIP != nil && *customer.AssignedIP != "" { return *customer.AssignedIP, nil }
			if customer.StaticIP != nil && *customer.StaticIP != "" { return *customer.StaticIP, nil }
			return "", fmt.Errorf("pppoe customer has no assigned IP")
		case "hotspot": ...
		case "static_ip": ...
	*/

	// Our model uses Enums for ServiceType?
	// model/customer.go -> ServiceType is string (likely).

	switch string(customer.ServiceType) {
	case "pppoe":
		if customer.AssignedIP != nil && *customer.AssignedIP != "" {
			return *customer.AssignedIP, nil
		}
		// Assuming StaticIP field exists in model? Let's check model/customer.go
		// If not, revert to just AssignedIP
		// User said: "Untuk ping itu berdasarkan assigned_ip pada table customers"
		// But in code provided: "if customer.StaticIP != nil"
		// I'll stick to AssignedIP if StaticIP missing.
		return "", fmt.Errorf("pppoe customer has no assigned IP")

	case "hotspot":
		if customer.AssignedIP != nil && *customer.AssignedIP != "" {
			return *customer.AssignedIP, nil
		}
		return "", fmt.Errorf("hotspot customer has no assigned IP")

	case "static_ip":
		// Check AssignedIP first as unified field? Or specific StaticIP field?
		// Existing schema usually puts static ip in assigned_ip or separate?
		// Checking previous files/migration...
		// `customers` table has `assigned_ip`.
		// If `static_ip` service, `assigned_ip` should be populated.
		if customer.AssignedIP != nil && *customer.AssignedIP != "" {
			return *customer.AssignedIP, nil
		}
		return "", fmt.Errorf("static IP not configured")

	default:
		// Fallback
		if customer.AssignedIP != nil && *customer.AssignedIP != "" {
			return *customer.AssignedIP, nil
		}
		return "", fmt.Errorf("unsupported service type or no IP: %s", customer.ServiceType)
	}
}
