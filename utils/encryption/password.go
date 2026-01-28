package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

const (
	// Character sets for password generation
	lowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars     = "0123456789"

	// For PPPoE: strong alphanumeric password
	pppoeChars = lowercaseChars + uppercaseChars + digitChars

	// For Hotspot: simpler alphanumeric password
	hotspotChars = lowercaseChars + digitChars
)

// GeneratePassword generates a random password with the specified length and character set
func GeneratePassword(length int, charset string) (string, error) {
	password := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		password[i] = charset[randomIndex.Int64()]
	}

	return string(password), nil
}

// GeneratePPPoEPassword generates a strong 16-character alphanumeric password for PPPoE
func GeneratePPPoEPassword() (string, error) {
	return GeneratePassword(16, pppoeChars)
}

// GenerateHotspotPassword generates a simpler 8-character alphanumeric password for Hotspot
func GenerateHotspotPassword() (string, error) {
	return GeneratePassword(8, hotspotChars)
}

// GenerateToken generates a random token for use in URLs or API keys
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
