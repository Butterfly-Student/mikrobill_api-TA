// pkg/queue/inspector.go
package queue

import (
	"fmt"

	"github.com/hibiken/asynq"
)

// Inspector wrapper untuk monitoring dan management
type Inspector struct {
	inspector *asynq.Inspector
}

// NewInspector membuat instance baru inspector
func NewInspector(cfg Config) *Inspector {
	return &Inspector{
		inspector: asynq.NewInspector(asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		}),
	}
}

// GetQueueInfo mendapatkan informasi tentang queue
func (i *Inspector) GetQueueInfo(queueName string) (*asynq.QueueInfo, error) {
	info, err := i.inspector.GetQueueInfo(queueName)
	if err != nil {
		return nil, fmt.Errorf("get queue info: %w", err)
	}
	return info, nil
}

// ListQueues mendapatkan list semua queues
func (i *Inspector) ListQueues() ([]*asynq.QueueInfo, error) {
	queues, err := i.inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("list queues: %w", err)
	}

	var infos []*asynq.QueueInfo
	for _, q := range queues {
		info, err := i.inspector.GetQueueInfo(q)
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// ListPendingTasks mendapatkan list pending tasks di queue
func (i *Inspector) ListPendingTasks(queueName string, pageSize int) ([]*asynq.TaskInfo, error) {
	tasks, err := i.inspector.ListPendingTasks(queueName, asynq.PageSize(pageSize))
	if err != nil {
		return nil, fmt.Errorf("list pending tasks: %w", err)
	}
	return tasks, nil
}

// ListActiveTasks mendapatkan list active tasks di queue
func (i *Inspector) ListActiveTasks(queueName string, pageSize int) ([]*asynq.TaskInfo, error) {
	tasks, err := i.inspector.ListActiveTasks(queueName, asynq.PageSize(pageSize))
	if err != nil {
		return nil, fmt.Errorf("list active tasks: %w", err)
	}
	return tasks, nil
}

// ListScheduledTasks mendapatkan list scheduled tasks di queue
func (i *Inspector) ListScheduledTasks(queueName string, pageSize int) ([]*asynq.TaskInfo, error) {
	tasks, err := i.inspector.ListScheduledTasks(queueName, asynq.PageSize(pageSize))
	if err != nil {
		return nil, fmt.Errorf("list scheduled tasks: %w", err)
	}
	return tasks, nil
}

// ListRetryTasks mendapatkan list retry tasks di queue
func (i *Inspector) ListRetryTasks(queueName string, pageSize int) ([]*asynq.TaskInfo, error) {
	tasks, err := i.inspector.ListRetryTasks(queueName, asynq.PageSize(pageSize))
	if err != nil {
		return nil, fmt.Errorf("list retry tasks: %w", err)
	}
	return tasks, nil
}

// ListArchivedTasks mendapatkan list archived (dead) tasks di queue
func (i *Inspector) ListArchivedTasks(queueName string, pageSize int) ([]*asynq.TaskInfo, error) {
	tasks, err := i.inspector.ListArchivedTasks(queueName, asynq.PageSize(pageSize))
	if err != nil {
		return nil, fmt.Errorf("list archived tasks: %w", err)
	}
	return tasks, nil
}

// DeleteTask menghapus task berdasarkan queue dan task ID
func (i *Inspector) DeleteTask(queueName, taskID string) error {
	if err := i.inspector.DeleteTask(queueName, taskID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// RunTask menjalankan task yang scheduled/retry immediately
func (i *Inspector) RunTask(queueName, taskID string) error {
	if err := i.inspector.RunTask(queueName, taskID); err != nil {
		return fmt.Errorf("run task: %w", err)
	}
	return nil
}

// ArchiveTask memindahkan task ke archived state
func (i *Inspector) ArchiveTask(queueName, taskID string) error {
	if err := i.inspector.ArchiveTask(queueName, taskID); err != nil {
		return fmt.Errorf("archive task: %w", err)
	}
	return nil
}

// DeleteAllPendingTasks menghapus semua pending tasks di queue
func (i *Inspector) DeleteAllPendingTasks(queueName string) (int, error) {
	n, err := i.inspector.DeleteAllPendingTasks(queueName)
	if err != nil {
		return 0, fmt.Errorf("delete all pending tasks: %w", err)
	}
	return n, nil
}

// DeleteAllScheduledTasks menghapus semua scheduled tasks di queue
func (i *Inspector) DeleteAllScheduledTasks(queueName string) (int, error) {
	n, err := i.inspector.DeleteAllScheduledTasks(queueName)
	if err != nil {
		return 0, fmt.Errorf("delete all scheduled tasks: %w", err)
	}
	return n, nil
}

// DeleteAllRetryTasks menghapus semua retry tasks di queue
func (i *Inspector) DeleteAllRetryTasks(queueName string) (int, error) {
	n, err := i.inspector.DeleteAllRetryTasks(queueName)
	if err != nil {
		return 0, fmt.Errorf("delete all retry tasks: %w", err)
	}
	return n, nil
}

// DeleteAllArchivedTasks menghapus semua archived tasks di queue
func (i *Inspector) DeleteAllArchivedTasks(queueName string) (int, error) {
	n, err := i.inspector.DeleteAllArchivedTasks(queueName)
	if err != nil {
		return 0, fmt.Errorf("delete all archived tasks: %w", err)
	}
	return n, nil
}

// PauseQueue pause queue processing
func (i *Inspector) PauseQueue(queueName string) error {
	if err := i.inspector.PauseQueue(queueName); err != nil {
		return fmt.Errorf("pause queue: %w", err)
	}
	return nil
}

// UnpauseQueue resume queue processing
func (i *Inspector) UnpauseQueue(queueName string) error {
	if err := i.inspector.UnpauseQueue(queueName); err != nil {
		return fmt.Errorf("unpause queue: %w", err)
	}
	return nil
}

// Close menutup inspector
func (i *Inspector) Close() error {
	return i.inspector.Close()
}