package mikrotik

import (
	"context"
	"fmt"
	"time"

	"github.com/go-routeros/routeros/v3"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
	UseTLS   bool
	Queue    int // opsional: default 100
}

// Client wraps *routeros.Client to make it reusable and configurable.
type Client struct {
	*routeros.Client // embedded → semua method bawaan tersedia!
}

// NewClient creates and returns a new MikroTik client.
func NewClient(cfg Config) (*Client, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var (
		conn *routeros.Client
		err  error
	)

	if cfg.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		if cfg.UseTLS {
			conn, err = routeros.DialTLSContext(ctx, address, cfg.Username, cfg.Password, nil)
		} else {
			conn, err = routeros.DialContext(ctx, address, cfg.Username, cfg.Password)
		}
	} else {
		// ⬅️ TANPA CONTEXT (AMAN)
		if cfg.UseTLS {
			conn, err = routeros.DialTLS(address, cfg.Username, cfg.Password, nil)
		} else {
			conn, err = routeros.Dial(address, cfg.Username, cfg.Password)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to MikroTik: %w", err)
	}

	if cfg.Queue > 0 {
		conn.Queue = cfg.Queue
	}

	return &Client{Client: conn}, nil
}
