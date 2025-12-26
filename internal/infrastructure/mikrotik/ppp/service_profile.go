package ppp

import (
	"fmt"
 "mikrobill/internal/infrastructure/mikrotik/model"
)

// ========== PPP PROFILE MANAGEMENT ==========

// CreateProfile membuat PPP profile baru
func (s *Service) CreateProfile(config model.PPPProfileRequest) (*model.PPPProfileResponse, error) {
	// Prepare profile data
	args := []string{
		"=name=" + config.Name,
	}

	if config.LocalAddress != "" {
		args = append(args, "=local-address="+config.LocalAddress)
	}
	if config.RemoteAddress != "" {
		args = append(args, "=remote-address="+config.RemoteAddress)
	}
	if config.DNSServer != "" {
		args = append(args, "=dns-server="+config.DNSServer)
	}
	if config.WINSServer != "" {
		args = append(args, "=wins-server="+config.WINSServer)
	}
	if config.RateLimit != "" {
		args = append(args, "=rate-limit="+config.RateLimit)
	}
	if config.SessionTimeout != "" {
		args = append(args, "=session-timeout="+config.SessionTimeout)
	}
	if config.IdleTimeout != "" {
		args = append(args, "=idle-timeout="+config.IdleTimeout)
	}
	if config.OnlyOne != "" {
		args = append(args, "=only-one="+config.OnlyOne)
	}
	if config.ChangeTCP_MSS != "" {
		args = append(args, "=change-tcp-mss="+config.ChangeTCP_MSS)
	}
	if config.UseEncryption != "" {
		args = append(args, "=use-encryption="+config.UseEncryption)
	}
	if config.UseCompression != "" {
		args = append(args, "=use-compression="+config.UseCompression)
	}
	if config.UseVJ_Compression != "" {
		args = append(args, "=use-vj-compression="+config.UseVJ_Compression)
	}
	if config.UseMPLS != "" {
		args = append(args, "=use-mpls="+config.UseMPLS)
	}
	if config.UseIPv6 != "" {
		args = append(args, "=use-ipv6="+config.UseIPv6)
	}
	if config.AddressList != "" {
		args = append(args, "=address-list="+config.AddressList)
	}
	if config.IncomingFilter != "" {
		args = append(args, "=incoming-filter="+config.IncomingFilter)
	}
	if config.OutgoingFilter != "" {
		args = append(args, "=outgoing-filter="+config.OutgoingFilter)
	}
	if config.Comment != "" {
		args = append(args, "=comment="+config.Comment)
	}

	// Add profile
	sentence := append([]string{"/ppp/profile/add"}, args...)
	reply, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.PPPProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Get created profile
	profileID := reply.Done.Map["ret"]
	if profileID != "" {
		profile, err := s.GetProfile(config.Name)
		if err == nil {
			return profile, nil
		}
	}

	return &model.PPPProfileResponse{
		Message: "success",
		Data: model.SimpleSuccessData{
			Created: true,
			Name:    config.Name,
		},
	}, nil
}

// GetProfile mendapatkan detail PPP profile
func (s *Service) GetProfile(name string) (*model.PPPProfileResponse, error) {
	reply, err := s.client.Run("/ppp/profile/print", "?name="+name)
	if err != nil {
		return &model.PPPProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	if len(reply.Re) == 0 {
		return &model.PPPProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "PPP profile not found"},
		}, fmt.Errorf("ppp profile not found")
	}

	return &model.PPPProfileResponse{
		Message: "success",
		Data:    reply.Re[0].Map,
	}, nil
}

// ListProfiles mendapatkan semua PPP profiles
func (s *Service) ListProfiles() (*model.PPPProfileResponse, error) {
	reply, err := s.client.Run("/ppp/profile/print")
	if err != nil {
		return &model.PPPProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	profiles := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		profiles[i] = re.Map
	}

	return &model.PPPProfileResponse{
		Message: "success",
		Data:    profiles,
	}, nil
}

// UpdateProfile mengupdate PPP profile
func (s *Service) UpdateProfile(profileID string, updates model.PPPProfileUpdateRequest) (*model.PPPProfileResponse, error) {
	args := []string{"=.id=" + profileID}

	if updates.LocalAddress != "" {
		args = append(args, "=local-address="+updates.LocalAddress)
	}
	if updates.RemoteAddress != "" {
		args = append(args, "=remote-address="+updates.RemoteAddress)
	}
	if updates.DNSServer != "" {
		args = append(args, "=dns-server="+updates.DNSServer)
	}
	if updates.WINSServer != "" {
		args = append(args, "=wins-server="+updates.WINSServer)
	}
	if updates.RateLimit != "" {
		args = append(args, "=rate-limit="+updates.RateLimit)
	}
	if updates.SessionTimeout != "" {
		args = append(args, "=session-timeout="+updates.SessionTimeout)
	}
	if updates.IdleTimeout != "" {
		args = append(args, "=idle-timeout="+updates.IdleTimeout)
	}
	if updates.OnlyOne != "" {
		args = append(args, "=only-one="+updates.OnlyOne)
	}
	if updates.ChangeTC_MSS != "" {
		args = append(args, "=change-tcp-mss="+updates.ChangeTC_MSS)
	}
	if updates.UseEncryption != "" {
		args = append(args, "=use-encryption="+updates.UseEncryption)
	}
	if updates.UseCompression != "" {
		args = append(args, "=use-compression="+updates.UseCompression)
	}
	if updates.UseVJ_Compression != "" {
		args = append(args, "=use-vj-compression="+updates.UseVJ_Compression)
	}
	if updates.UseMPLS != "" {
		args = append(args, "=use-mpls="+updates.UseMPLS)
	}
	if updates.UseIPv6 != "" {
		args = append(args, "=use-ipv6="+updates.UseIPv6)
	}
	if updates.AddressList != "" {
		args = append(args, "=address-list="+updates.AddressList)
	}
	if updates.IncomingFilter != "" {
		args = append(args, "=incoming-filter="+updates.IncomingFilter)
	}
	if updates.OutgoingFilter != "" {
		args = append(args, "=outgoing-filter="+updates.OutgoingFilter)
	}
	if updates.Comment != "" {
		args = append(args, "=comment="+updates.Comment)
	}

	sentence := append([]string{"/ppp/profile/set"}, args...)
	_, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.PPPProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.PPPProfileResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DeleteProfile menghapus PPP profile
func (s *Service) DeleteProfile(profileID string) (*model.PPPProfileResponse, error) {
	_, err := s.client.Run("/ppp/profile/remove", "=.id="+profileID)
	if err != nil {
		return &model.PPPProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.PPPProfileResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Deleted: true},
	}, nil
}
