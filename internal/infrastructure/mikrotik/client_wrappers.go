package mikrotik

import (
	"context"
	"fmt"
	"log"
)

// PPPoEProfileParams holds parameters for creating/updating PPPoE profiles
type PPPoEProfileParams struct {
	Name             string
	LocalAddress     string
	RemoteAddress    string
	RateLimitUp      string
	RateLimitDown    string
	IdleTimeout      string
	SessionTimeout   string
	KeepaliveTimeout string
	OnlyOne          bool
	DNSServer        string
	AddressPool      string
}

// ==================== PPPoE Secret Methods ====================

// CreatePPPoESecret creates a new PPPoE secret
func (c *Client) CreatePPPoESecret(username, password, profile, localAddress, remoteAddress string) (string, error) {
	cmd := []string{
		"/ppp/secret/add",
		"=name=" + username,
		"=password=" + password,
		"=service=pppoe",
	}

	if profile != "" {
		cmd = append(cmd, "=profile="+profile)
	}
	if localAddress != "" {
		cmd = append(cmd, "=local-address="+localAddress)
	}
	if remoteAddress != "" {
		cmd = append(cmd, "=remote-address="+remoteAddress)
	}

	r, err := c.RunArgs(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to create ppp secret: %w", err)
	}

	// return ID of created item
	return r.Done.Map["ret"], nil
}

// UpdatePPPoESecret updates an existing PPPoE secret
func (c *Client) UpdatePPPoESecret(id, username, password, profile, localAddress, remoteAddress string) error {
	cmd := []string{
		"/ppp/secret/set",
		"=.id=" + id,
	}

	if username != "" {
		cmd = append(cmd, "=name="+username)
	}
	if password != "" {
		cmd = append(cmd, "=password="+password)
	}
	if profile != "" {
		cmd = append(cmd, "=profile="+profile)
	}
	if localAddress != "" {
		cmd = append(cmd, "=local-address="+localAddress)
	}
	if remoteAddress != "" {
		cmd = append(cmd, "=remote-address="+remoteAddress)
	}

	_, err := c.RunArgs(cmd)
	if err != nil {
		return fmt.Errorf("failed to update ppp secret: %w", err)
	}
	return nil
}

// DeletePPPoESecret deletes a PPPoE secret by ID
func (c *Client) DeletePPPoESecret(id string) error {
	cmd := []string{
		"/ppp/secret/remove",
		"=.id=" + id,
	}
	_, err := c.RunArgs(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete ppp secret: %w", err)
	}
	return nil
}

// FindPPPoESecretID returns the ID of a PPPoE secret by username
func (c *Client) FindPPPoESecretID(username string) (string, error) {
	cmd := []string{
		"/ppp/secret/print",
		"?name=" + username,
		"=.proplist=.id",
	}

	r, err := c.RunArgs(cmd)
	if err != nil {
		return "", err
	}

	if len(r.Re) == 0 {
		return "", nil // Not found
	}

	return r.Re[0].Map[".id"], nil
}

// ==================== PPPoE Profile Methods ====================

// CreatePPPoEProfile creates a new PPPoE profile on MikroTik
func (c *Client) CreatePPPoEProfile(params PPPoEProfileParams) error {
	log.Printf("[MikroTik] CreatePPPoEProfile - Creating profile: %s", params.Name)

	cmd := []string{
		"/ppp/profile/add",
		"=name=" + params.Name,
	}

	if params.LocalAddress != "" {
		cmd = append(cmd, "=local-address="+params.LocalAddress)
	}
	if params.RemoteAddress != "" {
		cmd = append(cmd, "=remote-address="+params.RemoteAddress)
	}
	if params.RateLimitUp != "" || params.RateLimitDown != "" {
		rateLimit := params.RateLimitUp + "/" + params.RateLimitDown
		cmd = append(cmd, "=rate-limit="+rateLimit)
	}
	if params.IdleTimeout != "" {
		cmd = append(cmd, "=idle-timeout="+params.IdleTimeout)
	}
	if params.SessionTimeout != "" {
		cmd = append(cmd, "=session-timeout="+params.SessionTimeout)
	}
	if params.KeepaliveTimeout != "" {
		cmd = append(cmd, "=keepalive-timeout="+params.KeepaliveTimeout)
	}
	if params.OnlyOne {
		cmd = append(cmd, "=only-one=yes")
	}
	if params.DNSServer != "" {
		cmd = append(cmd, "=dns-server="+params.DNSServer)
	}

	_, err := c.RunArgs(cmd)
	if err != nil {
		log.Printf("[MikroTik] CreatePPPoEProfile - ERROR: %v", err)
		return fmt.Errorf("failed to create pppoe profile: %w", err)
	}

	log.Printf("[MikroTik] CreatePPPoEProfile - SUCCESS: Created profile %s", params.Name)
	return nil
}

// UpdatePPPoEProfile updates an existing PPPoE profile on MikroTik
func (c *Client) UpdatePPPoEProfile(params PPPoEProfileParams) error {
	log.Printf("[MikroTik] UpdatePPPoEProfile - Updating profile: %s", params.Name)

	profileID, err := c.FindPPPoEProfileID(params.Name)
	if err != nil {
		log.Printf("[MikroTik] UpdatePPPoEProfile - ERROR finding profile: %v", err)
		return fmt.Errorf("failed to find profile: %w", err)
	}

	if profileID == "" {
		return fmt.Errorf("profile not found: %s", params.Name)
	}

	cmd := []string{
		"/ppp/profile/set",
		"=.id=" + profileID,
	}

	if params.LocalAddress != "" {
		cmd = append(cmd, "=local-address="+params.LocalAddress)
	}
	if params.RemoteAddress != "" {
		cmd = append(cmd, "=remote-address="+params.RemoteAddress)
	}
	if params.RateLimitUp != "" || params.RateLimitDown != "" {
		rateLimit := params.RateLimitUp + "/" + params.RateLimitDown
		cmd = append(cmd, "=rate-limit="+rateLimit)
	}
	if params.IdleTimeout != "" {
		cmd = append(cmd, "=idle-timeout="+params.IdleTimeout)
	}
	if params.SessionTimeout != "" {
		cmd = append(cmd, "=session-timeout="+params.SessionTimeout)
	}
	if params.KeepaliveTimeout != "" {
		cmd = append(cmd, "=keepalive-timeout="+params.KeepaliveTimeout)
	}
	if params.OnlyOne {
		cmd = append(cmd, "=only-one=yes")
	} else {
		cmd = append(cmd, "=only-one=no")
	}
	if params.DNSServer != "" {
		cmd = append(cmd, "=dns-server="+params.DNSServer)
	}

	_, err = c.RunArgs(cmd)
	if err != nil {
		log.Printf("[MikroTik] UpdatePPPoEProfile - ERROR: %v", err)
		return fmt.Errorf("failed to update pppoe profile: %w", err)
	}

	log.Printf("[MikroTik] UpdatePPPoEProfile - SUCCESS: Updated profile %s", params.Name)
	return nil
}

// DeletePPPoEProfile deletes a PPPoE profile from MikroTik
func (c *Client) DeletePPPoEProfile(name string) error {
	log.Printf("[MikroTik] DeletePPPoEProfile - Deleting profile: %s", name)

	profileID, err := c.FindPPPoEProfileID(name)
	if err != nil {
		log.Printf("[MikroTik] DeletePPPoEProfile - ERROR finding profile: %v", err)
		return fmt.Errorf("failed to find profile: %w", err)
	}

	if profileID == "" {
		log.Printf("[MikroTik] DeletePPPoEProfile - Profile not found: %s", name)
		return fmt.Errorf("profile not found: %s", name)
	}

	cmd := []string{
		"/ppp/profile/remove",
		"=.id=" + profileID,
	}

	_, err = c.RunArgs(cmd)
	if err != nil {
		log.Printf("[MikroTik] DeletePPPoEProfile - ERROR: %v", err)
		return fmt.Errorf("failed to delete pppoe profile: %w", err)
	}

	log.Printf("[MikroTik] DeletePPPoEProfile - SUCCESS: Deleted profile %s", name)
	return nil
}

// FindPPPoEProfileID finds a PPPoE profile ID by name
func (c *Client) FindPPPoEProfileID(name string) (string, error) {
	cmd := []string{
		"/ppp/profile/print",
		"?name=" + name,
		"=.proplist=.id",
	}

	r, err := c.RunArgs(cmd)
	if err != nil {
		return "", err
	}

	if len(r.Re) == 0 {
		return "", nil // Not found
	}

	return r.Re[0].Map[".id"], nil
}

// GetPPPoEProfiles retrieves all PPPoE profiles from MikroTik
func (c *Client) GetPPPoEProfiles() ([]map[string]string, error) {
	log.Printf("[MikroTik] GetPPPoEProfiles - Fetching all profiles")

	cmd := []string{
		"/ppp/profile/print",
		"=.proplist=.id,name,local-address,remote-address,rate-limit,idle-timeout,session-timeout,keepalive-timeout,only-one,dns-server",
	}

	r, err := c.RunArgs(cmd)
	if err != nil {
		log.Printf("[MikroTik] GetPPPoEProfiles - ERROR: %v", err)
		return nil, fmt.Errorf("failed to get pppoe profiles: %w", err)
	}

	profiles := make([]map[string]string, len(r.Re))
	for i, re := range r.Re {
		profiles[i] = re.Map
	}

	log.Printf("[MikroTik] GetPPPoEProfiles - SUCCESS: Found %d profiles", len(profiles))
	return profiles, nil
}

// GetPPPoEProfile retrieves a single PPPoE profile by name
func (c *Client) GetPPPoEProfile(name string) (map[string]string, error) {
	log.Printf("[MikroTik] GetPPPoEProfile - Fetching profile: %s", name)

	cmd := []string{
		"/ppp/profile/print",
		"?name=" + name,
		"=.proplist=.id,name,local-address,remote-address,rate-limit,idle-timeout,session-timeout,keepalive-timeout,only-one,dns-server",
	}

	r, err := c.RunArgs(cmd)
	if err != nil {
		log.Printf("[MikroTik] GetPPPoEProfile - ERROR: %v", err)
		return nil, fmt.Errorf("failed to get pppoe profile: %w", err)
	}

	if len(r.Re) == 0 {
		log.Printf("[MikroTik] GetPPPoEProfile - Profile not found: %s", name)
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	log.Printf("[MikroTik] GetPPPoEProfile - SUCCESS: Found profile %s", name)
	return r.Re[0].Map, nil
}

// ==================== Ping Methods ====================

// PingResponse represents a single ping response from MikroTik
type PingResponse struct {
	Seq        string `json:"seq"`
	Host       string `json:"host"`
	Size       string `json:"size"`
	TTL        string `json:"ttl"`
	Time       string `json:"time"`
	Status     string `json:"status"` // "timeout", "net-unreachable", etc.
	Sent       string `json:"sent"`
	Received   string `json:"received"`
	PacketLoss string `json:"packet_loss"`
	AvgRtt     string `json:"avg_rtt"`
	MinRtt     string `json:"min_rtt"`
	MaxRtt     string `json:"max_rtt"`
	IsSummary  bool   `json:"is_summary"`
}

// StreamPing starts a continuous ping to the specified address
func (c *Client) StreamPing(
	ctx context.Context,
	address string,
	size string,
	interval string,
) (<-chan PingResponse, error) {

	args := []string{
		"/ping",
		"=address=" + address,
	}

	if size != "" {
		args = append(args, "=size="+size)
	}
	if interval != "" {
		args = append(args, "=interval="+interval)
	}

	reply, err := c.Client.ListenArgsContext(ctx, args)
	if err != nil {
		if IsConnectionError(err) {
			if recErr := c.Reconnect(); recErr == nil {
				reply, err = c.Client.ListenArgsContext(ctx, args)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	out := make(chan PingResponse)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-reply.Chan():
				if !ok {
					return
				}
				if r == nil || r.Map == nil {
					continue
				}
				out <- mapToPingResponse(r.Map)
			}
		}
	}()

	return out, nil
}

func mapToPingResponse(m map[string]string) PingResponse {
	isSummary := false
	if _, hasSeq := m["seq"]; !hasSeq {
		if _, hasSent := m["sent"]; hasSent {
			isSummary = true
		}
	}

	return PingResponse{
		Seq:        m["seq"],
		Host:       m["host"],
		Size:       m["size"],
		TTL:        m["ttl"],
		Time:       m["time"],
		Status:     m["status"],
		Sent:       m["sent"],
		Received:   m["received"],
		PacketLoss: m["packet-loss"],
		AvgRtt:     m["avg-rtt"],
		MinRtt:     m["min-rtt"],
		MaxRtt:     m["max-rtt"],
		IsSummary:  isSummary,
	}
}

// ==================== Helper Methods ====================

// IsConnectionError checks if an error is a connection error (exported version)
func IsConnectionError(err error) bool {
	return isConnectionError(err)
}
