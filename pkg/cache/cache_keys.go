// pkg/cache/cache_keys.go
package cache

import "fmt"

const (
	// Cache key prefixes
	PrefixGenieACSDevices = "genieacs:devices"
	PrefixGenieACSDevice  = "genieacs:device"
	PrefixSession         = "session"
	PrefixUser            = "user"
)

// DevicesKey key untuk cache semua devices
func DevicesKey() string {
	return PrefixGenieACSDevices
}

// DeviceKey key untuk cache device tertentu
func DeviceKey(deviceID string) string {
	return fmt.Sprintf("%s:%s", PrefixGenieACSDevice, deviceID)
}

// SessionKey key untuk cache session
func SessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", PrefixSession, sessionID)
}

// UserKey key untuk cache user data
func UserKey(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixUser, userID)
}

