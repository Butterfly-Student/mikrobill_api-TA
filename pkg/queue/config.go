// pkg/queue/config.go
package queue

import "time"

// Config konfigurasi untuk queue client dan server
type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// ServerConfig konfigurasi untuk queue server
type ServerConfig struct {
	Config
	Concurrency    int
	Queues         map[string]int
	StrictPriority bool
	ShutdownTimeout time.Duration
}

// DefaultServerConfig mengembalikan konfigurasi default untuk server
func DefaultServerConfig(cfg Config) ServerConfig {
	return ServerConfig{
		Config:      cfg,
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		StrictPriority:  false,
		ShutdownTimeout: 30 * time.Second,
	}
}

// SchedulerConfig konfigurasi untuk scheduler
type SchedulerConfig struct {
	Config
	Location *time.Location
}

// DefaultSchedulerConfig mengembalikan konfigurasi default untuk scheduler
func DefaultSchedulerConfig(cfg Config) SchedulerConfig {
	return SchedulerConfig{
		Config:   cfg,
		Location: time.UTC,
	}
}