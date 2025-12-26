package firewall

import (
	"fmt"
	"mikrobill/internal/infrastructure/mikrotik/model"
)

// ========== FIREWALL NAT MANAGEMENT ==========

// AddNATRule menambahkan NAT rule baru
func (s *Service) AddNATRule(config model.NATRequest) (*model.NATResponse, error) {
	// Prepare NAT data
	args := []string{
		"=chain=" + string(config.Chain),
		"=action=" + string(config.Action),
		"=disabled=" + boolToYesNo(config.Disabled),
	}

	if config.SrcAddress != "" {
		args = append(args, "=src-address="+config.SrcAddress)
	}
	if config.SrcAddressList != "" {
		args = append(args, "=src-address-list="+config.SrcAddressList)
	}
	if config.DstAddress != "" {
		args = append(args, "=dst-address="+config.DstAddress)
	}
	if config.DstAddressList != "" {
		args = append(args, "=dst-address-list="+config.DstAddressList)
	}
	if config.Protocol != "" {
		args = append(args, "=protocol="+config.Protocol)
	}
	if config.SrcPort != "" {
		args = append(args, "=src-port="+config.SrcPort)
	}
	if config.DstPort != "" {
		args = append(args, "=dst-port="+config.DstPort)
	}
	if config.InInterface != "" {
		args = append(args, "=in-interface="+config.InInterface)
	}
	if config.InInterfaceList != "" {
		args = append(args, "=in-interface-list="+config.InInterfaceList)
	}
	if config.OutInterface != "" {
		args = append(args, "=out-interface="+config.OutInterface)
	}
	if config.OutInterfaceList != "" {
		args = append(args, "=out-interface-list="+config.OutInterfaceList)
	}
	if config.ToAddresses != "" {
		args = append(args, "=to-addresses="+config.ToAddresses)
	}
	if config.ToPorts != "" {
		args = append(args, "=to-ports="+config.ToPorts)
	}
	if config.Comment != "" {
		args = append(args, "=comment="+config.Comment)
	}
	if config.Log {
		args = append(args, "=log=yes")
		if config.LogPrefix != "" {
			args = append(args, "=log-prefix="+config.LogPrefix)
		}
	}

	// Add NAT rule
	sentence := append([]string{"/ip/firewall/nat/add"}, args...)
	reply, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Get created rule
	ruleID := reply.Done.Map["ret"]
	if ruleID != "" {
		rule, err := s.GetNATRule(ruleID)
		if err == nil {
			return rule, nil
		}
	}

	return &model.NATResponse{
		Message: "success",
		Data: model.SimpleSuccessData{
			Created: true,
			ID:      ruleID,
		},
	}, nil
}

// GetNATRule mendapatkan detail NAT rule
func (s *Service) GetNATRule(ruleID string) (*model.NATResponse, error) {
	reply, err := s.client.Run("/ip/firewall/nat/print", "?.id="+ruleID)
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	if len(reply.Re) == 0 {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "NAT rule not found"},
		}, fmt.Errorf("nat rule not found")
	}

	return &model.NATResponse{
		Message: "success",
		Data:    reply.Re[0].Map,
	}, nil
}

// ListNATRules mendapatkan semua NAT rules
func (s *Service) ListNATRules() (*model.NATResponse, error) {
	reply, err := s.client.Run("/ip/firewall/nat/print")
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	rules := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		rules[i] = re.Map
	}

	return &model.NATResponse{
		Message: "success",
		Data:    rules,
	}, nil
}

// ListNATRulesByChain mendapatkan NAT rules berdasarkan chain
func (s *Service) ListNATRulesByChain(chain model.NATChain) (*model.NATResponse, error) {
	reply, err := s.client.Run("/ip/firewall/nat/print", "?chain="+string(chain))
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	rules := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		rules[i] = re.Map
	}

	return &model.NATResponse{
		Message: "success",
		Data:    rules,
	}, nil
}

// UpdateNATRule mengupdate NAT rule
func (s *Service) UpdateNATRule(ruleID string, updates model.NATUpdateRequest) (*model.NATResponse, error) {
	args := []string{"=.id=" + ruleID}

	if updates.Action != "" {
		args = append(args, "=action="+string(updates.Action))
	}
	if updates.SrcAddress != "" {
		args = append(args, "=src-address="+updates.SrcAddress)
	}
	if updates.SrcAddressList != "" {
		args = append(args, "=src-address-list="+updates.SrcAddressList)
	}
	if updates.DstAddress != "" {
		args = append(args, "=dst-address="+updates.DstAddress)
	}
	if updates.DstAddressList != "" {
		args = append(args, "=dst-address-list="+updates.DstAddressList)
	}
	if updates.Protocol != "" {
		args = append(args, "=protocol="+updates.Protocol)
	}
	if updates.SrcPort != "" {
		args = append(args, "=src-port="+updates.SrcPort)
	}
	if updates.DstPort != "" {
		args = append(args, "=dst-port="+updates.DstPort)
	}
	if updates.InInterface != "" {
		args = append(args, "=in-interface="+updates.InInterface)
	}
	if updates.InInterfaceList != "" {
		args = append(args, "=in-interface-list="+updates.InInterfaceList)
	}
	if updates.OutInterface != "" {
		args = append(args, "=out-interface="+updates.OutInterface)
	}
	if updates.OutInterfaceList != "" {
		args = append(args, "=out-interface-list="+updates.OutInterfaceList)
	}
	if updates.ToAddresses != "" {
		args = append(args, "=to-addresses="+updates.ToAddresses)
	}
	if updates.ToPorts != "" {
		args = append(args, "=to-ports="+updates.ToPorts)
	}
	if updates.Comment != "" {
		args = append(args, "=comment="+updates.Comment)
	}
	if updates.Disabled != nil {
		args = append(args, "=disabled="+boolToYesNo(*updates.Disabled))
	}
	if updates.Log != nil {
		args = append(args, "=log="+boolToYesNo(*updates.Log))
		if *updates.Log && updates.LogPrefix != "" {
			args = append(args, "=log-prefix="+updates.LogPrefix)
		}
	}

	sentence := append([]string{"/ip/firewall/nat/set"}, args...)
	_, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.NATResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DeleteNATRule menghapus NAT rule
func (s *Service) DeleteNATRule(ruleID string) (*model.NATResponse, error) {
	_, err := s.client.Run("/ip/firewall/nat/remove", "=.id="+ruleID)
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.NATResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Deleted: true},
	}, nil
}

// EnableNATRule mengaktifkan NAT rule
func (s *Service) EnableNATRule(ruleID string) (*model.NATResponse, error) {
	_, err := s.client.Run("/ip/firewall/nat/enable", "=.id="+ruleID)
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.NATResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DisableNATRule menonaktifkan NAT rule
func (s *Service) DisableNATRule(ruleID string) (*model.NATResponse, error) {
	_, err := s.client.Run("/ip/firewall/nat/disable", "=.id="+ruleID)
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.NATResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// AddMasqueradeRule menambahkan masquerade rule (shortcut helper)
func (s *Service) AddMasqueradeRule(config model.NATMasqueradeRequest) (*model.NATResponse, error) {
	natConfig := model.NATRequest{
		Chain:        model.NATChainSrcNAT,
		Action:       model.NATActionMasquerade,
		OutInterface: config.OutInterface,
		Comment:      config.Comment,
	}

	if config.SrcAddress != "" {
		natConfig.SrcAddress = config.SrcAddress
	}

	if natConfig.Comment == "" {
		natConfig.Comment = "Masquerade for " + config.OutInterface
	}

	return s.AddNATRule(natConfig)
}

// MoveNATRule memindahkan posisi NAT rule
func (s *Service) MoveNATRule(ruleID string, destination string) (*model.NATResponse, error) {
	_, err := s.client.Run("/ip/firewall/nat/move",
		"=.id="+ruleID,
		"=destination="+destination,
	)
	if err != nil {
		return &model.NATResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.NATResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// boolToYesNo converts bool to "yes" or "no"
func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
