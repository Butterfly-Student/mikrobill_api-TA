package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken creates SHA-256 hash of a token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
