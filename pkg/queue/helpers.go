// pkg/queue/helpers.go
package queue

import (
	"time"

	"github.com/hibiken/asynq"
)

// QueueOptions adalah helper untuk membuat asynq options dengan mudah
type QueueOptions struct {
	Queue           string
	MaxRetry        int
	Timeout         time.Duration
	Deadline        time.Time
	UniqueFor       time.Duration
	ProcessIn       time.Duration
	ProcessAt       time.Time
	Retention       time.Duration
	TaskID          string
	GroupKey        string
}

// ToAsynqOptions mengkonversi QueueOptions ke []asynq.Option
func (o QueueOptions) ToAsynqOptions() []asynq.Option {
	var opts []asynq.Option

	if o.Queue != "" {
		opts = append(opts, asynq.Queue(o.Queue))
	}
	if o.MaxRetry > 0 {
		opts = append(opts, asynq.MaxRetry(o.MaxRetry))
	}
	if o.Timeout > 0 {
		opts = append(opts, asynq.Timeout(o.Timeout))
	}
	if !o.Deadline.IsZero() {
		opts = append(opts, asynq.Deadline(o.Deadline))
	}
	if o.UniqueFor > 0 {
		opts = append(opts, asynq.Unique(o.UniqueFor))
	}
	if o.ProcessIn > 0 {
		opts = append(opts, asynq.ProcessIn(o.ProcessIn))
	}
	if !o.ProcessAt.IsZero() {
		opts = append(opts, asynq.ProcessAt(o.ProcessAt))
	}
	if o.Retention > 0 {
		opts = append(opts, asynq.Retention(o.Retention))
	}
	if o.TaskID != "" {
		opts = append(opts, asynq.TaskID(o.TaskID))
	}
	if o.GroupKey != "" {
		opts = append(opts, asynq.Group(o.GroupKey))
	}

	return opts
}

// TaskBuilder adalah builder pattern untuk membuat task dengan mudah
type TaskBuilder struct {
	taskType string
	payload  interface{}
	options  QueueOptions
}

// NewTaskBuilder membuat instance baru TaskBuilder
func NewTaskBuilder(taskType string) *TaskBuilder {
	return &TaskBuilder{
		taskType: taskType,
		options:  QueueOptions{},
	}
}

// WithPayload set payload
func (b *TaskBuilder) WithPayload(payload interface{}) *TaskBuilder {
	b.payload = payload
	return b
}

// InQueue set queue name
func (b *TaskBuilder) InQueue(queue string) *TaskBuilder {
	b.options.Queue = queue
	return b
}

// WithRetry set max retry
func (b *TaskBuilder) WithRetry(maxRetry int) *TaskBuilder {
	b.options.MaxRetry = maxRetry
	return b
}

// WithTimeout set timeout
func (b *TaskBuilder) WithTimeout(timeout time.Duration) *TaskBuilder {
	b.options.Timeout = timeout
	return b
}

// WithDelay set delay before processing
func (b *TaskBuilder) WithDelay(delay time.Duration) *TaskBuilder {
	b.options.ProcessIn = delay
	return b
}

// ScheduleAt set specific time to process
func (b *TaskBuilder) ScheduleAt(processAt time.Time) *TaskBuilder {
	b.options.ProcessAt = processAt
	return b
}

// AsUnique make task unique for duration
func (b *TaskBuilder) AsUnique(duration time.Duration) *TaskBuilder {
	b.options.UniqueFor = duration
	return b
}

// WithRetention set retention period
func (b *TaskBuilder) WithRetention(retention time.Duration) *TaskBuilder {
	b.options.Retention = retention
	return b
}

// WithTaskID set custom task ID
func (b *TaskBuilder) WithTaskID(taskID string) *TaskBuilder {
	b.options.TaskID = taskID
	return b
}

// InGroup set group key for task aggregation
func (b *TaskBuilder) InGroup(groupKey string) *TaskBuilder {
	b.options.GroupKey = groupKey
	return b
}

// Enqueue mengirim task ke queue menggunakan client
func (b *TaskBuilder) Enqueue(client *Client) (*asynq.TaskInfo, error) {
	return client.Enqueue(b.taskType, b.payload, b.options.ToAsynqOptions()...)
}

// Build membuat task dan options untuk digunakan manual
func (b *TaskBuilder) Build() (*asynq.Task, []asynq.Option, error) {
	task, err := NewTask(b.taskType, b.payload)
	if err != nil {
		return nil, nil, err
	}
	return task, b.options.ToAsynqOptions(), nil
}

// Common queue names
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

// Predefined QueueOptions untuk kemudahan
var (
	// CriticalTask - High priority, 5 retries, 30 min timeout
	CriticalTask = QueueOptions{
		Queue:    QueueCritical,
		MaxRetry: 5,
		Timeout:  30 * time.Minute,
	}

	// DefaultTask - Normal priority, 3 retries, 15 min timeout
	DefaultTask = QueueOptions{
		Queue:    QueueDefault,
		MaxRetry: 3,
		Timeout:  15 * time.Minute,
	}

	// LowPriorityTask - Low priority, 2 retries, 10 min timeout
	LowPriorityTask = QueueOptions{
		Queue:    QueueLow,
		MaxRetry: 2,
		Timeout:  10 * time.Minute,
	}

	// QuickTask - Quick execution, no retry, 1 min timeout
	QuickTask = QueueOptions{
		Queue:    QueueDefault,
		MaxRetry: 0,
		Timeout:  1 * time.Minute,
	}

	// IdempotentTask - Unique for 1 hour, prevents duplicates
	IdempotentTask = QueueOptions{
		Queue:     QueueDefault,
		MaxRetry:  3,
		Timeout:   15 * time.Minute,
		UniqueFor: 1 * time.Hour,
	}
)