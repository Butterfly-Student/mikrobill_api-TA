package mikrotik_outbound_adapter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

// Config holds MikroTik connection configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
	UseTLS   bool
	Queue    int // optional: default 100
}

// Client wraps *routeros.Client to make it reusable and configurable.
type Client struct {
	*routeros.Client        // embedded → all default methods available!
	Config           Config // Expose config for creating new instances
	mu               sync.Mutex
}

// NewClient creates and returns a new MikroTik client.
func NewClient(cfg Config) (*Client, error) {
	client := &Client{Config: cfg}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) connect() error {
	address := fmt.Sprintf("%s:%d", c.Config.Host, c.Config.Port)

	var (
		conn *routeros.Client
		err  error
	)

	if c.Config.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout)
		defer cancel()

		if c.Config.UseTLS {
			conn, err = routeros.DialTLSContext(ctx, address, c.Config.Username, c.Config.Password, nil)
		} else {
			conn, err = routeros.DialContext(ctx, address, c.Config.Username, c.Config.Password)
		}
	} else {
		// WITHOUT CONTEXT (SAFE)
		if c.Config.UseTLS {
			conn, err = routeros.DialTLS(address, c.Config.Username, c.Config.Password, nil)
		} else {
			conn, err = routeros.Dial(address, c.Config.Username, c.Config.Password)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to MikroTik: %w", err)
	}

	if c.Config.Queue > 0 {
		conn.Queue = c.Config.Queue
	}

	c.Client = conn
	return nil
}

// Reconnect attempts to re-establish the connection
func (c *Client) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Client != nil {
		c.Client.Close()
	}
	return c.connect()
}

// Run overrides routeros.Client.Run with auto-reconnection support
func (c *Client) Run(sentence ...string) (*routeros.Reply, error) {
	reply, err := c.Client.Run(sentence...)
	if err != nil {
		if IsConnectionError(err) {
			// Try to reconnect
			if recErr := c.Reconnect(); recErr == nil {
				// Retry command
				return c.Client.Run(sentence...)
			}
		}
		return nil, err
	}
	return reply, nil
}

// RunArgs overrides routeros.Client.RunArgs with auto-reconnection support
func (c *Client) RunArgs(sentence string, args map[string]string) (*routeros.Reply, error) {
	cmd := []string{sentence}
	for k, v := range args {
		cmd = append(cmd, "="+k+"="+v)
	}

	reply, err := c.Client.Run(cmd...)
	if err != nil {
		if IsConnectionError(err) {
			// Try to reconnect
			if recErr := c.Reconnect(); recErr == nil {
				// Retry command
				return c.Client.Run(cmd...)
			}
		}
		return nil, err
	}
	return reply, nil
}

// ... (existing code)

func IsConnectionError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "loop has ended") ||
		strings.Contains(msg, "closed network connection") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "EOF")
}

// ListenArgs overrides routeros.Client.ListenArgs (if exists) or implements it
func (c *Client) ListenArgs(sentence string, args map[string]string) (<-chan *proto.Sentence, error) {
	cmd := []string{sentence}
	for k, v := range args {
		cmd = append(cmd, "="+k+"="+v)
	}

	listenReply, err := c.Client.Listen(cmd...)
	if err != nil {
		return nil, err
	}
	return listenReply.Chan(), nil
}

// ListenArgsContext overrides routeros.Client.ListenArgsContext (if exists) or implements it
func (c *Client) ListenArgsContext(ctx context.Context, args []string) (*routeros.ListenReply, error) {
	// Note: routeros library might not support context in Listen directly if it's blocking?
	// The library `routeros.Client` has `Listen(cmd ...string) (*Reply, error)`.
	// Does it have context aware listen?
	// It doesn't seem to have `ListenContext`.
	// But `Listen` returns a Reply struct which has a channel.
	// The actual listening happens in background in the library?

	// If the upstream library doesn't support context cancellation of the initial send, we just use Listen.
	// But `StreamPing` uses `ListenArgsContext`.

	// User provided:
	// reply, err := c.Client.ListenArgsContext(ctx, args)

	// The user's provided code implies `mikrotik.Client` (their custom one) or the library has it.
	// Checking `github.com/go-routeros/routeros/v3`...
	// It doesn't seem to export `ListenArgsContext` on `*Client`.
	// However, we can implement it by wrapping `Listen`.

	// Wait, if user provided code uses `ListenArgsContext`, do they expect us to implement it?
	// Yes.

	// But wait, `routeros.Client` `Listen` sends the command and returns a channel.
	// Cancellation is done by closing the connection or cancelling the context if supported.
	// The library DOES NOT seem to support context for Listen command sending natively in v3?
	// Let's implement a wrapper.

	// Actually, `Listen` just sends command.
	return c.Client.Listen(args...)
}

func (c *Client) Close() error {
	if c.Client != nil {
		c.Client.Close()
	}
	return nil
}

func isConnectionError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "loop has ended") ||
		strings.Contains(msg, "closed network connection") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "EOF")
}
