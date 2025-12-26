// pkg/queue/handler.go
package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// HandlerFunc adalah signature untuk task handler
type HandlerFunc func(ctx context.Context, payload []byte) error

// TypedHandlerFunc adalah generic handler dengan payload yang sudah di-unmarshal
type TypedHandlerFunc[T any] func(ctx context.Context, payload T) error

// HandlerRegistry menyimpan semua task handlers
type HandlerRegistry struct {
	handlers map[string]HandlerFunc
}

// NewHandlerRegistry membuat instance baru handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]HandlerFunc),
	}
}

// Register mendaftarkan handler untuk task type tertentu
func (r *HandlerRegistry) Register(taskType string, handler HandlerFunc) {
	r.handlers[taskType] = handler
}

// RegisterTyped mendaftarkan typed handler dengan automatic unmarshaling
func RegisterTyped[T any](r *HandlerRegistry, taskType string, handler TypedHandlerFunc[T]) {
	r.Register(taskType, func(ctx context.Context, payload []byte) error {
		var data T
		if err := json.Unmarshal(payload, &data); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}
		return handler(ctx, data)
	})
}

// Get mengambil handler berdasarkan task type
func (r *HandlerRegistry) Get(taskType string) (HandlerFunc, bool) {
	handler, exists := r.handlers[taskType]
	return handler, exists
}

// GetAll mengembalikan semua task types yang terdaftar
func (r *HandlerRegistry) GetAll() []string {
	types := make([]string, 0, len(r.handlers))
	for taskType := range r.handlers {
		types = append(types, taskType)
	}
	return types
}

// ToAsynqHandler mengkonversi HandlerFunc ke asynq.HandlerFunc
func (r *HandlerRegistry) ToAsynqHandler(taskType string) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		handler, exists := r.Get(taskType)
		if !exists {
			return fmt.Errorf("handler not found for task type: %s", taskType)
		}

		// Add task type to context for middleware
		ctx = context.WithValue(ctx, taskTypeKey, taskType)
		return handler(ctx, task.Payload())
	}
}


