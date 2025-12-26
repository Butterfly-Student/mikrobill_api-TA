// pkg/queue/middleware.go
package queue

import (
	"context"
	"fmt"
	"time"

	pkg_logger "mikrobill/pkg/logger"

	"go.uber.org/zap"
)

// Middleware adalah function yang wrap HandlerFunc
type Middleware func(HandlerFunc) HandlerFunc

// Chain menggabungkan multiple middleware
func Chain(middlewares ...Middleware) Middleware {
	return func(handler HandlerFunc) HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// LoggingMiddleware mencatat eksekusi task
func LoggingMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, payload []byte) error {
			start := time.Now()
			taskType := GetTaskTypeFromContext(ctx)

			pkg_logger.Info("Task started",
				zap.String("task_type", taskType),
				zap.Time("started_at", start))

			err := next(ctx, payload)

			duration := time.Since(start)
			if err != nil {
				pkg_logger.Error("Task failed",
					zap.String("task_type", taskType),
					zap.Duration("duration", duration),
					zap.Error(err))
			} else {
				pkg_logger.Info("Task completed successfully",
					zap.String("task_type", taskType),
					zap.Duration("duration", duration))
			}

			return err
		}
	}
}

// RecoveryMiddleware menangkap panic dari handler
func RecoveryMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, payload []byte) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic recovered: %v", r)
					taskType := GetTaskTypeFromContext(ctx)
					pkg_logger.Error("Panic recovered in task",
						zap.String("task_type", taskType),
						zap.Any("panic", r),
						zap.Stack("stack"))
				}
			}()

			return next(ctx, payload)
		}
	}
}

// TimeoutMiddleware menambahkan timeout pada handler
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, payload []byte) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			errChan := make(chan error, 1)
			go func() {
				errChan <- next(ctx, payload)
			}()

			select {
			case err := <-errChan:
				return err
			case <-ctx.Done():
				taskType := GetTaskTypeFromContext(ctx)
				pkg_logger.Warn("Task timeout",
					zap.String("task_type", taskType),
					zap.Duration("timeout", timeout))
				return fmt.Errorf("task timeout after %s", timeout)
			}
		}
	}
}

// RetryMiddleware untuk custom retry logic (selain built-in asynq retry)
func RetryMiddleware(maxRetries int) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, payload []byte) error {
			var err error
			taskType := GetTaskTypeFromContext(ctx)

			for i := 0; i <= maxRetries; i++ {
				err = next(ctx, payload)
				if err == nil {
					return nil
				}

				if i < maxRetries {
					pkg_logger.Warn("Retry attempt",
						zap.String("task_type", taskType),
						zap.Int("attempt", i+1),
						zap.Int("max_retries", maxRetries),
						zap.Error(err))
					time.Sleep(time.Second * time.Duration(i+1)) // exponential backoff
				}
			}
			return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
		}
	}
}

// MetricsMiddleware untuk tracking metrics
func MetricsMiddleware(onComplete func(taskType string, duration time.Duration, err error)) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, payload []byte) error {
			start := time.Now()
			err := next(ctx, payload)
			duration := time.Since(start)

			taskType := GetTaskTypeFromContext(ctx)
			onComplete(taskType, duration, err)

			return err
		}
	}
}

// GetTaskTypeFromContext helper function
func GetTaskTypeFromContext(ctx context.Context) string {
	if v := ctx.Value(taskTypeKey); v != nil {
		if t, ok := v.(string); ok {
			return t
		}
	}
	return "unknown"
}
