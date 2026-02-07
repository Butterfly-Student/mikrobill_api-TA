package isolation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/logger"
)

const (
	CheckInterval = 24 * time.Hour
	MaxBatchSize  = 100
)

// IsolationWorker checks for expired customers and isolates them
type IsolationWorker struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	logger       *zap.Logger
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewIsolationWorker creates a new isolation worker
func NewIsolationWorker(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
) *IsolationWorker {
	return &IsolationWorker{
		databasePort: databasePort,
		messagePort:  messagePort,
		logger:       logger.GetLogger().With(zap.String("worker", "isolation")),
	}
}

// Start begins the isolation worker loop
func (w *IsolationWorker) Start(ctx context.Context) error {
	w.ctx, w.cancel = context.WithCancel(ctx)

	w.logger.Info("Starting automated isolation worker")

	go w.run()

	return nil
}

// Stop gracefully stops the isolation worker
func (w *IsolationWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}

	w.logger.Info("Stopping automated isolation worker")
}

func (w *IsolationWorker) run() {
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Isolation worker context cancelled")
			return
		case <-time.After(CheckInterval):
			if err := w.checkExpiredCustomers(); err != nil {
				w.logger.Error("Failed to check expired customers",
					zap.Error(err))
			}
		}
	}
}

func (w *IsolationWorker) checkExpiredCustomers() error {
	w.logger.Debug("Checking for expired customers")

	expiredCustomers, err := w.databasePort.Customer().ListExpired(w.ctx, model.CustomerStatusActive)
	if err != nil {
		return err
	}

	if len(expiredCustomers) == 0 {
		w.logger.Debug("No expired customers found")
		return nil
	}

	w.logger.Info("Found expired customers", zap.Int("count", len(expiredCustomers)))

	for _, customer := range expiredCustomers {
		if err := w.isolateCustomer(customer); err != nil {
			w.logger.Error("Failed to isolate customer",
				zap.String("customer_id", customer.ID),
				zap.Error(err))
		}
	}

	return nil
}

func (w *IsolationWorker) isolateCustomer(customer model.Customer) error {
	w.logger.Info("Isolating customer",
		zap.String("customer_id", customer.ID),
		zap.String("username", customer.ServiceUsername),
		zap.String("service_type", string(customer.ServiceType)))

	// Publish disconnect message via RabbitMQ for async MikroTik execution
	msg := model.DisconnectMessage{
		CustomerID:      customer.ID,
		TenantID:        customer.TenantID,
		ServiceUsername: customer.ServiceUsername,
		ServiceType:     string(customer.ServiceType),
	}

	if err := w.messagePort.Provisioning().PublishDisconnect(w.ctx, msg); err != nil {
		w.logger.Error("Failed to publish disconnect message",
			zap.String("customer_id", customer.ID),
			zap.Error(err))
		return err
	}

	// Update customer status to suspended
	customerID, parseErr := uuid.Parse(customer.ID)
	if parseErr != nil {
		w.logger.Error("Failed to parse customer ID", zap.String("customer_id", customer.ID), zap.Error(parseErr))
		return parseErr
	}
	if err := w.databasePort.Customer().UpdateStatus(w.ctx, customerID, model.CustomerStatusSuspended, nil, nil, nil); err != nil {
		w.logger.Error("Failed to update customer status to suspended",
			zap.String("customer_id", customer.ID),
			zap.Error(err))
		return err
	}

	w.logger.Info("Customer isolated successfully",
		zap.String("customer_id", customer.ID),
		zap.String("username", customer.ServiceUsername),
		zap.String("service_type", string(customer.ServiceType)))

	return nil
}
