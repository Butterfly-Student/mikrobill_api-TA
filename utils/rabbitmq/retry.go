package rabbitmq

import (
	"context"
	"fmt"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"MikrOps/utils/logger"
)

type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

func RetryWithBackoff(ctx context.Context, operation func() error, config RetryConfig, operationName string) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(config.Multiplier, float64(attempt))) * config.InitialDelay
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			logger.GetLogger().Warn("Retry operation",
				zap.String("operation", operationName),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		logger.GetLogger().Error("Operation failed",
			zap.String("operation", operationName),
			zap.Int("attempt", attempt+1),
			zap.Error(err))
	}

	return fmt.Errorf("%s failed after %d attempts: %w",
		operationName, config.MaxRetries, lastErr)
}

func DeclareDeadLetterQueue(ch *amqp.Channel) error {
	dlq, err := ch.QueueDeclare(
		"dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %w", err)
	}

	if err := ch.QueueBind(
		dlq.Name,
		"",
		"dlq",
		false,
		nil,
	); err != nil {
		logger.GetLogger().Warn("Failed to bind dead letter queue",
			zap.Error(err))
	}

	return nil
}

func CreateQueueWithDLQ(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	args := amqp.Table{
		"x-dead-letter-exchange": "dlq",
	}

	return ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args,
	)
}

func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	temporaryErrors := []string{
		"CONNECTION_CLOSED",
		"CHANNEL_CLOSED",
		"INTERNAL_ERROR",
	}

	errStr := err.Error()
	for _, tempErr := range temporaryErrors {
		if len(errStr) >= len(tempErr) && errStr[:len(tempErr)] == tempErr {
			return true
		}
	}

	return false
}

func RetryOperation(ctx context.Context, operation func() error, operationName string) error {
	return RetryWithBackoff(ctx, operation, DefaultRetryConfig(), operationName)
}
