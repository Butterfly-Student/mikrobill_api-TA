// pkg/queue/server.go
package queue

import (
	"context"

	pkg_logger "mikrobill/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Server wrapper untuk Asynq server
type Server struct {
	server   *asynq.Server
	mux      *asynq.ServeMux
	registry *HandlerRegistry
}

// NewServer membuat instance baru queue server
func NewServer(cfg ServerConfig, registry *HandlerRegistry) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency:     cfg.Concurrency,
			Queues:          cfg.Queues,
			StrictPriority:  cfg.StrictPriority,
			ShutdownTimeout: cfg.ShutdownTimeout,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				// Extract task type dari context jika ada
				taskType := "unknown"
				if v := ctx.Value(taskTypeKey); v != nil {
					if t, ok := v.(string); ok {
						taskType = t
					}
				}
				pkg_logger.Error("Task failed",
					zap.String("task_type", taskType),
					zap.Error(err))
			}),
		},
	)

	mux := asynq.NewServeMux()

	server := &Server{
		server:   srv,
		mux:      mux,
		registry: registry,
	}

	// Register all handlers dari registry
	server.registerHandlers()

	return server
}

type contextKey string

const taskTypeKey contextKey = "task_type"

// registerHandlers mendaftarkan semua handlers dari registry ke mux
func (s *Server) registerHandlers() {
	for _, taskType := range s.registry.GetAll() {
		s.mux.HandleFunc(taskType, s.registry.ToAsynqHandler(taskType))
		pkg_logger.Info("Registered handler for task type", zap.String("task_type", taskType))
	}
}

// RegisterHandler mendaftarkan handler baru (dapat dipanggil setelah server dibuat)
func (s *Server) RegisterHandler(taskType string, handler HandlerFunc) {
	s.registry.Register(taskType, handler)
	s.mux.HandleFunc(taskType, s.registry.ToAsynqHandler(taskType))
	pkg_logger.Info("Registered handler for task type", zap.String("task_type", taskType))
}

// Start memulai server
func (s *Server) Start() error {
	pkg_logger.Info("Starting queue server...")
	return s.server.Start(s.mux)
}

// Stop menghentikan server dengan graceful shutdown
func (s *Server) Stop() {
	pkg_logger.Info("Stopping queue server...")
	s.server.Stop()
	s.server.Shutdown()
}

// Run menjalankan server dengan graceful shutdown pada signal
func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := s.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		s.Stop()
		return ctx.Err()
	}
}
