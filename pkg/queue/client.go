// pkg/queue/client.go
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkg_logger "mikrobill/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Client wrapper untuk Asynq client
type Client struct {
	client *asynq.Client
}

// NewClient membuat instance baru queue client
func NewClient(cfg Config) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &Client{
		client: client,
	}
}

// Enqueue menambahkan task ke queue dengan payload apapun
func (c *Client) Enqueue(taskType string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task, err := NewTask(taskType, payload)
	if err != nil {
		pkg_logger.Error("Failed to create task",
			zap.String("task_type", taskType),
			zap.Error(err))
		return nil, fmt.Errorf("create task: %w", err)
	}

	info, err := c.client.Enqueue(task, opts...)
	if err != nil {
		pkg_logger.Error("Failed to enqueue task",
			zap.String("task_type", taskType),
			zap.Error(err))
		return nil, fmt.Errorf("enqueue task: %w", err)
	}

	pkg_logger.Info("Task enqueued successfully",
		zap.String("task_type", taskType),
		zap.String("task_id", info.ID))
	return info, nil
}

// EnqueueContext menambahkan task ke queue dengan context
func (c *Client) EnqueueContext(ctx context.Context, taskType string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task, err := NewTask(taskType, payload)
	if err != nil {
		pkg_logger.Error("Failed to create task",
			zap.String("task_type", taskType),
			zap.Error(err))
		return nil, fmt.Errorf("create task: %w", err)
	}

	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		pkg_logger.Error("Failed to enqueue task with context",
			zap.String("task_type", taskType),
			zap.Error(err))
		return nil, fmt.Errorf("enqueue task: %w", err)
	}

	pkg_logger.Info("Task enqueued successfully with context",
		zap.String("task_type", taskType),
		zap.String("task_id", info.ID))
	return info, nil
}

// EnqueueIn menambahkan task yang akan dieksekusi setelah delay tertentu
func (c *Client) EnqueueIn(taskType string, payload interface{}, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.ProcessIn(delay))
	return c.Enqueue(taskType, payload, opts...)
}

// EnqueueAt menambahkan task yang akan dieksekusi pada waktu tertentu
func (c *Client) EnqueueAt(taskType string, payload interface{}, processAt time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.ProcessAt(processAt))
	return c.Enqueue(taskType, payload, opts...)
}

// Close menutup koneksi client
func (c *Client) Close() error {
	err := c.client.Close()
	if err != nil {
		pkg_logger.Error("Failed to close queue client", zap.Error(err))
	}
	return err
}

// GetInspector mengembalikan inspector untuk monitoring queue
func (c *Client) GetInspector(cfg Config) *asynq.Inspector {
	return asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}

// NewTask membuat task baru dengan payload
func NewTask(taskType string, payload interface{}) (*asynq.Task, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(taskType, jsonPayload), nil
}
