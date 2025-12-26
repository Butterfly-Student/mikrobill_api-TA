// pkg/queue/scheduler.go
package queue

import (
	"fmt"

	pkg_logger "mikrobill/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Scheduler untuk periodic tasks
type Scheduler struct {
	scheduler *asynq.Scheduler
	entries   map[string]string // map[entryID]cronspec untuk tracking
}

// PeriodicTask definisi untuk periodic task
type PeriodicTask struct {
	EntryID  string
	CronSpec string
	TaskType string
	Payload  interface{}
	Options  []asynq.Option
}

// NewScheduler membuat instance baru scheduler
func NewScheduler(cfg SchedulerConfig) *Scheduler {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		&asynq.SchedulerOpts{
			Location: cfg.Location,
		},
	)

	return &Scheduler{
		scheduler: scheduler,
		entries:   make(map[string]string),
	}
}

// Register mendaftarkan periodic task
func (s *Scheduler) Register(pt PeriodicTask) error {
	task, err := NewTask(pt.TaskType, pt.Payload)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	entryID, err := s.scheduler.Register(pt.CronSpec, task, pt.Options...)
	if err != nil {
		return fmt.Errorf("register periodic task: %w", err)
	}

	s.entries[entryID] = pt.CronSpec
	pkg_logger.Info("Registered periodic task",
		zap.String("task_type", pt.TaskType),
		zap.String("entry_id", entryID),
		zap.String("cron", pt.CronSpec))

	return nil
}

// RegisterMultiple mendaftarkan multiple periodic tasks sekaligus
func (s *Scheduler) RegisterMultiple(tasks []PeriodicTask) error {
	for _, task := range tasks {
		if err := s.Register(task); err != nil {
			return fmt.Errorf("register task %s: %w", task.TaskType, err)
		}
	}
	return nil
}

// Unregister menghapus periodic task berdasarkan entry ID
func (s *Scheduler) Unregister(entryID string) error {
	if err := s.scheduler.Unregister(entryID); err != nil {
		return fmt.Errorf("unregister task: %w", err)
	}
	delete(s.entries, entryID)
	pkg_logger.Info("Unregistered periodic task", zap.String("entry_id", entryID))
	return nil
}

// GetEntries mengembalikan semua entry IDs yang terdaftar
func (s *Scheduler) GetEntries() map[string]string {
	return s.entries
}

// Start memulai scheduler
func (s *Scheduler) Start() error {
	pkg_logger.Info("Starting scheduler...")
	return s.scheduler.Start()
}

// Stop menghentikan scheduler
func (s *Scheduler) Stop() {
	pkg_logger.Info("Stopping scheduler...")
	s.scheduler.Shutdown()
}

// Helper functions untuk membuat PeriodicTask dengan mudah

// NewPeriodicTask membuat PeriodicTask baru
func NewPeriodicTask(entryID, cronSpec, taskType string, payload interface{}, opts ...asynq.Option) PeriodicTask {
	return PeriodicTask{
		EntryID:  entryID,
		CronSpec: cronSpec,
		TaskType: taskType,
		Payload:  payload,
		Options:  opts,
	}
}

// Common cron specifications
const (
	EveryMinute    = "* * * * *"
	Every5Minutes  = "*/5 * * * *"
	Every15Minutes = "*/15 * * * *"
	Every30Minutes = "*/30 * * * *"
	EveryHour      = "0 * * * *"
	Every6Hours    = "0 */6 * * *"
	Every12Hours   = "0 */12 * * *"
	EveryDay       = "0 0 * * *"
	EveryWeek      = "0 0 * * 0"
	EveryMonth     = "0 0 1 * *"
)
