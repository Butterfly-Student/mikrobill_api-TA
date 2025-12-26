# Queue Package - Reusable Asyncq Wrapper

Package ini adalah wrapper yang reusable dan production-ready untuk [asynq](https://github.com/hibiken/asynq) yang memudahkan pengelolaan background jobs dan scheduled tasks.

## Fitur

- ✅ **Reusable & Flexible** - Tidak hardcode task types, dynamic handler registration
- ✅ **Type-Safe Handlers** - Generic handler dengan automatic unmarshaling
- ✅ **Task Builder** - Fluent API untuk membuat tasks dengan mudah
- ✅ **Predefined Options** - CriticalTask, DefaultTask, QuickTask, dll
- ✅ **Middleware Support** - Logging, recovery, timeout, retry, metrics
- ✅ **Scheduler** - Periodic tasks dengan cron expressions
- ✅ **Inspector** - Queue monitoring dan management tools
- ✅ **Graceful Shutdown** - Context-based shutdown
- ✅ **Multiple Queues** - Priority-based queue management
- ✅ **Easy Configuration** - Default configs dengan customization options

## Instalasi

```bash
go get github.com/hibiken/asynq
```

## Struktur Package

```
pkg/queue/
├── config.go       # Konfigurasi untuk client, server, dan scheduler
├── client.go       # Queue client untuk enqueue tasks
├── server.go       # Queue server untuk process tasks
├── handler.go      # Handler registry dan typed handlers
├── scheduler.go    # Periodic task scheduler
├── middleware.go   # Middleware untuk handlers
├── inspector.go    # Queue monitoring dan management
└── helpers.go      # Helper functions dan builder pattern
```

## Quick Start

### 1. Setup Configuration

```go
cfg := queue.Config{
    RedisAddr:     "localhost:6379",
    RedisPassword: "",
    RedisDB:       0,
}
```

### 2. Define Task Types & Payloads

```go
const (
    TaskSendEmail = "email:send"
    TaskProcessImage = "image:process"
)

type SendEmailPayload struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}
```

### 3. Create Worker (Server)

```go
// Buat handler registry
registry := queue.NewHandlerRegistry()

// Register typed handlers
queue.RegisterTyped(registry, TaskSendEmail, func(ctx context.Context, payload SendEmailPayload) error {
    log.Printf("Sending email to: %s", payload.To)
    // Your email sending logic here
    return nil
})

// Buat server
serverCfg := queue.DefaultServerConfig(cfg)
server := queue.NewServer(serverCfg, registry)

// Start server
if err := server.Start(); err != nil {
    log.Fatal(err)
}
defer server.Stop()
```

### 4. Enqueue Tasks (Client)

Ada 3 cara untuk enqueue tasks:

**Cara 1: Direct Enqueue**
```go
client := queue.NewClient(cfg)
defer client.Close()

info, err := client.Enqueue(
    TaskSendEmail,
    SendEmailPayload{
        To:      "user@example.com",
        Subject: "Welcome",
        Body:    "Hello!",
    },
    asynq.Queue("critical"),
    asynq.MaxRetry(3),
)
```

**Cara 2: Using TaskBuilder (Recommended)**
```go
info, err := queue.NewTaskBuilder(TaskSendEmail).
    WithPayload(SendEmailPayload{...}).
    InQueue("critical").
    WithRetry(3).
    WithTimeout(5 * time.Minute).
    Enqueue(client)
```

**Cara 3: Using Predefined Options**
```go
opts := queue.CriticalTask // Preset: critical queue, 5 retries, 30min timeout
info, err := client.Enqueue(TaskSendEmail, payload, opts.ToAsynqOptions()...)
```

## Advanced Usage

### Task Builder Pattern

Task Builder membuat kode lebih readable dan maintainable:

```go
// Simple task
info, err := queue.NewTaskBuilder(TaskSendEmail).
    WithPayload(emailPayload).
    InQueue("critical").
    Enqueue(client)

// Delayed task
info, err := queue.NewTaskBuilder(TaskSendEmail).
    WithPayload(emailPayload).
    WithDelay(5 * time.Minute).
    Enqueue(client)

// Scheduled task
info, err := queue.NewTaskBuilder(TaskGenerateReport).
    WithPayload(reportPayload).
    ScheduleAt(time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC)).
    Enqueue(client)

// Unique task (prevent duplicates)
info, err := queue.NewTaskBuilder(TaskProcessImage).
    WithPayload(imagePayload).
    AsUnique(1 * time.Hour).
    WithTaskID("process-image-123").
    Enqueue(client)

// Complex configuration
info, err := queue.NewTaskBuilder(TaskComplexJob).
    WithPayload(payload).
    InQueue("default").
    WithRetry(5).
    WithTimeout(30 * time.Minute).
    AsUnique(24 * time.Hour).
    WithRetention(7 * 24 * time.Hour).
    InGroup("batch-process").
    Enqueue(client)
```

### Predefined Task Options

Package menyediakan preset options untuk use case umum:

```go
// High priority, 5 retries, 30 min timeout
queue.CriticalTask

// Normal priority, 3 retries, 15 min timeout
queue.DefaultTask

// Low priority, 2 retries, 10 min timeout
queue.LowPriorityTask

// Quick execution, no retry, 1 min timeout
queue.QuickTask

// Unique for 1 hour, prevents duplicates
queue.IdempotentTask
```

Usage:
```go
opts := queue.CriticalTask
opts.MaxRetry = 10 // Customize jika perlu
client.Enqueue(TaskSendEmail, payload, opts.ToAsynqOptions()...)
```

### Queue Inspector (Monitoring)

Inspector untuk monitoring dan management queue:

```go
inspector := queue.NewInspector(cfg)
defer inspector.Close()

// Get queue info
info, err := inspector.GetQueueInfo("critical")
fmt.Printf("Pending: %d, Active: %d\n", info.Pending, info.Active)

// List all queues
queues, err := inspector.ListQueues()
for _, q := range queues {
    fmt.Printf("%s: %d pending\n", q.Queue, q.Pending)
}

// List pending tasks
tasks, err := inspector.ListPendingTasks("default", 100)

// Manage tasks
inspector.DeleteTask("default", "task-id-123")
inspector.RunTask("default", "scheduled-task-id")
inspector.ArchiveTask("default", "failed-task-id")

// Pause/Resume queue
inspector.PauseQueue("default")
inspector.UnpauseQueue("default")

// Bulk operations
n, err := inspector.DeleteAllPendingTasks("default")
fmt.Printf("Deleted %d tasks\n", n)
```

### Periodic Tasks (Scheduler)

```go
schedulerCfg := queue.DefaultSchedulerConfig(cfg)
scheduler := queue.NewScheduler(schedulerCfg)

// Register periodic task
err := scheduler.Register(queue.PeriodicTask{
    EntryID:  "daily-cleanup",
    CronSpec: queue.EveryDay,  // "0 0 * * *"
    TaskType: TaskCleanup,
    Payload:  CleanupPayload{Directory: "/tmp"},
    Options:  []asynq.Option{asynq.Queue("low")},
})

scheduler.Start()
defer scheduler.Stop()
```

### Cron Specifications

Package ini menyediakan konstanta untuk cron yang umum:

```go
queue.EveryMinute      // "* * * * *"
queue.Every5Minutes    // "*/5 * * * *"
queue.Every15Minutes   // "*/15 * * * *"
queue.Every30Minutes   // "*/30 * * * *"
queue.EveryHour        // "0 * * * *"
queue.Every6Hours      // "0 */6 * * *"
queue.Every12Hours     // "0 */12 * * *"
queue.EveryDay         // "0 0 * * *"
queue.EveryWeek        // "0 0 * * 0"
queue.EveryMonth       // "0 0 1 * *"
```

### Delayed & Scheduled Tasks

```go
// Enqueue dengan delay
client.EnqueueIn(TaskSendEmail, payload, 5*time.Minute)

// Enqueue pada waktu tertentu
client.EnqueueAt(TaskSendEmail, payload, time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC))
```

### Middleware

```go
// Logging middleware
loggingMw := queue.LoggingMiddleware(logger)

// Recovery middleware untuk handle panic
recoveryMw := queue.RecoveryMiddleware(logger)

// Timeout middleware
timeoutMw := queue.TimeoutMiddleware(30 * time.Second)

// Chain middleware
combinedMw := queue.Chain(recoveryMw, loggingMw, timeoutMw)

// Apply ke handler
registry.Register(TaskSendEmail, combinedMw(yourHandler))
```

### Custom Server Configuration

```go
serverCfg := queue.ServerConfig{
    Config:      cfg,
    Concurrency: 20,
    Queues: map[string]int{
        "critical": 10,
        "default":  5,
        "low":      1,
    },
    StrictPriority:  false,
    ShutdownTimeout: 30 * time.Second,
}

server := queue.NewServer(serverCfg, registry)
```

### Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Handle signals
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    cancel()
}()

// Run server dengan context
if err := server.Run(ctx); err != nil && err != context.Canceled {
    log.Fatal(err)
}
```

## Best Practices

1. **Task Types**: Gunakan konstanta untuk task types dengan naming convention `domain:action`
2. **Payloads**: Definisikan struct dengan json tags untuk setiap payload
3. **Error Handling**: Selalu return error dari handler, jangan panic
4. **Idempotency**: Buat handlers yang idempotent (aman dijalankan multiple kali)
5. **Retry Logic**: Set `MaxRetry` sesuai kebutuhan task
6. **Queue Priority**: Gunakan queue berbeda untuk task dengan priority berbeda
7. **Monitoring**: Gunakan middleware untuk logging dan metrics

## Queue Priority

Queue dengan nilai lebih tinggi mendapat priority lebih tinggi:

```go
Queues: map[string]int{
    "critical": 6,  // Highest priority
    "default":  3,  // Medium priority
    "low":      1,  // Lowest priority
}
```

## Task Options

```go
asynq.Queue("critical")           // Set queue
asynq.MaxRetry(5)                 // Max retry attempts
asynq.Timeout(5 * time.Minute)   // Task timeout
asynq.ProcessIn(10 * time.Second) // Delay execution
asynq.ProcessAt(time.Now().Add(1 * time.Hour)) // Schedule at specific time
asynq.Unique(1 * time.Hour)       // Prevent duplicate tasks
asynq.Retention(24 * time.Hour)   // Keep task info after completion
```

## Example Project Structure

```
your-project/
├── pkg/
│   └── queue/          # Queue package
├── cmd/
│   ├── worker/         # Worker application
│   │   └── main.go
│   └── api/            # API application
│       └── main.go
├── internal/
│   ├── tasks/          # Task definitions
│   │   ├── email.go
│   │   └── report.go
│   └── handlers/       # Task handlers
│       ├── email.go
│       └── report.go
└── go.mod
```

## Troubleshooting

### Task tidak diproses
- Pastikan worker sedang running
- Check Redis connection
- Verify task type terdaftar di registry
- Check queue priority configuration

### Memory leak
- Pastikan selalu defer client.Close()
- Pastikan server.Stop() dipanggil saat shutdown
- Gunakan context.WithTimeout untuk long-running tasks

### Task timeout
- Increase timeout di task options atau middleware
- Split task besar menjadi subtasks lebih kecil
- Gunakan queue dengan priority lebih tinggi

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License