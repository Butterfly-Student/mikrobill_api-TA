package ppp

import (
	"mikrobill/internal/infrastructure/mikrotik/model"
)

// ConvertToPPPSecretRequest converts parameters to PPPSecretRequest
func ConvertToPPPSecretRequest(username, password, profile, localAddress, remoteAddress string) model.PPPSecretRequest {
	return model.PPPSecretRequest{
		Name:          username,
		Password:      password,
		Service:       "pppoe",
		Profile:       profile,
		LocalAddress:  localAddress,
		RemoteAddress: remoteAddress,
		Disabled:      false,
	}
}

// ConvertToPPPSecretUpdateRequest converts parameters to PPPSecretUpdateRequest
func ConvertToPPPSecretUpdateRequest(username, password, profile, localAddress, remoteAddress string) model.PPPSecretUpdateRequest {
	return model.PPPSecretUpdateRequest{
		Password:      password,
		Service:       "pppoe",
		Profile:       profile,
		LocalAddress:  localAddress,
		RemoteAddress: remoteAddress,
	}
}
