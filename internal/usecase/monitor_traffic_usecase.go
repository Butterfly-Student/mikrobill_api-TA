package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mikrobill/internal/entity"
	"mikrobill/internal/infrastructure/mikrotik"
	mon "mikrobill/internal/infrastructure/mikrotik/monitor"
	"sync"
	"time"
)

// OnDemandTrafficService monitors traffic only for requested customers
type OnDemandTrafficService struct {
	client    *mikrotik.Client
	db        entity.CustomerRepository
	publisher entity.RedisPublisher

	// Active monitors: key = customerID, value = monitor context
	activeMonitors map[string]*CustomerMonitor
	mu             sync.Mutex

	// Lock for preventing duplicate start/stops per customer
	monitorLocks map[string]*sync.Mutex
	locksMu      sync.Mutex
}

// CustomerMonitor represents a monitored customer session
type CustomerMonitor struct {
	CustomerID    string
	InterfaceName string
	Cancel        context.CancelFunc
	Clients       int
	Observers     map[chan entity.CustomerTrafficData]bool
	restartCount  int // Track restart attempts
}

// NewOnDemandTrafficService creates a new on-demand traffic service
func NewOnDemandTrafficService(
	client *mikrotik.Client,
	db entity.CustomerRepository,
	publisher entity.RedisPublisher,
) *OnDemandTrafficService {
	return &OnDemandTrafficService{
		client:         client,
		db:             db,
		publisher:      publisher,
		activeMonitors: make(map[string]*CustomerMonitor),
		monitorLocks:   make(map[string]*sync.Mutex),
	}
}

// StartMonitoring starts monitoring a specific customer if not already started
func (s *OnDemandTrafficService) StartMonitoring(ctx context.Context, customerID string) (<-chan entity.CustomerTrafficData, error) {
	// 1. Get lock for this customer to prevent race conditions
	s.locksMu.Lock()
	if _, ok := s.monitorLocks[customerID]; !ok {
		s.monitorLocks[customerID] = &sync.Mutex{}
	}
	lock := s.monitorLocks[customerID]
	s.locksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	s.mu.Lock()
	custMonitor, exists := s.activeMonitors[customerID]
	if exists {
		// Already monitoring, just increment client count
		custMonitor.Clients++
		s.mu.Unlock()
		log.Printf("[OnDemand] Customer %s: Client count incremented to %d", customerID, custMonitor.Clients)

		// Subscribe to existing monitor
		return s.addObserver(ctx, customerID)
	}
	s.mu.Unlock()

	// 2. Not monitoring yet, need to start.
	// Get customer details first
	customer, err := s.db.GetCustomerByID(customerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Validate customer data
	if customer.PPPoEUsername == nil || *customer.PPPoEUsername == "" {
		return nil, fmt.Errorf("customer has no PPPoE username configured")
	}

	// Use the interface from the customer record directly
	if customer.Interface == nil || *customer.Interface == "" {
		return nil, fmt.Errorf("customer has no interface assigned")
	}
	interfaceName := *customer.Interface

	// Create monitor context
	monitorCtx, cancel := context.WithCancel(context.Background())

	custMonitor = &CustomerMonitor{
		CustomerID:    customerID,
		InterfaceName: interfaceName,
		Cancel:        cancel,
		Clients:       1,
		Observers:     make(map[chan entity.CustomerTrafficData]bool),
		restartCount:  0,
	}

	s.mu.Lock()
	s.activeMonitors[customerID] = custMonitor
	s.mu.Unlock()

	// Start the actual background monitoring for this customer
	go s.runMonitorLoop(monitorCtx, customer, interfaceName)

	log.Printf("[OnDemand] Started monitoring for customer %s (%s) on interface %s",
		customer.Name, *customer.PPPoEUsername, interfaceName)

	return s.addObserver(ctx, customerID)
}

// StopMonitoring decrements client count and stops monitoring if zero
func (s *OnDemandTrafficService) StopMonitoring(customerID string) {
	s.locksMu.Lock()
	if _, ok := s.monitorLocks[customerID]; !ok {
		s.locksMu.Unlock()
		return
	}
	lock := s.monitorLocks[customerID]
	s.locksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	s.mu.Lock()
	custMonitor, exists := s.activeMonitors[customerID]
	if !exists {
		s.mu.Unlock()
		return
	}

	custMonitor.Clients--
	log.Printf("[OnDemand] Customer %s: Client count decremented to %d", customerID, custMonitor.Clients)

	if custMonitor.Clients <= 0 {
		// No more clients, stop monitoring
		custMonitor.Cancel()
		delete(s.activeMonitors, customerID)

		// Close all observers
		for ch := range custMonitor.Observers {
			close(ch)
		}

		log.Printf("[OnDemand] Stopped monitoring for customer %s", customerID)
	}
	s.mu.Unlock()
}

// runMonitorLoop runs the actual MikroTik monitoring command with auto-restart
func (s *OnDemandTrafficService) runMonitorLoop(ctx context.Context, customer *entity.Customer, interfaceName string) {
	maxRestarts := 3
	restartDelay := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Printf("[OnDemand] Monitor context cancelled for %s", customer.Name)
			return
		default:
			// Check restart count
			s.mu.Lock()
			custMonitor, exists := s.activeMonitors[customer.ID]
			if !exists {
				s.mu.Unlock()
				return
			}

			if custMonitor.restartCount >= maxRestarts {
				log.Printf("[OnDemand] Max restart attempts (%d) reached for %s, stopping monitor",
					maxRestarts, customer.Name)
				custMonitor.Cancel()
				delete(s.activeMonitors, customer.ID)

				// Notify observers about disconnection
				for ch := range custMonitor.Observers {
					select {
					case ch <- entity.CustomerTrafficData{
						CustomerID:   customer.ID,
						CustomerName: customer.Name,
						Timestamp:    time.Now(),
					}:
					default:
					}
					close(ch)
				}
				s.mu.Unlock()
				return
			}
			s.mu.Unlock()

			// Start monitoring stream
			trafficChan, err := mon.MonitorTraffic(ctx, s.client, interfaceName)
			if err != nil {
				log.Printf("[OnDemand] Failed to start monitor for %s on %s: %v",
					customer.Name, interfaceName, err)

				s.mu.Lock()
				if mon, ok := s.activeMonitors[customer.ID]; ok {
					mon.restartCount++
				}
				s.mu.Unlock()

				time.Sleep(restartDelay)
				continue
			}

			log.Printf("[OnDemand] Monitor stream active for %s on %s", customer.Name, interfaceName)

			// Process traffic data
			streamClosed := s.processTrafficStream(ctx, customer, trafficChan)

			if streamClosed {
				log.Printf("[OnDemand] Stream closed for %s, attempting restart...", customer.Name)

				s.mu.Lock()
				if mon, ok := s.activeMonitors[customer.ID]; ok {
					mon.restartCount++
				}
				s.mu.Unlock()

				time.Sleep(restartDelay)
			}
		}
	}
}

// processTrafficStream processes traffic data from the stream
func (s *OnDemandTrafficService) processTrafficStream(
	ctx context.Context,
	customer *entity.Customer,
	trafficChan <-chan mon.InterfaceTraffic,
) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case traffic, ok := <-trafficChan:
			if !ok {
				log.Printf("[OnDemand] Traffic channel closed for %s", customer.Name)
				return true // Stream closed
			}

			// Validate traffic data
			if traffic.Name == "" {
				log.Printf("[OnDemand] WARNING: Received traffic data with empty interface name for %s",
					customer.Name)
				continue
			}

			data := s.mapToCustomerTraffic(customer, traffic)
			s.publishTrafficData(data)
		}
	}
}

// addObserver creates a channel and adds it to the monitor's observers
func (s *OnDemandTrafficService) addObserver(ctx context.Context, customerID string) (<-chan entity.CustomerTrafficData, error) {
	s.mu.Lock()
	custMonitor, exists := s.activeMonitors[customerID]
	if !exists {
		s.mu.Unlock()
		return nil, fmt.Errorf("monitor not running for customer %s", customerID)
	}

	// Create buffered channel to prevent blocking
	ch := make(chan entity.CustomerTrafficData, 50)
	custMonitor.Observers[ch] = true
	s.mu.Unlock()

	// Cleanup routine: remove observer when context is done
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		if m, ok := s.activeMonitors[customerID]; ok && m.Observers != nil {
			delete(m.Observers, ch)
			close(ch)
		}
		s.mu.Unlock()
	}()

	return ch, nil
}

func (s *OnDemandTrafficService) mapToCustomerTraffic(c *entity.Customer, t mon.InterfaceTraffic) entity.CustomerTrafficData {
	return entity.CustomerTrafficData{
		CustomerID:         c.ID,
		CustomerName:       c.Name,
		Username:           c.Username,
		ServiceType:        c.ServiceType,
		InterfaceName:      t.Name,
		RxBitsPerSecond:    t.RxBitsPerSecond,
		TxBitsPerSecond:    t.TxBitsPerSecond,
		RxPacketsPerSecond: t.RxPacketsPerSecond,
		TxPacketsPerSecond: t.TxPacketsPerSecond,
		DownloadSpeed:      formatSpeed(t.RxBitsPerSecond),
		UploadSpeed:        formatSpeed(t.TxBitsPerSecond),
		Timestamp:          time.Now(),
	}
}

func (s *OnDemandTrafficService) publishTrafficData(data entity.CustomerTrafficData) {
	// 1. Publish to Redis (optional, for history/other consumers)
	jsonData, _ := json.Marshal(data)
	s.publisher.PublishStream("mikrotik:traffic:customers", string(jsonData))

	// 2. Broadcast to in-memory observers (active websockets)
	s.mu.Lock()
	defer s.mu.Unlock()

	if custMonitor, ok := s.activeMonitors[data.CustomerID]; ok {
		for ch := range custMonitor.Observers {
			select {
			case ch <- data:
			default:
				// Skip if channel full to prevent blocking
			}
		}
	}
}

// formatSpeed converts bits per second to human-readable format
func formatSpeed(bps string) string {
	if bps == "" || bps == "0" {
		return "0 bps"
	}

	// Simple implementation - use the one from continuous service if available
	return bps + " bps"
}
