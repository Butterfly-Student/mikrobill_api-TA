package hotspot

// ========== SYSTEM INFO ==========

// GetParentQueues mendapatkan parent queue
func (s *Service) GetParentQueues() ([]map[string]string, error) {
	reply, err := s.client.Run("/queue/simple/print", "?dynamic=false")
	if err != nil {
		return nil, err
	}

	queues := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		queues[i] = re.Map
	}

	return queues, nil
}

// GetHotspotServers mendapatkan hotspot servers
func (s *Service) GetHotspotServers() ([]map[string]string, error) {
	reply, err := s.client.Run("/ip/hotspot/print")
	if err != nil {
		return nil, err
	}

	servers := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		servers[i] = re.Map
	}

	return servers, nil
}

// GetHotspotHosts mendapatkan hotspot hosts
func (s *Service) GetHotspotHosts() ([]map[string]string, error) {
	reply, err := s.client.Run("/ip/hotspot/host/print")
	if err != nil {
		return nil, err
	}

	hosts := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		hosts[i] = re.Map
	}

	return hosts, nil
}

// GetFirewallNat mendapatkan firewall NAT rules
func (s *Service) GetFirewallNat() ([]map[string]string, error) {
	reply, err := s.client.Run("/ip/firewall/nat/print")
	if err != nil {
		return nil, err
	}

	rules := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		rules[i] = re.Map
	}

	return rules, nil
}