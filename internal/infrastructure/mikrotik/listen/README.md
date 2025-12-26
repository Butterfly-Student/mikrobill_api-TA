# Listen Package

Package untuk mendengarkan event real-time dari Mikrotik RouterOS menggunakan API Listen.

## Quick Start

```go
package main

import (
    "context"
    "sync"
    
    "mikrobill/pkg/listen"
    "mikrobill/pkg/routeros"
    pkg_logger "mikrobill/pkg/logger"
    "go.uber.org/zap"
)

func main() {
    // Init logger
    pkg_logger.InitLogger("development")
    defer pkg_logger.Sync()
    
    // Connect ke Mikrotik
    client, err := routeros.NewClient(routeros.Config{
        Host:     "192.168.1.1",
        Port:     8728,
        Username: "admin",
        Password: "password",
    })
    if err != nil {
        pkg_logger.Fatal("connection failed", zap.Error(err))
    }
    
    ctx := context.Background()
    eventCh := make(chan listen.ListenEvent, 256)
    state := make(map[string]map[string]map[string]string)
    var mu sync.Mutex
    
    // Start listening
    listen.StartListener(ctx, client, "/interface/listen", eventCh, state, &mu)
    
    // Process events
    for event := range eventCh {
        pkg_logger.Info("event received",
            zap.String("type", string(event.Type)),
            zap.String("id", event.ID),
            zap.Any("data", event.Data),
        )
    }
}
```

## Usage

### 1. Simple Command

```go
listen.StartListener(ctx, client, "/interface/listen", eventCh, state, &mu)
```

### 2. Command dengan Arguments

```go
listen.StartListener(ctx, client, 
    []string{"/log/listen", "follow=yes"}, 
    eventCh, state, &mu,
)
```

### 3. Command dengan Filter

```go
listen.StartListener(ctx, client, 
    []string{"/interface/listen", "where=type=ether"}, 
    eventCh, state, &mu,
)
```

### 4. Custom Configuration

```go
cfg := listen.ListenerConfig{
    Command:  []string{"/queue/simple/listen"},
    EventCh:  eventCh,
    State:    state,
    Mu:       &mu,
    QueueLen: 200, // custom queue size
}
listen.StartListenerWithConfig(ctx, client, cfg)
```

### 5. Multiple Listeners

```go
commands := []interface{}{
    "/interface/listen",
    "/ip/address/listen",
    "/queue/simple/listen",
    []string{"/log/listen", "follow=yes"},
}

for _, cmd := range commands {
    listen.StartListener(ctx, client, cmd, eventCh, state, &mu)
}
```

## Event Types

| Type | Deskripsi |
|------|-----------|
| `EventCreate` | Resource baru dibuat |
| `EventUpdate` | Resource diupdate |
| `EventDelete` | Resource dihapus |
| `EventEnable` | Resource diaktifkan |
| `EventDisable` | Resource dinonaktifkan |

## Event Structure

```go
type ListenEvent struct {
    Command string              // Command yang dijalankan
    ID      string              // Resource ID
    Type    EventType           // Tipe event
    Data    map[string]string   // Data lengkap resource
    Time    time.Time           // Timestamp event
}
```

## Supported Commands

### System & Logging
- `/log/listen`
- `/system/script/environment/listen`
- `/system/resource/listen`

### Interfaces
- `/interface/listen`
- `/interface/wireless/registration-table/listen`
- `/interface/ethernet/switch/port/listen`
- `/interface/ethernet/switch/monitor/listen`

### IP Addresses & Routes
- `/ip/address/listen`
- `/ip/route/listen`
- `/ipv6/route/listen`

### Firewall
- `/ip/firewall/filter/listen`
- `/ip/firewall/nat/listen`
- `/ip/firewall/connection/listen`
- `/ip/firewall/raw/listen`
- `/ip/firewall/service-port/listen`
- `/ipv6/firewall/filter/listen`
- `/ipv6/firewall/connection/listen`

### DHCP
- `/ip/dhcp-server/lease/listen`
- `/ip/dhcp-server/event/listen`

### Wireless
- `/interface/wireless/access-list/listen`

### PPP & VPN
- `/ppp/active/listen`
- `/ppp/secret/listen`
- `/interface/l2tp-server/server/listen`
- `/interface/sstp-server/server/listen`
- `/interface/ovpn-server/server/listen`

### Queues
- `/queue/simple/listen`
- `/queue/tree/listen`
- `/queue/type/listen`

### Hotspot
- `/ip/hotspot/active/listen`
- `/ip/hotspot/host/listen`
- `/ip/hotspot/ip-binding/listen`

### Neighbor Discovery
- `/ip/neighbor/listen`
- `/ipv6/neighbor/listen`

### Tools & Monitoring
- `/tool/bandwidth-server/listen`
- `/tool/romon/listen`
- `/ip/traffic-flow/active/listen`
- `/ip/traffic-flow/ipfix/listen`

### MPLS
- `/mpls/ldp/neighbor/listen`
- `/mpls/ldp/session/listen`

### Other Services
- `/ip/kid-control/device/listen`
- `/ip/socks/listen`
- `/ip/cloud/listen`

**Catatan:**
1. Beberapa command mungkin hanya tersedia di versi RouterOS tertentu
2. Untuk `go-routeros`, pastikan menggunakan method `Listen()` atau `ListenArgsQueueContext()` daripada `Run()` untuk command di atas
3. Listen akan mengembalikan data real-time ketika ada perubahan pada entitas tersebut
4. Tidak semua command di RouterOS support listen - hanya command yang terdaftar di atas yang dijamin bekerja

## State Management

State disimpan dalam struktur:
```go
state[command][id] = map[string]string
```

- **command**: Command path (e.g., "/interface/listen")
- **id**: Resource ID dari Mikrotik
- **map**: Snapshot lengkap data resource

State digunakan untuk mendeteksi perubahan dan menentukan event type.

## Configuration

### ListenerConfig

| Field | Type | Required | Default | Deskripsi |
|-------|------|----------|---------|-----------|
| `Command` | `interface{}` | Yes | - | Command string atau []string |
| `EventCh` | `chan<- ListenEvent` | Yes | - | Channel untuk menerima event |
| `State` | `map[string]map[string]map[string]string` | Yes | - | State storage |
| `Mu` | `*sync.Mutex` | Yes | - | Mutex untuk state |
| `QueueLen` | `int` | No | 100 | Queue size untuk listener |

## Best Practices

1. **Gunakan buffered channel** untuk `eventCh` agar tidak blocking:
   ```go
   eventCh := make(chan listen.ListenEvent, 256)
   ```

2. **Gunakan context untuk graceful shutdown**:
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()
   ```

3. **Handle channel close** saat context dibatalkan:
   ```go
   go func() {
       <-ctx.Done()
       close(eventCh)
   }()
   ```

4. **Proses event di goroutine terpisah** untuk performance:
   ```go
   go func() {
       for event := range eventCh {
           processEvent(event)
       }
   }()
   ```

## Example: Complete Application

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "sync"
    "syscall"
    
    "mikrobill/pkg/listen"
    "mikrobill/pkg/routeros"
    pkg_logger "mikrobill/pkg/logger"
    "go.uber.org/zap"
)

func main() {
    pkg_logger.InitLogger("production")
    defer pkg_logger.Sync()
    
    client, err := routeros.NewClient(routeros.Config{
        Host:     os.Getenv("MIKROTIK_HOST"),
        Port:     8728,
        Username: os.Getenv("MIKROTIK_USER"),
        Password: os.Getenv("MIKROTIK_PASS"),
    })
    if err != nil {
        pkg_logger.Fatal("connection failed", zap.Error(err))
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    eventCh := make(chan listen.ListenEvent, 256)
    state := make(map[string]map[string]map[string]string)
    var mu sync.Mutex
    
    // Start listeners
    commands := []string{
        "/interface/listen",
        "/ip/address/listen",
        "/queue/simple/listen",
    }
    
    for _, cmd := range commands {
        listen.StartListener(ctx, client, cmd, eventCh, state, &mu)
    }
    
    // Process events
    go func() {
        for event := range eventCh {
            pkg_logger.Info("event",
                zap.String("cmd", event.Command),
                zap.String("type", string(event.Type)),
                zap.String("id", event.ID),
            )
        }
    }()
    
    // Graceful shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    <-sigCh
    
    pkg_logger.Info("shutting down...")
    cancel()
    close(eventCh)
}
```

## Troubleshooting

### Listener tidak menerima event

1. Pastikan command path benar
2. Cek koneksi ke Mikrotik
3. Verifikasi permission user Mikrotik
4. Cek log untuk error

### Memory leak

1. Pastikan menggunakan context untuk cancel
2. Close channel saat tidak digunakan
3. Monitor goroutine count

### Event tertunda

1. Tingkatkan `QueueLen`
2. Gunakan buffered channel yang lebih besar
3. Process event di goroutine terpisah
